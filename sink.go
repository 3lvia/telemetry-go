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

const (
	logTypeAppInsights = "AppInsights"
	logTypeMetrics = "Metrics"
)

type sink interface {
	logEvent(name string, data map[string]string)
	error(err error)
	debug(d string)
	handleCounter(m Metric)
	handleGauge(m Metric)
	handleHistogram(m Metric)
}

func newSink(collector *OptionsCollector) sink {
	hbs := map[string][]float64{}
	if collector.histogramBucketSpecs != nil {
		for k, v := range collector.histogramBucketSpecs {
			hbs[promoMetricName(k)] = v
		}
	}

	m := &metricVectors{
		mux:                  &sync.Mutex{},
		counters:             map[string]*prometheus.CounterVec{},
		gauges:               map[string]*prometheus.GaugeVec{},
		histograms:           map[string]*prometheus.HistogramVec{},
		histogramBucketSpecs: hbs,
	}

	logInfo := map[string]string{}
	if collector.systemName != "" {
		logInfo["system"] = collector.systemName
	}
	if collector.appName != "" {
		logInfo["app"] = collector.appName
	}

	s :=  &standardSink{
		m:                        m,
		sendMetricsToAppInsights: collector.sendMetricsToAppInsights,
		appInsightsSecretPath:    collector.appInsightsSecretPath,
		writer:                   collector.writer,
		capture:                  collector.capture,
		logInfo:                  logInfo,
	}
	if collector.instrumentationKey != "" {
		s.setInstrumentationKey(collector.instrumentationKey)
	} else if !collector.empty && collector.appInsightsSecretPath != "" {
		vault.RegisterDynamicSecretDependency(s, collector.v, nil)
	}

	return s
}

type standardSink struct {
	appInsightsSecretPath    string
	sendMetricsToAppInsights bool
	capture                  EventCapture
	client                   appinsights.TelemetryClient
	logInfo                  map[string]string
	m                        *metricVectors
	writer                   io.Writer
}

func (s *standardSink) logEvent(name string, data map[string]string) {
	event := appinsights.NewEventTelemetry(name)
	d := s.merge(data)
	event.Properties = d
	if s.client != nil {
		s.client.Track(event)
	}
	if s.writer != nil {
		s.writer.Write([]byte(fmt.Sprintf("%s  EVENT(%s) %v\n", time.Now().Format("2006-01-02 15:04:05"), name, d)))
	}
	if s.capture != nil {
		ce := &CapturedEvent{
			SinkType: logTypeAppInsights,
			Type:     "Event",
			Event:    event,
		}
		s.capture.Capture(ce)
	}
}

func (s *standardSink) error(err error) {
	if s.client != nil {
		s.client.TrackException(err)
	}
	if s.writer != nil {
		s.writer.Write([]byte(fmt.Sprintf("%v\n", err)))
	}
	if s.capture != nil {
		ce := &CapturedEvent{
			SinkType: logTypeAppInsights,
			Type:     "Error",
			Event:    err,
		}
		s.capture.Capture(ce)
	}
}

func (s *standardSink) debug(d string) {
	if s.writer != nil {
		out := fmt.Sprintf("%s\n", d)
		s.writer.Write([]byte(out))
	}
}

func (s *standardSink) handleCounter(m Metric) {
	if m.Value < 0 {
		fmt.Printf("counter %s cannot decrease, value: %v\n", m.Name, m.Value)
		return
	}

	c := s.m.getCounter(m)
	c.Add(m.Value)

	if s.sendMetricsToAppInsights {
		s.logMetric(m)
	}

	if s.capture != nil {
		ce := &CapturedEvent{
			SinkType: logTypeMetrics,
			Type:     "Counter",
			Event:    c,
		}
		s.capture.Capture(ce)
	}
}

func (s *standardSink) handleGauge(m Metric) {
	g := s.m.getGauge(m)
	g.Set(m.Value)

	if s.sendMetricsToAppInsights {
		s.logMetric(m)
	}

	if s.capture != nil {
		ce := &CapturedEvent{
			SinkType: logTypeMetrics,
			Type:     "Gauge",
			Event:    g,
		}
		s.capture.Capture(ce)
	}
}

func (s *standardSink) handleHistogram(m Metric) {
	h := s.m.getHistogram(m)
	h.Observe(m.Value)

	if s.capture != nil {
		ce := &CapturedEvent{
			SinkType: logTypeMetrics,
			Type:     "Histogram",
			Event:    h,
		}
		s.capture.Capture(ce)
	}
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

func (s *standardSink) GetSubscriptionSpec() vault.SecretSubscriptionSpec {
	return vault.SecretSubscriptionSpec{
		Paths: []string{s.appInsightsSecretPath},
	}
}

func (s *standardSink) ReceiveAtStartup(secret vault.UpdatedSecret) {
	d := secret.GetAllData()
	instrumentationKey := d["instrumentation-key"]
	s.setInstrumentationKey(instrumentationKey)
}

func (s *standardSink) StartSecretsListener(){}

func (s *standardSink) setInstrumentationKey(instrumentationKey string) {
	telemetryConfig := appinsights.NewTelemetryConfiguration(instrumentationKey)

	// Configure how many items can be sent in one call to the data collector:
	telemetryConfig.MaxBatchSize = 8192

	// Configure the maximum delay before sending queued telemetry:
	telemetryConfig.MaxBatchInterval = 2 * time.Second

	client := appinsights.NewTelemetryClientFromConfig(telemetryConfig)
	s.client = client
}