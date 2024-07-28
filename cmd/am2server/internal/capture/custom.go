package capture

import (
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/walterwanderley/am2manager"
	"github.com/walterwanderley/am2manager/cmd/am2server/internal/server"
	"github.com/walterwanderley/am2manager/cmd/am2server/internal/server/htmx"
	"github.com/walterwanderley/am2manager/cmd/am2server/templates"
)

type CustomService struct {
	Service
}

func (s *CustomService) RegisterHandlers(mux *http.ServeMux) {
	//s.Service.RegisterHandlers(mux)
	mux.HandleFunc("GET /find-capture", s.handleFindCapture())
	mux.HandleFunc("POST /captures-convert", s.handleConvertCapture())
	mux.HandleFunc("POST /captures-upload", s.handleAddCapture())
	mux.HandleFunc("GET /captures/{id}/file", s.handleGetCaptureFile())
	mux.HandleFunc("GET /captures", s.handleSearchCaptures())
	mux.HandleFunc("GET /captures/{id}", s.handleGetCapture())
}

func (s *CustomService) handleGetCapture() http.HandlerFunc {
	type request struct {
		Id int64 `form:"id" json:"id"`
	}
	type response struct {
		ID          int64      `json:"id,omitempty"`
		UserID      *int64     `json:"user_id,omitempty"`
		Name        string     `json:"name,omitempty"`
		Description *string    `json:"description,omitempty"`
		Type        string     `json:"type,omitempty"`
		HasCab      *bool      `json:"has_cab,omitempty"`
		Am2Hash     string     `json:"am2_hash,omitempty"`
		DataHash    string     `json:"data_hash,omitempty"`
		Data        []byte     `json:"data,omitempty"`
		DemoLink    string     `json:"demo_link,omitempty"`
		Downloads   int64      `json:"downloads,omitempty"`
		CreatedAt   time.Time  `json:"created_at,omitempty"`
		UpdatedAt   *time.Time `json:"updated_at,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(templates.ContextWithTemplates(r.Context(), "captures/{id}.html"))
		var req request
		if str := r.PathValue("id"); str != "" {
			if v, err := strconv.ParseInt(str, 10, 64); err != nil {
				htmx.Error(w, r, http.StatusBadRequest, err.Error())
				return
			} else {
				req.Id = v
			}
		}
		id := req.Id

		result, err := s.querier.GetCapture(r.Context(), id)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "GetCapture")
			htmx.Error(w, r, http.StatusInternalServerError, err.Error())
			return
		}
		var res response
		res.ID = result.ID
		if result.UserID.Valid {
			res.UserID = &result.UserID.Int64
		}
		res.Name = result.Name
		if result.Description.Valid {
			res.Description = &result.Description.String
		}
		res.Type = result.Type
		if result.HasCab.Valid {
			res.HasCab = &result.HasCab.Bool
		}
		res.Am2Hash = result.Am2Hash
		res.DataHash = result.DataHash
		res.Data = result.Data
		res.Downloads = result.Downloads
		res.DemoLink = result.DemoLink.String
		res.CreatedAt = result.CreatedAt
		if result.UpdatedAt.Valid {
			res.UpdatedAt = &result.UpdatedAt.Time
		}
		server.Encode(w, r, http.StatusOK, res)
	}
}

func (s *CustomService) handleFindCapture() http.HandlerFunc {
	type response struct {
		MostDownloaded []mostDownloadedCapturesRow
		MostRecent     []mostRecentCapturesRow
	}
	return func(w http.ResponseWriter, r *http.Request) {
		mostDownloaded, err := s.querier.mostDownloadedCaptures(r.Context())
		if err != nil {
			slog.Warn("most downloaded query", "error", err)
			htmx.Warning(w, r, http.StatusInternalServerError, err.Error())
			return
		}

		mostRecent, err := s.querier.mostRecentCaptures(r.Context())
		if err != nil {
			slog.Warn("most recent query", "error", err)
			htmx.Warning(w, r, http.StatusInternalServerError, err.Error())
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
		ctx := templates.ContextWithTemplates(r.Context(), "convert-capture.html")
		r = r.WithContext(ctx)
		req, err := server.Decode[request](r)
		if err != nil {
			server.Encode(w, r, http.StatusUnprocessableEntity, htmx.ErrorMessage(err.Error()))
			return
		}

		if req.Level < 0 || req.Level > 255 {
			server.Encode(w, r, http.StatusUnprocessableEntity, htmx.ErrorMessage("level must be a value between 0 and 255"))
			return
		}

		if req.Mix < 0 || req.Mix > 100 {
			server.Encode(w, r, http.StatusUnprocessableEntity, htmx.ErrorMessage("mix must be a value between 0 and 100"))
			return
		}

		if req.GainMin < 0 || req.GainMin > 100 {
			server.Encode(w, r, http.StatusUnprocessableEntity, htmx.ErrorMessage("gain-min must be a value between 0 and 100"))
			return
		}

		if req.GainMax < 0 || req.GainMax > 100 {
			server.Encode(w, r, http.StatusUnprocessableEntity, htmx.ErrorMessage("gain-max must be a value between 0 and 100"))
			return
		}

		if req.GainMin > req.GainMax {
			server.Encode(w, r, http.StatusUnprocessableEntity, htmx.ErrorMessage("gain-min can't be greater than gain-max"))
			return
		}

		file, handler, err := r.FormFile("data")
		if err != nil {
			slog.Error("upload file", "error", err)
			server.Encode(w, r, http.StatusUnprocessableEntity, htmx.ErrorMessage("upload file error:"+err.Error()))
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			slog.Error("read file", "error", err)
			server.Encode(w, r, http.StatusUnprocessableEntity, htmx.ErrorMessage("read file error:"+err.Error()))
			return
		}

		var am2data am2manager.Am2Data
		if err := am2data.UnmarshalBinary(data); err != nil {
			slog.Error("unmarshal binary", "error", err)
			server.Encode(w, r, http.StatusUnprocessableEntity, htmx.ErrorMessage("unmarshal binary error:"+err.Error()))
			return
		}

		am2data.Level = byte(req.Level)
		am2data.Mix = byte(req.Mix)
		am2data.GainMin = byte(req.GainMin)
		am2data.GainMax = byte(req.GainMax)

		result, err := am2data.MarshalBinary()
		if err != nil {
			slog.Error("marshal file", "error", err)
			server.Encode(w, r, http.StatusUnprocessableEntity, htmx.ErrorMessage("marshal binary error:"+err.Error()))
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
		DemoLink    *string `form:"demo_link" json:"demo_link"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := server.Decode[request](r)
		if err != nil {
			htmx.Error(w, r, http.StatusUnprocessableEntity, err.Error())
			return
		}
		var arg addCaptureParams
		if req.UserID != nil {
			arg.UserID = sql.NullInt64{Valid: true, Int64: *req.UserID}
		}
		arg.Name = req.Name
		arg.Type = req.Type
		if req.Description != nil {
			if len(*req.Description) > 1000 {
				htmx.Warning(w, r, http.StatusUnprocessableEntity, "the description can't be greater than 1000 characters")
				return
			}
			arg.Description = sql.NullString{Valid: true, String: *req.Description}
		}
		if req.HasCab != nil {
			arg.HasCab = sql.NullBool{Valid: true, Bool: *req.HasCab}
		}
		if req.DemoLink != nil {
			if len(*req.DemoLink) > 200 {
				htmx.Warning(w, r, http.StatusUnprocessableEntity, "the demonstration link can't be greater than 200 characters")
				return
			}
			arg.DemoLink = sql.NullString{Valid: true, String: *req.DemoLink}
		}

		file, handler, err := r.FormFile("data")
		if err != nil {
			slog.Error("upload file", "error", err)
			htmx.Error(w, r, http.StatusInternalServerError, "upload file error: "+err.Error())
			return
		}
		defer file.Close()
		if arg.Name == "" {
			arg.Name = handler.Filename
		}

		if len(arg.Name) > 200 {
			htmx.Warning(w, r, http.StatusUnprocessableEntity, "the filename can't be greater than 200 characters")
			return
		}

		data, err := io.ReadAll(file)
		if err != nil {
			slog.Error("read file", "error", err)
			htmx.Error(w, r, http.StatusInternalServerError, "read file error: "+err.Error())
			return
		}

		var am2data am2manager.Am2Data
		if err := am2data.UnmarshalBinary(data); err != nil {
			htmx.Error(w, r, http.StatusUnprocessableEntity, "unmarshal binary error: "+err.Error())
			return
		}

		arg.Data, err = am2data.MarshalBinary()
		if err != nil {
			htmx.Error(w, r, http.StatusInternalServerError, "marshal binary error: "+err.Error())
			return
		}

		arg.Am2Hash = am2data.HashAm2()
		arg.DataHash = am2data.HashData()

		protected, _ := s.querier.protectedTrainer(r.Context(), arg.Am2Hash)
		if protected.Ref != "" {
			slog.Warn("am2 protected", "filename", arg.Name, "am2_hash", arg.Am2Hash)
			htmx.Warning(w, r, http.StatusBadRequest, "am2 file protected by "+protected.Ref)
			return
		}

		result, err := s.querier.addCapture(r.Context(), arg)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "AddCapture")
			if strings.Contains(err.Error(), "UNIQUE constraint failed: capture.data_hash") {
				htmx.Warning(w, r, http.StatusBadRequest, fmt.Sprintf("Capture %q already registered", arg.DataHash))
				return
			}
			htmx.Error(w, r, http.StatusInternalServerError, "error: "+err.Error())
			return
		}
		lastInsertId, _ := result.LastInsertId()
		htmx.Success(w, r, http.StatusOK, fmt.Sprintf("Capture saved! %d", lastInsertId))
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

func (s *CustomService) handleSearchCaptures() http.HandlerFunc {
	type request struct {
		Arg    *string `form:"arg" json:"arg"`
		Offset int64   `form:"offset" json:"offset"`
		Limit  int64   `form:"limit" json:"limit"`
	}
	type response struct {
		ID          int64     `json:"id,omitempty"`
		Name        string    `json:"name,omitempty"`
		Description *string   `json:"description,omitempty"`
		Downloads   int64     `json:"downloads,omitempty"`
		HasCab      *bool     `json:"has_cab,omitempty"`
		Type        string    `json:"type,omitempty"`
		CreatedAt   time.Time `json:"created_at,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		if str := r.URL.Query().Get("arg"); str != "" {
			req.Arg = &str
		}
		if str := r.URL.Query().Get("offset"); str != "" {
			if v, err := strconv.ParseInt(str, 10, 64); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			} else {
				req.Offset = v
			}
		}
		if str := r.URL.Query().Get("limit"); str != "" {
			if v, err := strconv.ParseInt(str, 10, 64); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			} else {
				req.Limit = v
			}
		}
		var arg SearchCapturesParams
		if req.Arg != nil {
			arg.Arg = sql.NullString{Valid: true, String: *req.Arg}
		}
		arg.Offset = req.Offset
		arg.Limit = req.Limit

		result, err := s.querier.SearchCaptures(r.Context(), arg)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "SearchCaptures")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		slog.Info("RESULT", "len", len(result))
		total, _ := s.querier.totalSearchCaptures(r.Context(), arg.Arg)

		r = r.WithContext(templates.ContextWithPagination(r.Context(), &templates.Pagination{
			Limit:  req.Limit,
			Offset: req.Offset,
			Total:  total,
		}))
		res := make([]response, 0)
		for _, r := range result {
			var item response
			item.ID = r.ID
			item.Name = r.Name
			if r.Description.Valid {
				item.Description = &r.Description.String
			}
			item.Downloads = r.Downloads
			if r.HasCab.Valid {
				item.HasCab = &r.HasCab.Bool
			}
			item.Type = r.Type
			item.CreatedAt = r.CreatedAt
			res = append(res, item)
		}
		server.Encode(w, r, http.StatusOK, res)
	}
}

func toAm2DataExtension(filename string) string {
	i := strings.LastIndex(filename, ".")
	if i < 0 {
		return filename + ".am2data"
	}
	return filename[0:i] + ".am2data"
}
