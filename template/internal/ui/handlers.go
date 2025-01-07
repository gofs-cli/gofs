package ui

import (
	"net/http"

	"github.com/gofs-cli/gofs/template/internal/ui/components/toast"

	"github.com/a-h/templ"
)

func Success() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		toast.Success(w, r, "Success!")
	})
}

func Info() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		toast.Info(w, r, "Info!")
	})
}

func Warning() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		toast.Warning(w, r, "Warning!")
	})
}

func Error() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		toast.Error(w, r, http.StatusInternalServerError, "Error!")
	})
}

func Modal() http.Handler {
	return templ.Handler(ModalDemo())
}
