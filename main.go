package main

import (
	"fmt"
	"net/http"
	"os"

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

	GOFI_PORT := os.Getenv("GOFI_PORT")
	if GOFI_PORT == "" {
		GOFI_PORT = "8083"
	}
	fmt.Printf("running on port: %v\n", GOFI_PORT)
	http.ListenAndServe(":" + GOFI_PORT, s.Router)
}
