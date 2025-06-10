package ovs

import (
	"fmt"
)

type IpService struct {
	exec Client
}

func NewIpService() IpService {

	serviceClient := NewClient(IpClient)
	return IpService{exec: serviceClient}
}

func (ipService *IpService) SetInterfaceUp(interfaceName string) error {

	output, err := ipService.exec.CombinedOutput("link", "set", interfaceName, "up")
	if err != nil {
		return fmt.Errorf("command error: %v\nOutput: %s", err, output)
	}
	return nil

}

func (ipService *IpService) AddVethPair(vethName, peerName string) error {

	output, err := ipService.exec.CombinedOutput("link", "add", vethName, "type", "veth", "peer", "name", peerName)
	if err != nil {
		return fmt.Errorf("command error: %v\nOutput: %s", err, output)
	}
	return nil

}

func (ipService *IpService) AddInterfaceToLinuxBridge(interfaceName, bridgeName string) error {

	output, err := ipService.exec.CombinedOutput("link", "set", interfaceName, "master", bridgeName)
	if err != nil {
		return fmt.Errorf("command error: %v\nOutput: %s", err, output)
	}
	return nil

}
