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
	Identifier  string
	Temperature float64
	Usage       int
}

type FanSpeed struct {
	Value float64
}

type PSU struct {
	Connected bool
	Current   *float64
	Power     *float64
	PsuType   string
	Voltage   *float64
}

type RAM struct {
	Free  float64
	Total float64
	Usage float64
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
	RxBroadcast *float64
	RxBytes     *float64
	RxMulticast *float64
	RxPackets   *float64
	RxRate      *float64
	TxBroadcast *float64
	TxBytes     *float64
	TxMulticast *float64
	TxPackets   *float64
	TxRate      *float64
	SFP         *SfpStatistics
}

type SfpStatistics struct {
	Current     *float64
	RxPower     *float64
	Temperature *float64
	TxPower     *float64
	Voltage     *float64
}
