package telemetry

// LogChans a set of channels used for communicating events, metrics, errors and
// other telemetry types to the logger.
type LogChans struct {
	MetricChan   chan Metric
	GaugeChan    chan Metric
	ErrorChan    chan error
	EventChan    chan Event
	DebugChan    chan string
	//FlushSigChan chan struct{}
}

// Metric is a named numeric value.
type Metric struct {
	Name  string
	Value float64
}

// Event is raised when something interesting happens in the application. Consists
// of a name an a map of key/value pairs.
type Event struct {
	Name string
	Data map[string]string
}