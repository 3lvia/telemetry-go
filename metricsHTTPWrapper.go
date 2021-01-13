package telemetry

import (
	"fmt"
	"net/http"
	"time"
)

// RoundTrip contains the results og a single http request. It is the responsibility of the RequestHandler to return
// an instance of RoundTrip with appropriate values.
type RoundTrip struct {
	HandlerName      string
	HTTPResponseCode int
	Contents         []byte
}

// RequestHandler the instance that is responsible for the business logic to be performed as a result of the incoming
// http request. Clients wishing t
type RequestHandler interface {
	Handle(r *http.Request) RoundTrip
}

// Wrap the handler so that it can be presented as http.Handler. The wrapper will automatically set the correct
// Prometheus metrics for each handled request.
func Wrap(h RequestHandler, logChannels LogChannels) http.Handler {
	return &wrapper{
		handler:     h,
		logChannels: logChannels,
	}
}

type wrapper struct {
	handler     RequestHandler
	logChannels LogChannels
}

func (w *wrapper) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	start := time.Now()

	r := w.handler.Handle(req)

	elapsed := time.Since(start)
	latency := float64(elapsed.Milliseconds())

	w.registerMetrics(r, latency)

	rw.WriteHeader(r.HTTPResponseCode)
	rw.Write(r.Contents)
}

func (w *wrapper) registerMetrics(r RoundTrip, latency float64) {
	w.logChannels.CountChan <- Metric{
		Name:        fmt.Sprintf("http_%s_requests_total", r.HandlerName),
		Value:       1,
		ConstLabels: map[string]string{
			"code": fmt.Sprintf("%d", r.HTTPResponseCode),
		},
	}
	w.logChannels.HistogramChan <- Metric{
		Name:        fmt.Sprintf("http_%s_latency", r.HandlerName),
		Value:       latency,
		ConstLabels: map[string]string{
			"code": fmt.Sprintf("%d", r.HTTPResponseCode),
		},
	}
	//if r.HTTPResponseCode >= 500 {
	//	w.logChannels.CountChan <- Metric{
	//		Name:  "http_responses_500_total",
	//		Value: 1,
	//	}
	//	w.logChannels.CountChan <- Metric{
	//		Name:  fmt.Sprintf("http_responses_500_%s", r.HandlerName),
	//		Value: 1,
	//	}
	//}
	//w.logChannels.CountChan <- Metric{
	//	Name:  "http_responses_total",
	//	Value: 1,
	//}
	//w.logChannels.CountChan <- Metric{
	//	Name:  fmt.Sprintf("http_responses_%s", r.HandlerName),
	//	Value: 1,
	//}
	//
	//w.logChannels.CountChan <- Metric{
	//	Name:  "http_latency_total",
	//	Value: latency,
	//}
	//w.logChannels.CountChan <- Metric{
	//	Name:  fmt.Sprintf("http_latency_%s", r.HandlerName),
	//	Value: latency,
	//}
}