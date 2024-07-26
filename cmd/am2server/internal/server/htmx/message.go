package htmx

import (
	"encoding/json"
	"net/http"
	"strings"
)

const messagesSelector = "#messages"

type Message struct {
	Code int
	Text string
}

func (m Message) Render(w http.ResponseWriter, r *http.Request) error {
	if strings.Contains(r.Header.Get("accept"), "application/json") {
		w.WriteHeader(m.Code)
		return json.NewEncoder(w).Encode(m)
	}

	if HXRequest(r) {
		w.Header().Set("HX-Retarget", messagesSelector)
		w.Header().Set("HX-Reswap", "beforeend")
	}

	//TODO render message

	return nil
}
