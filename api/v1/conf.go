package v1

const (
	DEFAULT_CONFIG_PATH = "/etc/l2sm"
	SETTINGS_FILE       = "config.json"
	TOPOLOGY_FILE       = "topology.json"
	NEIGHBOR_FILE       = "neighbors.json"
	DEFAULT_VXLAN_PORT  = "7000"
)

type Settings struct {
	ControllerIP     []string `json:"controllerIp"`
	ControllerPort   string   `json:"controllerPort"`
	NodeName         string   `json:"nodeName,omitempty"`
	SwitchName       string   `json:"switchName"`
	InterfacesNumber int      `json:"interfacesNumber,omitempty"`
}
