package tests

import (
	"ImageProcessor/pipeline"
	"fmt"
	"net/http"
	"testing"

	"github.com/VictoriaMetrics/metrics"
)

func TestExampleWritePrometheus(t *testing.T) {
	// Export all the registered metrics in Prometheus format at `/metrics` http path.
	http.HandleFunc("/metrics", func(w http.ResponseWriter, req *http.Request) {
		metrics.WritePrometheus(w, true)
	})

	err := http.ListenAndServe(":8070", nil)

	fmt.Println(err)
}

func TestSnt(t *testing.T) {
	fmt.Println(pipeline.IncValue(`OCR_done{host="111"}`))

}
