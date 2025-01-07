package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gofs-cli/gofs/template/internal/config"
)

type contextKey string

var userContextKey contextKey = "user"

func UserFromContext(ctx context.Context) *User {
	if user, ok := ctx.Value(userContextKey).(*User); ok {
		return user
	}
	return nil
}

func WithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func Middleware(env config.Environment) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if env.Local() {
				r = r.WithContext(WithUser(r.Context(), &LocalUser))
			} else {
				fmt.Println("TODO: implement token parsing")
				// get user from token
				// TODO: implement token parsing
			}
			next.ServeHTTP(w, r)
		})
	}
}
