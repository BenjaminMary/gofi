package routes

import (
	"fmt"
	"net/http"

	// 3 fileserver req
	"os"
	"path/filepath"
	"strings"

	"gofi/gofi/back/api"
	"gofi/gofi/back/appmiddleware"
	"gofi/gofi/data/appdata"
	"gofi/gofi/front"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// FileServer conveniently sets up a http.FileServer handler to serve static files from a http.FileSystem.
// go chi example: https://github.com/go-chi/chi/blob/master/_examples/fileserver/main.go
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}
	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"
	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}

type Server struct {
	Router *chi.Mux
	// Db     *sql.DB //, config can be added here
}

func CreateNewServer() *Server {
	s := &Server{}
	s.Router = chi.NewRouter()
	fmt.Println("------------------ROUTER START HERE------------------")
	fmt.Println("start DB from server")
	appdata.DB = OpenDbCon()
	return s
}

func (s *Server) MountBackHandlers() {
	// go-chi example: https://github.com/go-chi/chi/blob/master/_examples/rest/main.go
	// Mount all Middleware here
	s.Router.Use(middleware.Logger)
	s.Router.Use(middleware.Recoverer)
	s.Router.Use(appmiddleware.MaintenanceMode)
	s.Router.Use(appmiddleware.AddContextUserAndTimeout)

	// CSV
	s.Router.Route("/api/csv", func(r chi.Router) {
		r.Use(appmiddleware.CheckHeader)
		r.Use(appmiddleware.AuthenticatedUserOnly)
		r.Post("/import", func(w http.ResponseWriter, r *http.Request) { api.PostCSVimport(w, r, false) })
		r.Post("/export", func(w http.ResponseWriter, r *http.Request) { api.PostCSVexport(w, r, false) })
		r.Post("/export/reset", func(w http.ResponseWriter, r *http.Request) { api.PostCSVexportReset(w, r, false) })
	})

	s.Router.Route("/api", func(r chi.Router) {
		r.Use(appmiddleware.HeaderContentType)
		r.Use(appmiddleware.CheckHeader)

		// Mount all handlers here
		// OTHERS PUBLIC
		r.Group(func(r chi.Router) {
			r.Get("/isadmin", api.IsAdmin)                 // GET /api/isadmin
			r.Get("/isauthenticated", api.IsAuthenticated) // GET /api/isauthenticated
		})
		// OTHERS PRIVATE
		r.Group(func(r chi.Router) {
			r.Use(appmiddleware.AuthenticatedUserOnly)
			r.Get("/dbpath", api.GetFullDbPath) // GET /api/dbpath
			r.Get("/shutdown", api.Shutdown)    // GET /api/shutdown
		})
		// r.With(appmiddleware.AuthenticatedUserOnly).Get("/shutdown", api.Shutdown) // auth needed: GET /api/shutdown

		// USERS
		r.Route("/user", func(r chi.Router) {
			r.Post("/create", func(w http.ResponseWriter, r *http.Request) { api.UserCreate(w, r, false) }) // POST /api/user/create
			r.Post("/login", func(w http.ResponseWriter, r *http.Request) { api.UserLogin(w, r, false) })   // POST /api/user/login

			// PRIVATE
			r.Group(func(r chi.Router) {
				r.Use(appmiddleware.AuthenticatedUserOnly)                                                     // auth needed
				r.Get("/logout", func(w http.ResponseWriter, r *http.Request) { api.UserLogout(w, r, false) }) // auth needed: GET /api/user/logout (use the header -H "sessionID: XYZ")
			})
			r.Route("/{userID}", func(r chi.Router) {
				r.Use(appmiddleware.AuthenticatedUserOnly)                                                              // auth needed
				r.Get("/refresh", func(w http.ResponseWriter, r *http.Request) { api.UserRefreshSession(w, r, false) }) // GET /api/user/123/refresh (use the header -H "sessionID: XYZ")
				r.Delete("/delete", func(w http.ResponseWriter, r *http.Request) { api.UserDelete(w, r, false) })       // DELETE /api/user/123/delete
			})
		})
		r.Group(func(r chi.Router) {
			// PRIVATE
			r.Use(appmiddleware.AuthenticatedUserOnly)
			// PARAMS
			r.Route("/param", func(r chi.Router) {
				r.Get("/", func(w http.ResponseWriter, r *http.Request) { api.GetParam(w, r, false, "", "", false) })
				r.Post("/account", func(w http.ResponseWriter, r *http.Request) { api.PostParamAccount(w, r, false) })
				r.Get("/category/{categoryName}", func(w http.ResponseWriter, r *http.Request) {
					api.GetCategoryIcon(w, r, false, "", &appdata.CategoryDetails{})
				})
				r.Put("/category", func(w http.ResponseWriter, r *http.Request) { api.PutParamCategory(w, r, false) })
				r.Patch("/category/in-use", func(w http.ResponseWriter, r *http.Request) { api.PatchParamCategoryInUse(w, r, false) })
				r.Patch("/category/order", func(w http.ResponseWriter, r *http.Request) { api.PatchParamCategoryOrder(w, r, false) })
				r.Post("/category-rendering", func(w http.ResponseWriter, r *http.Request) { api.PostParamCategoryRendering(w, r, false) })
			})
			// RECORDS
			r.Route("/record", func(r chi.Router) {
				r.Get("/{orderby}-{ordersort}-{limit}", func(w http.ResponseWriter, r *http.Request) { api.GetRecords(w, r, false) })
				r.Post("/getviapost", func(w http.ResponseWriter, r *http.Request) {
					api.GetRecordsViaPost(w, r, false, &appdata.FilterRows{})
				})
				r.Post("/edit/{idft}", func(w http.ResponseWriter, r *http.Request) { api.PostRecordEdit(w, r, false, "") })
				r.Post("/insert", func(w http.ResponseWriter, r *http.Request) { api.PostRecordInsert(w, r, false) })
				r.Post("/lend-or-borrow", func(w http.ResponseWriter, r *http.Request) { api.PostLendOrBorrowRecords(w, r, false) })
				r.Post("/lender-borrower-state-change", func(w http.ResponseWriter, r *http.Request) { api.PostLenderBorrowerStateChange(w, r, false) })
				r.Post("/lend-or-borrow-unlink", func(w http.ResponseWriter, r *http.Request) { api.PostUnlinkLendOrBorrowRecords(w, r, false) })
				r.Post("/transfer", func(w http.ResponseWriter, r *http.Request) { api.PostRecordTransfer(w, r, false) })
				r.Get("/recurrent", func(w http.ResponseWriter, r *http.Request) { api.RecordRecurrentRead(w, r, false) })
				r.Post("/recurrent/create", func(w http.ResponseWriter, r *http.Request) { api.RecordRecurrentCreate(w, r, false) })
				r.Post("/recurrent/save", func(w http.ResponseWriter, r *http.Request) { api.RecordRecurrentSave(w, r, false) })
				r.Put("/recurrent/update", func(w http.ResponseWriter, r *http.Request) { api.RecordRecurrentUpdate(w, r, false) })
				r.Delete("/recurrent/{idrr}/delete", func(w http.ResponseWriter, r *http.Request) { api.RecordRecurrentDelete(w, r, false, "") })
				r.Put("/validate", func(w http.ResponseWriter, r *http.Request) { api.RecordValidate(w, r, false) })
				r.Put("/cancel", func(w http.ResponseWriter, r *http.Request) { api.RecordCancel(w, r, false) })
			})
			// SAVE
			r.Route("/save", func(r chi.Router) {
				r.Get("/", api.SaveRead)    // GET /api/save
				r.Post("/", api.SaveCreate) // POST /api/save
				// r.Put("/", api.SaveEdit)      // PUT /api/save
				r.Delete("/delete/{id}", api.SaveDelete)     // DELETE /api/save/delete/1
				r.Delete("/keep/{num}", api.SaveDeleteKeepX) // DELETE /api/save/keep/3
			})
		})
	})
}

func (s *Server) MountFrontHandlers() {
	// Mount all front handlers here
	s.Router.Route("/", func(r chi.Router) {
		r.Use(appmiddleware.CheckCookie)
		r.NotFound(front.Lost)
		r.Get("/", front.TemplIndex)

		// USERS
		r.Route("/user", func(r chi.Router) {
			r.Get("/create", front.GetCreateUser)
			r.Post("/create", front.PostCreateUser)
			r.Get("/login", front.GetLogin)
			r.Post("/login", front.PostLogin)
			r.Get("/logout", front.GetLogout)
		})
		r.Group(func(r chi.Router) {
			// PRIVATE
			r.Use(appmiddleware.AuthenticatedUserOnly)

			// PARAMS
			r.Route("/param", func(r chi.Router) {
				r.Get("/", front.GetParam)
				r.Get("/account", front.GetParamAccount)
				r.Post("/account", front.PostParamAccount)
				r.Post("/category-rendering", front.PostParamCategoryRendering)
				r.Get("/category", front.GetParamCategory)
				r.Put("/category", front.PutParamCategory)
				r.Patch("/category/in-use", front.PatchParamCategoryInUse)
				r.Patch("/category/order", front.PatchParamCategoryOrder)
			})
			// RECORDS
			r.Route("/record", func(r chi.Router) {
				// r.Get("/insert/{isDefault}-{account}-{category}-{product}-{priceDirection}-{price}", front.GetRecordInsert)
				r.Get("/insert/*", front.GetRecordInsert)
				r.Post("/insert", front.PostRecordInsert)
				r.Get("/lend-or-borrow", front.GetLendBorrowRecord)
				r.Post("/lend-or-borrow", front.PostLendOrBorrowRecord)
				r.Get("/transfer", front.GetRecordTransfer)
				r.Post("/transfer", front.PostRecordTransfer)
				r.Get("/recurrent", front.GetRecordRecurrent)
				r.Post("/recurrent/create", front.PostRecordRecurrentCreate)
				r.Post("/recurrent/save", front.PostRecordRecurrentSave)
				r.Post("/recurrent/update", front.PostRecordRecurrentUpdate)
				r.Post("/recurrent/delete", front.PostRecordRecurrentDelete)
				r.Get("/alter/{alterMode}", front.GetRecordAlter)
				r.Get("/edit/{id}", front.GetRecordEdit)
				r.Post("/edit/{idft}", front.PostRecordEdit)
				r.Post("/validate", front.PostRecordValidate)
				r.Post("/cancel", front.PostRecordCancel)
				r.Post("/getviapost", front.PostFullRecordRefresh)
			})
			// CSV
			r.Route("/csv", func(r chi.Router) {
				r.Get("/export", front.GetCSVexport)
				r.Post("/export/reset", front.PostCSVexportReset)
				r.Get("/import", front.GetCSVimport)
				r.Post("/import", front.PostCSVimport)
			})
			// STATS
			r.Get("/stats/{checkedValidData}-{year}-{checkedYearStats}-{checkedGainsStats}", front.GetStats)
			r.Get("/budget", front.GetBudget)
			r.Get("/stats/lender-borrower/{lbID}", front.GetLenderBorrowerStats)
			r.Post("/stats/lender-borrower/{lbID}/state-change", front.PostLenderBorrowerStateChange)
			r.Post("/stats/lender-borrower/{lbID}/unlink", front.PostUnlinkLendOrBorrowRecords)
		})
	})
}

func (s *Server) MountFileServer() {
	// Mount static files
	workDir := os.Getenv("EXE_PATH")
	// Create a route along /img that will serve contents from the ./assets/img/ folder.
	imgDir := http.Dir(filepath.Join(workDir, "assets", "img"))
	FileServer(s.Router, "/img", imgDir)
	jsDir := http.Dir(filepath.Join(workDir, "assets", "js"))
	FileServer(s.Router, "/js", jsDir)
	fontsDir := http.Dir(filepath.Join(workDir, "assets", "fonts"))
	FileServer(s.Router, "/fonts", fontsDir)
}
