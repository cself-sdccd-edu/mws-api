package api

import (
	"fmt"
	"net/http"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
}

func scheduleHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "/schedule endpoint")
}
