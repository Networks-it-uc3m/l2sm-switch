package main

import (
	"errors"
	"flag"
	"fmt"

	plsv1 "github.com/Networks-it-uc3m/l2sm-switch/api/v1"
	"github.com/Networks-it-uc3m/l2sm-switch/pkg/ovs"
)

// Script that takes two required arguments:
// the first one is the name in the cluster of the node where the script is running
// the second one is the path to the configuration file, in reference to the code.
func main() {

	portName, err := takeArguments()

	if err != nil {
		fmt.Println("Error with arguments: ", err)
		return
	}
	_, err = ovs.UpdateVirtualSwitch(ovs.WithName("brtun"), ovs.WithPorts([]plsv1.Port{{Name: portName}}))

	if err != nil {
		fmt.Println("Port not added: ", err)
		return
	}
}

func takeArguments() (string, error) {

	portName := flag.String("port_name", "", "port you want to add. Required.")

	flag.Parse()

	if *portName == "" {
		return "", errors.New("port name is not defined")

	}

	return *portName, nil
}
