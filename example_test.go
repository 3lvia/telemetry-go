package telemetry

import (
	"context"
	"errors"
	"github.com/3lvia/hn-config-lib-go/vault"
	"log"
	"os"
	"testing"
)

func TestExample(t *testing.T) {

	////////////////////// SETUP

	ctx := context.Background()

	vaultSecretsManager, err := vault.New()
	if err != nil {
		log.Fatal()
	}

	var capture EventCapture

	// Start starts a go routine listening to the different logging channels that are returned.
	logChannels := Start(ctx,
		// These names will be added as custom dimensions in all logs to application insights and also
		// Example: EVENT(Start) map[app:cost-monitor handler:h system:monitoring]
		Named("monitoring", "cost-monitor"),

		// Will ensure that a connection to Application Insights is not set up, and that it will not be
		// written to. Overrides both WithAppInsightsSecretPath and WithAppInsightsInstrumentationKey.
		Empty(),

		// If you want to write to Application Insights, and you have its instrumentation key in Hashicorp Vault
		WithAppInsightsSecretPath("monitoring/kv/app/appinsights/monitoring", vaultSecretsManager),

		// If you want to write to Application Insights, and you have the instrumentation key at hand
		WithAppInsightsInstrumentationKey("579a01b9-65c4-4070-b523-a76ade6a49c3"),

		// If you want all logs to be written to an instance of io.Writer, for instance to standard out (as shown
		// here) or to a string buffer for testing purposes.
		WithWriter(os.Stdout),

		// Metrics are normally just incremented internally as Prometheus data. If you want to also send metrics
		// to Application Insights, this can be used.
		SendMetricsToAppInsights(),

		// All logging events are sent to the given capture. This is implemented as a feature that us useful
		// during unit testing when it may be desirable to be able to examine the logging events that application
		// raises.
		WithCapture(capture),

		// Gives the ability to tailor which buckets are used for named Prometheus histograms. NB! Must be invoked
		// before a histogram event of that name is ever raised.
		AddHistogramBucketSpec("my_histogram", []float64{50, 60, 70, 80, 90, 100, 110}),
		AddHistogramBucketSpec("my_other_histogram", []float64{1000, 2000, 3000, 4000, 5000}),
		)

	////////////////////// USAGE

	// Raise an event! Is sent to Application Insights (if configured). Events are meant to be low frequency. Typical
	// usage scenarios include lifecycle events, i.e. when the service was started/stopped etc,
	logChannels.EventChan <- Event{
		Name: "Start",
		Data: map[string]string { "handler": "cost-handler"},
	}

	// Send an error! Is sent to Application Insights (if configured).
	logChannels.ErrorChan <- errors.New("an error has occurred")

	// Send some debug information! This is only sent to io.Writer if it is configured.
	logChannels.DebugChan <- "some debug information"

	// Increment a Prometheus counter!
	logChannels.CountChan <- Metric {
		Name:        "Events handled", // will be transformed to 'events_handled'
		Value:       2,
		ConstLabels: map[string]string{"handler": "cost"}, // will be added as labels to the metric
	}

	// Set a Prometheus gauge!
	logChannels.GaugeChan <- Metric{
		Name:        "Concurrent handlers", // will be transformed to 'concurrent_handlers'
		Value:       13,
		ConstLabels: nil,
	}

	logChannels.HistogramChan <- Metric{
		Name:        "http_handler_latency",
		Value:       182.12,
		ConstLabels: map[string]string{"code": "200"},
	}
}
