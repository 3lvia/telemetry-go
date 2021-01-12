package telemetry

import (
	"context"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_wrapper_ServeHTTP(t *testing.T) {
	// Arrange
	expectedMetrics := `# HELP http_myhandler_requests_total 
# TYPE http_myhandler_requests_total counter
http_myhandler_requests_total{code="200"} 1`

	ctx := context.Background()
	logChannels := Start(ctx,
		Empty())
	impl := &mockHandler{
		r: RoundTrip{
			HandlerName:      "myhandler",
			HTTPResponseCode: 200,
			Contents:         nil,
		},
	}
	handler := Wrap(impl, logChannels)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/myhandler", nil)
	handler.ServeHTTP(rr, req)

	prr := httptest.NewRecorder()
	pHandler := promhttp.Handler()
	req = httptest.NewRequest("GET", "/metrics", nil)
	pHandler.ServeHTTP(prr, req)

	body := prr.Body.String()

	// Assert
	if !strings.Contains(body, expectedMetrics) {
		t.Error("did not contain expected metrics")
	}
}


type mockHandler struct {
	r RoundTrip
}

func (m *mockHandler) Handle(r *http.Request) RoundTrip {
	return m.r
}