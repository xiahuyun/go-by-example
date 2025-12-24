// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package middleware

import (
	"context"
	"net/http"
)

type UserAgentMiddleware struct {
}

func NewUserAgentMiddleware() *UserAgentMiddleware {
	return &UserAgentMiddleware{}
}

func (m *UserAgentMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val := r.Header.Get("User-Agent")
		reqCtx := r.Context()
		ctx := context.WithValue(reqCtx, "User-Agent", val)
		newReq := r.WithContext(ctx)
		next(w, newReq)
	}
}
