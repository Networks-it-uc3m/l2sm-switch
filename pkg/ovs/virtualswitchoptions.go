package ovs

import plsv1 "github.com/Networks-it-uc3m/l2sm-switch/api/v1"

// ConfigurableField is the settable field in BridgeConf
type ConfigurableField string

const (
	FieldController ConfigurableField = "controller"
	FieldName       ConfigurableField = "name"
	FieldProtocol   ConfigurableField = "protocol"
	FieldDatapathId ConfigurableField = "datapathid"
	FieldPorts      ConfigurableField = "ports"
	FieldVxlans     ConfigurableField = "vxlans"
)

type BridgeConf struct {
	bridge    plsv1.Bridge
	setFields map[ConfigurableField]bool
}

func WithController(controller []string) func(*BridgeConf) {
	return func(v *BridgeConf) {
		v.bridge.Controller = controller
		v.setFields[FieldController] = true
	}
}

func WithName(name string) func(*BridgeConf) {
	return func(v *BridgeConf) {
		v.bridge.Name = name
		v.setFields[FieldName] = true
	}
}

func WithProtocol(protocol string) func(*BridgeConf) {
	return func(v *BridgeConf) {
		v.bridge.Protocol = protocol
		v.setFields[FieldProtocol] = true
	}
}

func WithDatapathId(datapathid string) func(*BridgeConf) {
	return func(v *BridgeConf) {
		v.bridge.DatapathId = datapathid
		v.setFields[FieldDatapathId] = true
	}
}

func WithPorts(ports []plsv1.Port) func(*BridgeConf) {
	return func(v *BridgeConf) {
		portMap := make(map[string]plsv1.Port)
		for _, p := range ports {
			if p.Name != "" {
				portMap[p.Name] = p
			}
		}
		v.bridge.Ports = portMap
		v.setFields[FieldPorts] = true
	}
}

func WithVxlans(vxlans []plsv1.Vxlan) func(*BridgeConf) {
	return func(v *BridgeConf) {
		vxlanMap := make(map[string]plsv1.Vxlan)
		for _, vx := range vxlans {
			if vx.VxlanId != "" {
				vxlanMap[vx.VxlanId] = vx
			}
		}
		v.bridge.Vxlans = vxlanMap
		v.setFields[FieldVxlans] = true
	}
}
