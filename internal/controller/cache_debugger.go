package controller

import (
	"context"
	"sort"
	"time"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CacheDebugger struct {
	Client client.Client
	Cache  cache.Cache
	Log    logr.Logger
}

func (d *CacheDebugger) Start(ctx context.Context) error {
	logger := d.Log.WithName("cache-debugger")

	if ok := d.Cache.WaitForCacheSync(ctx); !ok {
		logger.Info("cache sync did not complete before shutdown")
		return nil
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	logger.Info("cache debugger started")

	for {
		select {
		case <-ctx.Done():
			logger.Info("cache debugger stopped")
			return nil
		case <-ticker.C:
			stsList := &appsv1.StatefulSetList{}
			if err := d.Client.List(ctx, stsList); err != nil {
				logger.Error(err, "failed to list cached StatefulSets")
				continue
			}

			names := make([]string, 0, len(stsList.Items))
			for _, sts := range stsList.Items {
				names = append(names, sts.Namespace+"/"+sts.Name)
			}
			sort.Strings(names)

			logger.Info("cached StatefulSets snapshot", "count", len(names), "items", names)
		}
	}
}
