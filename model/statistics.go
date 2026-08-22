package model

type Statistics struct {
	Device     Device
	Interfaces []StatisticsInterface
}

type Device struct {
	CPU          []CPU
	FanSpeeds    []FanSpeed
	Power        []PSU
	RAM          RAM
	Temperatures []Temperature
	Uptime       float64
}

type CPU struct {
	Identifier string
	Usage      int
}

type FanSpeed struct {
	Value float64
}

type PSU struct {
	Connected bool
	Power     *float64
	Voltage   *float64
}

type RAM struct {
	Free  float64
	Total float64
}

type Temperature struct {
	Value float64
}

type StatisticsInterface struct {
	ID         string
	Name       string
	Statistics InterfaceStatistics
}

type InterfaceStatistics struct {
	RxBytes   *float64
	RxPackets *float64
	TxBytes   *float64
	TxPackets *float64
	SFP       *SfpStatistics
}

type SfpStatistics struct {
	RxPower     *float64
	Temperature *float64
	TxPower     *float64
}
