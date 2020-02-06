package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	civoController "github.com/gabeduke/civo-controller/pkg/controller"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spotahome/gontroller/controller"
	"github.com/spotahome/gontroller/log"
	"github.com/spotahome/gontroller/metrics"
)


const (
	metricsListenAddr = ":8081"
)

func main() {
	logger := log.Std{Debug: true}

	// Create prometheus metrics and serve the metrics.
	promreg := prometheus.NewRegistry()
	go func() {
		logger.Infof("serving metrics on %s", metricsListenAddr)
		http.ListenAndServe(metricsListenAddr, promhttp.HandlerFor(promreg, promhttp.HandlerOpts{}))
	}()

	// Create all required components for the controller.
	lw := civoController.CreateListeWatcher()
	st := civoController.CreateStorage()
	h := civoController.CreateHandler(logger)
	metricsrecorder := metrics.NewPrometheus(promreg)

	// Create and run the controller.
	ctrl, err := controller.New(controller.Config{
		Name:            "civo-controller",
		Workers:         3,
		MaxRetries:      2,
		ListerWatcher:   lw,
		Handler:         h,
		Storage:         st,
		MetricsRecorder: metricsrecorder,
		Logger:          logger,
	})
	if err != nil {
		logger.Errorf("error creating controller: %s", err)
		os.Exit(1)
	}

	go func() {
		err = ctrl.Run(context.Background())
		if err != nil {
			logger.Errorf("error running controller: %s", err)
			os.Exit(1)
		}
	}()

	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGTERM, syscall.SIGINT)
	<-sigC
	logger.Infof("signal captured, exiting...")
}

