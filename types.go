package telemetry

import "strings"

// LogChans a set of channels used for communicating events, metrics, errors and
// other telemetry types to the logger.
type LogChans struct {
	// CountChan increases the named Prometheus counter.
	CountChan chan Metric

	// GaugeChan increases the named Prometheus gauge.
	GaugeChan chan Metric

	// ErrorChan sends the error to Application Insights.
	ErrorChan chan error

	// EventChan sends the event to Application Insights.
	EventChan chan Event

	// DebugChan prints a debug message to the console.
	DebugChan chan string
}

// Metric is a named numeric value.
type Metric struct {
	Name  string
	Value float64
}

func (m Metric) toPromoMetricName() string {
	ss := strings.ReplaceAll(m.Name, " ", "_")
	return strings.ToLower(ss)
}

// Event is raised when something interesting happens in the application. Consists
// of a name an a map of key/value pairs.
type Event struct {
	Name string
	Data map[string]string
}

// Options is used to configure the functionality of this entire module.
type Options struct {
	// SystemName the name of the containing system.
	SystemName string

	// The name of the running application/micro-service.
	AppName string

	// AppInsightsSecretPath the path to Application Insights instrumentation key in Vault.
	AppInsightsSecretPath string

	// SendMetricsToAppInsights indicates whether metrics should be sent to Application Insights
	// in addition to Prometheus.
	SendMetricsToAppInsights bool
}