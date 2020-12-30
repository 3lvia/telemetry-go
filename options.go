package telemetry

import "io"

// Option specifies options for configuring the logging.
type Option func(*OptionsCollector)

// OptionsCollector collects all options before they are set.
type OptionsCollector struct {
	writer io.Writer
}

// WithWriter lets clients set a writer which will receive logging events (in addition to the events being written
// to the standard destinations).
func WithWriter(w io.Writer) Option {
	return func(collector *OptionsCollector) {
		collector.writer = w
	}
}