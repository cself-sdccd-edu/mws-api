package api

import (
	"fmt"
	"github.com/cself-sdccd-edu/mws-api/internal/cache"
	"net/http"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
}

func (s *Server) scheduleHandler(w http.ResponseWriter, r *http.Request) {
	term := r.URL.Query().Get("term")
	career := r.URL.Query().Get("career")

	if term == "" {
		http.Error(w, "missing term parameter", http.StatusBadRequest)
		return
	}

	if career == "" {
		http.Error(w, "missing career parameter", http.StatusBadRequest)
		return
	}

	if career != "ugrd" && career != "ce" {
		http.Error(w, "invalid career parameter", http.StatusBadRequest)
		return
	}

	// setup the refresh request based on parameters
	var refresh cache.RefreshRequest

	switch career {
	case "ugrd":
		refresh = cache.RefreshRequest{
			QueryName: s.config.Queries["ugrd"],
			Term:      term,
		}
	case "ce":
		refresh = cache.RefreshRequest{
			QueryName: s.config.Queries["ce"],
			Term:      term,
		}
	}

	// set the key for caching
	key := fmt.Sprintf("schedule:%s:%s", term, career)

	// try to get a cached request for this key
	entry, err := s.cache.Get(r.Context(), key, refresh)
	if err != nil {
		http.Error(w, fmt.Sprintf("cache error: %v", err), http.StatusInternalServerError)
		return
	}

	// I don't want to write this to the client/browser. I want to write the JSON form entry.Data
	//fmt.Fprintln(w, "/schedule endpoint. params: term=", term, "career=", career)
	//fmt.Fprintln(w, "entry updated at", entry.UpdatedAt)

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(entry.Data); err != nil {
		return
	}
}
