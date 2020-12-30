package telemetry

import (
	"fmt"
	"io"
)

// StartEmpty starts a logger that doesn't log anything, but that will not
// block when log events are sent on the logging channels, the purpose being
// to provide support for testing scenarios when logging is not in focus.
func StartEmpty(opts ...Option) LogChans {
	logChannels := LogChans{
		CountChan:     make(chan Metric),
		GaugeChan:     make(chan Metric),
		ErrorChan:     make(chan error),
		EventChan:     make(chan Event),
		DebugChan:     make(chan string),
		HistogramChan: make(chan Metric),
	}

	logger := &emptyLogger{logChannels: logChannels}

	if opts != nil {
		collector := &OptionsCollector{}
		for _, opt := range opts {
			opt(collector)
		}
		logger.withCollector(collector)
	}

	go logger.start()

	return logChannels
}

type emptyLogger struct {
	writer io.Writer
	logChannels LogChans
}

func (l *emptyLogger) start() {
	for {
		select {
		case m := <-l.logChannels.CountChan:
			l.metric(m, "count")
		case m := <-l.logChannels.GaugeChan:
			l.metric(m, "gauge")
		case e := <-l.logChannels.ErrorChan:
			l.error(e)
		case e := <-l.logChannels.EventChan:
			l.event(e)
		case s := <-l.logChannels.DebugChan:
			l.debug(s)
		case m := <-l.logChannels.HistogramChan:
			l.metric(m, "histogram")
		}
	}
}

func (l *emptyLogger) error(err error) {
	if l.writer == nil {
		return
	}

	l.writer.Write([]byte(fmt.Sprintf("ERROR: %v\n", err)))
}

func (l *emptyLogger) metric(m Metric, t string) {
	if l.writer == nil {
		return
	}

	l.writer.Write([]byte(fmt.Sprintf("METRIC(%s): %v\n", t, m)))
}

func (l *emptyLogger) event(e Event) {
	if l.writer == nil {
		return
	}

	l.writer.Write([]byte(fmt.Sprintf("EVENT(%s): %v\n", e.Name, e.Data)))
}

func (l *emptyLogger) debug(s string) {
	if l.writer == nil {
		return
	}

	l.writer.Write([]byte(fmt.Sprintf("DEBUG: %s\n", s)))
}

func (l *emptyLogger) withCollector(collector *OptionsCollector) {
	l.writer = collector.writer
}