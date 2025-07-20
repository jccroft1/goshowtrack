package main

import (
	"bytes"
	"log"
	"net/http"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	lrw.body.Write(b) // capture body for logging
	return lrw.ResponseWriter.Write(b)
}

func loggingMiddleware(next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     200, // default status
		}

		// Call the next handler
		next(lrw, r)

		// Log duration
		duration := time.Since(start)

		// lrw.body.String()
		log.Printf("[%s] %s %s from %s | Status: %d | Duration: %v\n",
			r.Method,
			r.URL.Path,
			r.Proto,
			r.RemoteAddr,
			lrw.statusCode,
			duration,
		)
	})
}
