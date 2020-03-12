package controller

import (
	"context"
	"fmt"

	"github.com/110y/bootes/internal/k8s/store"
	"github.com/110y/bootes/internal/xds/cache"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewClusterReconciler(s store.Store, c cache.Cache, l logr.Logger) reconcile.Reconciler {
	return &ClusterReconciler{
		store:  s,
		cache:  c,
		logger: l,
	}
}

type ClusterReconciler struct {
	store  store.Store
	cache  cache.Cache
	logger logr.Logger
}

func (r *ClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	version := uuid.New().String()
	logger := r.logger.WithValues("version", version)

	logger.Info(fmt.Sprintf("reconciling: %s", req.NamespacedName))

	clusters, err := r.store.ListClustersByNamespace(ctx, req.Namespace)
	if err != nil {
		logger.Error(err, "failed to list clusters")
		return ctrl.Result{}, err
	}

	pods, err := r.store.ListPodsByNamespace(ctx, req.Namespace)
	if err != nil {
		logger.Error(err, "failed to list pods")
		return ctrl.Result{}, err
	}

	for _, pod := range pods.Items {
		if err := r.cache.UpdateClusters(toNodeName(pod.Name, pod.Namespace), version, clusters.Items); err != nil {
			logger.Error(err, "failed to update clusuters")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}
