package main

import (
	"fmt"
	"net/http"

	"gofi/gofi/back/routes"
	"gofi/gofi/data/appdata"
)

func main() {
	s := routes.CreateNewServer()
	s.MountBackHandlers()
	s.MountFrontHandlers()
	s.MountFileServer()
	defer routes.CloseDbCon(appdata.DB)
	defer fmt.Println("closing DB conn from main")

	http.ListenAndServe(":8083", s.Router)
	// 8082 gosheets
}
