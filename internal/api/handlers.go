package api

import (
	"fmt"
	"github.com/cself-sdccd-edu/mws-api/internal/cache"
	mwslog "github.com/cself-sdccd-edu/mws-api/internal/log"
	"log"
	"net/http"
	"strings"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
}

func (s *Server) endpointHandler(w http.ResponseWriter, r *http.Request) {
	term := r.URL.Query().Get("term")
	career := r.URL.Query().Get("career")

	if term == "" {
		http.Error(w, "missing term parameter", http.StatusBadRequest)
		return
	}

	if !validTerm(term) {
		http.Error(w, "invalid term parameter", http.StatusBadRequest)
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

	endpoint := strings.TrimPrefix(r.URL.Path, "/")
	endpointConfig, ok := s.config.Endpoints[endpoint]
	if !ok {
		http.Error(w, "endpoint not configured", http.StatusNotFound)
		return
	}
	// setup the refresh request based on parameters

	var refresh cache.RefreshRequest

	switch career {
	case "ugrd":
		refresh = cache.RefreshRequest{
			EndpointName: endpoint,
			QueryName:    endpointConfig.Queries["ugrd"],
			Term:         term,
		}
	case "ce":
		refresh = cache.RefreshRequest{
			EndpointName: endpoint,
			QueryName:    endpointConfig.Queries["ce"],
			Term:         term,
		}
	}

	// set the key for caching
	key := fmt.Sprintf("%s:%s:%s", endpoint, term, career)

	// try to get a cached request for this key
	entry, err := s.cache.Get(r.Context(), key, refresh)
	if err != nil {
		//http.Error(w, fmt.Sprintf("cache error: %v", err), http.StatusInternalServerError)
		s.writeInternalError(w, r, "unable to retrieve data", err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err := w.Write(entry.Data); err != nil {
		return
	}

}
func (s *Server) scheduleHandler(w http.ResponseWriter, r *http.Request) {
	term := r.URL.Query().Get("term")
	career := r.URL.Query().Get("career")

	if term == "" {
		http.Error(w, "missing term parameter", http.StatusBadRequest)
		return
	}

	if !validTerm(term) {
		http.Error(w, "invalid term parameter", http.StatusBadRequest)
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
			EndpointName: "schedule",
			QueryName:    s.config.Endpoints["schedule"].Queries["ugrd"],
			Term:         term,
		}
	case "ce":
		refresh = cache.RefreshRequest{
			EndpointName: "schedule",
			QueryName:    s.config.Endpoints["schedule"].Queries["ce"],
			Term:         term,
		}
	}

	// set the key for caching
	key := fmt.Sprintf("schedule:%s:%s", term, career)

	// try to get a cached request for this key
	entry, err := s.cache.Get(r.Context(), key, refresh)
	if err != nil {
		//http.Error(w, fmt.Sprintf("cache error: %v", err), http.StatusInternalServerError)
		s.writeInternalError(w, r, "unable to retrieve schedule data", err)
		return
	}

	// I don't want to write this to the client/browser. I want to write the JSON form entry.Data
	//fmt.Fprintln(w, "/schedule endpoint. params: term=", term, "career=", career)
	//fmt.Fprintln(w, "entry updated at", entry.UpdatedAt)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err := w.Write(entry.Data); err != nil {
		return
	}
}

func (s *Server) writeInternalError(w http.ResponseWriter, r *http.Request, publicMessage string, internalErr error) {
	requestID := mwslog.RequestID(r.Context())

	event := mwslog.LogEvent{
		RequestID:  requestID,
		Event:      "request_error",
		Method:     r.Method,
		Endpoint:   r.URL.Path,
		StatusCode: http.StatusInternalServerError,
		Message:    publicMessage,
		Details:    internalErr.Error(),
	}

	if err := s.logger.Log(r.Context(), event); err != nil {
		log.Printf("requestid %s error logging failed: %v", requestID, err)
	}

	http.Error(w, publicMessage, http.StatusInternalServerError)
}
