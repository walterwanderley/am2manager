package user

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/walterwanderley/am2manager/cmd/am2server/internal/server"
	"github.com/walterwanderley/am2manager/cmd/am2server/internal/server/htmx"
	"github.com/walterwanderley/am2manager/cmd/am2server/templates"
)

type CustomService struct {
	Service
}

func (s *CustomService) RegisterHandlers(mux *http.ServeMux) {
	s.Service.RegisterHandlers(mux)
	mux.HandleFunc("GET /users/{id}", s.handleGetUser())
	mux.HandleFunc("PATCH /users/{id}/name", s.handleUpdateUserName())
}

func (s *CustomService) handleGetUser() http.HandlerFunc {
	type request struct {
		Id int64 `form:"id" json:"id"`
	}
	type response struct {
		ID        int64      `json:"id,omitempty"`
		Email     string     `json:"email,omitempty"`
		Name      string     `json:"name,omitempty"`
		Status    string     `json:"status,omitempty"`
		CreatedAt time.Time  `json:"created_at,omitempty"`
		UpdatedAt *time.Time `json:"updated_at,omitempty"`
		Picture   *string    `json:"picture,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(templates.ContextWithTemplates(r.Context(), "users/{id}.html"))
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
		if user.ID <= 0 {
			htmx.Error(w, r, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			return
		}
		if user.ID != id {
			htmx.Error(w, r, http.StatusForbidden, http.StatusText(http.StatusForbidden))
			return
		}

		result, err := s.querier.getUser(r.Context(), id)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "GetUser")
			htmx.Error(w, r, http.StatusInternalServerError, err.Error())
			return
		}
		var res response
		res.ID = result.ID
		res.Email = result.Email
		res.Name = result.Name
		res.Status = result.Status
		res.CreatedAt = result.CreatedAt
		if result.UpdatedAt.Valid {
			res.UpdatedAt = &result.UpdatedAt.Time
		}
		if result.Picture.Valid {
			res.Picture = &result.Picture.String
		}
		server.Encode(w, r, http.StatusOK, res)
	}
}

func (s *CustomService) handleUpdateUserName() http.HandlerFunc {
	type request struct {
		Name string `form:"name" json:"name"`
		ID   int64  `form:"id" json:"id"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := server.Decode[request](r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		if str := r.PathValue("id"); str != "" {
			if v, err := strconv.ParseInt(str, 10, 64); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			} else {
				req.ID = v
			}
		}

		user := templates.UserFromContext(r.Context())
		if user.ID <= 0 {
			htmx.Error(w, r, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			return
		}
		if user.ID != req.ID {
			htmx.Error(w, r, http.StatusForbidden, http.StatusText(http.StatusForbidden))
			return
		}
		var arg updateUserNameParams
		arg.Name = strings.TrimSpace(req.Name)
		arg.ID = req.ID

		if arg.Name == "" {
			htmx.Error(w, r, http.StatusUnprocessableEntity, "Name can't be empty")
			return
		}

		if len(arg.Name) > 200 {
			htmx.Error(w, r, http.StatusUnprocessableEntity, "Name > 200 characters")
			return
		}

		_, err = s.querier.updateUserName(r.Context(), arg)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "UpdateUserName")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		htmx.Success(w, r, http.StatusOK, "User data updated!")
	}
}

type UserRequest struct {
	Email   string
	Name    string
	Picture string
}

func (s *Service) GetOrInsert(ctx context.Context, r UserRequest, db *sql.DB) (*User, error) {

	tx, err := db.BeginTx(ctx, &sql.TxOptions{
		ReadOnly:  false,
		Isolation: sql.LevelReadCommitted,
	})
	defer tx.Rollback()

	if err != nil {
		return nil, err
	}

	queries := New(tx)

	var picture sql.NullString
	if r.Picture != "" {
		picture.Valid = true
		picture.String = r.Picture
	}

	user, _ := queries.GetUserByEmail(ctx, r.Email)
	if user.ID > 0 {
		if picture.Valid {
			_, err := queries.updateUserPicture(ctx, updateUserPictureParams{
				Picture: picture,
				ID:      user.ID,
			})
			if err == nil {
				tx.Commit()
			}
		}
		return &user, nil
	}

	if r.Name == "" {
		r.Name = r.Email
	}

	result, err := queries.addUser(ctx, addUserParams{
		Name:    r.Name,
		Email:   r.Email,
		Picture: picture,
		Status:  "VALID",
	})
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	user, err = queries.getUser(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &user, err

}
