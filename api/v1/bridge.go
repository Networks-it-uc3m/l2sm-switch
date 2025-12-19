package v1

type Port struct {
	Name     string
	Status   string
	Id       *int
	Internal bool
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
