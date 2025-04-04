package capture

import (
	"context"
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
	mux.HandleFunc("POST /captures", s.handleAddCapture())
	mux.HandleFunc("GET /captures/{id}/file", s.handleGetCaptureFile())
	mux.HandleFunc("GET /captures", s.handleSearchCaptures())
	mux.HandleFunc("GET /captures/{id}", s.handleGetCapture())
	mux.HandleFunc("PATCH /captures/{id}", s.handleUpdateCapture())
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
		Rate        float64    `json:"rate,omitempty"`
		Reviews     []Review   `json:"-"`
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

		res.Reviews, err = s.querier.listReviewsByCapture(r.Context(), res.ID)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "ListReviewsByCapture")
			htmx.Error(w, r, http.StatusInternalServerError, err.Error())
			return
		}
		if total := len(res.Reviews); total > 0 {
			var sum int64
			for _, review := range res.Reviews {
				sum += review.Rate
			}
			res.Rate = float64(sum) / float64(total)
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
	return func(w http.ResponseWriter, r *http.Request) {
		htmx.Warning(w, r, http.StatusInternalServerError, "Upload your captures to the m-vave community at https://community.m-vave.com/")
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
		Arg     string `form:"arg" json:"arg"`
		Offset  int64  `form:"offset" json:"offset"`
		Limit   int64  `form:"limit" json:"limit"`
		OrderBy int64  `form:"orderBy" json:"order_by"`
	}
	type response struct {
		ID          int64     `json:"id,omitempty"`
		Name        string    `json:"name,omitempty"`
		Description *string   `json:"description,omitempty"`
		Downloads   int64     `json:"downloads,omitempty"`
		HasCab      *bool     `json:"has_cab,omitempty"`
		Type        string    `json:"type,omitempty"`
		DemoLink    string    `json:"demo_link,omitempty"`
		Rate        float64   `json:"rate,omitempty"`
		Favorite    bool      `json:"favorite,omitempty"`
		CreatedAt   time.Time `json:"created_at,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		if str := r.URL.Query().Get("arg"); str != "" {
			req.Arg = str
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
		if str := r.URL.Query().Get("orderBy"); str != "" {
			if v, err := strconv.ParseInt(str, 10, 64); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			} else {
				req.OrderBy = v
			}
		}

		var arg SearchCapturesParams
		arg.Arg = sql.NullString{Valid: true, String: req.Arg}
		arg.Offset = req.Offset
		arg.Limit = req.Limit
		if arg.Limit == 0 {
			arg.Limit = 10
		}
		arg.OrderBy = req.OrderBy
		user := templates.UserFromContext(r.Context())
		arg.User = user.ID

		result, err := s.querier.SearchCaptures(r.Context(), arg)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "SearchCaptures")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		total, _ := s.querier.totalSearchCaptures(r.Context(), arg.Arg)

		r = r.WithContext(templates.ContextWithPagination(r.Context(), &templates.Pagination{
			Limit:  arg.Limit,
			Offset: arg.Offset,
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
			item.DemoLink = r.DemoLink.String
			if r.Rate.Valid {
				item.Rate = r.Rate.Float64
			}
			if r.Fav.Valid && user.Logged() {
				item.Favorite = r.Fav.Int64 == user.ID
			}
			res = append(res, item)
		}
		server.Encode(w, r, http.StatusOK, res)
	}
}

func (s *CustomService) handleUpdateCapture() http.HandlerFunc {
	type request struct {
		Name        string  `form:"name" json:"name"`
		Description *string `form:"description" json:"description"`
		Type        string  `form:"type" json:"type"`
		HasCab      *bool   `form:"has_cab" json:"has_cab"`
		DemoLink    *string `form:"demo_link" json:"demo_link"`
		ID          int64   `form:"id" json:"id"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req, err := server.Decode[request](r)
		if err != nil {
			htmx.Error(w, r, http.StatusUnprocessableEntity, err.Error())
			return
		}
		if str := r.PathValue("id"); str != "" {
			if v, err := strconv.ParseInt(str, 10, 64); err != nil {
				htmx.Error(w, r, http.StatusBadRequest, err.Error())
				return
			} else {
				req.ID = v
			}
		}

		var arg updateCaptureParams
		arg.Name = req.Name
		if req.Description != nil {
			arg.Description = sql.NullString{Valid: true, String: *req.Description}
		}
		arg.Type = req.Type
		if req.HasCab != nil {
			arg.HasCab = sql.NullBool{Valid: true, Bool: *req.HasCab}
		}
		if req.DemoLink != nil {
			arg.DemoLink = sql.NullString{Valid: true, String: *req.DemoLink}
		}
		arg.ID = req.ID

		user := templates.UserFromContext(r.Context())
		if !user.CanEdit(arg.ID) {
			htmx.Error(w, r, http.StatusForbidden, "Forbidden")
			return
		}

		_, err = s.querier.updateCapture(r.Context(), arg)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "UpdateCapture")
			htmx.Error(w, r, http.StatusInternalServerError, err.Error())
			return
		}
		htmx.Toast(w, r, http.StatusOK, "Capture updated")
	}
}

func toAm2DataExtension(filename string) string {
	filename = strings.ReplaceAll(filename, " ", "_")
	i := strings.LastIndex(filename, ".")
	if i < 0 {
		return filename + ".am2data"
	}
	return filename[0:i] + ".am2data"
}

func toFriendlyName(filename string) string {
	filename = strings.ReplaceAll(filename, "_", " ")
	i := strings.LastIndex(filename, ".")
	if i > 1 {
		return filename[0:i]
	}
	return filename
}

const searchCaptures = `-- name: SearchCaptures :many
SELECT c.id, c.name, c.description, c.downloads, c.has_cab, c.type, c.created_at, c.demo_link, AVG(r.rate) rate, uf.user_id fav
FROM capture c LEFT OUTER JOIN review r ON (c.id = r.capture_id)
LEFT OUTER JOIN user_favorite uf ON (c.id = uf.capture_id)
WHERE c.description LIKE '%'||?1||'%' OR c.name LIKE '%'||?1||'%' 
OR c.data_hash = ?1 OR c.am2_hash = ?1
AND (uf.user_id = ?2 OR uf.user_id IS NULL)
GROUP BY c.id
ORDER BY %order_by% DESC, c.downloads DESC
LIMIT ?4 OFFSET ?3
`

type SearchCapturesParams struct {
	Arg     sql.NullString
	User    int64
	Offset  int64
	Limit   int64
	OrderBy int64
}

type SearchCapturesRow struct {
	ID          int64
	Name        string
	Description sql.NullString
	Downloads   int64
	HasCab      sql.NullBool
	Type        string
	CreatedAt   time.Time
	DemoLink    sql.NullString
	Rate        sql.NullFloat64
	Fav         sql.NullInt64
}

// http: GET /captures
func (q *Queries) SearchCaptures(ctx context.Context, arg SearchCapturesParams) ([]SearchCapturesRow, error) {
	orderBy := "c.downloads"
	if arg.OrderBy == 2 {
		orderBy = "c.created_at"
	}
	query := strings.Replace(searchCaptures, "%order_by%", orderBy, 1)
	rows, err := q.db.QueryContext(ctx, query,
		arg.Arg,
		arg.User,
		arg.Offset,
		arg.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SearchCapturesRow
	for rows.Next() {
		var i SearchCapturesRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Downloads,
			&i.HasCab,
			&i.Type,
			&i.CreatedAt,
			&i.DemoLink,
			&i.Rate,
			&i.Fav,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
