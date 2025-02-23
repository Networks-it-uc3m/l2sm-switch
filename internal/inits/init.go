package inits

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"time"

	switchv1 "github.com/Networks-it-uc3m/l2sm-switch/api/v1"
	"github.com/Networks-it-uc3m/l2sm-switch/pkg/ovs"
)

func InitializeSwitch(switchName, controllerIP, controllerPort string) (ovs.Bridge, error) {

	re := regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)

	if !re.MatchString(controllerIP) {

		out, _ := exec.Command("host", controllerIP).Output()

		controllerIP = re.FindString(string(out))

	}
	controller := fmt.Sprintf("tcp:%s:%s", controllerIP, controllerPort)

	datapathId := ovs.GenerateDatapathID(switchName)
	bridge, err := ovs.NewBridge(ovs.Bridge{Name: switchName, Controller: controller, Protocol: "OpenFlow13", DatapathId: datapathId})

	return bridge, err
}

func ReadFile(configDir string, dataStruct interface{}) error {

	/// Read file and save in memory the JSON info
	data, err := os.ReadFile(configDir)
	if err != nil {
		fmt.Println("No input file was found.", err)
		return err
	}

	err = json.Unmarshal(data, &dataStruct)
	if err != nil {
		return err
	}

	return nil

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
func ConnectToNeighbors(bridge ovs.Bridge, node switchv1.Node) error {
	for vxlanNumber, neighborIp := range node.NeighborNodes {
		vxlanId := fmt.Sprintf("vxlan%d", vxlanNumber)
		err := bridge.CreateVxlan(ovs.Vxlan{VxlanId: vxlanId, LocalIp: node.NodeIP, RemoteIp: neighborIp, UdpPort: "7000"})

		if err != nil {
			return fmt.Errorf("could not create vxlan with neighbor %s", neighborIp)
		} else {
			fmt.Printf("Created vxlan with neighbor %s", neighborIp)
		}
	}
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
func CreateTopology(bridge ovs.Bridge, topology switchv1.Topology, nodeName string) error {

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

	localIp := nodeMap[nodeName]

	for vxlanNumber, link := range topology.Links {
		vxlanId := fmt.Sprintf("vxlan%d", vxlanNumber)
		var remoteIp string
		switch nodeName {
		case link.EndpointNodeA:
			remoteIp = nodeMap[link.EndpointNodeB]
		case link.EndpointNodeB:
			remoteIp = nodeMap[link.EndpointNodeA]
		default:
			continue
		}
		err := bridge.CreateVxlan(ovs.Vxlan{VxlanId: vxlanId, LocalIp: localIp, RemoteIp: remoteIp, UdpPort: "7000"})

		if err != nil {
			return fmt.Errorf("could not create vxlan between node %s and node %s. Error:%s", link.EndpointNodeA, link.EndpointNodeB, err)
		} else {
			fmt.Printf("Created vxlan between node %s and node %s.\n", link.EndpointNodeA, link.EndpointNodeB)
		}

	}
	return nil

}

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
