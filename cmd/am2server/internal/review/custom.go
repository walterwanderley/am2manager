package review

import (
	"database/sql"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/walterwanderley/am2manager/cmd/am2server/internal/server"
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
	type response struct {
		LastInsertId int64 `json:"last_insert_id"`
		RowsAffected int64 `json:"rows_affected"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req, err := server.Decode[request](r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		user := templates.UserFromContext(r.Context())
		if !user.Logged() {
			http.Error(w, "Sign-in to write a review", http.StatusUnauthorized)
			return
		}
		var arg addReviewParams

		arg.UserID = sql.NullInt64{Valid: true, Int64: user.ID}
		arg.CaptureID = req.CaptureID
		arg.Rate = req.Rate
		if req.Comment != nil {
			arg.Comment = sql.NullString{Valid: true, String: *req.Comment}
		}

		result, err := s.querier.addReview(r.Context(), arg)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "AddReview")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lastInsertId, _ := result.LastInsertId()
		rowsAffected, _ := result.RowsAffected()
		server.Encode(w, r, http.StatusOK, response{
			LastInsertId: lastInsertId,
			RowsAffected: rowsAffected,
		})
	}
}

func (s *CustomService) handleRemoveReview() http.HandlerFunc {
	type request struct {
		Id int64 `form:"id" json:"id"`
	}
	type response struct {
		LastInsertId int64 `json:"last_insert_id"`
		RowsAffected int64 `json:"rows_affected"`
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
		if !user.Logged() {
			http.Error(w, "Sign-in to remove a review", http.StatusUnauthorized)
			return
		}

		review, err := s.querier.getReview(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if review.UserID.Int64 != user.ID {
			http.Error(w, "Can't remove other user's review", http.StatusForbidden)
			return
		}

		result, err := s.querier.removeReview(r.Context(), id)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "RemoveReview")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lastInsertId, _ := result.LastInsertId()
		rowsAffected, _ := result.RowsAffected()
		server.Encode(w, r, http.StatusOK, response{
			LastInsertId: lastInsertId,
			RowsAffected: rowsAffected,
		})
	}
}
