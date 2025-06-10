package main

import (
	"errors"
	"flag"
	"fmt"
	"path/filepath"

	plsv1 "github.com/Networks-it-uc3m/l2sm-switch/api/v1"
	"github.com/Networks-it-uc3m/l2sm-switch/internal/inits"
)

// Script that takes two required arguments:
// the first one is the name in the cluster of the node where the script is running
// the second one is the path to the configuration file, in reference to the code.
func main() {

	//configDir, _, fileType, err := takeArguments()

	// configDir, nodeName, fileType, err := takeArguments()
	nodeName, bridgeName, err := takeArguments()

	if err != nil {
		fmt.Println("Error with the arguments provided. Error:", err)
		return
	}

	configDir := filepath.Join(plsv1.DEFAULT_CONFIG_PATH, plsv1.SETTINGS_FILE)
	topologyDir := filepath.Join(plsv1.DEFAULT_CONFIG_PATH, plsv1.TOPOLOGY_FILE)

	var topology plsv1.Topology

	err = inits.ReadFile(topologyDir, &topology)

	if err != nil {
		fmt.Println("Error with the provided file. Error:", err)
		return
	}

	var settings plsv1.Settings

	err = inits.ReadFile(configDir, &settings)

	if err != nil {
		fmt.Println("Error with the provided file. Error:", err)
		return
	}

	err = inits.CreateTopology(bridgeName, topology, nodeName)

	if err != nil {
		fmt.Println("Error creating the topology. Error:", err)
		return
	}
}

func takeArguments() (string, string, error) {

	nodeName := flag.String("node_name", "", "name of the node the script is executed in. Required.")
	bridgeName := flag.String("bridge_name", "brtun", "name of the ovs bridge.")
	flag.Parse()

	switch {
	case *nodeName == "":
		return "", "", errors.New("node name is not defined")
	}

	return *nodeName, *bridgeName, nil
}
