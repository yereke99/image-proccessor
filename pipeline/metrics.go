package pipeline

import (
	"ImageProcessor/config"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

const (
	portHTTPMetrics = 9097
)

var (
	ErrCouldNotCreateRequest = errors.New("client: could not create request: %s\n")
	ErrSendMetric            = errors.New("error send metric: %s\n")
	ErrHttpRequest           = errors.New("error do sends an HTTP request: %s\n")
	ErrPushMetric            = errors.New("error push metric: %s\n")
)

type MetricManager struct {
	metricAddressHTTP  string
	metricsHTTPChannel chan httpMetric
	httpclient         *http.Client
	conf               *config.Config
	ctx                context.Context
}

type httpMetric struct {
	MetricName  string  `json:"metricName"`
	MetricValue float64 `json:"metricValue"`
}

// SendMetric Metric string format.
//
// Example: `device_call{instance="cloud1",device_id="19007"}`
func (m *MetricManager) SendMetric(metricName string, metricValue float64) {
	if strings.Contains("ocr_error", metricName) {
		log.Println("TEST OK")
	}
	m.metricsHTTPChannel <- httpMetric{MetricName: metricName, MetricValue: metricValue}
}

func NewMetrics(ctx context.Context, conf *config.Config) *MetricManager {

	addressHTTPMetrics := fmt.Sprintf("%s:%d%s", conf.MetricReceiverAddress, portHTTPMetrics, "/metrics")

	metricsManager := &MetricManager{
		metricAddressHTTP:  addressHTTPMetrics,
		metricsHTTPChannel: make(chan httpMetric, 10000),
		httpclient:         &http.Client{},
		conf:               conf,
		ctx:                ctx,
	}

	go metricsManager.listenerHTTPMetric()

	return metricsManager
}

// Metric listener http
func (m *MetricManager) listenerHTTPMetric() {

	for _metric := range m.metricsHTTPChannel {

		if m.ctx.Err() != nil {
			return
		}

		jsonMetric, err := json.Marshal(_metric)
		if err != nil {
			m.conf.Logger.Error("error send metric", zap.Error(err), zap.Any("metric", _metric))
			continue
		}
		bodyReader := bytes.NewReader(jsonMetric)
		req, err := http.NewRequest(http.MethodPatch, m.metricAddressHTTP, bodyReader)
		if err != nil {
			m.conf.Logger.Error("could not create request", zap.Error(err), zap.Any("metricAddressHTTP", m.metricAddressHTTP), zap.Any("bodyReader", bodyReader))
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		go func() {
			resp, err := m.httpclient.Do(req)
			if err != nil {
				m.conf.Logger.Error("error do sends an HTTP request", zap.Error(err), zap.Any("resp", resp))
				return
			}
			resp.Body.Close()
		}()
	}
}
