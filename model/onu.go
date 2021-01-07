package model

type ONU struct {
	// only these fields are always returned from the API
	Connected bool
	Serial    string

	// connected
	Authorized      *bool
	ConnectionTime  *float64
	Distance        *float64
	Error           string
	FirmwareHash    string
	FirmwareVersion string
	LaserBias       *float64
	MAC             string
	OLTPort         *float64
	Ports           *[]ONUPort
	PortsStat       *[]ONUStatistics
	RxPower         *float64
	TxPower         *float64
	Statistics      *ONUStatistics
	System          *ONUSystem
	UpgradeStatus   *ONUUpgradeStatus

	// disconnected
	DyingGasp string
}

type ONUPort struct {
	ID      string
	Plugged bool
	Speed   string
}

type ONUStatistics struct {
	RxBytes float64
	RxRate  float64
	TxBytes float64
	TxRate  float64
}

type ONUSystem struct {
	CPU         float64
	Mem         float64
	Temperature map[string]float64
	Uptime      float64
	Voltage     float64
}

type ONUUpgradeStatus struct {
	FailureReason string
	Status        string
}
