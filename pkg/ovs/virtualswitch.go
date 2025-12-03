package ovs

import (
	"fmt"
	"time"

	plsv1 "github.com/Networks-it-uc3m/l2sm-switch/api/v1"
	"github.com/Networks-it-uc3m/l2sm-switch/pkg/utils"
)

type VirtualSwitch struct {
	bridge     plsv1.Bridge
	ovsService OvsService
	ipService  IpService
}

func GetVirtualSwitch(bridgeOptions ...func(*BridgeConf)) (VirtualSwitch, error) {

	bridgeConf := &BridgeConf{
		setFields: make(map[ConfigurableField]bool),
	}

	// Apply options to populate bridgeConf
	for _, opt := range bridgeOptions {
		opt(bridgeConf)

	}

	ovsService := NewOvsService()

	if bridgeConf.setFields[FieldSudo] {
		ovsService = NewSudoOvsService()
	}
	ipService := NewIpService()

	if bridgeConf.setFields[FieldSudo] {
		ipService = NewSudoIpService()
	}
	vs := VirtualSwitch{ovsService: ovsService, ipService: ipService, bridge: plsv1.Bridge{Name: bridgeConf.bridge.Name}}

	if !bridgeConf.setFields[FieldName] || bridgeConf.bridge.Name == "" {
		return vs, fmt.Errorf("bridge name must be set using WithName")
	}
	if !vs.exists() {
		return vs, fmt.Errorf("bridge with name: %s not found", bridgeConf.bridge.Name)
	}

	vs.getPorts()

	vs.getController()

	vs.getVxlans()

	return vs, nil
}

func UpdateVirtualSwitch(bridgeOptions ...func(*BridgeConf)) (VirtualSwitch, error) {
	var err error
	bridgeConf := &BridgeConf{
		setFields: make(map[ConfigurableField]bool),
	}

	// Apply options to populate bridgeConf
	for _, opt := range bridgeOptions {
		opt(bridgeConf)
	}

	if !bridgeConf.setFields[FieldName] || bridgeConf.bridge.Name == "" {
		return VirtualSwitch{}, fmt.Errorf("bridge name must be set using WithName")
	}

	// Attempt to retrieve the existing bridge
	vs, err := GetVirtualSwitch(WithName(bridgeConf.bridge.Name), WithSudo(bridgeConf.setFields[FieldSudo]))
	if err != nil {
		// Bridge does not exist, fallback to creation
		fmt.Println("si que entra")
		return NewVirtualSwitch(bridgeOptions...)
	}

	ovs := vs.ovsService
	ip := vs.ipService
	name := bridgeConf.bridge.Name

	// Update only the explicitly provided fields

	if bridgeConf.setFields[FieldController] {
		if err := ovs.SetController(name, bridgeConf.bridge.Controller...); err != nil {
			return vs, fmt.Errorf("failed to update controller: %v", err)
		}
		vs.bridge.Controller = bridgeConf.bridge.Controller
	}

	if bridgeConf.setFields[FieldProtocol] {
		if err := ovs.SetProtocol(name, bridgeConf.bridge.Protocol); err != nil {
			return vs, fmt.Errorf("failed to update protocol: %v", err)
		}
		vs.bridge.Protocol = bridgeConf.bridge.Protocol
	}

	if bridgeConf.setFields[FieldDatapathId] {
		if err := ovs.SetDatapathID(name, bridgeConf.bridge.DatapathId); err != nil {
			return vs, fmt.Errorf("failed to update datapath ID: %v", err)
		}
		vs.bridge.DatapathId = bridgeConf.bridge.DatapathId
	}

	if bridgeConf.setFields[FieldPorts] {
		for id, port := range bridgeConf.bridge.Ports {
			if _, exists := vs.bridge.Ports[id]; !exists {
				i := NO_DEFAULT_ID
				if err = ip.SetInterfaceUp(port.Name); err != nil {
					return vs, fmt.Errorf("failed to add port %s: %v", port.Name, err)
				}
				if port.Id != nil {
					i = *port.Id
				}
				if err = ovs.AddPort(name, port.Name, i); err != nil {
					return vs, fmt.Errorf("failed to add port %s: %v", port.Name, err)
				}
				vs.bridge.Ports[id] = bridgeConf.bridge.Ports[id]
			}
		}
	}

	if bridgeConf.setFields[FieldVxlans] {
		vxs, err := ovs.GetVxlans(name)
		if err != nil {
			return vs, fmt.Errorf("failed to get vxlans: %w", err)
		}
		requiredVxlans := bridgeConf.bridge.Vxlans

		for vxID, vx := range requiredVxlans {
			if _, ok := vxs[vxID]; !ok {
				if err = ovs.CreateVxlan(name, vx); err != nil {
					return vs, fmt.Errorf("failed to create vxlan %s: %v", vxID, err)
				}

			} else {
				delete(vxs, vxID)
			}
		}
		for vxID := range vxs {
			if err = ovs.DeleteVxlan(name, vxID); err != nil {
				return vs, fmt.Errorf("failed to delete vxlan %s: %v", vxID, err)
			}
		}

	}

	return vs, nil
}

func NewVirtualSwitch(bridgeOptions ...func(*BridgeConf)) (VirtualSwitch, error) {
	var err error
	bridgeConf := &BridgeConf{
		setFields: make(map[ConfigurableField]bool),
	}

	// Apply functional options
	for _, opt := range bridgeOptions {
		opt(bridgeConf)
	}

	ovsService := NewOvsService()

	if bridgeConf.setFields[FieldSudo] {
		ovsService = NewSudoOvsService()
	}
	ipService := NewIpService()

	if bridgeConf.setFields[FieldSudo] {
		ipService = NewSudoIpService()
	}
	vs := VirtualSwitch{ovsService: ovsService, ipService: ipService,
		bridge: plsv1.Bridge{Name: bridgeConf.bridge.Name},
	}

	// Validate name
	if !bridgeConf.setFields[FieldName] || bridgeConf.bridge.Name == "" {
		return vs, fmt.Errorf("bridge name must be set using WithName")
	}

	// If bridge exists, delete it
	if vs.exists() {
		err = vs.ovsService.DeleteBridge(vs.bridge.Name)
		if err != nil {
			return vs, fmt.Errorf("could not delete existing bridge %s: %v", vs.bridge.Name, err)
		}
	}

	ovs := vs.ovsService
	ip := vs.ipService
	name := bridgeConf.bridge.Name
	// Create the bridge
	err = ovs.AddBridge(vs.bridge.Name)
	if err != nil {
		return vs, fmt.Errorf("could not create bridge %s: %v", vs.bridge.Name, err)
	}

	// Bring interface up
	err = vs.ipService.SetInterfaceUp(vs.bridge.Name)
	if err != nil {
		return vs, fmt.Errorf("could not set %s interface up: %v", vs.bridge.Name, err)
	}

	// Apply only explicitly set fields
	if bridgeConf.setFields[FieldDatapathId] {
		err = ovs.SetDatapathID(vs.bridge.Name, bridgeConf.bridge.DatapathId)
		if err != nil {
			return vs, fmt.Errorf("could not set datapath ID: %v", err)
		}
		vs.bridge.DatapathId = bridgeConf.bridge.DatapathId
	}

	if bridgeConf.setFields[FieldProtocol] {
		err = ovs.SetProtocol(vs.bridge.Name, bridgeConf.bridge.Protocol)
		if err != nil {
			return vs, fmt.Errorf("could not set protocol: %v", err)
		}
		vs.bridge.Protocol = bridgeConf.bridge.Protocol
	}

	if bridgeConf.setFields[FieldController] {
		err = ovs.SetController(vs.bridge.Name, bridgeConf.bridge.Controller...)
		if err != nil {
			return vs, fmt.Errorf("could not set controller: %v", err)
		}
		vs.bridge.Controller = bridgeConf.bridge.Controller
	}

	// TODO: interfaces exist in the void? Create new ones? Specific for NED
	if bridgeConf.setFields[FieldPorts] {
		for _, port := range bridgeConf.bridge.Ports {
			i := NO_DEFAULT_ID
			if err = ip.SetInterfaceUp(port.Name); err != nil {
				return vs, fmt.Errorf("failed to add port %s: %v", port.Name, err)
			}
			if port.Id != nil {
				i = *port.Id
			}
			if err = ovs.AddPort(name, port.Name, i); err != nil {
				return vs, fmt.Errorf("failed to add port %s: %v", port.Name, err)
			}

		}
		vs.bridge.Ports = bridgeConf.bridge.Ports

	}
	if bridgeConf.setFields[FieldVxlans] {
		for _, vx := range bridgeConf.bridge.Vxlans {
			err := vs.createVxlan(vx)
			if err != nil {
				return vs, fmt.Errorf("could not create vxlan %s: %v", vx.VxlanId, err)
			}
		}
		vs.bridge.Vxlans = bridgeConf.bridge.Vxlans
	}

	return vs, nil
}

func (vs *VirtualSwitch) createVxlan(vxlan plsv1.Vxlan) error {

	err := vs.ovsService.CreateVxlan(vs.bridge.Name, vxlan)

	if err != nil {
		return fmt.Errorf("could not create vxlan from bridge %s to %s: %v", vs.bridge.Name, vxlan.RemoteIp, err)
	}

	return nil

}

// func (vs *VirtualSwitch) addPort(portName string) error {

// 	err := vs.ipService.SetInterfaceUp(portName)

// 	if err != nil {
// 		return fmt.Errorf("could not set interface %s up: %v", portName, err)
// 	}

// 	err = vs.ovsService.AddPort(vs.bridge.Name, portName, NO_DEFAULT_ID)
// 	if err != nil {
// 		return fmt.Errorf("could not add interface %s as a port: %v", portName, err)
// 	}
// 	vs.bridge.Ports[portName] = plsv1.Port{Name: portName, Status: "UP"}
// 	return nil
// }

func (vs *VirtualSwitch) getPorts() error {

	var err error
	vs.bridge.Ports, err = vs.ovsService.GetPorts(vs.bridge.Name)

	return err
}

func (vs *VirtualSwitch) getController() error {
	var err error

	vs.bridge.Controller, err = vs.ovsService.GetController(vs.bridge.Name)

	return err
}

func (vs *VirtualSwitch) getVxlans() error {

	return nil
}

func (vs *VirtualSwitch) GetPortNumber(portName string) (int64, error) {
	ofport, err := vs.ovsService.GetPortNumber(portName)

	if err != nil {
		return 0, fmt.Errorf("failed to parse port number: %v", err)
	}

	return ofport, nil
}

func (vs *VirtualSwitch) exists() bool {
	err := vs.ovsService.exec.Run("br-exists", vs.bridge.Name)

	return err == nil
}

// AddInterfaceToBridge creates a new veth pair, attaches one end to the specified bridge,
// and returns the name of the other end.
func AddInterfaceToBridge(bridgeName string) (string, error) {
	var err error
	ipService := NewIpService()
	// Generate unique interface names
	timestamp := time.Now().UnixNano()
	vethName, _ := utils.GenerateInterfaceName("veth", fmt.Sprintf("%s%d", bridgeName, timestamp))
	peerName, _ := utils.GenerateInterfaceName("vpeer", fmt.Sprintf("%s%d", bridgeName, timestamp))

	// Create the veth pair
	err = ipService.AddVethPair(vethName, peerName)

	if err != nil {
		return "", fmt.Errorf("failed to create veth pair: %v", err)
	}
	// Set both interfaces up
	err = ipService.SetInterfaceUp(vethName)
	if err != nil {
		return "", fmt.Errorf("failed to set %s up: %v", vethName, err)
	}

	err = ipService.SetInterfaceUp(peerName)
	if err != nil {
		return "", fmt.Errorf("failed to set %s up: %v", peerName, err)
	}

	err = ipService.AddInterfaceToLinuxBridge(peerName, bridgeName)

	if err != nil {
		return "", fmt.Errorf("failed to add %s to bridge %s: %v", peerName, bridgeName, err)
	}

	return vethName, nil
}
