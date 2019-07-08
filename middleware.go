package main

import (
	"net/http"
)

func MiddlewareCheckCredentials(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userid, login, err := ReadCookies(w, r)
		if err != nil {
			if err == http.ErrNoCookie {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
		}

		if !CheckAuth(userid, login) {
			http.Redirect(w, r, "/", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
