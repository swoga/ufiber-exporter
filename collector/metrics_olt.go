package collector

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/swoga/ufiber-exporter/model"
)

func AddMetricsOlt(registry prometheus.Registerer, statistics model.Statistics, interfacesInterfaces []model.InterfacesInterface) error {
	err := addMetricsOltDevice(registry, statistics.Device)
	if err != nil {
		return err
	}
	err = addMetricsOltInterfaces(prometheus.WrapRegistererWithPrefix("interface_", registry), statistics.Interfaces, interfacesInterfaces)
	if err != nil {
		return err
	}
	return nil
}

func addMetricsOltDevice(registry prometheus.Registerer, device model.Device) error {
	// CPU
	cpuGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cpu_usage",
	}, []string{"cpu"})
	registry.MustRegister(cpuGaugeVec)

	for _, cpu := range device.CPU {
		if cpu.Identifier == "cpu" {
			continue
		}
		cpuGaugeVec.WithLabelValues(cpu.Identifier).Set(float64(cpu.Usage))
	}

	// FANs
	fanSpeedGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "fan_speed",
	}, []string{"fan"})
	registry.MustRegister(fanSpeedGaugeVec)

	for i, fanSpeed := range device.FanSpeeds {
		fanSpeedGaugeVec.WithLabelValues(strconv.Itoa(i)).Set(fanSpeed.Value)
	}

	// PSUs
	psuConnectedGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "psu_connected",
	}, []string{"psu"})
	registry.MustRegister(psuConnectedGaugeVec)
	psuVoltageGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "psu_voltage",
	})
	registry.MustRegister(psuVoltageGauge)
	psuPowerGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "psu_power",
	})
	registry.MustRegister(psuPowerGauge)

	for i, psu := range device.Power {
		var connected float64
		if psu.Connected {
			connected = 1
		}
		psuConnectedGaugeVec.WithLabelValues(strconv.Itoa(i)).Set(connected)

		// voltage and power is only reported by one PSU
		if psu.Voltage != nil {
			psuVoltageGauge.Set(*psu.Voltage)
		}
		if psu.Power != nil {
			psuPowerGauge.Set(*psu.Power)
		}
	}

	// RAM
	ramTotalGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ram_total",
	})
	registry.MustRegister(ramTotalGauge)
	ramFreeGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ram_free",
	})
	registry.MustRegister(ramFreeGauge)

	ramTotalGauge.Set(device.RAM.Total)
	ramFreeGauge.Set(device.RAM.Free)

	// Temperatures
	temperatureGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "temperature",
	}, []string{"sensor"})
	registry.MustRegister(temperatureGaugeVec)

	for i, temperature := range device.Temperatures {
		temperatureGaugeVec.WithLabelValues(strconv.Itoa(i)).Set(temperature.Value)
	}

	// Uptime
	uptimeCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "uptime",
	})
	registry.MustRegister(uptimeCounter)

	uptimeCounter.Add(device.Uptime)
	return nil
}

func addMetricsOltInterfaces(registry prometheus.Registerer, statisticsInterfaces []model.StatisticsInterface, interfacesInterfaces []model.InterfacesInterface) error {
	interfaceRxBytesCounterVec := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "rx_bytes",
	}, []string{"name"})
	registry.MustRegister(interfaceRxBytesCounterVec)
	interfaceRxPacketsCounterVec := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "rx_packets",
	}, []string{"name"})
	registry.MustRegister(interfaceRxPacketsCounterVec)
	interfaceTxBytesCounterVec := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tx_bytes",
	}, []string{"name"})
	registry.MustRegister(interfaceTxBytesCounterVec)
	interfaceTxPacketsCounterVec := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tx_packets",
	}, []string{"name"})
	registry.MustRegister(interfaceTxPacketsCounterVec)

	interfaceRxPowerGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "rx_power",
	}, []string{"name"})
	registry.MustRegister(interfaceRxPowerGaugeVec)
	interfaceSfpTemperatureGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sfp_temperature",
	}, []string{"name"})
	registry.MustRegister(interfaceSfpTemperatureGaugeVec)

	interfaceNameGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "name",
	}, []string{"name", "given_name"})
	registry.MustRegister(interfaceNameGaugeVec)

	for _, interf := range statisticsInterfaces {
		if interf.Statistics.RxBytes != nil {
			interfaceRxBytesCounterVec.WithLabelValues(interf.ID).Add(*interf.Statistics.RxBytes)
		}
		if interf.Statistics.RxPackets != nil {
			interfaceRxPacketsCounterVec.WithLabelValues(interf.ID).Add(*interf.Statistics.RxPackets)
		}
		if interf.Statistics.TxBytes != nil {
			interfaceTxBytesCounterVec.WithLabelValues(interf.ID).Add(*interf.Statistics.TxBytes)
		}
		if interf.Statistics.TxPackets != nil {
			interfaceTxPacketsCounterVec.WithLabelValues(interf.ID).Add(*interf.Statistics.TxPackets)
		}

		if interf.Statistics.SFP != nil {
			if interf.Statistics.SFP.RxPower != nil {
				interfaceRxPowerGaugeVec.WithLabelValues(interf.ID).Set(*interf.Statistics.SFP.RxPower)
			}
			if interf.Statistics.SFP.Temperature != nil {
				interfaceSfpTemperatureGaugeVec.WithLabelValues(interf.ID).Set(*interf.Statistics.SFP.Temperature)
			}
		}

		name := interf.Name
		if name == "" {
			name = interf.ID
		}
		interfaceNameGaugeVec.WithLabelValues(interf.ID, name).Set(1)
	}

	interfaceStatusEnabledGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "enabled",
	}, []string{"name"})
	registry.MustRegister(interfaceStatusEnabledGaugeVec)
	interfaceStatusPluggedGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "plugged",
	}, []string{"name"})
	registry.MustRegister(interfaceStatusPluggedGaugeVec)
	interfaceSfpPresentGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sfp_present",
	}, []string{"name"})
	registry.MustRegister(interfaceSfpPresentGaugeVec)

	for _, interf := range interfacesInterfaces {
		var enabled float64 = 0
		if interf.Status.Enabled {
			enabled = 1
		}
		interfaceStatusEnabledGaugeVec.WithLabelValues(interf.Identification.ID).Set(enabled)

		var plugged float64 = 0
		if interf.Status.Plugged {
			plugged = 1
		}
		interfaceStatusPluggedGaugeVec.WithLabelValues(interf.Identification.ID).Set(plugged)

		if interf.Port != nil {
			var present float64 = 0
			if interf.Port.SFP.Present {
				present = 1
			}
			interfaceSfpPresentGaugeVec.WithLabelValues(interf.Identification.ID).Set(present)
		}

		if interf.PON != nil {
			var present float64 = 0
			if interf.PON.SFP.Present {
				present = 1
			}
			interfaceSfpPresentGaugeVec.WithLabelValues(interf.Identification.ID).Set(present)
		}
	}
	return nil
}
