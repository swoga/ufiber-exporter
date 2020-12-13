package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

func AddMetricsOnu(registry prometheus.Registerer) {
	registry = prometheus.WrapRegistererWithPrefix("onu_", registry)
}
