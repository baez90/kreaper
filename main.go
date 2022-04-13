package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/baez90/kreaper/reaper"
)

const defaultKreaperLifetime = 5 * time.Minute

var (
	kubeconfig string
	dryRun     bool
	logLevel   *zapcore.Level
	kreaper    = reaper.Kreaper{
		Target: lookupEnvOr("KREAPER_TARGET", "", reaper.ParseTarget),
	}
)

func main() {
	prepareFlags()
	setupLogging()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	if err := run(ctx); err != nil {
		cancel()
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	logger := zap.L()

	restCfg, err := loadRestConfig()
	if err != nil {
		logger.Error("Failed to get cluster config", zap.Error(err))
		return err
	}

	if kreaper.Client, err = client.NewWithWatch(restCfg, client.Options{}); err != nil {
		logger.Error("failed to prepare Kubernetes API client", zap.Error(err))
		return err
	}

	return kreaper.Kill(ctx)
}

func setupLogging() {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(*logLevel)
	logger, err := cfg.Build()
	if err != nil {
		log.Fatalf("Failed to setup logging: %v", err)
	}

	zap.ReplaceGlobals(logger)
}

func prepareFlags() {
	logLevel = zap.LevelFlag("log-level", zapcore.InfoLevel, "Log level to use")
	flag.Var(&kreaper.Target, "target", "Target that should be monitored")

	flag.BoolVar(
		&dryRun,
		"dry-run",
		lookupEnvOr("KREAPER_DRY_RUN", false, strconv.ParseBool),
		"Don't actually delete anything but only list all found pods matching the target - env variable: KREAPER_DRY_RUN",
	)

	flag.DurationVar(
		&kreaper.Lifetime,
		"lifetime",
		lookupEnvOr("KREAPER_LIFETIME", defaultKreaperLifetime, time.ParseDuration),
		"Lifetime after which all matching targets will be deleted - env variable: KREAPER_LIFETIME",
	)

	flag.StringVar(
		&kreaper.TargetNamespace,
		"target-namespace",
		lookupEnvOr("KREAPER_TARGET_NAMESPACE", "default", identity[string]),
		"Set target namespace in which kreaper will look for pods - env variable: KREAPER_TARGET_NAMESPACE",
	)

	if home := homedir.HomeDir(); home != "" {
		flag.StringVar(&kubeconfig, "kubeconfig", lookupEnvOr("KUBECONFIG", filepath.Join(home, ".kube", "config"), identity[string]), "(optional) absolute path to the kubeconfig file")
	} else {
		flag.StringVar(&kubeconfig, "kubeconfig", lookupEnvOr("KUBECONFIG", "", identity[string]), "absolute path to the kubeconfig file")
	}

	flag.Parse()
}

func loadRestConfig() (cfg *rest.Config, err error) {
	if cfg, err = rest.InClusterConfig(); err == nil {
		return cfg, nil
	}

	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

func lookupEnvOr[T any](envKey string, fallback T, parse func(envVal string) (T, error)) T {
	envVal := os.Getenv(envKey)
	if envVal == "" {
		return fallback
	}

	if parsed, err := parse(envVal); err != nil {
		return fallback
	} else {
		return parsed
	}
}

func identity[T any](in T) (T, error) {
	return in, nil
}
