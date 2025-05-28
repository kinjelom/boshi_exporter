package collectors

import (
	"boshi_exporter/config"
	"boshi_exporter/fetchers"
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type BoshInstanceCollector struct {
	fetchers      *fetchers.Fetchers
	baseMetrics   *BaseMetrics
	monitMetrics  *MonitMetrics
	systemMetrics *SystemMetrics
	instanceSpec  *fetchers.InstanceSpec
}

func NewBoshInstanceCollector(programName, programVersion string, metricsContext *config.MetricsContext, fetchers *fetchers.Fetchers) (*BoshInstanceCollector, error) {
	instanceSpec, err := fetchers.SpecFetcher.Fetch(context.Background())
	if err != nil {
		return nil, err
	}
	return &BoshInstanceCollector{
		fetchers:      fetchers,
		baseMetrics:   NewBaseMetrics(programName, programVersion, metricsContext, instanceSpec),
		monitMetrics:  NewMonitMetrics(metricsContext, instanceSpec),
		systemMetrics: NewSystemMetrics(metricsContext, instanceSpec),
		instanceSpec:  instanceSpec,
	}, nil
}

func (b *BoshInstanceCollector) Describe(ch chan<- *prometheus.Desc) {
	b.describeAllMetrics(b.baseMetrics, ch)
	b.describeAllMetrics(b.monitMetrics, ch)
	b.describeAllMetrics(b.systemMetrics, ch)
}

func (b *BoshInstanceCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()
	b.baseMetrics.Emit()

	monitStat, err := b.fetchers.MonitFetcher.Fetch(ctx)
	if err != nil {
		zap.L().Error("Failed to fetch monit stat, some monitMetrics won't be updated", zap.Error(err))
	} else {
		b.monitMetrics.Emit(monitStat)
	}

	systemStat, err := b.fetchers.SystemFetcher.Fetch(ctx)
	if err != nil {
		zap.L().Error("Failed to fetch system stat, some monitMetrics won't be updated", zap.Error(err))
	} else {
		b.systemMetrics.Emit(systemStat)
	}

	b.collectAllMetrics(b.baseMetrics, ch)
	b.collectAllMetrics(b.monitMetrics, ch)
	b.collectAllMetrics(b.systemMetrics, ch)
}

func (b *BoshInstanceCollector) collectAllMetrics(metrics Metrics, ch chan<- prometheus.Metric) {
	for _, c := range metrics.Collectors() {
		c.Collect(ch)
	}
}

func (b *BoshInstanceCollector) describeAllMetrics(metrics Metrics, ch chan<- *prometheus.Desc) {
	for _, c := range metrics.Collectors() {
		c.Describe(ch)
	}
}
