package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func randomHTTPStatus() int {
	codes := []int{
		http.StatusOK,
		http.StatusCreated,
		http.StatusAccepted,
		http.StatusNoContent,
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusInternalServerError,
	}

	return codes[rand.IntN(len(codes))]
}

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"endpoint", "method", "status"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
}

func main() {
	opts := &slog.HandlerOptions{
		AddSource: true,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		status := randomHTTPStatus()

		w.WriteHeader(status)

		if status == http.StatusOK {
			w.Write([]byte(" { message: hello } "))
		}

		httpRequestsTotal.WithLabelValues("/", "GET", fmt.Sprintf("%d", status)).Inc()

		logger.Info("Request received",
			slog.String("method", "GET"),
			slog.String("path", "/"),
			slog.Int("status", status),
		)
	})

	mux.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		status := randomHTTPStatus()

		w.WriteHeader(status)

		httpRequestsTotal.WithLabelValues("/", "POST", fmt.Sprintf("%d", status)).Inc()

		logger.Info("Request received",
			slog.String("method", "POST"),
			slog.String("path", "/"),
			slog.Int("status", status),
		)
	})

	mux.Handle("GET /metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	logger.Info("Starting HTTP server",
		slog.String("host", "localhost"),
		slog.Int("port", 8080),
	)

	err := server.ListenAndServe()

	if err != nil {
		ctx := context.Background()
		logger.ErrorContext(ctx, "Failed to start the HTTP server",
			slog.String("host", "localhost"),
			slog.Int("port", 8080),
		)
	}
}
