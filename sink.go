package telemetry

import (
	"fmt"
	"github.com/3lvia/hn-config-lib-go/vault"
	"github.com/microsoft/ApplicationInsights-Go/appinsights"
	"github.com/prometheus/client_golang/prometheus"
	"io"
	"sync"
	"time"
)

type sink interface {
	logEvent(name string, data map[string]string)
	error(err error)
	debug(d string)
	handleCounter(m Metric)
	handleGauge(m Metric)
	handleHistogram(m Metric)
}

func newSink(collector *OptionsCollector, v vault.SecretsManager) sink {
	logInfo := map[string]string{
		"system": collector.systemName,
		"app":    collector.appName,
	}

	m := &metrics{
		mux:             &sync.Mutex{},
		promoGauges:     map[string]prometheus.Gauge{},
		promoCounters:   map[string]prometheus.Counter{},
		promoHistograms: map[string]prometheus.Histogram{},
		logInfo:         logInfo,
	}

	s :=  &standardSink{
		m:                        m,
		sendMetricsToAppInsights: collector.sendMetricsToAppInsights,
		appInsightsSecretPath:    collector.appInsightsSecretPath,
		logInfo:                  logInfo,
		writer:                   collector.writer,
	}
	if collector.sendMetricsToAppInsights && collector.appInsightsSecretPath != "" {
		vault.RegisterDynamicSecretDependency(s, v, nil)
	}

	return s
}

type standardSink struct {
	appInsightsSecretPath    string
	sendMetricsToAppInsights bool
	client                   appinsights.TelemetryClient
	logInfo                  map[string]string
	m                        *metrics
	writer                   io.Writer
}

func (s *standardSink) GetSubscriptionSpec() vault.SecretSubscriptionSpec {
	return vault.SecretSubscriptionSpec{
		Paths: []string{s.appInsightsSecretPath},
	}
}

func (s *standardSink) logEvent(name string, data map[string]string) {
	event := appinsights.NewEventTelemetry(name)
	d := s.merge(data)
	event.Properties = d
	if s.client != nil {
		s.client.Track(event)
	}
	if s.writer != nil {
		s.writer.Write([]byte(fmt.Sprintf("EVENT(%s) %v\n", name, d)))
	}
}

func (s *standardSink) error(err error) {
	if s.client != nil {
		s.client.TrackException(err)
	}
	if s.writer != nil {
		s.writer.Write([]byte(fmt.Sprintf("%v\n", err)))
	}
}

func (s *standardSink) debug(d string) {
	if s.writer != nil {
		s.writer.Write([]byte(d))
	}
}

func (s *standardSink) handleCounter(m Metric) {
	c := s.m.getCounter(m)
	c.Add(m.Value)

	if s.sendMetricsToAppInsights {
		s.logMetric(m)
	}
}

func (s *standardSink) handleGauge(m Metric) {
	g := s.m.getGauge(m)
	g.Set(m.Value)

	if s.sendMetricsToAppInsights {
		s.logMetric(m)
	}
}

func (s *standardSink) handleHistogram(m Metric) {
	h := s.m.getHistogram(m)
	h.Observe(m.Value)
}

func (s *standardSink) logMetric(m Metric) {
	name := m.toPromoMetricName()
	aiMetric := appinsights.NewMetricTelemetry(name, m.Value)
	if s.client != nil {
		s.client.Track(aiMetric)
	}
}

func (s *standardSink) merge(data map[string]string) map[string]string {
	if data == nil || len(data) == 0 {
		return s.logInfo
	}
	for k, v := range s.logInfo {
		data[k] = v
	}
	return data
}

func (s *standardSink) ReceiveAtStartup(secret vault.UpdatedSecret) {
	d := secret.GetAllData()
	instrumentationKey := d["instrumentation-key"]
	telemetryConfig := appinsights.NewTelemetryConfiguration(instrumentationKey)

	// Configure how many items can be sent in one call to the data collector:
	telemetryConfig.MaxBatchSize = 8192

	// Configure the maximum delay before sending queued telemetry:
	telemetryConfig.MaxBatchInterval = 2 * time.Second

	client := appinsights.NewTelemetryClientFromConfig(telemetryConfig)
	s.client = client
}

func (s *standardSink) StartSecretsListener(){}