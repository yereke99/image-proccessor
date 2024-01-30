package main

import (
	"ImageProcessor/app"
	"ImageProcessor/components"
	"ImageProcessor/config"
	"ImageProcessor/pipeline"
	"context"
	"fmt"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {

	// Initialize logger for application-wide logging.
	logger, err := app.NewLogger()
	if err != nil {
		panic(err) // Panic if there's an error creating the logger.
	}
	defer logger.Sync() // Defer the synchronization of the logger's output resources.

	logger.Info("create configuration")

	conf, err := config.NewConfig(logger)
	if err != nil {
		return
	}

	args := os.Args[1:]
	if len(args) > 0 {
		if strings.Contains(args[0], "--method=") {
			tmp := strings.Split(args[0], "=")
			conf.OCRMethod = components.DefineMethod(tmp[1])
		}
	}

	ctx, contextCancel := context.WithCancel(context.Background())
	metric := pipeline.NewMetrics(ctx, conf)

	go func() {
		for {
			if err := ctx.Err(); err != nil {
				logger.Info("main goroutine closed")
				logger.Info("current number of processes", zap.Any("pool", len(conf.PoolChannel)))
				close(conf.PoolChannel)
				return
			}
			// Create an OCR process instance with the provided logger.
			ocrProcess := pipeline.NewOCRProcess(ctx, logger, conf, metric)
			ocrProcess.PoolChannel <- struct{}{}
			go ocrProcess.RunStages()
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	logger.Info("context cancel")
	contextCancel()

	for i := 5; i > 0; i-- {
		if len(conf.PoolChannel) > 0 {
			i += 1
		} else {
			fmt.Println(i)
		}
		time.Sleep(1 * time.Second)
	}

	logger.Info("application closed")
}
