package telemetry

import (
	"bytes"
	"errors"
	"testing"
	"time"
)

func TestStartEmpty(t *testing.T) {
	// Arrange
	expectedOutput := `METRIC(count): {test 3.14}
METRIC(gauge): {my-gauge 1}
EVENT(my-event): map[a:b c:d]
DEBUG: debug information
METRIC(histogram): {my-hist 12.34}
ERROR: unexpected error
`
	buf := new(bytes.Buffer)

	err := errors.New("unexpected error")

	// Act
	logChannels := StartEmpty(WithWriter(buf))

	go func(lc LogChans) {
		lc.CountChan <- Metric{
			Name:  "test",
			Value: 3.14,
		}
		lc.GaugeChan <- Metric{
			Name:  "my-gauge",
			Value: 1,
		}
		lc.EventChan <- Event{
			Name: "my-event",
			Data: map[string]string{
				"a": "b",
				"c": "d",
			},
		}
		lc.DebugChan <- "debug information"
		lc.HistogramChan <- Metric{
			Name:  "my-hist",
			Value: 12.34,
		}
		lc.ErrorChan <- err
	}(logChannels)

	<- time.After(10 * time.Millisecond)

	// Assert
	s := buf.String()

	if s != expectedOutput {
		t.Errorf("unexpected output, got %s", s)
	}
}

func TestStartEmpty_noOptions(t *testing.T) {
	// Arrange

	// Act
	logChannels := StartEmpty()

	go func(lc LogChans) {
		lc.CountChan <- Metric{
			Name:  "test",
			Value: 3.14,
		}
		lc.GaugeChan <- Metric{
			Name:  "my-gauge",
			Value: 1,
		}
	}(logChannels)

	<- time.After(1 * time.Millisecond)

}
