package telemetry

import (
	"context"
	"fmt"
	"github.com/3lvia/hn-config-lib-go/vault"
	"github.com/microsoft/ApplicationInsights-Go/appinsights"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"sync"
	"time"
)

// Start starts the logger in a go routine and returns a set of channels
// that can be used to send telemetry to the logger.
func Start(ctx context.Context, v vault.SecretsManager, opts ...Option) LogChans {
	collector := &OptionsCollector{
		sendMetricsToAppInsights: true,
	}
	for _, opt := range opts {
		opt(collector)
	}

	l := &logger{
		logInfo: map[string]string{
			"system": collector.systemName,
			"app":    collector.appName,
		},
		appInsightsSecretPath:    collector.appInsightsSecretPath,
		mux:                      &sync.Mutex{},
		promoGauges:              map[string]prometheus.Gauge{},
		promoCounters:            map[string]prometheus.Counter{},
		promoHistograms:          map[string]prometheus.Histogram{},
		sendMetricsToAppInsights: collector.sendMetricsToAppInsights,
	}
	vault.RegisterDynamicSecretDependency(l, v, nil)
	lg := l.getLogChannels()
	go l.start(ctx)
	return lg
}

type logger struct {
	logInfo                  map[string]string
	appInsightsSecretPath    string
	client                   appinsights.TelemetryClient
	gaugeChan                <-chan Metric
	counterChan              <-chan Metric
	histogramChan            <-chan Metric
	errorChan                <-chan error
	eventChan                <-chan Event
	debugChan                <-chan string
	mux                      *sync.Mutex
	promoGauges              map[string]prometheus.Gauge
	promoCounters            map[string]prometheus.Counter
	promoHistograms          map[string]prometheus.Histogram
	sendMetricsToAppInsights bool
}

func (l *logger) start(ctx context.Context) {
	for {
		select {
		case c := <-l.counterChan:
			l.handleCounter(c)
		case g := <-l.gaugeChan:
			l.handleGauge(g)
		case h := <-l.histogramChan:
			l.handleHistogram(h)
		case err := <-l.errorChan:
			l.error(err)
		case e := <-l.eventChan:
			l.logEvent(e.Name, e.Data)
		case d := <-l.debugChan:
			fmt.Printf("%s\n", d)
		}
	}
}

func (l *logger) handleCounter(m Metric) {
	c := l.getCounter(m)
	c.Add(m.Value)

	if l.sendMetricsToAppInsights {
		l.logMetric(m)
	}
}

func (l *logger) handleGauge(m Metric) {
	g := l.getGauge(m)
	g.Set(m.Value)

	if l.sendMetricsToAppInsights {
		l.logMetric(m)
	}
}

func (l *logger) handleHistogram(m Metric) {
	h := l.getHistogram(m)
	h.Observe(m.Value)
}

func (l *logger) getCounter(m Metric) prometheus.Counter {
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

func (l *logger) getGauge(m Metric) prometheus.Gauge {
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

func (l *logger) getHistogram(m Metric) prometheus.Histogram {
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

func (l *logger) flush() {
	l.client.Channel().Flush()
}

func (l *logger) logMetric(m Metric) {
	name := m.toPromoMetricName()
	aiMetric := appinsights.NewMetricTelemetry(name, m.Value)
	l.client.Track(aiMetric)
}

func (l *logger) logEvent(name string, data map[string]string)  {
	event := appinsights.NewEventTelemetry(name)
	event.Properties = data
	l.client.Track(event)
}

func (l *logger) error(err error) {
	fmt.Printf("%v\n", err)
	l.client.TrackException(err)
}

func (l *logger) GetSubscriptionSpec() vault.SecretSubscriptionSpec {
	return vault.SecretSubscriptionSpec{
		Paths: []string{l.appInsightsSecretPath},
	}
}

func (l *logger) ReceiveAtStartup(secret vault.UpdatedSecret) {
	d := secret.GetAllData()
	instrumentationKey := d["instrumentation-key"]
	telemetryConfig := appinsights.NewTelemetryConfiguration(instrumentationKey)

	// Configure how many items can be sent in one call to the data collector:
	telemetryConfig.MaxBatchSize = 8192

	// Configure the maximum delay before sending queued telemetry:
	telemetryConfig.MaxBatchInterval = 2 * time.Second

	client := appinsights.NewTelemetryClientFromConfig(telemetryConfig)
	l.client = client
}

func (l *logger) StartSecretsListener(){}

func (l *logger) getLogChannels() LogChans {
	gaugeChan := make(chan Metric)
	histogramChan := make(chan Metric)
	errorChan := make(chan error)
	eventChan := make(chan Event)
	debugChan := make(chan string)
	counterChan := make(chan Metric)
	l.gaugeChan = gaugeChan
	l.errorChan = errorChan
	l.eventChan = eventChan
	l.debugChan = debugChan
	l.counterChan = counterChan
	l.histogramChan = histogramChan
	return LogChans{
		GaugeChan:     gaugeChan,
		ErrorChan:     errorChan,
		EventChan:     eventChan,
		DebugChan:     debugChan,
		CountChan:     counterChan,
		HistogramChan: histogramChan,
	}
}
