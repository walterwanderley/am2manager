package templates

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"math"
	"net/http"
	"os"
	"strings"
	"unicode"

	"github.com/walterwanderley/am2manager/cmd/am2server/internal/server/htmx"
	"github.com/walterwanderley/am2manager/cmd/am2server/internal/server/watcher"
)

// Content-Security-Policy
const csp = "default-src 'self'; img-src 'self' *.googleusercontent.com; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; frame-src https://www.youtube.com https://youtube.com;"

type key int

const (
	// templatesContext key to store templates on context
	templatesContext key = iota
	// messageContext key to store messages on context
	messageContext
	// paginationContext key to store pagination config on context
	paginationContext
	// userContext key to store logged user
	userContext
)

var (
	//go:embed *
	templatesFS embed.FS
	funcs       = template.FuncMap{}

	Commit, Version string

	provider templatesProvider

	adminMail = os.Getenv("ADMIN")
)

type User struct {
	ID      int64
	Email   string
	Name    string
	Picture string
}

func (u User) Logged() bool {
	return u.ID > 0
}

func (u User) Admin() bool {
	return u.Logged() && u.Email == adminMail
}

func (u User) CanEdit(resouceOwner int64) bool {
	if u.Admin() {
		return true
	}
	return u.Logged() && u.ID == resouceOwner
}

type Pagination struct {
	request *http.Request
	Limit   int64
	Offset  int64
	Total   int64
}

func (p *Pagination) TotalPages() int64 {
	if p.Total == 0 || p.Limit == 0 {
		return 0
	}
	total := float64(p.Total)
	limit := float64(p.Limit)
	return int64(math.Ceil(total / limit))
}

func (p *Pagination) PageNumbers() []int64 {
	total := p.TotalPages()
	pages := make([]int64, total)
	var i int64
	for i = 0; i < total; i++ {
		pages[i] = i + 1
	}
	return pages
}

func (p *Pagination) CurrentPage() int64 {
	return (p.Offset / p.Limit) + 1
}

func (p *Pagination) Next() int64 {
	limit := p.Limit
	if limit == 0 {
		limit = 10
	}
	return p.Offset + limit
}

func (p *Pagination) Prev() int64 {
	limit := p.Limit
	if limit == 0 {
		limit = 10
	}
	offset := p.Offset - limit
	if offset < 0 {
		offset = 0
	}
	return offset
}

func (p *Pagination) URLPage(page int64) string {
	if p == nil {
		return ""
	}
	if page == 1 {
		return p.URL(p.Limit, 0)
	}
	return p.URL(p.Limit, page*p.Limit-p.Limit)
}

func (p *Pagination) URL(limit, offset int64) string {
	if p == nil {
		return ""
	}
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 10
	}
	var url strings.Builder
	url.WriteString(p.request.URL.Path)
	url.WriteString("?")
	for k := range p.request.URL.Query() {
		if k == "limit" || k == "offset" {
			continue
		}
		url.WriteString(fmt.Sprintf("%s=%s&", k, p.request.URL.Query().Get(k)))
	}
	url.WriteString(fmt.Sprintf("limit=%d&offset=%d", limit, offset))
	return url.String()
}

type templatesProvider interface {
	Full() *template.Template
	Dynamic() *template.Template
	TemplatesFS() fs.FS
	DevMode() bool
}

func RegisterHandlers(mux *http.ServeMux, devMode bool) error {
	if devMode {
		templatesPath := "templates"
		provider = &templateDevRender{
			templatesFS: os.DirFS(templatesPath),
		}
		watch, err := watcher.New(templatesPath, "web")
		if err != nil {
			return err
		}
		watch.Start(context.Background())
		mux.Handle("GET /reload", watch)
	} else {
		base := template.New("base.html").Funcs(contextFunctions(context.Background(), nil))
		provider = &templateRender{
			templatesFS: templatesFS,
			dynamic: template.Must(template.ParseFS(templatesFS,
				"layout/content.html")),
			full: template.Must(base.ParseFS(templatesFS,
				"layout/base.html",
				"layout/header.html",
				"layout/footer.html")),
		}
	}
	mux.Handle("GET /", templatesHandler(nil))
	return nil
}

func templatesHandler(content any, templates ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(templates) > 0 {
			r = r.WithContext(ContextWithTemplates(r.Context(), templates...))
		}
		if err := RenderHTML(w, r, content); err != nil {
			slog.Error("render html", "error", err, "path", r.URL.Path)
		}
	})
}

type templateOpts struct {
	Content any
	Title   string
	Version string
	DevMode bool
}

func defaultTemplateOpts(content any) templateOpts {
	return templateOpts{
		Title:   "am2 Server",
		Version: Version,
		Content: content,
		DevMode: provider.DevMode(),
	}
}

func ContextWithPagination(ctx context.Context, pagination *Pagination) context.Context {
	return context.WithValue(ctx, paginationContext, pagination)
}

func ContextWithMessage(ctx context.Context, msg htmx.Message) context.Context {
	return context.WithValue(ctx, messageContext, msg)
}

func ContextWithTemplates(ctx context.Context, templates ...string) context.Context {
	return context.WithValue(ctx, templatesContext, templates)
}

func ContextWithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userContext, user)
}

func UserFromContext(ctx context.Context) User {
	u, _ := ctx.Value(userContext).(User)
	return u
}

func RenderHTML(w http.ResponseWriter, r *http.Request, content any) (err error) {
	if msg, ok := r.Context().Value(messageContext).(htmx.Message); ok {
		return msg.Render(w, r)
	}
	templates := contextTemplates(r)
	if len(templates) == 0 {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return nil
	}
	var base *template.Template
	var obj any
	if r.Header.Get("hx-request") == "true" {
		base, err = provider.Dynamic().Clone()
		obj = content
	} else {
		base, err = provider.Full().Clone()
		obj = defaultTemplateOpts(content)
	}
	if err != nil {
		return err
	}

	t, err := base.Funcs(contextFunctions(r.Context(), r)).ParseFS(provider.TemplatesFS(), templates...)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/html;charset=utf-8")
	w.Header().Set("Content-Security-Policy", csp)
	w.WriteHeader(http.StatusOK)
	err = t.Execute(w, obj)
	return err
}

type templateDevRender struct {
	templatesFS fs.FS
}

func (t *templateDevRender) DevMode() bool {
	return true
}

func (t *templateDevRender) Full() *template.Template {
	base := template.New("base.html").Funcs(contextFunctions(context.Background(), nil))
	return template.Must(base.ParseFS(t.templatesFS,
		"layout/base.html",
		"layout/header.html",
		"layout/footer.html"))
}

func (t *templateDevRender) Dynamic() *template.Template {
	return template.Must(template.ParseFS(t.templatesFS,
		"layout/content.html"))
}

func (t *templateDevRender) TemplatesFS() fs.FS {
	return t.templatesFS
}

type templateRender struct {
	templatesFS fs.FS

	full    *template.Template
	dynamic *template.Template
}

func (t *templateRender) DevMode() bool {
	return false
}

func (t *templateRender) Full() *template.Template {
	return t.full
}

func (t *templateRender) Dynamic() *template.Template {
	return t.dynamic
}

func (t *templateRender) TemplatesFS() fs.FS {
	return t.templatesFS
}

func contextFunctions(ctx context.Context, r *http.Request) template.FuncMap {
	copy := make(template.FuncMap)
	for k, v := range funcs {
		copy[k] = v
	}
	if r != nil {
		copy["Header"] = func(key string) string {
			return r.Header.Get(key)
		}
		copy["HasQuery"] = func(key string) bool {
			return r.URL.Query().Has(key)
		}
		copy["Query"] = func(key string) string {
			return r.URL.Query().Get(key)
		}
		pagination, _ := r.Context().Value(paginationContext).(*Pagination)
		if pagination != nil {
			pagination.request = r
		}
		copy["Pagination"] = func() *Pagination {
			return pagination
		}
	}
	user, _ := ctx.Value(userContext).(User)
	copy["User"] = func() User {
		return user
	}
	copy["LoggedIn"] = func() bool {
		return user.ID > 0
	}
	copy["EmbedFrame"] = func(link string) bool {
		return strings.Contains(link, "youtube.com/embed/")
	}
	copy["FirstChar"] = func(str string) string {
		for _, c := range str {
			return string(unicode.ToUpper(c))
		}
		return str
	}
	return copy
}

func contextTemplates(r *http.Request) []string {
	if v, ok := r.Context().Value(templatesContext).([]string); ok {
		return v
	}
	var templates []string
	path := r.URL.Path
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		path = "index"
	}
	if !strings.HasSuffix(path, ".html") {
		path = path + ".html"
	}
	f, err := provider.TemplatesFS().Open(path)
	if err == nil {
		templates = append(templates, path)
		f.Close()
	}

	return templates
}
