package main

import (
	"flag"
	"fmt"
	"path/filepath"

	plsv1 "github.com/Networks-it-uc3m/l2sm-switch/api/v1"
	"github.com/Networks-it-uc3m/l2sm-switch/internal/inits"
	"github.com/Networks-it-uc3m/l2sm-switch/pkg/ovs"
)

// Script that takes two required arguments:
// the first one is the name in the cluster of the node where the script is running
// the second one is the path to the configuration file, in reference to the code.
func main() {

	nodeName, bridgeName := takeArguments()

	configDir := filepath.Join(plsv1.DEFAULT_CONFIG_PATH, plsv1.SETTINGS_FILE)
	var settings plsv1.Settings

	err := inits.ReadFile(configDir, &settings)

	if err != nil {
		fmt.Println("Error with the config file. Error:", err)
		return
	}

	fmt.Println("Initializing switch, connecting to controller: ", settings.ControllerIP)

	vSwitch, err := inits.ConfigureSwitch(
		nodeName,
		bridgeName,
		settings.ControllerPort,
		settings.ControllerIP,
	)

	if err != nil {

		fmt.Println("Could not initialize switch. Error:", err)
		return
	}

	fmt.Println("Switch initialized and connected to the controller.")

	ports := []plsv1.Port{}
	// Set all virtual interfaces up, and connect them to the tunnel bridge:
	for i := 1; i <= settings.InterfacesNumber; i++ {
		ports = append(ports, plsv1.Port{Name: fmt.Sprintf("net%d", i)})
	}
	_, err = ovs.UpdateVirtualSwitch(ovs.WithName(bridgeName), ovs.WithPorts(ports))
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Printf("\nSwitch initialized, current state: %v", vSwitch)
}

func takeArguments() (string, string) {

	switchName := flag.String("switch_name", "", "name of the switch that will be used to set a custom datapath id. If not set, a randome datapath will be assigned")
	bridgeName := flag.String("bridge_name", "brtun", "name of the ovs bridge.")

	flag.Parse()

	return *switchName, *bridgeName
}
