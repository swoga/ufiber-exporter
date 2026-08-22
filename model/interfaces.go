package model

type InterfacesInterface struct {
	Identification Identification
	PON            *PON
	Port           *Port
	Status         Status
}

type Identification struct {
	ID string
}

type PON struct {
	SFP SfpModule
}

type SfpModule struct {
	Present bool
}

type Status struct {
	Enabled bool
	Plugged bool
}

type Port struct {
	SFP SfpModule
}
