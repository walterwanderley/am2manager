package capture

import (
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/walterwanderley/am2manager"
	"github.com/walterwanderley/am2manager/cmd/am2server/internal/server"
)

type CustomService struct {
	Service
}

func (s *CustomService) RegisterHandlers(mux *http.ServeMux) {
	s.Service.RegisterHandlers(mux)
	mux.HandleFunc("POST /captures-upload", s.handleAddCapture())
	mux.HandleFunc("GET /captures/{id}/file", s.handleGetCaptureFile())
}

func (s *CustomService) handleAddCapture() http.HandlerFunc {
	type request struct {
		UserID      *int64  `form:"user_id" json:"user_id"`
		Name        string  `form:"name" json:"name"`
		Type        string  `form:"type" json:"type"`
		HasCab      *bool   `form:"has_cab" json:"has_cab"`
		Description *string `form:"description" json:"description"`
		Data        []byte  `form:"data" json:"data"`
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
		var arg addCaptureParams
		if req.UserID != nil {
			arg.UserID = sql.NullInt64{Valid: true, Int64: *req.UserID}
		}
		arg.Name = req.Name
		arg.Type = req.Type
		if req.Description != nil {
			arg.Description = sql.NullString{Valid: true, String: *req.Description}
		}
		if req.HasCab != nil {
			arg.HasCab = sql.NullBool{Valid: true, Bool: *req.HasCab}
		}

		file, handler, err := r.FormFile("data")
		if err != nil {
			slog.Error("upload file", "error", err)
			http.Error(w, "upload file error", http.StatusInternalServerError)
			return
		}
		defer file.Close()
		if arg.Name == "" {
			arg.Name = handler.Filename
		}

		data, err := io.ReadAll(file)
		if err != nil {
			slog.Error("read file", "error", err)
			http.Error(w, "read file error", http.StatusInternalServerError)
			return
		}

		var am2data am2manager.Am2Data
		if err := am2data.UnmarshalBinary(data); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		arg.Data, err = am2data.MarshalBinary()
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		arg.Am2Hash = am2data.HashAm2()
		arg.DataHash = am2data.HashData()

		result, err := s.querier.addCapture(r.Context(), arg)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "AddCapture")
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

func (s *CustomService) handleGetCaptureFile() http.HandlerFunc {
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

		result, err := s.querier.GetCaptureFile(r.Context(), id)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "GetCaptureFile")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		filename := strings.TrimSuffix(result.Name, ".am2data")
		filename = strings.TrimSuffix(strings.TrimSuffix(filename, ".am2Data"), ".am2")

		filename += ".am2data"

		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		w.Header().Set("Content-Length", fmt.Sprint(len(result.Data)))
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(result.Data)
	}
}
