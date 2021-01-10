package telemetry

import (
	"context"
	"github.com/3lvia/hn-config-lib-go/vault"
)

// Start starts the logger in a go routine and returns a set of channels
// that can be used to send telemetry to the logger.
func Start(ctx context.Context, v vault.SecretsManager, opts ...Option) LogChannels {
	collector := &OptionsCollector{
		sendMetricsToAppInsights: true,
		empty:                    false,
	}
	for _, opt := range opts {
		opt(collector)
	}

	s := newSink(collector, v)

	l := &logger{
		sink:                     s,
		sendMetricsToAppInsights: collector.sendMetricsToAppInsights,
	}

	lg := l.getLogChannels()
	go l.start(ctx)
	return lg
}

type logger struct {
	sink                     sink
	gaugeChan                <-chan Metric
	counterChan              <-chan Metric
	histogramChan            <-chan Metric
	errorChan                <-chan error
	eventChan                <-chan Event
	debugChan                <-chan string

	sendMetricsToAppInsights bool
}

func (l *logger) start(ctx context.Context) {
	for {
		select {
		case c := <-l.counterChan:
			l.sink.handleCounter(c)
		case g := <-l.gaugeChan:
			l.sink.handleGauge(g)
		case h := <-l.histogramChan:
			l.sink.handleHistogram(h)
		case err := <-l.errorChan:
			l.sink.error(err)
		case e := <-l.eventChan:
			l.sink.logEvent(e.Name, e.Data)
		case d := <-l.debugChan:
			l.sink.debug(d)
		}
	}
}

func (l *logger) getLogChannels() LogChannels {
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
	return LogChannels{
		GaugeChan:     gaugeChan,
		ErrorChan:     errorChan,
		EventChan:     eventChan,
		DebugChan:     debugChan,
		CountChan:     counterChan,
		HistogramChan: histogramChan,
	}
}
