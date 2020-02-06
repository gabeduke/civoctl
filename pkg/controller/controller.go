package controller

import (
	"context"
	"fmt"
	"github.com/gabeduke/civo-controller/pkg/civo"
	"github.com/gabeduke/civo-controller/pkg/config"
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

func missing(a, b []string) string {
	ma := make(map[string]bool, len(a))
	for _, ka := range a {
		ma[ka] = true
	}
	for _, kb := range b {
		if !ma[kb] {
			return kb
		}
	}
	return ""
}

func Run(app config.App) {
	log.Info("Beginning Civo control loop")

	// Create prometheus metrics and serve the metrics.
	promreg := prometheus.NewRegistry()
	go func() {
		log.Infof("serving metrics on %s", metricsListenAddr)
		http.ListenAndServe(metricsListenAddr, promhttp.HandlerFor(promreg, promhttp.HandlerOpts{}))
	}()

	// Create all required components for the controller.
	lw := CreateListerWatcher(".civo-controller.yaml")
	st := CreateStorage()
	h := CreateHandler(log.StandardLogger())
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

func clustersWant() ([]string, error) {

	var clusterNames []string
	for _, cluster := range c.Clusters {
		clusterNames = append(clusterNames, cluster.Name)
	}

	return clusterNames, nil
}

func CreateListerWatcher(cfgFile string) controller.ListerWatcher {
	clustersWant, err := clustersWant(cfgFile)
	if err != nil {
		log.Error(err)
	}

	return &controller.ListerWatcherFunc{
		ListFunc: func(_ context.Context) ([]string, error) {

			var clusterNames []string
			for _, cluster := range listConfig.Clusters {
				clustersWant = append(clustersWant, cluster.Name)
			}

			return clusterNames, nil
		},
		WatchFunc: func(_ context.Context) (<-chan controller.Event, error) {
			c := make(chan controller.Event)
			go func() {

				for {
					// create a local copy so we can get fresh state
					watchConfig := config.Config{}
					v.Unmarshal(&watchConfig)
					var watchClusterNames []string
					for _, cluster := range watchConfig.Clusters {
						watchClusterNames = append(watchClusterNames, cluster.Name)
					}

					x := missing(watchClusterNames, clusterNames)
					if x != "" {
						id, err := civo.GetCluster(x)
						if err != nil {
							fmt.Println(err.Error())
						}
						c <- controller.Event{
							ID:   id,
							Kind: controller.EventDeleted,
						}
					}
					clusterNames = watchClusterNames
					time.Sleep(10 * time.Second)

				}
			}()
			return c, nil
		},
	}
}

func CreateStorage() controller.Storage {
	return controller.StorageFunc(func(_ context.Context, name string) (interface{}, error) {
		id, err := civo.GetCluster(name)
		if err != nil {
			return nil, err
		}

		return &civoCluster{ID: id, Name: name}, nil
	})
}

func CreateHandler(logger *log.Logger) controller.Handler {
	return &controller.HandlerFunc{
		AddFunc: func(_ context.Context, obj interface{}) error {
			cluster := obj.(*civoCluster)
			if cluster.ID == "" {
				logger.Infof("attempt create object %s", cluster.Name)
				civo.CreateCluster(cluster.Name)
				return nil
			}
			return nil
		},
		DeleteFunc: func(_ context.Context, id string) error {
			logger.Infof("attempt delete object %s", id)
			civo.DeleteCluster(id)
			return nil
		},
	}
}
