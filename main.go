package main

import (
	"context"

	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func main() {
	restCfg, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}
	k8sClient, err := client.NewWithWatch(restCfg, client.Options{})

	labels := client.MatchingLabels{
		"from": "value",
	}
	k8sClient.Watch(context.Background(), nil, labels)
}
