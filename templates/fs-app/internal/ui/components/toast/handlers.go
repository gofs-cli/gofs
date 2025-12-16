package toast

import (
	"net/http"

	"github.com/a-h/templ"
)

func Success(w http.ResponseWriter, r *http.Request, msg string) {
	templ.Handler(success(msg)).ServeHTTP(w, r)
}

func Info(w http.ResponseWriter, r *http.Request, msg string) {
	templ.Handler(info(msg)).ServeHTTP(w, r)
}

func Warning(w http.ResponseWriter, r *http.Request, msg string) {
	templ.Handler(warning(msg)).ServeHTTP(w, r)
}

func Error(w http.ResponseWriter, r *http.Request, status int, msg string) {
	w.Header().Add("HX-Reswap", "none")
	w.WriteHeader(status)
	templ.Handler(err(msg)).ServeHTTP(w, r)
}
