package pipeline

import (
	"bytes"
	"context"

	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

func (p OCRProcess) StageToS3(b []byte) error {
	r := bytes.NewReader(b)

	ctx := context.Background()

	bucketFileName := p.task.MessageId + "/" + p.task.Fname

	_, err := p.appConfig.S3.PutObject(ctx, p.appConfig.AwsScreenBucket, bucketFileName, r, int64(len(b)), minio.PutObjectOptions{})
	if err != nil {
		p.logger.Info("S3 put object", zap.Any(processId, p.ProcessId), zap.Any("bucket name", p.appConfig.AwsScreenBucket), zap.Any("fileName", bucketFileName))
		return err
	}
	return nil
}

// StageToS3CallFailed TODO: Temporary solution for debugging
func (p OCRProcess) StageToS3CallFailed(b []byte) error {

	r := bytes.NewReader(b)

	ctx := context.Background()

	bucketFileName := p.task.MessageId + "/" + p.task.Fname

	_, err := p.appConfig.S3.PutObject(ctx, "calleridrep-screenshots-failed", p.task.MessageId+"/call_failed_"+p.task.Fname, r, int64(len(b)), minio.PutObjectOptions{})
	if err != nil {
		p.logger.Info("S3 put object", zap.Any(processId, p.ProcessId), zap.Any("bucket name", p.appConfig.AwsScreenBucket), zap.Any("fileName", bucketFileName))
		return err
	}
	return nil
}
