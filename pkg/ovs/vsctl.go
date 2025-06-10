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
	vSwitch := VirtualSwitch{ovsService: NewOvsService(), ipService: NewIpService(), bridge: plsv1.Bridge{Name: bridgeConf.bridge.Name}}

	if !bridgeConf.setFields[FieldName] || bridgeConf.bridge.Name == "" {
		return vSwitch, fmt.Errorf("bridge name must be set using WithName")
	}
	if !vSwitch.exists() {
		return vSwitch, fmt.Errorf("bridge with name: %s not found", bridgeConf.bridge.Name)
	}

	vSwitch.getPorts()

	vSwitch.getController()

	vSwitch.getVxlans()

	return vSwitch, nil
}

func UpdateVirtualSwitch(bridgeOptions ...func(*BridgeConf)) (VirtualSwitch, error) {
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
	existing, err := GetVirtualSwitch(WithName(bridgeConf.bridge.Name))
	if err != nil {
		// Bridge does not exist, fallback to creation
		return NewVirtualSwitch(bridgeOptions...)
	}

	ovs := existing.ovsService
	ip := existing.ipService
	name := bridgeConf.bridge.Name

	// Update only the explicitly provided fields

	if bridgeConf.setFields[FieldController] {
		if err := ovs.SetController(name, bridgeConf.bridge.Controller...); err != nil {
			return existing, fmt.Errorf("failed to update controller: %v", err)
		}
		existing.bridge.Controller = bridgeConf.bridge.Controller
	}

	if bridgeConf.setFields[FieldProtocol] {
		if err := ovs.SetProtocol(name, bridgeConf.bridge.Protocol); err != nil {
			return existing, fmt.Errorf("failed to update protocol: %v", err)
		}
		existing.bridge.Protocol = bridgeConf.bridge.Protocol
	}

	if bridgeConf.setFields[FieldDatapathId] {
		if err := ovs.SetDatapathID(name, bridgeConf.bridge.DatapathId); err != nil {
			return existing, fmt.Errorf("failed to update datapath ID: %v", err)
		}
		existing.bridge.DatapathId = bridgeConf.bridge.DatapathId
	}

	if bridgeConf.setFields[FieldPorts] {
		for port := range bridgeConf.bridge.Ports {
			if _, exists := existing.bridge.Ports[port]; !exists {
				_ = ip.SetInterfaceUp(port)
				if err := ovs.AddPort(name, port); err != nil {
					return existing, fmt.Errorf("failed to add port %s: %v", port, err)
				}
				existing.bridge.Ports[port] = bridgeConf.bridge.Ports[port]
			}
		}
	}

	// TODO: if vxlan is not anymore in the desired ones, delete it
	if bridgeConf.setFields[FieldVxlans] {
		currentVxlans, _ := ovs.GetVxlans(name)

		for vxID, vx := range bridgeConf.bridge.Vxlans {
			if _, exists := currentVxlans[vxID]; !exists {
				if err := ovs.CreateVxlan(name, vx); err != nil {
					return existing, fmt.Errorf("failed to create vxlan %s: %v", vxID, err)
				}
				if existing.bridge.Vxlans == nil {
					existing.bridge.Vxlans = make(map[string]plsv1.Vxlan)
				}
				existing.bridge.Vxlans[vxID] = vx
			} else {
				if err := ovs.ModifyVxlan(vx); err != nil {
					return existing, fmt.Errorf("failed to modify vxlan %s: %v", vxID, err)
				}
				if existing.bridge.Vxlans == nil {
					existing.bridge.Vxlans = make(map[string]plsv1.Vxlan)
				}
			}
		}
	}

	return existing, nil
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

	// Create the base switch object
	vSwitch := VirtualSwitch{
		ovsService: NewOvsService(),
		ipService:  NewIpService(),
		bridge:     plsv1.Bridge{Name: bridgeConf.bridge.Name},
	}

	// Validate name
	if !bridgeConf.setFields[FieldName] || bridgeConf.bridge.Name == "" {
		return vSwitch, fmt.Errorf("bridge name must be set using WithName")
	}

	// If bridge exists, delete it
	if vSwitch.exists() {
		err = vSwitch.ovsService.DeleteBridge(vSwitch.bridge.Name)
		if err != nil {
			return vSwitch, fmt.Errorf("could not delete existing bridge %s: %v", vSwitch.bridge.Name, err)
		}
	}

	// Create the bridge
	err = vSwitch.ovsService.AddBridge(vSwitch.bridge.Name)
	if err != nil {
		return vSwitch, fmt.Errorf("could not create bridge %s: %v", vSwitch.bridge.Name, err)
	}

	// Bring interface up
	err = vSwitch.ipService.SetInterfaceUp(vSwitch.bridge.Name)
	if err != nil {
		return vSwitch, fmt.Errorf("could not set %s interface up: %v", vSwitch.bridge.Name, err)
	}

	// Apply only explicitly set fields
	if bridgeConf.setFields[FieldDatapathId] {
		err = vSwitch.ovsService.SetDatapathID(vSwitch.bridge.Name, bridgeConf.bridge.DatapathId)
		if err != nil {
			return vSwitch, fmt.Errorf("could not set datapath ID: %v", err)
		}
		vSwitch.bridge.DatapathId = bridgeConf.bridge.DatapathId
	}

	if bridgeConf.setFields[FieldProtocol] {
		err = vSwitch.ovsService.SetProtocol(vSwitch.bridge.Name, bridgeConf.bridge.Protocol)
		if err != nil {
			return vSwitch, fmt.Errorf("could not set protocol: %v", err)
		}
		vSwitch.bridge.Protocol = bridgeConf.bridge.Protocol
	}

	if bridgeConf.setFields[FieldController] {
		err = vSwitch.ovsService.SetController(vSwitch.bridge.Name, bridgeConf.bridge.Controller...)
		if err != nil {
			return vSwitch, fmt.Errorf("could not set controller: %v", err)
		}
		vSwitch.bridge.Controller = bridgeConf.bridge.Controller
	}

	if bridgeConf.setFields[FieldPorts] {
		for name := range bridgeConf.bridge.Ports {
			err := vSwitch.addPort(name)
			if err != nil {
				return vSwitch, fmt.Errorf("could not add port %s: %v", name, err)
			}
		}
		vSwitch.bridge.Ports = bridgeConf.bridge.Ports
	}

	if bridgeConf.setFields[FieldVxlans] {
		for _, vx := range bridgeConf.bridge.Vxlans {
			err := vSwitch.createVxlan(vx)
			if err != nil {
				return vSwitch, fmt.Errorf("could not create vxlan %s: %v", vx.VxlanId, err)
			}
		}
		vSwitch.bridge.Vxlans = bridgeConf.bridge.Vxlans
	}

	return vSwitch, nil
}

func (vSwitch *VirtualSwitch) createVxlan(vxlan plsv1.Vxlan) error {

	err := vSwitch.ovsService.CreateVxlan(vSwitch.bridge.Name, vxlan)

	if err != nil {
		return fmt.Errorf("could not create vxlan from bridge %s to %s: %v", vSwitch.bridge.Name, vxlan.RemoteIp, err)
	}

	return nil

}

func (vSwitch *VirtualSwitch) addPort(portName string) error {

	err := vSwitch.ipService.SetInterfaceUp(portName)

	if err != nil {
		return fmt.Errorf("could not set interface %s up: %v", portName, err)
	}

	err = vSwitch.ovsService.AddPort(vSwitch.bridge.Name, portName)
	if err != nil {
		return fmt.Errorf("could not add interface %s as a port: %v", portName, err)
	}
	vSwitch.bridge.Ports[portName] = plsv1.Port{Name: portName, Status: "UP"}
	return nil
}

func (vSwitch *VirtualSwitch) getPorts() error {

	var err error
	vSwitch.bridge.Ports, err = vSwitch.ovsService.GetPorts(vSwitch.bridge.Name)

	return err
}

func (vSwitch *VirtualSwitch) getController() error {
	var err error

	vSwitch.bridge.Controller, err = vSwitch.ovsService.GetController(vSwitch.bridge.Name)

	return err
}

func (vSwitch *VirtualSwitch) getVxlans() error {

	return nil
}

func (vSwitch *VirtualSwitch) GetPortNumber(portName string) (int64, error) {
	ofport, err := vSwitch.ovsService.GetPortNumber(portName)

	if err != nil {
		return 0, fmt.Errorf("failed to parse port number: %v", err)
	}

	return ofport, nil
}

func (vSwitch *VirtualSwitch) exists() bool {
	err := vSwitch.ovsService.exec.Run("br-exists", vSwitch.bridge.Name)

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
