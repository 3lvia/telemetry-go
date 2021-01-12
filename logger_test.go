package telemetry

import (
	"context"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestStart_forMetrics(t *testing.T) {
	// Arrange
	expectedGcp := `telemetry_app_cost{cloud="gcp"} 3.14`
	expectedAzure := `telemetry_app_cost{cloud="azure"} 100.11`
	expectedGauge := `telemetry_app_temp{room="bathroom"} 12.12`
	expectedHistogram := `# HELP telemetry_app_latency 
# TYPE telemetry_app_latency histogram
telemetry_app_latency_bucket{handler="metrics",le="0.005"} 0
telemetry_app_latency_bucket{handler="metrics",le="0.01"} 0
telemetry_app_latency_bucket{handler="metrics",le="0.025"} 0
telemetry_app_latency_bucket{handler="metrics",le="0.05"} 0
telemetry_app_latency_bucket{handler="metrics",le="0.1"} 0
telemetry_app_latency_bucket{handler="metrics",le="0.25"} 0
telemetry_app_latency_bucket{handler="metrics",le="0.5"} 0
telemetry_app_latency_bucket{handler="metrics",le="1"} 0
telemetry_app_latency_bucket{handler="metrics",le="2.5"} 0
telemetry_app_latency_bucket{handler="metrics",le="5"} 0
telemetry_app_latency_bucket{handler="metrics",le="10"} 0
telemetry_app_latency_bucket{handler="metrics",le="+Inf"} 2
telemetry_app_latency_sum{handler="metrics"} 294.34000000000003
telemetry_app_latency_count{handler="metrics"} 2`

	ctx := context.Background()
	system := "telemetry"
	app := "app"

	doneChan := make(chan struct{})
	cpt := &mockCapture{
		ch: doneChan,
	}

	// Act
	logChannels := Start(ctx,
		Named(system, app),
		WithCapture(cpt))

	go func() {
		logChannels.CountChan <- Metric{
			Name:        "cost",
			Value:       2.14,
			ConstLabels: map[string]string{
				"cloud": "gcp",
			},
		}
		logChannels.CountChan <- Metric{
			Name:        "cost",
			Value:       1.0,
			ConstLabels: map[string]string{
				"cloud": "gcp",
			},
		}
		logChannels.CountChan <- Metric{
			Name:        "cost",
			Value:       99.11,
			ConstLabels: map[string]string{
				"cloud": "azure",
			},
		}
		logChannels.CountChan <- Metric{
			Name:        "cost",
			Value:       1.0,
			ConstLabels: map[string]string{
				"cloud": "azure",
			},
		}
		logChannels.GaugeChan <- Metric{
			Name:        "temp",
			Value:       12.12,
			ConstLabels: map[string]string{
				"room": "bathroom",
			},
		}
		logChannels.HistogramChan <- Metric{
			Name:        "latency",
			Value:       193.11,
			ConstLabels: map[string]string{
				"handler": "metrics",
			},
		}
		logChannels.HistogramChan <- Metric{
			Name:        "latency",
			Value:       101.23,
			ConstLabels: map[string]string{
				"handler": "metrics",
			},
		}
	}()

	rr := httptest.NewRecorder()
	handler := promhttp.Handler()

	wg := &sync.WaitGroup{}
	wg.Add(7)
	go func(dc <-chan struct{}, wg *sync.WaitGroup) {
		for {
			<- dc
			wg.Done()
		}
	}(doneChan, wg)

	wg.Wait()

	req := httptest.NewRequest("GET", "/metrics", nil)

	handler.ServeHTTP(rr, req)

	body := rr.Body.String()

	if !strings.Contains(body, expectedGcp) {
		t.Error("did not contain gcp cost metric")
	}
	if !strings.Contains(body, expectedAzure) {
		t.Error("did not contain azure cost metric")
	}
	if !strings.Contains(body, expectedGauge) {
		t.Error("did not contain gauge metric")
	}
	if !strings.Contains(body, expectedHistogram) {
		t.Error("did not contain histogram metric")
	}
	//
	//fmt.Print(body)
}


//////////////////////////
///
/// Mocks
///
//////////////////////////

type mockCapture struct {
	captured []*CapturedEvent
	ch       chan<- struct{}
}

func (m *mockCapture) Capture(ce *CapturedEvent) {
	m.captured = append(m.captured, ce)
	m.ch <- struct{}{}
}