package controller

import (
	"context"
	"github.com/gabeduke/civo-controller/pkg/civo"
	"github.com/spotahome/gontroller/controller"
)

type civoCluster struct {
	ID string
	Name string
}

func CreateListeWatcher() controller.ListerWatcher {
	return &controller.ListerWatcherFunc{
		ListFunc: func(_ context.Context) ([]string, error) {
			return civo.WantClusters, nil
		},
		WatchFunc: func(_ context.Context) (<-chan controller.Event, error) {
			c := make(chan controller.Event)
			go func() {
				//t := time.NewTicker(10 * time.Millisecond)
				//for range t.C {
				//	c <- controller.Event{
				//		ID:   "faked-obj1",
				//		Kind: controller.EventAdded,
				//	}
				//}
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

func CreateHandler(logger log.Logger) controller.Handler {
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
			return nil
		},
	}
}

