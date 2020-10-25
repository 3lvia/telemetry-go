# telemetry
  
This module exposes channels that are meant be used for logging to Azure Application Insights and Prometheus. The information that is conveyed through the channel 
is either to **Azure Application Insights** or used to set/increase an internal **Prometheus** metric. 

This package is currently dependent on the modules "github.com/3lvia/hn-config-lib-go/vault" which again means that
you need a running instance of Hashicorp Vault in order to use this functionality. This dependency will probably 
become optional in a future version. The Vault instance is used for the sole purpose of fetching the instrumentation
key for Application Insights.

The functionality is made available through the func *Start* which internally starts a go-routine listening to the
different channels that are contained in the returned instance of LogChans. The following code sample shows how to
bootstrap the functionality:

    import (
        "context"
        "github.com/3lvia/telemetry-go"
        "github.com/3lvia/hn-config-lib-go/vault"
        "github.com/prometheus/client_golang/prometheus/promhttp"
    )
    
    ctx := context.Background()
    v, err := vault.New()
    if err != nil {
        log.Fatal(err)
    }
    systemName := "my-system"
    appName := "my-app"
    applicationInsightsVaultPath := "mount/kv/path/to/appinsights-instrumentationkey"
    logChannels := telemetry.Start(ctx, systemName, appName, applicationInsightsVaultPath, v)
    // logChannels are ready to use!
    
    // Start Prometheus metrics API
    metricsPort := "2112"
    http.Handle("/metrics", promhttp.Handler())
    http.ListenAndServe(metricsPort, nil)
    
  
### Log Channels
* **CountChan** Increases the named Prometheus counter.
* **GaugeChan** Sets the named Prometheus counter.
* **ErrorChan** Sends the error to Application Insights. It is registered as an *Exception* in App Insights.
* **EventChan** Sends the event to Application Insights. It is registered as a *Cusom Event* in App Insights.
* **DebugChan** Prints the debug string to the console.

### About Prometheus Names
The metric instances that are used in the two channels *CountChan* and *GaugeChan* contain the element *Name*. It is 
assumed that this name contains a human readable sentence that describes the metric, for instance *Number of
succesful API requests*. 

This sentence is transformed internally as follows:
*Number of succesful API requests* -> *number_of_succesful_api_requests*

Simultaneously, the given sentence-based name is used as the "Help" of the metric.

### Empty Logger
Sometimes it's useful to be able to spin up log channels that don't do anything, but that also don't block as logging
events are sent on the logging channels. Unit testing is on such scenario.

This package implements the concept of empty logging to handle this scenario, the code sample below shows how it's used:

 ```
 import (
     "github.com/3lvia/telemetry-go"
 )
 
 logChannels := telemetry.StartEmpty()
 // logChannels are ready to use!
 ```

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
* **http_responses_total** (count) The total number of registered http requests
* **http_responses_<handlerName>** (count) The total number of registered http requests for the given handler
* **http_latency_total** (histogram) Measures the latency across all API requests
* **http_latency_<handlerName>** (histogram) Measures the latency across all API requests for the named handler