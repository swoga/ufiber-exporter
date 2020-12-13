package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/swoga/ufiber-exporter/api"
	"github.com/swoga/ufiber-exporter/cache"
	"github.com/swoga/ufiber-exporter/collector"
	"github.com/swoga/ufiber-exporter/config"
	"github.com/swoga/ufiber-exporter/model"
	"github.com/swoga/ufiber-exporter/version"
	"go.uber.org/zap"
)

var (
	sc        config.SafeConfig
	authCache = cache.New()
	log       *zap.Logger
)

func main() {
	// parse command line args
	configFile := flag.String("config.file", "", "")
	debug := flag.Bool("debug", false, "")
	flag.Parse()

	level := zap.InfoLevel
	if *debug {
		level = zap.DebugLevel
	}

	zapConfig := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      false,
		Encoding:         "console",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	log, _ = zapConfig.Build()
	defer log.Sync()
	log.Info("starting ufiber-exporter", zap.String("version", version.Version), zap.String("revision", version.Revision))

	// inital config load
	sc = config.New(*configFile)
	err := sc.LoadConfig()
	if err != nil {
		log.Fatal("error loading config", zap.Any("err", err))
	}

	// setup config reload
	hup := make(chan os.Signal, 1)
	signal.Notify(hup, syscall.SIGHUP)
	reloadRequest := make(chan chan error)
	go func() {
		for {
			var err error
			select {
			case <-hup:
				log.Debug("config reload triggerd by SIGHUP")
				err = sc.LoadConfig()
			case reloadResult := <-reloadRequest:
				log.Debug("config reload triggerd by API")
				err = sc.LoadConfig()
				reloadResult <- err
			}
			if err != nil {
				log.Error("error reloading config", zap.Any("err", err))
			} else {
				log.Info("reloaded config file")
			}
		}
	}()

	http.HandleFunc("/-/reload", func(w http.ResponseWriter, r *http.Request) {
		reloadResult := make(chan error)
		reloadRequest <- reloadResult
		err := <-reloadResult
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to reload config: %s", err), http.StatusInternalServerError)
		}
	})

	// start http server
	config := sc.Get()
	http.Handle(config.MetricsPath, promhttp.Handler())
	http.HandleFunc(config.ProbePath, handleRequest)

	log.Info("starting http server", zap.String("metrics_path", config.MetricsPath), zap.String("probe_path", config.ProbePath), zap.String("listen", config.Listen))

	err = http.ListenAndServe(config.Listen, nil)
	if err != nil {
		log.Fatal("error starting http server", zap.Any("err", err))
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	config := sc.Get()
	target := r.URL.Query().Get("target")
	if target == "" {
		log.Error("request with missing target")
		http.Error(w, "?target= missing", http.StatusBadRequest)
		return
	}

	log := log.With(zap.String("target", target))

	device, ok := config.Devices[target]
	if !ok {
		log.Error("unknown target")
		http.Error(w, "unknown target", http.StatusBadRequest)
		return
	}

	timeout := getTimeout(config, r)

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(timeout*float64(time.Second)))
	defer cancel()
	r = r.WithContext(ctx)

	start := time.Now()
	registry := prometheus.NewRegistry()
	exporterRegistry := prometheus.WrapRegistererWithPrefix("ufiber_exporter_", registry)

	err := probeDevice(ctx, log, target, device, exporterRegistry)
	var success float64 = 1
	if err != nil {
		log.Error("error probing device", zap.Any("err", err))
		success = 0
	}

	probeDurationGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "probe_duration_seconds",
		Help: "Returns how long the probe took to complete in seconds",
	})
	registry.MustRegister(probeDurationGauge)
	duration := time.Since(start).Seconds()
	probeDurationGauge.Set(duration)

	probeSuccessGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "probe_success",
		Help: "Displays whether or not the probe was a success",
	})
	registry.MustRegister(probeSuccessGauge)
	probeSuccessGauge.Set(success)

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func getTimeout(config *config.Config, r *http.Request) float64 {
	value := r.Header.Get("X-Prometheus-Scrape-Timeout-Seconds")
	if value != "" {
		timeout, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return timeout
		}
	}
	return config.Timeout
}

func probeDevice(ctx context.Context, log *zap.Logger, target string, device config.Device, registry prometheus.Registerer) error {
	statistics, interfaces, err := getFromAPI(ctx, log, target, device)
	if err != nil {
		// if there was an error somewhere, retry once

		// remove auth token, so login will be repeated
		authCache.Remove(target)
		statistics, interfaces, err = getFromAPI(ctx, log, target, device)
		if err != nil {
			return fmt.Errorf("error after retry: %s", err)
		}
	}

	collector.AddMetricsOlt(registry, *statistics, *interfaces)
	collector.AddMetricsOnu(registry)

	return nil
}

func getFromAPI(ctx context.Context, log *zap.Logger, target string, device config.Device) (*model.Statistics, *[]model.InterfacesInterface, error) {
	auth := authCache.Get(target)
	// if there is no X-Auth-Token in the cache try to login
	if auth == "" {
		res, err := api.DoLogin(ctx, log, device)

		if err != nil {
			return nil, nil, err
		}

		auth = res.Header.Get("X-Auth-Token")
		if auth == "" {
			return nil, nil, errors.New("no X-Auth-Token after login")
		}
		authCache.Set(target, auth)
	}

	statistics, err := api.GetStatistics(ctx, log, device, auth)
	if err != nil {
		return nil, nil, err
	}
	interfaces, err := api.GetInterfaces(ctx, log, device, auth)
	if err != nil {
		return nil, nil, err
	}

	return statistics, interfaces, nil
}
