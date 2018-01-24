package server

import (
	"hash"
	"hash/crc32"
	"net/http"
	"strings"

	"github.com/NYTimes/gziphandler"
	"github.com/fharding1/ezetag"
)

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

func cors(next http.Handler, allowedMethods []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(append(allowedMethods, "OPTIONS"), ","))

		if r.Method != "GET" {
			w.Header().Set("Access-Control-Allow-Headers", "Authentication")
		}

		if r.Method == "OPTIONS" {
			return
		} else if !existsIn(r.Method, allowedMethods) {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func limitBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1000000) // 1 MB

		next.ServeHTTP(w, r)
	})
}

func existsIn(str string, strs []string) bool {
	for _, s := range strs {
		if s == str {
			return true
		}
	}
	return false
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

func stdMiddleware(next http.Handler) http.Handler {
	return cors(cache(next), []string{"GET"})
}
