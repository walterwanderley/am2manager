package htmx

import (
	_ "embed"
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
)

const messagesSelector = "#messages"

var (
	//go:embed message.html
	messageHTML string

	messageTemplate = template.Must(template.New("message").Parse(messageHTML))
)

type MessageType string

const (
	TypeInfo    = MessageType("info")
	TypeSuccess = MessageType("success")
	TypeError   = MessageType("error")
	TypeWarning = MessageType("warning")
)

func (t MessageType) Icon() string {
	switch t {
	case TypeInfo:
		return "info_outline"
	case TypeSuccess:
		return "check"
	case TypeError:
		return "error_outline"
	case TypeWarning:
		return "warning"
	default:
		return "check_circle"
	}
}

type Message struct {
	Code int
	Text string
	Type MessageType
}

func NewMessage(code int, text string, typ MessageType) Message {
	return Message{Code: code,
		Text: text,
		Type: typ}
}

func ErrorMessage(text string) Message {
	return NewMessage(http.StatusInternalServerError, text, TypeError)
}

func InfoMessage(text string) Message {
	return NewMessage(http.StatusOK, text, TypeInfo)
}

func SuccessMessage(text string) Message {
	return NewMessage(http.StatusOK, text, TypeSuccess)
}

func WarningMessage(text string) Message {
	return NewMessage(http.StatusInternalServerError, text, TypeWarning)
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

	return messageTemplate.Execute(w, m)
}

func Info(w http.ResponseWriter, r *http.Request, code int, text string) error {
	return NewMessage(code, text, TypeInfo).Render(w, r)
}

func Success(w http.ResponseWriter, r *http.Request, code int, text string) error {
	return NewMessage(code, text, TypeSuccess).Render(w, r)
}

func Error(w http.ResponseWriter, r *http.Request, code int, text string) error {
	return NewMessage(code, text, TypeError).Render(w, r)
}

func Warning(w http.ResponseWriter, r *http.Request, code int, text string) error {
	return NewMessage(code, text, TypeWarning).Render(w, r)
}
