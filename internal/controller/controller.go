package controller

import (
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"time"

	plsv1 "github.com/Networks-it-uc3m/l2sm-switch/api/v1"
	"github.com/Networks-it-uc3m/l2sm-switch/pkg/datapath"
	"github.com/Networks-it-uc3m/l2sm-switch/pkg/ovs"
	"github.com/Networks-it-uc3m/l2sm-switch/pkg/utils"
)

type Controller struct {
	switchName string
	nodeName   string
	sudo       bool
}

func (ctr *Controller) GetNodeName() string {
	return ctr.nodeName
}

func NewSwitchManager(switchName, nodeName string, sudo bool) *Controller {

	return &Controller{switchName, nodeName, sudo}
}

func (ctr *Controller) ConfigureSwitch(controllerPort string, controllerIPs []string) (ovs.VirtualSwitch, error) {

	re := regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)

	controllers := []string{}

	for _, controllerIP := range controllerIPs {

		if !re.MatchString(controllerIP) {

			out, _ := exec.Command("host", controllerIP).Output()

			controllerIP = re.FindString(string(out))

		}
		cs := fmt.Sprintf("tcp:%s:%s", controllerIP, controllerPort)
		controllers = append(controllers, cs)
	}

	datapathId := datapath.GenerateID(ctr.switchName)

	var err error
	var vs ovs.VirtualSwitch

	_, err = ctr.getOvs()

	if err != nil {
		fmt.Println("Switch doesn't exist. Creating a new one.")
		vs, err = ctr.newOvs(
			ovs.WithController(controllers),
			ovs.WithProtocol("OpenFlow13"),
			ovs.WithDatapathId(datapathId),
		)

		return vs, err
	}

	vs, err = ctr.updateOvs(
		ovs.WithController(controllers),
		ovs.WithProtocol("OpenFlow13"),
		ovs.WithDatapathId(datapathId),
	)

	return vs, err
}

/*
*
Example:

	        {
	            "Name": "l2sm1",
	            "nodeIP": "10.1.14.53",
				"neighborNodes":["10.4.2.3","10.4.2.5"]
			}
*/
func (ctr *Controller) ConnectToNeighbors(node plsv1.Node) error {
	vxs := make([]plsv1.Vxlan, len(node.NeighborNodes))

	for _, neighIP := range node.NeighborNodes {
		vxID, err := utils.GenerateInterfaceName("vxlan-", fmt.Sprintf("%s%s", node.NodeIP, neighIP))
		if err != nil {
			return fmt.Errorf("error generating vxlan id: %v", err)
		}
		vxs = append(vxs, plsv1.Vxlan{VxlanId: vxID, LocalIp: node.NodeIP, RemoteIp: neighIP, UdpPort: plsv1.DEFAULT_VXLAN_PORT})

	}
	_, err := ctr.updateOvs(ovs.WithVxlans(vxs))

	if err != nil {
		return fmt.Errorf("could not create vxlans with neighbors %s", node.NeighborNodes)
	}

	fmt.Printf("Created vxlan with neighbors %s\n", node.NeighborNodes)

	return nil
}

// TODO: not finished the getVxlans method and getting localip
func (ctr *Controller) ConnectNewNeighbor(ip string) error {
	vs, _ := ctr.getOvs()
	vxs, _ := vs.GetVxlans()

	vxID, err := utils.GenerateInterfaceName("vxlan-", fmt.Sprintf("%s%s", "", ip))
	if err != nil {
		return fmt.Errorf("error generating vxlan id: %v", err)
	}
	vxs = append(vxs, plsv1.Vxlan{VxlanId: vxID, LocalIp: "", RemoteIp: ip, UdpPort: plsv1.DEFAULT_VXLAN_PORT})

	_, err = ctr.updateOvs(ovs.WithVxlans(vxs))

	if err != nil {
		return fmt.Errorf("could not create vxlans with neighbor %s", ip)
	}

	fmt.Printf("Created vxlan with neighbor %s\n", ip)

	return nil
}

/*
*
Example:

	{
	    "Nodes": [
	        {
	            "name": "l2sm1",
	            "nodeIP": "10.1.14.53"
	        },
	        {
	            "name": "l2sm2",
	            "nodeIP": "10.1.14.90"
	        }
	    ],
	    "Links": [
	        {
	            "endpointA": "l2sm1",
	            "endpointB": "l2sm2"
	        }
	    ]
	}
*/
func (ctr *Controller) CreateTopology(topology plsv1.Topology) error {

	nodeMap := make(map[string]string)
	for _, node := range topology.Nodes {
		var nodeIP string
		if parsedIP := net.ParseIP(node.NodeIP); parsedIP != nil {
			nodeIP = node.NodeIP
		} else {
			ips, err := resolveWithRetry(node.NodeIP, 300)
			if err != nil {
				fmt.Printf("Failed to resolve %s after retries: %v\n", node.NodeIP, err)
				continue
			}
			nodeIP = ips[0]
		}
		nodeMap[node.Name] = nodeIP
	}

	localIp := nodeMap[ctr.nodeName]

	vxs := []plsv1.Vxlan{}
	for _, link := range topology.Links {
		var remoteIp string
		switch ctr.nodeName {
		case link.EndpointNodeA:
			remoteIp = nodeMap[link.EndpointNodeB]
		case link.EndpointNodeB:
			remoteIp = nodeMap[link.EndpointNodeA]
		default:
			continue
		}
		vxID, _ := utils.GenerateInterfaceName("vxlan-", remoteIp)

		vxs = append(vxs, plsv1.Vxlan{VxlanId: vxID, LocalIp: localIp, RemoteIp: remoteIp, UdpPort: plsv1.DEFAULT_VXLAN_PORT})

	}
	_, err := ctr.updateOvs(ovs.WithVxlans(vxs))

	if err != nil {
		return fmt.Errorf("could not update existing switch %s. Provided Vxlans: %s. Error:%s", ctr.switchName, vxs, err)
	} else {
		fmt.Printf("Created topology %s.\n", vxs)
	}
	return nil

}

// func UpdateSwitch(bridge ovs.Bridge) error {

// }

func resolveWithRetry(host string, maxDelay int) ([]string, error) {
	for i := 1; i <= maxDelay; i = i * 2 {
		if i > maxDelay {
			i = maxDelay
		}
		fmt.Printf("Retrying service resolution for %s, next retry in: %ds\n", host, i)
		time.Sleep(time.Duration(i) * time.Second)

		ips, err := net.LookupHost(host)
		if err == nil && len(ips) > 0 {
			return ips, nil
		}
	}
	return nil, fmt.Errorf("unable to resolve host: %s", host)
}

func (ctr *Controller) AddPorts(interfacesNumber int) error {
	if interfacesNumber <= 0 {
		return fmt.Errorf("interfacesNumber must be > 0")
	}

	ports := make([]plsv1.Port, 0, interfacesNumber)

	for i := 1; i <= interfacesNumber; i++ {
		id := i // create a new variable so the pointer is unique

		ports = append(ports, plsv1.Port{
			Name: fmt.Sprintf("net%d", i),
			Id:   &id,
		})
	}

	_, err := ctr.updateOvs(
		ovs.WithPorts(ports),
	)

	return err
}

func (ctr *Controller) AddProbingPort(ip net.IPNet) error {
	id := plsv1.RESERVED_PROBE_ID
	ports := []plsv1.Port{
		{
			Name:      "probe0",
			Id:        &id,
			Internal:  true,
			IpAddress: &ip,
		},
	}
	_, err := ctr.updateOvs(
		ovs.WithPorts(ports),
	)

	return err
}
func (ctr *Controller) AddCustomInterface(switchName string) (int64, error) {
	// Create a new interface and attach it to the bridge
	newPort, err := ovs.AddInterfaceToBridge(switchName)
	if err != nil {
		return 0, fmt.Errorf("failed to create interface: %v", err)
	}

	vs, err := ctr.updateOvs(ovs.WithPorts([]plsv1.Port{{Name: newPort}}))

	if err != nil {
		return 0, fmt.Errorf("failed to add port to switch: %v", err)
	}

	return vs.GetPortNumber(newPort)

}

// Wrapper for ovs.UpdateVirtualSwitch, including the default name and sudo option
func (ctr *Controller) updateOvs(opts ...func(*ovs.BridgeConf)) (ovs.VirtualSwitch, error) {
	allOpts := append([]func(*ovs.BridgeConf){
		ovs.WithName(ctr.switchName), ovs.WithSudo(ctr.sudo),
	}, opts...)

	return ovs.UpdateVirtualSwitch(allOpts...)
}

// Wrapper for ovs.UpdateVirtualSwitch, including the default name and sudo option
func (ctr *Controller) newOvs(opts ...func(*ovs.BridgeConf)) (ovs.VirtualSwitch, error) {
	allOpts := append([]func(*ovs.BridgeConf){
		ovs.WithName(ctr.switchName), ovs.WithSudo(ctr.sudo),
	}, opts...)

	return ovs.NewVirtualSwitch(allOpts...)
}

// Wrapper for ovs.UpdateVirtualSwitch, including the default name and sudo option
func (ctr *Controller) getOvs() (ovs.VirtualSwitch, error) {

	return ovs.GetVirtualSwitch(ovs.WithName(ctr.switchName), ovs.WithSudo(ctr.sudo))
}
