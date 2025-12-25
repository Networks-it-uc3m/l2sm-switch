package v1

import "net"

type Port struct {
	Name      string
	Status    string
	Id        *int
	Internal  bool
	IpAddress *net.IPNet
}

type Bridge struct {
	Controller []string
	Name       string
	Protocol   string
	DatapathId string
	Ports      map[string]Port
	Vxlans     map[string]Vxlan
}

type Vxlan struct {
	VxlanId  string
	LocalIp  string
	RemoteIp string
	UdpPort  string
}
