package telemetry

import "io"

// Option specifies options for configuring the logging.
type Option func(*OptionsCollector)

// OptionsCollector collects all options before they are set.
type OptionsCollector struct {
	systemName               string
	appName                  string
	appInsightsSecretPath    string
	sendMetricsToAppInsights bool
	empty                    bool
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

// WithSystemName lets clients set the name of the system. This value will be included in all logs.
func WithSystemName(systemName string) Option {
	return func(collector *OptionsCollector) {
		collector.systemName = systemName
	}
}

// WithAppName lets clients set the name of the application. This value will be included in all logs.
func WithAppName(app string) Option {
	return func(collector *OptionsCollector) {
		collector.appName = app
	}
}

// WithAppInsightsSecretPath lets clients set the Vault path to the secret containing the instrumentation key needed
// in order to write logs to application insights.
func WithAppInsightsSecretPath(path string) Option {
	return func(collector *OptionsCollector) {
		collector.appInsightsSecretPath = path
	}
}

// MuteApplicationInsights if used, logs won't be sent to application insights.
func MuteApplicationInsights() Option {
	return func(collector *OptionsCollector) {
		collector.sendMetricsToAppInsights = false
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