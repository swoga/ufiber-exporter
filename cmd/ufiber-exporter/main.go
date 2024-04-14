package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/expfmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/swoga/ufiber-exporter/api"
	"github.com/swoga/ufiber-exporter/cache"
	"github.com/swoga/ufiber-exporter/collector"
	"github.com/swoga/ufiber-exporter/config"
	"github.com/swoga/ufiber-exporter/model"
)

var (
	version       = "dev"
	sc            config.SafeConfig
	authCache     = cache.New()
	consoleWriter = zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
)

func main() {
	log.Logger = log.Output(consoleWriter)
	log.Info().Str("version", version).Msg("starting ufiber-exporter")

	// parse command line args
	configFile := flag.String("config.file", "config.yml", "")
	debug := flag.Bool("debug", false, "")
	flag.Parse()

	if *debug {
		log.Logger = log.Logger.Level(zerolog.DebugLevel)
	} else {
		log.Logger = log.Logger.Level(zerolog.InfoLevel)
	}

	// inital config load
	sc = config.New(*configFile)
	err := sc.LoadConfig()
	if err != nil {
		log.Panic().Err(err).Msg("error loading config")
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
				log.Debug().Msg("config reload triggerd by SIGHUP")
				err = sc.LoadConfig()
			case reloadResult := <-reloadRequest:
				log.Debug().Msg("config reload triggerd by API")
				err = sc.LoadConfig()
				reloadResult <- err
			}
			if err != nil {
				log.Err(err).Msg("error reloading config")
			} else {
				log.Info().Msg("reloaded config file")
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

	log.Info().Str("metrics_path", config.MetricsPath).Str("listen", config.Listen).Str("probe_path", config.ProbePath).Msg("starting http server")

	err = http.ListenAndServe(config.Listen, nil)
	if err != nil {
		log.Panic().Err(err).Msg("error starting http server")
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	debugStr := r.URL.Query().Get("debug")
	debug := false
	if debugStr != "" {
		debug = true
	}
	traceStr := r.URL.Query().Get("trace")
	trace := false
	if traceStr != "" {
		trace = true
	}

	requestLog := log.With().Logger()

	if debug || trace {
		debugWriter := zerolog.ConsoleWriter{Out: w, TimeFormat: time.RFC3339, NoColor: true}
		multi := zerolog.MultiLevelWriter(consoleWriter, debugWriter)
		requestLog = requestLog.Output(multi)
		w.Header().Set("Content-Type", "text/plain")
	}

	conf := sc.Get()
	target := r.URL.Query().Get("target")
	if target == "" {
		requestLog.Error().Msg("request with missing target")
		http.Error(w, "?target= missing", http.StatusBadRequest)
		return
	}

	requestLog = requestLog.With().Str("target", target).Logger()

	device, ok := conf.GetDevice(target)
	if !ok {
		_, err := net.LookupHost(target)
		if err != nil {
			requestLog.Err(err).Msg("invalid target")
			http.Error(w, "invalid target", http.StatusBadRequest)
			return
		}

		requestLog.Debug().Msg("unconfigured target, use param as address")
		device = &config.Device{
			Address:  target,
			Username: &conf.Global.Username,
			Password: &conf.Global.Password,
			Options:  &conf.Global.Options,
		}
	}

	deviceOptions := *device.Options

	paramExportOLT := r.URL.Query().Get("export_olt")
	if paramExportOLT != "" {
		deviceOptions.ExportOLT = paramExportOLT == "1"
	}
	paramExportONUs := r.URL.Query().Get("export_onus")
	if paramExportONUs != "" {
		deviceOptions.ExportONUs = paramExportONUs == "1"
	}
	paramExportMACTable := r.URL.Query().Get("export_mac_table")
	if paramExportMACTable != "" {
		deviceOptions.ExportMACTable = paramExportMACTable == "1"
	}

	timeout := getTimeout(conf, r)

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(timeout*float64(time.Second)))
	defer cancel()
	r = r.WithContext(ctx)

	start := time.Now()
	registry := prometheus.NewRegistry()

	var success float64 = 1

	data, err := getFromAPIWithRetry(ctx, requestLog, target, *device, deviceOptions)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		requestLog.Err(err).Msg("error getting data from API")
		success = 0
	} else {
		exporterRegistry := prometheus.WrapRegistererWithPrefix("ufiber_exporter_", registry)
		err = addMetrics(data, deviceOptions, exporterRegistry)
		if err != nil {
			requestLog.Err(err).Msg("error adding metrics")
			success = 0
		}

		probeDurationGauge := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "probe_duration_seconds",
			Help: "Returns how long the probe took to complete in seconds",
		})
		registry.MustRegister(probeDurationGauge)
		duration := time.Since(start).Seconds()
		probeDurationGauge.Set(duration)
	}

	probeSuccessGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "probe_success",
		Help: "Displays whether or not the probe was a success",
	})
	registry.MustRegister(probeSuccessGauge)
	probeSuccessGauge.Set(success)

	if debug || trace {
		mfs, _ := registry.Gather()
		for _, mf := range mfs {
			expfmt.MetricFamilyToText(w, mf)
		}
		return
	}

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

type apiData struct {
	statistics   *model.Statistics
	interfaces   *[]model.InterfacesInterface
	onus         *[]model.ONU
	onusSettings *[]model.ONUSettings
	macTable     *[]model.MACTable
}

func addMetrics(data apiData, deviceOptions config.Options, registry prometheus.Registerer) error {
	if deviceOptions.ExportOLT {
		err := collector.AddMetricsOlt(prometheus.WrapRegistererWithPrefix("olt_", registry), *data.statistics, *data.interfaces)
		if err != nil {
			return err
		}
	}
	if deviceOptions.ExportONUs {
		err := collector.AddMetricsOnu(prometheus.WrapRegistererWithPrefix("onu_", registry), *data.onus, *data.onusSettings)
		if err != nil {
			return err
		}
	}
	if deviceOptions.ExportMACTable {
		err := collector.AddMetricsOnuMACTable(prometheus.WrapRegistererWithPrefix("onu_", registry), *data.macTable)
		if err != nil {
			return err
		}
	}

	return nil
}

func getFromAPIWithRetry(ctx context.Context, log zerolog.Logger, target string, device config.Device, deviceOptions config.Options) (apiData, error) {
	data, err := getFromAPI(ctx, log, target, device, deviceOptions)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return data, err
		}
		// if there was an error somewhere, retry once
		log.Err(err).Msg("error on first try")

		// remove auth token, so login will be repeated
		authCache.Remove(target)
		data, err = getFromAPI(ctx, log, target, device, deviceOptions)
		if err != nil {
			return data, fmt.Errorf("error after retry: %w", err)
		}
	}

	return data, nil
}

func getFromAPI(ctx context.Context, log zerolog.Logger, target string, device config.Device, deviceOptions config.Options) (apiData, error) {
	data := apiData{}
	auth := authCache.Get(target)
	// if there is no X-Auth-Token in the cache try to login
	if auth == "" {
		res, err := api.DoLogin(ctx, log, device)

		if err != nil {
			return data, err
		}

		auth = res.Header.Get("X-Auth-Token")
		if auth == "" {
			return data, errors.New("no X-Auth-Token after login")
		}
		authCache.Set(target, auth)
	}

	if deviceOptions.ExportOLT {
		statistics, err := api.GetStatistics(ctx, log, device, auth)
		if err != nil {
			return data, err
		}
		data.statistics = statistics

		interfaces, err := api.GetInterfaces(ctx, log, device, auth)
		if err != nil {
			return data, err
		}
		data.interfaces = interfaces
	}
	if deviceOptions.ExportONUs {
		onus, err := api.GetONUs(ctx, log, device, auth)
		if err != nil {
			return data, err
		}
		data.onus = onus

		onusSettings, err := api.GetONUsSettings(ctx, log, device, auth)
		if err != nil {
			return data, err
		}
		data.onusSettings = onusSettings
	}
	if deviceOptions.ExportMACTable {
		macTable, err := api.GetMACTable(ctx, log, device, auth)
		if err != nil {
			return data, err
		}
		data.macTable = macTable
	}

	return data, nil
}
