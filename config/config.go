package config

import (
	"github.com/alecthomas/kingpin/v2"
	"os"
)

type Config struct {
	ListenAddress      *string
	TelemetryPath      *string
	BoshSpecPath       *string
	MonitPath          *string
	MetricsNamespace   *string
	MetricsEnvironment *string
	MetricsBoshName    *string
	MetricsBoshUuid    *string
	LogLevel           *string
	LogPath            *string
}

func ParseConfig(programName, programHelp, programVersion string) *Config {
	app := kingpin.New(programName, programHelp)
	config := &Config{
		ListenAddress: app.Flag(
			"web.listen-address", "Address to listen on for web interface and telemetry ($BOSHI_EXPORTER_WEB_LISTEN_ADDRESS)",
		).Envar("BOSHI_EXPORTER_WEB_LISTEN_ADDRESS").Default(":9191").String(),

		TelemetryPath: app.Flag(
			"web.telemetry-path", "Path under which to expose Prometheus metrics ($BOSHI_EXPORTER_WEB_TELEMETRY_PATH)",
		).Envar("BOSHI_EXPORTER_WEB_TELEMETRY_PATH").Default("/metrics").String(),

		BoshSpecPath: app.Flag(
			"bosh.spec-path", "Path to the Bosh instance spec.json, default: /var/vcap/bosh/spec.json ($BOSHI_EXPORTER_BOSH_SPEC_PATH)",
		).Envar("BOSHI_EXPORTER_BOSH_SPEC_PATH").Default("/var/vcap/bosh/spec.json").String(),

		MonitPath: app.Flag(
			"monit.path", "Path to the Monit program, default: /var/vcap/bosh/bin/monit ($BOSHI_EXPORTER_MONIT_PATH)",
		).Envar("BOSHI_EXPORTER_MONIT_PATH").Default("/var/vcap/bosh/bin/monit").String(),

		MetricsNamespace: app.Flag(
			"metrics.namespace", "Metrics namespace, default: boshi ($BOSHI_EXPORTER_METRICS_NAMESPACE)",
		).Envar("BOSHI_EXPORTER_METRICS_NAMESPACE").Default("boshi").String(),

		MetricsEnvironment: app.Flag(
			"metrics.environment", "Environment label (e.g. prod/dev) to be attached to metrics ($BOSHI_EXPORTER_METRICS_ENVIRONMENT)",
		).Envar("BOSHI_EXPORTER_METRICS_ENVIRONMENT").Default("").String(),

		MetricsBoshName: app.Flag(
			"metrics.bosh-name", "Bosh director name label to be attached to metrics ($BOSHI_EXPORTER_METRICS_BOSH_NAME)",
		).Envar("BOSHI_EXPORTER_METRICS_BOSH_NAME").Default("").String(),
		MetricsBoshUuid: app.Flag(
			"metrics.bosh-uuid", "Bosh director UUID label to be attached to metrics ($BOSHI_EXPORTER_METRICS_BOSH_UUID)",
		).Envar("BOSHI_EXPORTER_METRICS_BOSH_UUID").Default("").String(),

		LogLevel: app.Flag(
			"log.level", "Defines the minimum severity of messages that will be emitted, can be: debug, info, warn, error. ($BOSHI_EXPORTER_LOG_LEVEL)",
		).Envar("BOSHI_EXPORTER_LOG_LEVEL").Default("info").String(),
		LogPath: app.Flag(
			"log.path", "Specifies where logs are written, can be: stdout, stderr, any file path. ($BOSHI_EXPORTER_LOG_PATH)",
		).Envar("BOSHI_EXPORTER_LOG_PATH").Default("stdout").String(),
	}
	app.Version(programVersion)
	app.HelpFlag.Short('h')
	kingpin.MustParse(app.Parse(os.Args[1:]))
	return config
}

type MetricsContext struct {
	Namespace   string
	Environment string
	BoshName    string
	BoshUuid    string
}

func (c *Config) CreateMetricsContext() *MetricsContext {
	return &MetricsContext{
		Namespace:   *c.MetricsNamespace,
		Environment: *c.MetricsEnvironment,
		BoshName:    *c.MetricsBoshName,
		BoshUuid:    *c.MetricsBoshUuid,
	}
}
