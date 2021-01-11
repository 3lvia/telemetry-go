package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
	"testing"
)

func Test_metricVectors_getCounter(t *testing.T) {
	// Arrange
	m1 := Metric{
		Name:        "cost",
		Value:       2,
		ConstLabels: map[string]string{
			"cloud": "gcp",
		},
	}
	m2 := Metric{
		Name:        "cost",
		Value:       233,
		ConstLabels: map[string]string{
			"cloud": "gcp",
		},
	}
	m3 := Metric{
		Name:        "cost",
		Value:       233,
		ConstLabels: map[string]string{
			"cloud": "azure",
		},
	}
	mv := mVectors()

	// Act
	c1 := mv.getCounter(m1)
	c2 := mv.getCounter(m2)
	c3 := mv.getCounter(m3)

	// Assert
	if c1 != c2 {
		t.Error("expected same counter for cloud 'gcp'")
	}
	if c3 == c2 {
		t.Error("expected cloud 'azure' to be unequal to 'gcp'")
	}
}

func Test_metricVectors_getGauge(t *testing.T) {
	// Arrange
	m1 := Metric{
		Name:        "cost",
		Value:       2,
		ConstLabels: map[string]string{
			"cloud": "gcp",
		},
	}
	m2 := Metric{
		Name:        "cost",
		Value:       233,
		ConstLabels: map[string]string{
			"cloud": "gcp",
		},
	}
	m3 := Metric{
		Name:        "cost",
		Value:       233,
		ConstLabels: map[string]string{
			"cloud": "azure",
		},
	}
	mv := mVectors()

	// Act
	c1 := mv.getGauge(m1)
	c2 := mv.getGauge(m2)
	c3 := mv.getGauge(m3)

	// Assert
	if c1 != c2 {
		t.Error("expected same counter for cloud 'gcp'")
	}
	if c3 == c2 {
		t.Error("expected cloud 'azure' to be unequal to 'gcp'")
	}
}

func Test_metricVectors_getHistogram(t *testing.T) {
	// Arrange
	m1 := Metric{
		Name:        "cost",
		Value:       2,
		ConstLabels: map[string]string{
			"cloud": "gcp",
		},
	}
	m2 := Metric{
		Name:        "cost",
		Value:       233,
		ConstLabels: map[string]string{
			"cloud": "gcp",
		},
	}
	m3 := Metric{
		Name:        "cost",
		Value:       233,
		ConstLabels: map[string]string{
			"cloud": "azure",
		},
	}
	mv := mVectors()

	// Act
	c1 := mv.getHistogram(m1)
	c2 := mv.getHistogram(m2)
	c3 := mv.getHistogram(m3)

	// Assert
	if c1 != c2 {
		t.Error("expected same counter for cloud 'gcp'")
	}
	if c3 == c2 {
		t.Error("expected cloud 'azure' to be unequal to 'gcp'")
	}
}

func mVectors() *metricVectors {
	return &metricVectors{
		mux:       &sync.Mutex{},
		counters:  map[string]*prometheus.CounterVec{},
		gauges:    map[string]*prometheus.GaugeVec{},
		histograms: map[string]*prometheus.HistogramVec{},
		namespace: "system",
		subsystem: "app",
	}
}
