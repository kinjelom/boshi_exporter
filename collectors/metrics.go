package collectors

import (
	"boshi_exporter/config"
	"boshi_exporter/fetchers"
	"github.com/prometheus/client_golang/prometheus"
	"reflect"
	"strconv"
)

type Metrics interface {
	Collectors() []prometheus.Collector
}

func ListMetricsCollectors(m Metrics) []prometheus.Collector {
	var collectors []prometheus.Collector
	val := reflect.ValueOf(m).Elem()
	for i := 0; i < val.NumField(); i++ {
		fieldVal := val.Field(i)
		if fieldVal.Kind() == reflect.Ptr && !fieldVal.IsNil() {
			if collector, ok := fieldVal.Interface().(prometheus.Collector); ok {
				collectors = append(collectors, collector)
			}
		}
	}
	return collectors
}

const (
	namespace = "boshi"

	programNameLabel    = "program_name"
	programVersionLabel = "program_version"

	environmentLabel   = "environment"
	directorNameLabel  = "bosh_name"
	directorUuidLabel  = "bosh_uuid"
	deploymentLabel    = "bosh_deployment"
	instanceNameLabel  = "bosh_instance_name"
	instanceIdLabel    = "bosh_instance_id"
	instanceIndexLabel = "bosh_instance_index"
	instanceAzLabel    = "bosh_instance_az"
)

func NewInstanceLabels(metricsContext *config.MetricsContext, spec *fetchers.InstanceSpec) *prometheus.Labels {
	return &prometheus.Labels{
		environmentLabel:   metricsContext.Environment,
		directorNameLabel:  metricsContext.BoshName,
		directorUuidLabel:  metricsContext.BoshUuid,
		deploymentLabel:    spec.Deployment,
		instanceNameLabel:  spec.Name,
		instanceIdLabel:    spec.ID,
		instanceIndexLabel: strconv.Itoa(spec.Index),
		instanceAzLabel:    spec.AZ,
	}
}
