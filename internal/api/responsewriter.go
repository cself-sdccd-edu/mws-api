package api

import "net/http"

type detailRecorder interface {
	SetDetails(string)
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
	details    string
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *statusRecorder) Write(data []byte) (int, error) {
	if r.statusCode == 0 {
		r.statusCode = http.StatusOK
	}

	return r.ResponseWriter.Write(data)
}

func badRequest(w http.ResponseWriter, message string) {
	if recorder, ok := w.(detailRecorder); ok {
		recorder.SetDetails(message)
	}
	http.Error(w, message, http.StatusBadRequest)
}

func (r *statusRecorder) SetDetails(details string) {
	r.details = details
}
