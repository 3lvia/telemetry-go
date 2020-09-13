package telemetry

// StartEmpty starts a logger that doesn't log anything, but that will not
// block when log events are sent on the logging channels, the purpose being
// to provide support for testing scenarios when logging is not in focus.
func StartEmpty() LogChans {
	logChannels := LogChans{
		CountChan: make(chan Metric),
		GaugeChan: make(chan Metric),
		ErrorChan: make(chan error),
		EventChan: make(chan Event),
		DebugChan: make(chan string),
	}

	logger := &emptyLogger{logChannels: logChannels}
	go logger.start()

	return logChannels
}

type emptyLogger struct {
	logChannels LogChans
}

func (l *emptyLogger) start() {
	for {
		select {
		case <-l.logChannels.CountChan:
		case <-l.logChannels.GaugeChan:
		case <-l.logChannels.ErrorChan:
		case <-l.logChannels.EventChan:
		case <-l.logChannels.DebugChan:
		}
	}
}