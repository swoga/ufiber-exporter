package model

type InterfacesInterface struct {
	Identification Identification
	PON            *PON
	Port           *Port
	LAG            *LAG
	Status         Status
}

type Identification struct {
	ID   string
	MAC  string
	Name string
	Type string
}

type PON struct {
	SFP SfpModule
}

type SfpModule struct {
	LoS     *bool
	Part    string
	Present bool
	Serial  *string
	TxFault *string
	Vendor  *string
}

type Status struct {
	ArpProxy     bool
	CurrentSpeed string
	Enabled      bool
	MTU          float64
	Plugged      bool
	Speed        string
}

type Port struct {
	SFP SfpModule
}

type LAG struct {
	Interfaces  []Identification
	LoadBalance string
	Static      bool
}
