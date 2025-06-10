package ovs

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	plsv1 "github.com/Networks-it-uc3m/l2sm-switch/api/v1"
)

type VirtualSwitch struct {
	bridge     plsv1.Bridge
	ovsService OvsService
	ipService  IpService
}

func GetVirtualSwitch(options ...func(*BridgeConf)) (*VirtualSwitch, error) {

	vSwitch := VirtualSwitch{bridge: bridgeConf, ovsService: NewOvsService(), ipService: NewIpService()}

	if bridgeConf.Name == "" {
		return VirtualSwitch{}, fmt.Errorf("no bridge name provided. Please specify a name to correctly get the bridge")
	}

	if !vSwitch.exists() {
		return VirtualSwitch{}, fmt.Errorf("bridge with name: %s not found", bridgeConf.Name)
	}

	vSwitch.getPorts()

	vSwitch.getController()

	vSwitch.getVxlans()

	return vSwitch, nil
}

func UpdateVirtualSwitch(desired plsv1.Bridge) (*VirtualSwitch, error) {
	existingSwitch, err := GetVirtualSwitch(desired)
	if err != nil {
		// Bridge doesn't exist: create from scratch
		return NewVirtualSwitch(desired)
	}

	ovs := existingSwitch.ovsService
	ip := existingSwitch.ipService
	name := desired.Name

	// Update controller if changed
	if !equalStringSlices(existingSwitch.bridge.Controller, desired.Controller) {
		err := ovs.SetController(name, desired.Controller...)
		if err != nil {
			return existingSwitch, fmt.Errorf("failed to update controller: %v", err)
		}
		existingSwitch.bridge.Controller = desired.Controller
	}

	// Update protocol if needed
	if existingSwitch.bridge.Protocol != desired.Protocol {
		err := ovs.SetProtocol(name, desired.Protocol)
		if err != nil {
			return existingSwitch, fmt.Errorf("failed to update protocol: %v", err)
		}
		existingSwitch.bridge.Protocol = desired.Protocol
	}

	// Update datapath ID if changed
	if desired.DatapathId != "" && desired.DatapathId != existingSwitch.bridge.DatapathId {
		err := ovs.SetDatapathID(name, desired.DatapathId)
		if err != nil {
			return existingSwitch, fmt.Errorf("failed to update datapath ID: %v", err)
		}
		existingSwitch.bridge.DatapathId = desired.DatapathId
	}

	// Update Ports
	for portName := range desired.Ports {
		if _, exists := existingSwitch.bridge.Ports[portName]; !exists {
			_ = ip.SetInterfaceUp(portName)
			err := ovs.AddPort(name, portName)
			if err != nil {
				return existingSwitch, fmt.Errorf("failed to add port %s: %v", portName, err)
			}
		}
	}

	for portName := range existingSwitch.bridge.Ports {
		if _, desiredExists := desired.Ports[portName]; !desiredExists {
			// No delete method yet: consider implementing ovsService.DeletePort()
			// _ = ovs.DeletePort(name, portName)
		}
	}

	// Update VxLANs
	currentVxlans, _ := ovs.GetVxlans(name)
	currentMap := map[string]plsv1.Vxlan{}
	for _, v := range currentVxlans {
		currentMap[v.VxlanId] = v
	}

	for _, vx := range desired.Vxlans {
		if _, exists := currentMap[vx.VxlanId]; !exists {
			err := ovs.CreateVxlan(name, vx)
			if err != nil {
				return existingSwitch, fmt.Errorf("failed to create vxlan %s: %v", vx.VxlanId, err)
			}
		}
	}

	return existingSwitch, nil
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	aMap := map[string]bool{}
	for _, val := range a {
		aMap[val] = true
	}
	for _, val := range b {
		if !aMap[val] {
			return false
		}
	}
	return true
}

func NewVirtualSwitch(bridgeOptions ...func(*BridgeConf)) (VirtualSwitch, error) {

	var err error
	bridgeConf := &BridgeConf{}

	vSwitch := VirtualSwitch{ovsService: NewOvsService(), ipService: NewIpService()}
	for _, o := range bridgeOptions {
		o(bridgeConf)
	}

	for bridgeConf.setFields
	if vSwitch.exists() {
		err = vSwitch.ovsService.DeleteBridge(bridgeConf.Name)
		if err != nil {
			return VirtualSwitch{}, fmt.Errorf("could not create already existing bridge %s: %v", bridgeConf.Name, err)
		}
	}

	err = vSwitch.ovsService.AddBridge(bridgeConf.Name)
	if err != nil {
		return VirtualSwitch{}, fmt.Errorf("could not create %s interface: %v", bridgeConf.Name, err)
	}
	vSwitch.bridge.Name = bridgeConf.Name

	err = vSwitch.ipService.SetInterfaceUp(vSwitch.bridge.Name)
	if err != nil {
		return VirtualSwitch{}, fmt.Errorf("could not set %s interface up: %v", bridgeConf.Name, err)
	}

	if bridgeConf.DatapathId != "" {
		err = vSwitch.ovsService.SetDatapathID(vSwitch.bridge.Name, bridgeConf.DatapathId)
		if err != nil {
			return VirtualSwitch{}, fmt.Errorf("could not set custom datapath id: %v", err)
		}
	}

	err = vSwitch.ovsService.SetProtocol(vSwitch.bridge.Name, bridgeConf.Protocol)
	if err != nil {
		return VirtualSwitch{}, fmt.Errorf("could not set %s messaging protocol to OpenFlow13: %v", bridgeConf.Name, err)
	}
	vSwitch.bridge.Protocol = bridgeConf.Protocol // TODO: Check

	err = vSwitch.ovsService.SetController(vSwitch.bridge.Name, bridgeConf.Controller...)

	if err != nil {
		return VirtualSwitch{}, fmt.Errorf("could not connect to controller: %v", err)

	}

	vSwitch.bridge.Controller = bridgeConf.Controller

	return vSwitch, nil
}

func (vSwitch *VirtualSwitch) CreateVxlan(vxlan plsv1.Vxlan) error {

	err := vSwitch.ovsService.CreateVxlan(vSwitch.bridge.Name, vxlan)

	if err != nil {
		return fmt.Errorf("could not create vxlan from bridge %s to %s: %v", vSwitch.bridge.Name, vxlan.RemoteIp, err)
	}

	return nil

}

func (vSwitch *VirtualSwitch) AddPort(portName string) error {

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
	cmd := exec.Command("ovs-vsctl", "get", "Interface", portName, "ofport")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("failed to get port number for %s: %v", portName, err)
	}

	ofportStr := strings.TrimSpace(out.String())
	ofport, err := strconv.ParseInt(ofportStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse port number: %v", err)
	}

	return ofport, nil
}

func (vSwitch *VirtualSwitch) exists() bool {
	err := vSwitch.ovsService.exec.Run("br-exists", vSwitch.bridge.Name)

	return err == nil
}
