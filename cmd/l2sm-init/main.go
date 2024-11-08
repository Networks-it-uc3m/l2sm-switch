package main

import (
	"flag"
	"fmt"
	"os/exec"
	"regexp"

	switchv1 "github.com/Networks-it-uc3m/l2sm-switch/api/v1"
	"github.com/Networks-it-uc3m/l2sm-switch/internal/inits"
	"github.com/Networks-it-uc3m/l2sm-switch/pkg/ovs"
)

// Script that takes two required arguments:
// the first one is the name in the cluster of the node where the script is running
// the second one is the path to the configuration file, in reference to the code.
func main() {

	configDir, _, _, _ := takeArguments()

	var settings switchv1.OverlaySettings

	err := inits.ReadFile(configDir, &settings)

	if err != nil {
		fmt.Println("Error with the config file. Error:", err)
		return
	}

	fmt.Println("Initializing switch, connected to controller: ", settings.ControllerIp)

	bridge, err := initializeSwitch(settings.NodeName, settings.ControllerIp)

	if err != nil {

		fmt.Println("Could not initialize switch. Error:", err)
		return
	}

	fmt.Println("Switch initialized and connected to the controller.")

	// Set all virtual interfaces up, and connect them to the tunnel bridge:
	for i := 1; i <= settings.InterfacesNumber; i++ {
		veth := fmt.Sprintf("net%d", i)
		if err := bridge.AddPort(veth); err != nil {
			fmt.Println("Error:", err)
		}
	}
	fmt.Printf("Switch initialized, current state: ", bridge)
}

func takeArguments() (string, int, string, string) {

	vethNumber := flag.Int("n_veths", 0, "number of pod interfaces that are going to be attached to the switch")
	controllerIP := flag.String("controller_ip", "", "ip where the SDN controller is listening using the OpenFlow13 protocol. ")
	switchName := flag.String("switch_name", "", "name of the switch that will be used to set a custom datapath id. If not set, a randome datapath will be assigned")
	configDir := flag.String("config_dir", fmt.Sprintf("%s/config.json", switchv1.DEFAULT_CONFIG_PATH), "directory where the switch settings are specified. ")

	flag.Parse()

	return *configDir, *vethNumber, *controllerIP, *switchName
}

func initializeSwitch(switchName, controllerIP string) (ovs.Bridge, error) {

	re := regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
	if !re.MatchString(controllerIP) {
		out, _ := exec.Command("host", controllerIP).Output()
		controllerIP = re.FindString(string(out))
	}

	controller := fmt.Sprintf("tcp:%s:6633", controllerIP)

	datapathId := ovs.GenerateDatapathID(switchName)
	bridge, err := ovs.NewBridge(ovs.Bridge{Name: "brtun", Controller: controller, Protocol: "OpenFlow13", DatapathId: datapathId})

	return bridge, err
}
