package ovs

import (
	"fmt"
)

type IpService struct {
	exec Client
}

func NewIpService() IpService {
	return IpService{exec: NewClient(IpClient)}
}

func NewSudoIpService() IpService {
	return IpService{exec: NewSudoClient(IpClient)}
}

func (ipService *IpService) SetInterfaceUp(interfaceName string) error {

	o, err := ipService.exec.CombinedOutput("link", "set", interfaceName, "up")
	if err != nil {
		return fmt.Errorf("command error: %v\nOutput: %s", err, o)
	}
	return nil

}

func (ipService *IpService) AddVethPair(vethName, peerName string) error {

	o, err := ipService.exec.CombinedOutput("link", "add", vethName, "type", "veth", "peer", "name", peerName)
	if err != nil {
		return fmt.Errorf("command error: %v\nOutput: %s", err, o)
	}
	return nil

}

func (ipService *IpService) AddInterfaceToLinuxBridge(interfaceName, bridgeName string) error {

	o, err := ipService.exec.CombinedOutput("link", "set", interfaceName, "master", bridgeName)
	if err != nil {
		return fmt.Errorf("command error: %v\nOutput: %s", err, o)
	}
	return nil

}
