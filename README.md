# telemetry
  
This module exposes channels that are meant be used for logging to Azure Application Insights and Prometheus. The information that is conveyed through the channel 
is either to **Azure Application Insights** or used to set/increase an internal **Prometheus** metric. 

Additionally, functionality for wrapping http handlers so that http request telemetry is automatically maintained is exposed. 

This package is currently dependent on the modules "github.com/3lvia/hn-config-lib-go/vault" which again means that
you need a running instance of Hashicorp Vault in order to use this functionality. This dependency will probably 
become optional in a future version. The Vault instance is used for the sole purpose of fetching the instrumentation
key for Application Insights.

The functionality is made available through the func *Start* which internally starts a go-routine listening to the
different channels that are contained in the returned instance of LogChannels. Behavior is configured by use og the options pattern in the Start method.

The following code sample shows how to bootstrap the functionality, and it also lists all possible options in the code:

    import (
        "context"
        "github.com/3lvia/telemetry-go"
        "github.com/3lvia/hn-config-lib-go/vault"
        "github.com/prometheus/client_golang/prometheus/promhttp"
    )
    
    import (
    	"context"
    	"errors"
    	"github.com/3lvia/hn-config-lib-go/vault"
    	"github.com/3lvia/telemetry-go"
    	"log"
    	"os"

    )
    ctx := context.Background()
    
    vaultSecretsManager, err := vault.New()
    if err != nil {
    	log.Fatal()
    }
    
    var capture telemetry.EventCapture
    
    // Start starts a go routine listening to the different logging channels that are returned.
    logChannels := telemetry.Start(ctx,
    	// These names will be added as custom dimensions in all logs to application insights and also
    	// Example: EVENT(Start) map[app:cost-monitor handler:h system:monitoring]
    	telemetry.Named("monitoring", "cost-monitor"),
    
    	// Will ensure that a connection to Application Insights is not set up, and that it will not be
    	// written to. Overrides both WithAppInsightsSecretPath and WithAppInsightsInstrumentationKey.
    	telemetry.Empty(),
    
    	// If you want to write to Application Insights, and you have its instrumentation key in Hashicorp Vault
    	telemetry.WithAppInsightsSecretPath("monitoring/kv/app/appinsights/monitoring", vaultSecretsManager),
    
    	// If you want to write to Application Insights, and you have the instrumentation key at hand
    	telemetry.WithAppInsightsInstrumentationKey("579a01b9-65c4-4070-b523-a76ade6a49c3"),
    
    	// If you want all logs to be written to an instance of io.Writer, for instance to standard out (as shown
    	// here) or to a string buffer for testing purposes.
    	telemetry.WithWriter(os.Stdout),
    
    	// Metrics are normally just incremented internally as Prometheus data. If you want to also send metrics
    	// to Application Insights, this can be used.
    	telemetry.SendMetricsToAppInsights(),
    
    	// All logging events are sent to the given capture. This is implemented as a feature that us useful
    	// during unit testing when it may be desirable to be able to examine the logging events that application
    	// raises.
    	telemetry.WithCapture(capture),
    
    	// Gives the ability to tailor which buckets are used for named Prometheus histograms. NB! Must be invoked
    	// before a histogram event of that name is ever raised.
    	telemetry.AddHistogramBucketSpec("my_histogram", []float64{50, 60, 70, 80, 90, 100, 110}),
    	telemetry.AddHistogramBucketSpec("my_other_histogram", []float64{1000, 2000, 3000, 4000, 5000}),
    )
    
    ////////////////////// USAGE
    
    // Raise an event! Is sent to Application Insights (if configured). Events are meant to be low frequency. Typical
    // usage scenarios include lifecycle events, i.e. when the service was started/stopped etc,
    logChannels.EventChan <- telemetry.Event{
    	Name: "Start",
    	Data: map[string]string { "handler": "cost-handler" },
    }
    
    // Send an error! Is sent to Application Insights (if configured).
    logChannels.ErrorChan <- errors.New("an error has occurred")
    
    // Send some debug information! This is only sent to io.Writer if it is configured.
    logChannels.DebugChan <- "some debug information"
    
    // Increment a Prometheus counter!
    logChannels.CountChan <- telemetry.Metric {
    	Name:        "Events handled", // will be transformed to 'events_handled'
    	Value:       2,
    	ConstLabels: map[string]string{"handler": "cost"}, // will be added as labels to the metric
    }
    
    // Set a Prometheus gauge!
    logChannels.GaugeChan <- telemetry.Metric{
    	Name:        "Concurrent handlers", // will be transformed to 'concurrent_handlers'
    	Value:       13,
    	ConstLabels: nil,
    }
    
    logChannels.HistogramChan <- telemetry.Metric{
    	Name:        "http_handler_latency",
    	Value:       182.12,
    	ConstLabels: map[string]string{"code": "200"},
    }
    
    
  
### Log Channels
* **CountChan** Increases the named Prometheus counter.
* **GaugeChan** Sets the named Prometheus counter.
* **HistogramChan** Observes the value of the given histogram.
* **ErrorChan** Sends the error to Application Insights. It is registered as an *Exception* in App Insights.
* **EventChan** Sends the event to Application Insights. It is registered as a *Cusom Event* in App Insights.
* **DebugChan** Prints the debug-string to the console.

### About Prometheus Names
The metric instances that are used in the two channels *CountChan* and *GaugeChan* contain the element *Name*. It is 
assumed that this name contains a human readable sentence that describes the metric, for instance *Number of
succesful API requests*. 

This sentence is transformed internally as follows:
*Number of succesful API requests* -> *number_of_succesful_api_requests*

Simultaneously, the given sentence-based name is used as the "Help" of the metric.


### HTTP wrapper
This package implements HTTP wrapper functionality. The purpose of this is to provide automatic logging of metrics
for the number of handled requests (in total and failed) and for latency.

An interface **telemetry.RequestHandler** is exposed, and clients must implement this interface and then wrap it through
the function **telemetry.Wrap** 

```
 import (
     "net/http"
     "github.com/3lvia/telemetry-go"
 )

myHandler := handlerImplementingRequestHandler{}
httpHandler := telemetry.Wrap(myHandler)
http.Handle("/mypath", httpHandler)
```

It is the responsibility of the client application handler to return an instance of **telemetry.Rountrip** with the correct
values. The wrapper will automatically increse/set the correct metrics and also handle the http response correctly.

**telemetry.Roundtrip** contains the following fields:
* **HandlerName** a constant for the given handler. This value will be used in the metrics name, see below.
* **HTTPResponseCode** the http response code to be set on the response.
* **Contents** a byte slice containing the data to be added as contents on the http response.

The following metrics will be maintained automatically:
* **http_<handlerName>_requests_total** (count) The total number of registered http requests. The http response code is added as a tag.
* **http_<handlerName>_latency** (histogram) Measures the latency across all API requests. The http response code is added as a tag.