package telemetry

import (
	"context"
	"fmt"
	"github.com/3lvia/hn-config-lib-go/vault"
	"github.com/microsoft/ApplicationInsights-Go/appinsights"
	"time"
)

// Start starts the logger in a go routine and returns a set of channels
// that can be used to send telemetry to the logger.
func Start(ctx context.Context, appInsightsSecretPath string, v vault.SecretsManager) LogChans {
	l := &logger{appInsightsSecretPath: appInsightsSecretPath}
	vault.RegisterDynamicSecretDependency(l, v, nil)
	lg := l.getLogChannels()
	go l.start(ctx)
	return lg
}

type logger struct {
	appInsightsSecretPath string
	client                appinsights.TelemetryClient
	metricChan            <-chan Metric
	gaugeChan             <-chan Metric
	errorChan             <-chan error
	eventChan             <-chan Event
	debugChan             <-chan string
}

func (l *logger) start(ctx context.Context) {
	for {
		select {
		case m := <-l.metricChan:
			fmt.Printf("%v", m)
		case g := <-l.gaugeChan:
			fmt.Printf("%v", g)
		case err := <-l.errorChan:
			l.error(err)
		case e := <-l.eventChan:
			l.logEvent(e.Name, e.Data)
			fmt.Printf("%s %v \n", e.Name, e.Data)
		case d := <-l.debugChan:
			_ = d
			//fmt.Printf("%s\n", d)
		}
	}
}

func (l *logger) flush() {
	l.client.Channel().Flush()
}

func (l *logger) logEvent(name string, data map[string]string)  {
	event := appinsights.NewEventTelemetry(name)
	event.Properties = data
	l.client.Track(event)
}

func (l *logger) error(err error) {
	fmt.Printf("%v\n", err)
	l.client.TrackException(err)
}

func (l *logger) GetSubscriptionSpec() vault.SecretSubscriptionSpec {
	return vault.SecretSubscriptionSpec{
		Paths: []string{l.appInsightsSecretPath},
	}
}

func (l *logger) ReceiveAtStartup(secret vault.UpdatedSecret) {
	d := secret.GetAllData()
	instrumentationKey := d["instrumentation-key"]
	telemetryConfig := appinsights.NewTelemetryConfiguration(instrumentationKey)

	// Configure how many items can be sent in one call to the data collector:
	telemetryConfig.MaxBatchSize = 8192

	// Configure the maximum delay before sending queued telemetry:
	telemetryConfig.MaxBatchInterval = 2 * time.Second

	client := appinsights.NewTelemetryClientFromConfig(telemetryConfig)
	l.client = client
}

func (l *logger) StartSecretsListener(){}

func (l *logger) getLogChannels() LogChans {
	metricChan := make(chan Metric)
	gaugeChan := make(chan Metric)
	errorChan := make(chan error)
	eventChan := make(chan Event)
	debugChan := make(chan string)
	l.metricChan = metricChan
	l.gaugeChan = gaugeChan
	l.errorChan = errorChan
	l.eventChan = eventChan
	l.debugChan = debugChan

	return LogChans{
		MetricChan:   metricChan,
		GaugeChan:    gaugeChan,
		ErrorChan:    errorChan,
		EventChan:    eventChan,
		DebugChan:    debugChan,
	}
}