package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"sort"
	"sync"
)

type metricVectors struct {
	mux                  *sync.Mutex
	counters             map[string]*prometheus.CounterVec
	gauges               map[string]*prometheus.GaugeVec
	histograms           map[string]*prometheus.HistogramVec
	histogramBucketSpecs map[string][]float64
	namespace            string
	subsystem            string
}

func (v *metricVectors) getCounter(m Metric) prometheus.Counter {
	vector := v.ensureCountVector(m)
	return vector.With(m.ConstLabels)
}

func (v *metricVectors) ensureCountVector(m Metric) *prometheus.CounterVec {
	key := vectorKey(m)
	if vector, ok := v.counters[key]; ok {
		return vector
	}

	v.mux.Lock()
	defer v.mux.Unlock()

	if vector, ok := v.counters[key]; ok {
		return vector
	}

	opts := prometheus.CounterOpts{
		Namespace:   v.namespace,
		Subsystem:   v.subsystem,
		Name:        m.toPromoMetricName(),
	}
	vector := prometheus.NewCounterVec(opts, labelNames(m))
	_ = prometheus.Register(vector)
	v.counters[key] = vector


	return vector
}

func (v *metricVectors) getGauge(m Metric) prometheus.Gauge {
	vector := v.ensureGaugeVector(m)
	return vector.With(m.ConstLabels)
}

func (v *metricVectors) ensureGaugeVector(m Metric) *prometheus.GaugeVec {
	key := vectorKey(m)
	if vector, ok := v.gauges[key]; ok {
		return vector
	}

	v.mux.Lock()
	defer v.mux.Unlock()

	if vector, ok := v.gauges[key]; ok {
		return vector
	}

	opts := prometheus.GaugeOpts{
		Namespace:   v.namespace,
		Subsystem:   v.subsystem,
		Name:        m.toPromoMetricName(),
	}
	vector := prometheus.NewGaugeVec(opts, labelNames(m))
	_ = prometheus.Register(vector)
	v.gauges[key] = vector
	return vector
}

func (v *metricVectors) getHistogram(m Metric) prometheus.Observer {
	vector := v.ensureHistogramVector(m)
	return vector.With(m.ConstLabels)
}

func (v *metricVectors) ensureHistogramVector(m Metric) *prometheus.HistogramVec {
	key := vectorKey(m)
	if vector, ok := v.histograms[key]; ok {
		return vector
	}

	v.mux.Lock()
	defer v.mux.Unlock()

	if vector, ok := v.histograms[key]; ok {
		return vector
	}

	var buckets []float64
	k := m.toPromoMetricName()
	if b, ok := v.histogramBucketSpecs[k]; ok {
		buckets = b
	}

	opts := prometheus.HistogramOpts{
		Namespace: v.namespace,
		Subsystem: v.subsystem,
		Name:      m.toPromoMetricName(),
		Buckets:   buckets,
	}
	vector := prometheus.NewHistogramVec(opts, labelNames(m))
	_ = prometheus.Register(vector)
	v.histograms[key] = vector
	return vector
}

func vectorKey(m Metric) string {
	key := m.toPromoMetricName()
	ln := labelNames(m)
	sort.Strings(ln)

	key += "("
	started := false
	for k, _ := range m.ConstLabels {
		if started {
			key += ","
		}
		started = true
		key += k
	}
	key += ")"

	return key
}

func labelNames(m Metric) []string {
	var names []string
	if m.ConstLabels == nil || len(m.ConstLabels) == 0 {
		return names
	}
	for k, _ := range m.ConstLabels {
		names = append(names, k)
	}
	return names
}

