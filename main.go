package main

import (
	"context"
	"flag"
	"path/filepath"
	"time"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/baez90/kreaper/reaper"
)

type ReaperTarget string

var (
	kubeconfig string
	target     reaper.Target
	lifetime   time.Duration
)

func main() {
	prepareFlags()
	restCfg, err := loadRestConfig()
	if err != nil {
		panic(err)
	}
	k8sClient, err := client.NewWithWatch(restCfg, client.Options{})

	labels := client.MatchingLabels{
		"from": "value",
	}
	k8sClient.Watch(context.Background(), nil, labels)
}

func prepareFlags() {
	flag.Var(&target, "target", "Target that should be monitored")
	flag.DurationVar(&lifetime, "lifetime", 5*time.Minute, "Lifetime after which all matching targets will be deleted")
	if home := homedir.HomeDir(); home != "" {
		flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	}
}

func loadRestConfig() (cfg *rest.Config, err error) {
	if cfg, err = rest.InClusterConfig(); err == nil {
		return cfg, nil
	}

	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}
