package v1

const (
	DEFAULT_CONFIG_PATH = "/etc/l2sm"
)

type NedSettings struct {
	ConfigDir    string
	ControllerIP string
	NodeName     string
	NedName      string
}

type OverlaySettings struct {
	ControllerIp     string
	InterfacesNumber int
	NodeName         string
}
