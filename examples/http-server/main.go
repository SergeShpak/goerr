package main

import (
	"net/http"

	"github.com/SergeyShpak/goerr/examples/http/handlers"
)

func main() {
	http.HandleFunc("/hello", handlers.Hello)
	http.HandleFunc("/admin", handlers.AccessAdminConsole)
	http.ListenAndServe(":8080", nil)
}
