package telemetry

import (
	"github.com/3lvia/hn-config-lib-go/vault"
	"io"
)

// Option specifies options for configuring the logging.
type Option func(*OptionsCollector)

// OptionsCollector collects all options before they are set.
type OptionsCollector struct {
	systemName               string
	appName                  string
	appInsightsSecretPath    string
	sendMetricsToAppInsights bool
	empty                    bool
	v                        vault.SecretsManager
	capture                  EventCapture
	writer                   io.Writer
}

// WithWriter lets clients set a writer which will receive logging events (in addition to the events being written
// to the standard destinations).
func WithWriter(w io.Writer) Option {
	return func(collector *OptionsCollector) {
		collector.writer = w
	}
}

// Named lets clients set the name of the system and application. This value will be included in all logs.
func Named(systemName string, app string) Option {
	return func(collector *OptionsCollector) {
		collector.systemName = systemName
		collector.appName = app
	}
}

// WithAppInsightsSecretPath lets clients set the Vault path to the secret containing the instrumentation key needed
// in order to write logs to application insights.
func WithAppInsightsSecretPath(path string, v vault.SecretsManager) Option {
	return func(collector *OptionsCollector) {
		collector.appInsightsSecretPath = path
		collector.v = v
	}
}

// SendMetricsToAppInsights will send metrics to Application Insights (as well as registering it as a Prometheus
// metric.
func SendMetricsToAppInsights() Option {
	return func(collector *OptionsCollector) {
		collector.sendMetricsToAppInsights = true
	}
}

// Empty if used logs will not be sent to application insights, and also not to Prometheues.
func Empty() Option {
	return func(collector *OptionsCollector) {
		collector.empty = true
	}
}

func WithCapture(c EventCapture) Option {
	return func(collector *OptionsCollector) {
		collector.capture = c
	}
}