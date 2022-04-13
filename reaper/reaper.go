package reaper

import (
	"context"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Kreaper struct {
	labelSelector   labels.Selector
	Client          client.WithWatch
	Lifetime        time.Duration
	Target          Target
	TargetNamespace string
}

func (k Kreaper) Kill(ctx context.Context) (err error) {
	var (
		logger  = zap.L()
		podList corev1.PodList
	)

	if k.labelSelector, err = k.Target.Selector(); err != nil {
		return err
	}

	opts := []client.ListOption{
		client.MatchingLabelsSelector{Selector: k.labelSelector},
		client.InNamespace(k.TargetNamespace),
	}

	if err = k.Client.List(ctx, &podList, opts...); err != nil {
		logger.Error("failed to list", zap.Error(err))
		return err
	}

	if len(podList.Items) < 1 {
		logger.Warn("No pod targets found")
		return nil
	}

	for i := range podList.Items {
		logger.Info("Found pod", zap.String("pod_name", podList.Items[i].Name))
	}

	watcher, err := k.Client.Watch(ctx, &podList, opts...)
	if err != nil {
		logger.Error("failed to setup watch", zap.Error(err))
		return err
	}

	defer watcher.Stop()
	done, err := k.startPodWatcher(ctx, podList, opts)

	select {
	case <-time.After(k.Lifetime):
		logger.Info("Reached end of lifetime force delete all matching pods")
		if err := k.forceDeleteAll(ctx); err != nil {
			logger.Error("Failed to delete all pods", zap.Error(err))
			return err
		}
		return nil
	case <-done:
		logger.Info("All pods deleted")
		return nil
	}
}

func (k *Kreaper) startPodWatcher(ctx context.Context, podList corev1.PodList, opts []client.ListOption) (<-chan struct{}, error) {
	logger := zap.L()
	watcher, err := k.Client.Watch(ctx, &podList, opts...)
	if err != nil {
		logger.Error("failed to setup watch", zap.Error(err))
		return nil, err
	}

	done := make(chan struct{})

	go func() {
		defer watcher.Stop()

		for ev := range watcher.ResultChan() {
			if ev.Type != watch.Deleted {
				continue
			}

			if err := k.Client.List(ctx, &podList, opts...); err != nil {
				continue
			}

			if len(podList.Items) == 0 {
				close(done)
				break
			}
		}
	}()

	return done, nil
}

func (k *Kreaper) forceDeleteAll(ctx context.Context) error {
	return k.Client.DeleteAllOf(
		ctx,
		new(corev1.Pod),
		client.InNamespace(k.TargetNamespace),
		client.PropagationPolicy(metav1.DeletePropagationForeground),
		client.MatchingLabelsSelector{
			Selector: k.labelSelector,
		},
	)
}
