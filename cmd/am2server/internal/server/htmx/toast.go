package htmx

import (
	_ "embed"
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
)

var (
	//go:embed toast.html
	toastHTML string

	toastTemplate = template.Must(template.New("toast").Parse(toastHTML))
)

type toast struct {
	Message string `json:"message"`
}

func Toast(w http.ResponseWriter, r *http.Request, code int, text string) error {
	t := toast{
		Message: text,
	}
	if strings.Contains(r.Header.Get("accept"), "application/json") {
		w.WriteHeader(code)
		return json.NewEncoder(w).Encode(t)
	}

	if HXRequest(r) {
		w.Header().Set("HX-Retarget", messagesSelector)
		w.Header().Set("HX-Reswap", "beforeend")
	}

	err := toastTemplate.Execute(w, t)
	if err != nil {
		slog.Error("render toast", "err", err)
	}
	return err
}
