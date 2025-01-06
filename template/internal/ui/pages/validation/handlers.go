package validation

import "net/http"

func HandleNameValidation() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := r.PostFormValue("name")
		if name == "" {
			w.Write([]byte("Name is required"))
			return
		}
		if len(name) < 3 {
			w.Write([]byte("Name must be at least 3 characters"))
			return
		}
	})
}
