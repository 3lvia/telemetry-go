package telemetry

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func Test_wrapper_ServeHTTP(t *testing.T) {
	// Arrange
	handlerName := "costs"
	expectedMetrics := `# HELP http_costs_requests_total 
# TYPE http_costs_requests_total counter
http_costs_requests_total{code="200"} 2
http_costs_requests_total{code="500"} 1`

	expectedHistogram1 := `http_costs_latency_bucket{code="200",le="10"} 1
http_costs_latency_bucket{code="200",le="100"} 1
http_costs_latency_bucket{code="200",le="200"} 1
http_costs_latency_bucket{code="200",le="1000"} 2
http_costs_latency_bucket{code="200",le="5000"} 2
http_costs_latency_bucket{code="200",le="10000"} 2
http_costs_latency_bucket{code="200",le="+Inf"} 2`

	expectedHistogram2 := `http_costs_latency_count{code="200"} 2`

//	expectedHistogram3 := `http_costs_latency_bucket{code="500",le="10"} 0
//http_costs_latency_bucket{code="500",le="100"} 0
//http_costs_latency_bucket{code="500",le="200"} 0
//http_costs_latency_bucket{code="500",le="1000"} 0
//http_costs_latency_bucket{code="500",le="5000"} 1
//http_costs_latency_bucket{code="500",le="10000"} 1
//http_costs_latency_bucket{code="500",le="+Inf"} 1`

	expectedHistogram4 := `http_costs_latency_count{code="500"} 1`

	expected := []string {expectedMetrics, expectedHistogram1, expectedHistogram2, expectedHistogram4}

	ctx := context.Background()
	logChannels := Start(ctx,
		//Named("monitoring", "cost-monitor"),
		Empty(),
		AddHistogramBucketSpec(fmt.Sprintf("http_%s_latency", handlerName), []float64{10, 100, 200, 1000, 5000, 10000}))

	impl := &mockHandler{
		index: 0,
		elements: []roundTripTestElement{
			{
				r:     RoundTrip{
					HandlerName:      handlerName,
					HTTPResponseCode: 200,
					Contents:         nil,
				},
				delay: 1 * time.Millisecond,
			},
			{
				r:     RoundTrip{
					HandlerName:      handlerName,
					HTTPResponseCode: 200,
					Contents:         nil,
				},
				delay: 201 * time.Millisecond,
			},
			{
				r:     RoundTrip{
					HandlerName:      handlerName,
					HTTPResponseCode: 500,
					Contents:         nil,
				},
				delay: 1000 * time.Millisecond,
			},
		},
	}
	handler := Wrap(impl, logChannels)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/myhandler", nil)
	handler.ServeHTTP(rr, req)
	handler.ServeHTTP(rr, req)
	handler.ServeHTTP(rr, req)

	prr := httptest.NewRecorder()
	pHandler := promhttp.Handler()
	req = httptest.NewRequest("GET", "/metrics", nil)
	pHandler.ServeHTTP(prr, req)

	body := prr.Body.String()

	// Assert
	for _, e := range expected {
		if !strings.Contains(body, e) {
			t.Errorf("expected body to contain, %s", e)
		}
	}
}

type mockHandler struct {
	index    int
	elements []roundTripTestElement
}

func (m *mockHandler) Handle(r *http.Request) RoundTrip {
	if len(m.elements) >= (m.index - 1) {
		e := m.elements[m.index]
		m.index = m.index + 1
		<- time.After(e.delay)
		return e.r
	}
	return RoundTrip{HTTPResponseCode: 200}
}

type roundTripTestElement struct {
	r     RoundTrip
	delay time.Duration
}