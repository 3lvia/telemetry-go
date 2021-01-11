package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
	"testing"
)

func Test_metrics_getCounter_twice(t *testing.T) {
	// Arrange
	m1 := Metric{
		Name:        "my metric",
		Value:       3.14,
		ConstLabels: map[string]string{
			"a": "b",
		},
	}
	m2 := Metric{
		Name:        "my metric",
		Value:       1.0,
		ConstLabels: map[string]string{
			"a": "c",
		},
	}
	ms := metricService()

	// Act
	c1 := ms.getCounter(m1)
	c2 := ms.getCounter(m2)

	// Assert
	if c1 != c2 {
		t.Error("expected same counter")
	}


}

func metricService() *metrics  {
	return &metrics{
		mux:             &sync.Mutex{},
		promoGauges: map[string]prometheus.Gauge{},
		promoCounters: map[string]prometheus.Counter{},
		promoHistograms: map[string]prometheus.Histogram{},
		logInfo: map[string]string{
			"system": "sys",
			"application": "app",
		},
	}
}