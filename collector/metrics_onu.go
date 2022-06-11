package collector

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/swoga/ufiber-exporter/model"
)

func AddMetricsOnu(registry prometheus.Registerer, onus []model.ONU, onussettings []model.ONUSettings) error {
	authorizedGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "authorized",
	}, []string{"serial"})
	registry.MustRegister(authorizedGaugeVec)
	connectedGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "connected",
	}, []string{"serial"})
	registry.MustRegister(connectedGaugeVec)
	connectionTimeCounterVec := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "connection_time",
	}, []string{"serial"})
	registry.MustRegister(connectionTimeCounterVec)
	distanceGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "distance",
	}, []string{"serial"})
	registry.MustRegister(distanceGaugeVec)
	oltPortGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pon",
	}, []string{"serial", "pon"})
	registry.MustRegister(oltPortGaugeVec)
	infoGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "info",
	}, []string{"serial", "firmware_version", "mac", "error", "dying_gasp", "given_name", "model", "mode"})
	registry.MustRegister(infoGaugeVec)

	rxPowerGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "rx_power",
	}, []string{"serial"})
	registry.MustRegister(rxPowerGaugeVec)
	txPowerGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "tx_power",
	}, []string{"serial"})
	registry.MustRegister(txPowerGaugeVec)

	portPluggedGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "port_plugged",
	}, []string{"serial", "name"})
	registry.MustRegister(portPluggedGaugeVec)
	portRxBytesCounterVec := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "port_rx_bytes",
	}, []string{"serial", "name"})
	registry.MustRegister(portRxBytesCounterVec)
	portTxBytesCounterVec := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "port_tx_bytes",
	}, []string{"serial", "name"})
	registry.MustRegister(portTxBytesCounterVec)
	portInfoGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "port_info",
	}, []string{"serial", "name", "speed"})
	registry.MustRegister(portInfoGaugeVec)

	rxBytesCounterVec := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "rx_bytes",
	}, []string{"serial"})
	registry.MustRegister(rxBytesCounterVec)
	txBytesCounterVec := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tx_bytes",
	}, []string{"serial"})
	registry.MustRegister(txBytesCounterVec)

	systemCPUGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cpu",
	}, []string{"serial"})
	registry.MustRegister(systemCPUGaugeVec)
	systemMemGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "memory",
	}, []string{"serial"})
	registry.MustRegister(systemMemGaugeVec)
	systemTempGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "temperature",
	}, []string{"serial", "sensor"})
	registry.MustRegister(systemTempGaugeVec)
	systemUptimeCounterVec := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "uptime",
	}, []string{"serial"})
	registry.MustRegister(systemUptimeCounterVec)
	systemVoltageGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "voltage",
	}, []string{"serial"})
	registry.MustRegister(systemVoltageGaugeVec)

	upgradeStatusGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "upgrade_status",
	}, []string{"serial"})
	registry.MustRegister(upgradeStatusGaugeVec)

	var onuSettingsMap = map[string]model.ONUSettings{}
	for _, onusettings := range onussettings {
		onuSettingsMap[onusettings.Serial] = onusettings
	}

	for _, onu := range onus {
		onusettings, ok := onuSettingsMap[onu.Serial]
		if !ok {
			return fmt.Errorf("settings for ONU %s not found", onu.Serial)
		}
		infoGaugeVec.WithLabelValues(onu.Serial, onu.FirmwareVersion, onu.MAC, onu.Error, onu.DyingGasp, onusettings.Name, onusettings.Model, onusettings.Mode).Set(1)

		var connected float64
		if onu.Connected {
			connected = 1
		}
		connectedGaugeVec.WithLabelValues(onu.Serial).Set(connected)
		if onu.Authorized != nil {
			var authorized float64
			if *onu.Authorized {
				authorized = 1
			}
			authorizedGaugeVec.WithLabelValues(onu.Serial).Set(authorized)
		}
		if onu.ConnectionTime != nil {
			connectionTimeCounterVec.WithLabelValues(onu.Serial).Add(*onu.ConnectionTime)
		}
		if onu.Distance != nil {
			distanceGaugeVec.WithLabelValues(onu.Serial).Set(*onu.Distance)
		}
		if onu.OLTPort != nil {
			oltPortGaugeVec.WithLabelValues(onu.Serial, fmt.Sprintf("%.0f", *onu.OLTPort)).Set(1)
		}
		if onu.RxPower != nil {
			rxPowerGaugeVec.WithLabelValues(onu.Serial).Set(*onu.RxPower)
		}
		if onu.TxPower != nil {
			txPowerGaugeVec.WithLabelValues(onu.Serial).Set(*onu.TxPower)
		}
		if onu.Statistics != nil {
			rxBytesCounterVec.WithLabelValues(onu.Serial).Add(onu.Statistics.RxBytes)
			txBytesCounterVec.WithLabelValues(onu.Serial).Add(onu.Statistics.TxBytes)
		}
		if onu.Ports != nil {
			for i, port := range *onu.Ports {
				var portStat model.ONUStatistics
				if onu.PortsStat != nil {
					portsStat := *onu.PortsStat
					portStat = portsStat[i]
				}

				var plugged float64
				if port.Plugged {
					plugged = 1
				}
				portPluggedGaugeVec.WithLabelValues(onu.Serial, port.ID).Set(plugged)

				portRxBytesCounterVec.WithLabelValues(onu.Serial, port.ID).Add(portStat.RxBytes)
				portTxBytesCounterVec.WithLabelValues(onu.Serial, port.ID).Add(portStat.TxBytes)
				portInfoGaugeVec.WithLabelValues(onu.Serial, port.ID, port.Speed).Set(1)
			}
		}
		if onu.System != nil {
			systemCPUGaugeVec.WithLabelValues(onu.Serial).Set(onu.System.CPU)
			systemMemGaugeVec.WithLabelValues(onu.Serial).Set(onu.System.Mem)
			for sensor, value := range onu.System.Temperature {
				systemTempGaugeVec.WithLabelValues(onu.Serial, sensor).Set(value)
			}
			systemUptimeCounterVec.WithLabelValues(onu.Serial).Add(onu.System.Uptime)
			systemVoltageGaugeVec.WithLabelValues(onu.Serial).Set(onu.System.Voltage)
			var upgradeStatus float64
			switch onu.UpgradeStatus.Status {
			case "in_progress":
				upgradeStatus = 2
			case "finished":
				upgradeStatus = 1
			case "failed":
				upgradeStatus = 0
			default:
				upgradeStatus = -1
			}
			upgradeStatusGaugeVec.WithLabelValues(onu.Serial).Set(upgradeStatus)
		}
	}

	return nil
}

func AddMetricsOnuMACTable(registry prometheus.Registerer, macTable []model.MACTable) error {
	macGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "fdb",
	}, []string{"serial", "mac"})
	registry.MustRegister(macGaugeVec)

	for _, entry := range macTable {
		macGaugeVec.WithLabelValues(entry.ONU, entry.MAC).Set(1)
	}

	return nil
}
