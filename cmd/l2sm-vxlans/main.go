package main

import (
	"errors"
	"flag"
	"fmt"

	switchv1 "github.com/Networks-it-uc3m/l2sm-switch/api/v1"
	"github.com/Networks-it-uc3m/l2sm-switch/internal/inits"

	"github.com/Networks-it-uc3m/l2sm-switch/pkg/ovs"
)

// Script that takes two required arguments:
// the first one is the name in the cluster of the node where the script is running
// the second one is the path to the configuration file, in reference to the code.
func main() {

	//configDir, _, fileType, err := takeArguments()

	// configDir, nodeName, fileType, err := takeArguments()
	configDir, topologyDir, nodeName, err := takeArguments()

	if err != nil {
		fmt.Println("Error with the arguments provided. Error:", err)
		return
	}
	bridge := ovs.FromName("brtun")

	var topology switchv1.Topology

	err = inits.ReadFile(topologyDir, &topology)

	if err != nil {
		fmt.Println("Error with the provided file. Error:", err)
		return
	}

	var config switchv1.OverlaySettings

	err = inits.ReadFile(configDir, &config)

	if err != nil {
		fmt.Println("Error with the provided file. Error:", err)
		return
	}

	err = inits.CreateTopology(bridge, topology, nodeName)
}

func takeArguments() (string, string, string, error) {

	configDir := flag.String("config_dir", fmt.Sprintf("%s/config.json", switchv1.DEFAULT_CONFIG_PATH), "directory where the ned settings are specified. Required")
	topologyDir := flag.String("topology_dir", fmt.Sprintf("%s/topology.json", switchv1.DEFAULT_CONFIG_PATH), "directory where the ned's neighbors  are specified. Required")
	nodeName := flag.String("node_name", "", "name of the node the script is executed in. Required.")
	flag.Parse()

	switch {
	case *nodeName == "":
		return "", "", "", errors.New("node name is not defined")
	}

	return *configDir, *topologyDir, *nodeName, nil
}
