package http

import (
	"net/http"
	"os"

	"singularity.com/pr7/services/tasks/internal/repository"
	"singularity.com/pr7/services/tasks/internal/service"
	sharedlogger "singularity.com/pr7/shared/logger"
	"singularity.com/pr7/shared/metrics"
	"singularity.com/pr7/shared/middleware"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func NewRouter(log *logrus.Logger) http.Handler {
	serviceLog := sharedlogger.WithService(log, "tasks")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		serviceLog.WithFields(logrus.Fields{
			"component": "startup",
			"error":     "DATABASE_URL is empty",
		}).Error("failed to start tasks service")
		os.Exit(1)
	}

	db, err := repository.OpenPostgres(dbURL)
	if err != nil {
		serviceLog.WithFields(logrus.Fields{
			"component": "database",
			"error":     err.Error(),
		}).Error("failed to connect to database")
		os.Exit(1)
	}

	repo := repository.NewPostgresTaskRepository(db)
	svc := service.NewTaskService(repo)
	handler := NewHandler(svc, log)

	reg := prometheus.NewRegistry()
	httpMetrics := metrics.NewHTTPMetrics(reg)

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/tasks/search", handler.SearchTasks)
	mux.HandleFunc("/v1/tasks", handler.Tasks)
	mux.HandleFunc("/v1/tasks/", handler.TaskByID)
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	return middleware.RequestID(
		middleware.SecurityHeaders(
			middleware.Metrics(httpMetrics)(
				middleware.AccessLog(serviceLog)(
					middleware.CSRFMiddleware(mux),
				),
			),
		),
	)
}
