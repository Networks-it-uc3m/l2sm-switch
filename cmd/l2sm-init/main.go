package main

import (
	"flag"
	"fmt"

	switchv1 "github.com/Networks-it-uc3m/l2sm-switch/api/v1"
	"github.com/Networks-it-uc3m/l2sm-switch/internal/inits"
)

// Script that takes two required arguments:
// the first one is the name in the cluster of the node where the script is running
// the second one is the path to the configuration file, in reference to the code.
func main() {

	configDir, _, _, switchName := takeArguments()

	var settings switchv1.OverlaySettings

	err := inits.ReadFile(configDir, &settings)

	if err != nil {
		fmt.Println("Error with the config file. Error:", err)
		return
	}

	fmt.Println("Initializing switch, connected to controller: ", settings.ControllerIp)

	bridge, err := inits.InitializeSwitch(switchName, settings.ControllerIp, settings.ControllerPort)

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
