// Code generated by sqlc-http (https://github.com/walterwanderley/sqlc-http). DO NOT EDIT.

package review

import "net/http"

func (s *Service) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("GET /users/{user_id}/reviews", s.handleListReviewsByUser())
}
