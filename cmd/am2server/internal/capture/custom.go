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
	mux.HandleFunc("GET /find-capture", s.handleFindCapture())
	mux.HandleFunc("POST /captures-convert", s.handleConvertCapture())
	mux.HandleFunc("POST /captures-upload", s.handleAddCapture())
	mux.HandleFunc("GET /captures/{id}/file", s.handleGetCaptureFile())
}

func (s *CustomService) handleFindCapture() http.HandlerFunc {
	type response struct {
		MostDownloaded []mostDownloadedCapturesRow
		MostRecent     []mostRecentCapturesRow
	}
	return func(w http.ResponseWriter, r *http.Request) {
		mostDownloaded, err := s.querier.mostDownloadedCaptures(r.Context())
		if err != nil {
			slog.Error("most downloaded query", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		mostRecent, err := s.querier.mostRecentCaptures(r.Context())
		if err != nil {
			slog.Error("most recent query", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		server.Encode(w, r, http.StatusOK, response{
			MostDownloaded: mostDownloaded,
			MostRecent:     mostRecent,
		})
	}
}

func (s *CustomService) handleConvertCapture() http.HandlerFunc {
	type request struct {
		Level   int    `form:"level" json:"level"`
		Mix     int    `form:"mix" json:"mix"`
		GainMin int    `form:"gain_min" json:"gain_min"`
		GainMax int    `form:"gain_max" json:"gain_max"`
		Data    []byte `form:"file" json:"file"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := server.Decode[request](r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		if req.Level < 0 || req.Level > 255 {
			http.Error(w, "level must be a value between 0 and 255", http.StatusBadRequest)
			return
		}

		if req.Mix < 0 || req.Mix > 100 {
			http.Error(w, "mix must be a value between 0 and 100", http.StatusBadRequest)
			return
		}

		if req.GainMin < 0 || req.GainMin > 100 {
			http.Error(w, "gain-min must be a value between 0 and 100", http.StatusBadRequest)
			return
		}

		if req.GainMax < 0 || req.GainMax > 100 {
			http.Error(w, "gain-max must be a value between 0 and 100", http.StatusBadRequest)
			return
		}

		if req.GainMin > req.GainMax {
			http.Error(w, "gain-min can't be greater than gain-max", http.StatusBadRequest)
			return
		}

		file, handler, err := r.FormFile("data")
		if err != nil {
			slog.Error("upload file", "error", err)
			http.Error(w, "upload file error", http.StatusInternalServerError)
			return
		}
		defer file.Close()

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

		am2data.Level = byte(req.Level)
		am2data.Mix = byte(req.Mix)
		am2data.GainMin = byte(req.GainMin)
		am2data.GainMax = byte(req.GainMax)

		result, err := am2data.MarshalBinary()
		if err != nil {
			slog.Error("marshal file", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return

		}

		filename := toAm2DataExtension(handler.Filename)

		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		w.Header().Set("Content-Length", fmt.Sprint(len(result)))
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(result)
	}
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

		filename := toAm2DataExtension(result.Name)

		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		w.Header().Set("Content-Length", fmt.Sprint(len(result.Data)))
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(result.Data)
	}
}

func toAm2DataExtension(filename string) string {
	i := strings.LastIndex(filename, ".")
	if i < 0 {
		return filename
	}
	return filename[0:i] + ".am2data"
}
