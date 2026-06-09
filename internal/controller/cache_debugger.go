package controller

import (
	"context"
	"sort"
	"time"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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
			// List StatefulSets
			stsList := &appsv1.StatefulSetList{}
			if err := d.Client.List(ctx, stsList); err != nil {
				logger.Error(err, "failed to list cached StatefulSets")
			} else {
				stsNames := make([]string, 0, len(stsList.Items))
				for _, sts := range stsList.Items {
					stsNames = append(stsNames, sts.Namespace+"/"+sts.Name)
				}
				sort.Strings(stsNames)
				logger.Info("cached StatefulSets snapshot", "count", len(stsNames), "items", stsNames)
			}

			// List Pods
			podList := &corev1.PodList{}
			if err := d.Client.List(ctx, podList); err != nil {
				logger.Error(err, "failed to list cached Pods")
			} else {
				podNames := make([]string, 0, len(podList.Items))
				for _, pod := range podList.Items {
					podNames = append(podNames, pod.Namespace+"/"+pod.Name)
				}
				sort.Strings(podNames)
				logger.Info("cached Pods snapshot", "count", len(podNames), "items", podNames)
			}
		}
	}
}
