package main

import (
	"boshi_exporter/collectors"
	"boshi_exporter/config"
	"boshi_exporter/fetchers"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var ProgramVersion = "dev"

const (
	ProgramName = "boshi_exporter"
	ProgramHelp = "Bosh Instance Exporter"
)

func createPromHttpHandler(metricsContext *config.MetricsContext, fetchers *fetchers.Fetchers) (http.Handler, error) {
	registry := prometheus.NewRegistry()
	collector, err := collectors.NewBoshInstanceCollector(ProgramName, ProgramVersion, metricsContext, fetchers)
	if err != nil {
		return nil, err
	}
	registry.MustRegister(collector)
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	return handler, nil
}

func initLogger() *zap.Logger {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
	return logger
}

func main() {
	logger := initLogger()
	defer func() { _ = logger.Sync() }()
	cfg := config.ParseConfig(ProgramName, ProgramHelp, ProgramVersion)
	metricsCtx := cfg.CreateMetricsContext()
	allFetchers := fetchers.NewFetchers(*cfg.BoshSpecPath, *cfg.MonitPath)
	handler, err := createPromHttpHandler(metricsCtx, allFetchers)
	if err != nil {
		zap.L().Error("Failed to create prometheus handler", zap.Error(err))
		os.Exit(1)
	}

	zap.S().Infow("Starting application",
		"program", ProgramName,
		"version", ProgramVersion,
		"listen_address", *cfg.ListenAddress,
		"telemetry_path", *cfg.TelemetryPath,
		"pid", os.Getpid(),
	)

	http.Handle(*cfg.TelemetryPath, promhttp.Handler())
	err = http.ListenAndServe(*cfg.ListenAddress, handler)
	if err != nil {
		zap.L().Error("Failed to start http server", zap.Error(err))
		os.Exit(1)
	}
}
