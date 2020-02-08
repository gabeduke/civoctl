package controller

import (
	"context"
	"github.com/gabeduke/civoctl/pkg/civo"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spotahome/gontroller/controller"
	"github.com/spotahome/gontroller/metrics"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	metricsListenAddr = ":8081"
)

type civoCluster struct {
	ID   string
	Name string
}

// empty struct (0 bytes)
type void struct{}

// missing looks for strings in a that are missing from b
func missing(a, b []string) []string {
	// create map with length of the 'a' slice
	ma := make(map[string]void, len(a))
	diffs := []string{}
	// Convert first slice to map with empty struct (0 bytes)
	for _, ka := range a {
		ma[ka] = void{}
	}
	// find missing values in a
	for _, kb := range b {
		if _, ok := ma[kb]; !ok {
			diffs = append(diffs, kb)
		}
	}
	return diffs
}

// Run begins the CivoCtl loop
func Run(civoCtl *civo.CivoCtl) {
	log.Info("Beginning Civo control loop")

	// Create prometheus metrics and serve the metrics.
	promreg := prometheus.NewRegistry()
	go func() {
		log.Infof("serving metrics on %s", metricsListenAddr)
		http.ListenAndServe(metricsListenAddr, promhttp.HandlerFor(promreg, promhttp.HandlerOpts{}))
	}()

	// Create all required components for the controller.
	lw := listerWatcher(civoCtl)
	st := storage(civoCtl)
	h := handler(civoCtl, log.StandardLogger())
	metricsrecorder := metrics.NewPrometheus(promreg)

	// Create and run the controller.
	ctrl, err := controller.New(controller.Config{
		Name:            "civoctl",
		Workers:         3,
		MaxRetries:      2,
		ListerWatcher:   lw,
		Handler:         h,
		Storage:         st,
		MetricsRecorder: metricsrecorder,
		Logger:          log.StandardLogger(),
	})
	if err != nil {
		log.Errorf("error creating controller: %s", err)
		os.Exit(1)
	}

	go func() {
		err = ctrl.Run(context.Background())
		if err != nil {
			log.Errorf("error running controller: %s", err)
			os.Exit(1)
		}
	}()

	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGTERM, syscall.SIGINT)
	<-sigC
	log.Infof("signal captured, exiting...")
}

func getClustersFromCfg(civoCtl *civo.CivoCtl) []string {
	cfg := civoCtl.Config()
	var clusters []string
	for _, c := range cfg.Clusters {
		clusters = append(clusters, c.Name)
	}
	return clusters
}

func listerWatcher(civoCtl *civo.CivoCtl) controller.ListerWatcher {

	return &controller.ListerWatcherFunc{
		ListFunc: func(_ context.Context) ([]string, error) {
			c := getClustersFromCfg(civoCtl)
			return c, nil
		},
		WatchFunc: func(_ context.Context) (<-chan controller.Event, error) {
			c := make(chan controller.Event)
			go func() {
				for {
					want := getClustersFromCfg(civoCtl)
					have := civoCtl.Client.GetClusterNames()
					extras := missing(want, have)

					for _, name := range extras {
						id, err := civoCtl.Client.GetClusterId(name)
						if err != nil {
							log.Error(err)
						}

						c <- controller.Event{
							ID:   id,
							Kind: controller.EventDeleted,
						}

					}

					time.Sleep(10 * time.Second)

				}
			}()
			return c, nil
		},
	}
}

func storage(civoCtl *civo.CivoCtl) controller.Storage {
	return controller.StorageFunc(func(_ context.Context, name string) (interface{}, error) {
		id, err := civoCtl.Client.GetClusterId(name)
		if err != nil {
			return nil, err
		}

		return &civoCluster{ID: id, Name: name}, nil
	})
}

func handler(civoCtl *civo.CivoCtl, logger *log.Logger) controller.Handler {
	return &controller.HandlerFunc{
		AddFunc: func(_ context.Context, obj interface{}) error {
			cluster := obj.(*civoCluster)
			if cluster.ID == "" {
				logger.Infof("attempt create object %s", cluster.Name)
				civoCtl.Client.CreateCluster(cluster.Name)
				return nil
			}
			return nil
		},
		DeleteFunc: func(_ context.Context, id string) error {
			logger.Infof("attempt delete object %s", id)
			if civoCtl.Dangerous {
				civoCtl.Client.DeleteCluster(id)
				return nil
			}

			logger.Warn("delete blocked, enable dangerous mode to proceed")
			return nil
		},
	}
}
