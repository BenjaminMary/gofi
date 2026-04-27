package api

import (
	"fmt"
	"net/http"

	"gofi/gofi/data/appdata"

	"github.com/go-chi/chi/v5"
)

// MaintenanceReadOnly toggles appdata.ReadOnlyFlag based on the {state} URL param.
// state="on"  → enable read-only (block all mutating methods server-wide)
// state="off" → disable read-only (back to normal)
// Admin only. Served as GET (mirroring /api/shutdown) so it passes through
// the MaintenanceMode middleware naturally even when the flag is on.
func MaintenanceReadOnly(w http.ResponseWriter, r *http.Request) {
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	if !userContext.IsAdmin {
		appdata.RenderAPIorUI(w, r, false, false, false, http.StatusForbidden, "admin only", "")
		return
	}
	state := chi.URLParam(r, "state")
	switch state {
	case "on":
		appdata.ReadOnlyFlag = true
		fmt.Println("ReadOnlyFlag enabled — writes blocked")
		appdata.RenderAPIorUI(w, r, false, false, true, http.StatusOK, "read-only mode enabled", "")
	case "off":
		appdata.ReadOnlyFlag = false
		fmt.Println("ReadOnlyFlag disabled — writes allowed")
		appdata.RenderAPIorUI(w, r, false, false, true, http.StatusOK, "read-only mode disabled", "")
	default:
		appdata.RenderAPIorUI(w, r, false, false, false, http.StatusBadRequest, "state must be 'on' or 'off'", "")
	}
}
