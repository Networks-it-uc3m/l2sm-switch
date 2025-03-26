package v1

const (
	DEFAULT_CONFIG_PATH = "/etc/l2sm"
)

type NedSettings struct {
	ConfigDir      string
	ControllerIP   []string
	ControllerPort string
	NodeName       string
	NedName        string
}

type OverlaySettings struct {
	ControllerIp     []string
	ControllerPort   string
	InterfacesNumber int
	OverlayName      string
}
