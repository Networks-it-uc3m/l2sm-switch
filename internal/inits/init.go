package inits

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	plsv1 "github.com/Networks-it-uc3m/l2sm-switch/api/v1"
	"github.com/Networks-it-uc3m/l2sm-switch/pkg/ovs"
	"github.com/fsnotify/fsnotify"
)

func ConfigureSwitch(nodeName, switchName, controllerPort string, controllerIPs []string) (ovs.VirtualSwitch, error) {

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

	datapathId := ovs.GenerateDatapathID(nodeName)

	var err error
	var vs ovs.VirtualSwitch

	_, err = ovs.GetVirtualSwitch(ovs.WithName(switchName))

	if err != nil {
		fmt.Println("Switch doesn't exist. Creating a new one.")
		vs, err = ovs.NewVirtualSwitch(
			ovs.WithController(controllers),
			ovs.WithProtocol("OpenFlow13"),
			ovs.WithDatapathId(datapathId),
			ovs.WithName(switchName),
		)

		return vs, err
	}

	vs, err = ovs.UpdateVirtualSwitch(
		ovs.WithController(controllers),
		ovs.WithProtocol("OpenFlow13"),
		ovs.WithDatapathId(datapathId),
		ovs.WithName(switchName),
	)

	return vs, err
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
func ConnectToNeighbors(bridgeName string, node plsv1.Node) error {
	for vxlanNumber, neighborIp := range node.NeighborNodes {
		vxlanId := fmt.Sprintf("vxlan%d", vxlanNumber)
		_, err := ovs.UpdateVirtualSwitch(ovs.WithName(bridgeName), ovs.WithVxlans([]plsv1.Vxlan{{VxlanId: vxlanId, LocalIp: node.NodeIP, RemoteIp: neighborIp, UdpPort: "7000"}}))
		//err := bridge.CreateVxlan(ovs.Vxlan{VxlanId: vxlanId, LocalIp: node.NodeIP, RemoteIp: neighborIp, UdpPort: "7000"})

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
func CreateTopology(bridgeName string, topology plsv1.Topology, nodeName string) error {

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

	vxlans := []plsv1.Vxlan{}
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
		vxlans = append(vxlans, plsv1.Vxlan{VxlanId: vxlanId, LocalIp: localIp, RemoteIp: remoteIp, UdpPort: "7000"})

	}
	_, err := ovs.UpdateVirtualSwitch(ovs.WithName(bridgeName), ovs.WithVxlans(vxlans))

	if err != nil {
		return fmt.Errorf("could not update existing switch %s. Provided Vxlans: %s. Error:%s", bridgeName, vxlans, err)
	} else {
		fmt.Printf("Created topology %s.\n", vxlans)
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

func WatchDirectory(w *fsnotify.Watcher, topologyFileDir, configFileDirectory, neighborsFileDirectory string) {
	var node plsv1.Node
	var settings plsv1.Settings
	var topology plsv1.Topology
	var err error

	err = ReadFile(configFileDirectory, &settings)
	if err != nil {
		fmt.Println("Error with the settings file. Error:", err)
	}
	for {
		select {
		case event, ok := <-w.Events:
			if !ok {
				return
			}
			log.Println("event:", event)

			if event.Has(fsnotify.Write) {
				log.Println("modified file:", filepath.Base(event.Name))
				switch filepath.Base(event.Name) {
				case plsv1.SETTINGS_FILE:
					err = ReadFile(configFileDirectory, &settings)
					if err != nil {
						fmt.Println("Error with the provided file. Error:", err)
						break
					}
					_, err := ConfigureSwitch(
						settings.NodeName,
						settings.SwitchName,
						settings.ControllerPort,
						settings.ControllerIP,
					)
					if err != nil {
						fmt.Println("Could not configure switch. Error:", err)
						break
					}
				case plsv1.TOPOLOGY_FILE:
					err = ReadFile(topologyFileDir, &topology)

					if err != nil {
						fmt.Println("Error with the provided file. Errorr:", err)
						break
					}
					err = CreateTopology(settings.SwitchName, topology, settings.NodeName)
					if err != nil {
						fmt.Println("Could not update topology. Error:", err)
						break
					}
				case plsv1.NEIGHBOR_FILE:
					err = ReadFile(neighborsFileDirectory, &node)

					if err != nil {
						fmt.Println("Error with the provided file. Error:", err)
						break
					}

					err = ConnectToNeighbors(settings.SwitchName, node)
					if err != nil {
						fmt.Println("Could not connect neighbors: ", err)
						break
					}
				}

			}
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}

}

func AddPorts(switchName string, interfacesNumber int) error {
	ports := []plsv1.Port{}
	// Set all virtual interfaces up, and connect them to the tunnel bridge:
	for i := 1; i <= interfacesNumber; i++ {
		ports = append(ports, plsv1.Port{Name: fmt.Sprintf("net%d", i)})
	}
	_, err := ovs.UpdateVirtualSwitch(ovs.WithName(switchName), ovs.WithPorts(ports))
	return err

}
