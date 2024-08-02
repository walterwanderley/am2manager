package review

import (
	"database/sql"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/walterwanderley/am2manager/cmd/am2server/internal/server"
	"github.com/walterwanderley/am2manager/cmd/am2server/internal/server/htmx"
	"github.com/walterwanderley/am2manager/cmd/am2server/templates"
)

type CustomService struct {
	Service
}

func (s *CustomService) RegisterHandlers(mux *http.ServeMux) {
	s.Service.RegisterHandlers(mux)
	mux.HandleFunc("POST /reviews", s.handleAddReview())
	mux.HandleFunc("DELETE /reviews/{id}", s.handleRemoveReview())
}

func (s *CustomService) handleAddReview() http.HandlerFunc {
	type request struct {
		CaptureID int64   `form:"capture_id" json:"capture_id"`
		Rate      int64   `form:"rate" json:"rate"`
		Comment   *string `form:"comment" json:"comment"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := server.Decode[request](r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		user := templates.UserFromContext(r.Context())
		if !user.Logged() {
			htmx.Toast(w, r, http.StatusUnauthorized, "Sign-in to write a review")
			return
		}
		var arg addReviewParams

		arg.UserID = sql.NullInt64{Valid: true, Int64: user.ID}
		arg.CaptureID = req.CaptureID
		arg.Rate = req.Rate
		if req.Comment != nil {
			arg.Comment = sql.NullString{Valid: true, String: *req.Comment}
		}

		if arg.Rate < 0 || arg.Rate > 5 {
			htmx.Toast(w, r, http.StatusBadRequest, "Invalid rate")
			return
		}

		count, err := s.querier.existsReviewByUserCapture(r.Context(), existsReviewByUserCaptureParams{
			UserID:    arg.UserID,
			CaptureID: arg.CaptureID,
		})
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "existsReviewByUserCapture")
			htmx.Toast(w, r, http.StatusInternalServerError, err.Error())
			return
		}

		if count > 0 {
			htmx.Toast(w, r, http.StatusBadRequest, "Dulicated review")
			return
		}

		_, err = s.querier.addReview(r.Context(), arg)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "AddReview")
			htmx.Toast(w, r, http.StatusInternalServerError, err.Error())
			return
		}
		htmx.Toast(w, r, http.StatusOK, "Review created")
	}
}

func (s *CustomService) handleRemoveReview() http.HandlerFunc {
	type request struct {
		Id int64 `form:"id" json:"id"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		if str := r.PathValue("id"); str != "" {
			if v, err := strconv.ParseInt(str, 10, 64); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			} else {
				req.Id = v
			}
		}
		id := req.Id

		user := templates.UserFromContext(r.Context())
		if !user.CanEdit(id) {
			htmx.Toast(w, r, http.StatusForbidden, "Sign-in to remove a review")
			return
		}

		_, err := s.querier.removeReview(r.Context(), id)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "RemoveReview")
			htmx.Toast(w, r, http.StatusInternalServerError, err.Error())
			return
		}

	}
}
