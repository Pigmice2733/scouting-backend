package server

import (
	"context"
	"fmt"
	"hash"
	"hash/crc32"
	"net/http"
	"strings"

	"github.com/NYTimes/gziphandler"
	"github.com/Pigmice2733/scouting-backend/internal/server/logic"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/fharding1/ezetag"
)

type key int

const (
	keyUsernameCtx key = iota
	keyIsAdminCtx
)

func (s *Server) authHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ss := strings.TrimPrefix(r.Header.Get("Authentication"), "Bearer ")
		token, err := jwt.Parse(ss, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return s.jwtSecret, nil
		})

		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		var username string
		var isAdmin bool

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			var uOk, aOk bool
			username, uOk = claims[logic.SubjectClaim].(string)
			isAdmin, aOk = claims[logic.IsAdminClaim].(bool)

			if !uOk || !aOk {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
		} else {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), keyUsernameCtx, username)
		ctx = context.WithValue(ctx, keyIsAdminCtx, isAdmin)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func adminHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAdmin, ok := r.Context().Value(keyIsAdminCtx).(bool)
		if !ok {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if !isAdmin {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func cors(next http.Handler, origin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Headers", "Authentication")

		next.ServeHTTP(w, r)
	})
}

func limitBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1000000) // 1 MB

		next.ServeHTTP(w, r)
	})
}

var castagoliTable = crc32.MakeTable(crc32.Castagnoli)

func cache(next http.Handler) http.Handler {
	return gziphandler.GzipHandler(ezetag.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=180") // 3 minute max age, overriden by /events

		next.ServeHTTP(w, r)
	}), func() hash.Hash {
		return crc32.New(castagoliTable)
	}))
}
