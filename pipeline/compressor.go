package pipeline

import (
	"bytes"
	"fmt"

	"github.com/go-resty/resty/v2"
	//_ "github.com/minio/minio-go/v7"
	_ "github.com/nfnt/resize"
	"go.uber.org/zap"
)

func (p OCRProcess) StageCompressBytes(imageByte []byte) (compressedImage []byte, err error) {

	reader := bytes.NewReader(imageByte)

	client := resty.New()

	resp, err := client.R().
		SetFileReader("image", "image.png", reader).
		Post("http://192.168.0.106:2201/compress")

	if err != nil {
		p.logger.Info("stage compress bytes", zap.Any(processId, p.ProcessId), zap.Any("resp", resp))
		return
	}

	if resp.StatusCode() != 200 {
		err = fmt.Errorf("bad status code: %d - %s", resp.StatusCode(), resp.String())
		p.logger.Info("status code", zap.Any(processId, p.ProcessId), zap.Any("respStatusCode", resp.StatusCode()), zap.Any("respString", resp.String()))
		return
	}

	compressedImage = resp.Body()

	return
}
