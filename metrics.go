package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"sync"
)

type metrics struct {
	mux             *sync.Mutex
	promoGauges     map[string]prometheus.Gauge
	promoCounters   map[string]prometheus.Counter
	promoHistograms map[string]prometheus.Histogram
	logInfo         map[string]string
}

func (l *metrics) getCounter(m Metric) prometheus.Counter {
	name := m.toPromoMetricName()
	if pc, ok := l.promoCounters[name]; ok {
		return pc
	}

	l.mux.Lock()
	defer l.mux.Unlock()

	if pc, ok := l.promoCounters[name]; ok {
		return pc
	}

	counter := promauto.NewCounter(prometheus.CounterOpts{
		Name:        name,
		Help:        m.Name,
		ConstLabels: l.logInfo,
	})

	l.promoCounters[name] = counter

	return counter
}

func (l *metrics) getGauge(m Metric) prometheus.Gauge {
	name := m.toPromoMetricName()
	if pm, ok := l.promoGauges[name]; ok {
		return pm
	}

	l.mux.Lock()
	defer l.mux.Unlock()

	if pm, ok := l.promoGauges[name]; ok {
		return pm
	}

	gauge := promauto.NewGauge(prometheus.GaugeOpts{
		Name:        name,
		Help:        m.Name,
		ConstLabels: l.logInfo,
	})

	l.promoGauges[name] = gauge
	return gauge
}

func (l *metrics) getHistogram(m Metric) prometheus.Histogram {
	name := m.toPromoMetricName()
	if h, ok := l.promoHistograms[name]; ok {
		return h
	}

	l.mux.Lock()
	defer l.mux.Unlock()

	if h, ok := l.promoHistograms[name]; ok {
		return h
	}

	histogram := promauto.NewHistogram(prometheus.HistogramOpts{
		//Namespace:   "",
		//Subsystem:   "",
		Name:        name,
		Help:        m.Name,
		ConstLabels: l.logInfo,
		//Buckets:     nil,
	})

	l.promoHistograms[name] = histogram
	return histogram
}

