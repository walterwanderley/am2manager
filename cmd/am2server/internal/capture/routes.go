// Code generated by sqlc-http (https://github.com/walterwanderley/sqlc-http). DO NOT EDIT.

package capture

import "net/http"

func (s *Service) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("POST /captures", s.handleAddCapture())
	mux.HandleFunc("GET /captures/{id}", s.handleGetCapture())
	mux.HandleFunc("GET /captures/{id}/file", s.handleGetCaptureFile())
	mux.HandleFunc("GET /captures", s.handleListCaptures())
	mux.HandleFunc("DELETE /captures/{id}", s.handleRemoveCapture())
}
