# telemetry
  
This module exposes channels that are meant be used for logging to Azure Application Insights and Prometheus. The information that is conveyed through the channel 
is either to **Azure Application Insights** or used to set/increase an internal **Prometheus** metric. 

This package is currently dependent on the modules "github.com/3lvia/hn-config-lib-go/vault" which again means that
you need a running instance of Hashicorp Vault in order to use this functionality. This dependency will probably 
become optional in a future version. The Vault instance is used for the sole purpose of fetching the instrumentation
key for Application Insights.

The functionality is made available through the func *Start* which internally starts a go-routine listening to the
different channels that are contained in the returned instance of LogChans. The follwoing code sample shows how to
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
    
    // 
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

This sentence is internally transformed as follows:
*Number of succesful API requests* -> *number_of_succesful_api_requests*

Simultaneously, the given sentence-based name is used as the "Help" of the metric.