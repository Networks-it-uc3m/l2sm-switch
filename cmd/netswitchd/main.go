package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	plsv1 "github.com/Networks-it-uc3m/l2sm-switch/api/v1"
	"github.com/Networks-it-uc3m/l2sm-switch/internal/inits"
	"github.com/fsnotify/fsnotify"
)

func main() {

	configPath := takeArguments()

	topologyDir := filepath.Join(configPath, plsv1.TOPOLOGY_FILE)
	settingsDir := filepath.Join(configPath, plsv1.SETTINGS_FILE)

	var settings plsv1.Settings

	err := inits.ReadFile(settingsDir, &settings)

	if err != nil {
		fmt.Println("Error with the config file. Error:", err)
		return
	}

	fmt.Println("Initializing switch, connecting to controller: ", settings.ControllerIP)

	_, err = inits.ConfigureSwitch(
		settings.NodeName,
		settings.SwitchName,
		settings.ControllerPort,
		settings.ControllerIP,
	)

	if err != nil {

		fmt.Println("Could not initialize switch. Error:", err)
		return
	}

	err = inits.AddPorts(settings.SwitchName, settings.InterfacesNumber)
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Printf("\nSwitch initialized, current state: %v", settings.SwitchName)

	var topology plsv1.Topology

	err = inits.ReadFile(topologyDir, &topology)

	if err != nil {
		fmt.Println("Error with the provided file. Error:", err)
		return
	}

	err = inits.CreateTopology(settings.SwitchName, topology, settings.NodeName)

	if err != nil {
		log.Fatal("Error creating the topology: ", err)
	}
	//err = inits.CreateTopology(bridge, topology, nodeName)

	// Create new watcher.

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Error setting the topology watcher")
	}
	defer watcher.Close()

	// Start listening for events.
	//go inits.WatchDirectory(watcher, topologyDir, settingsDir, "")

	// Add a path.
	err = watcher.Add(configPath)
	if err != nil {
		log.Fatal(err)
	}

	// Block main goroutine forever.
	<-make(chan struct{})
}
func takeArguments() string {

	configPath := flag.String("config_path", plsv1.DEFAULT_CONFIG_PATH, "configuration path where config.json and topology.json are going to be used.")

	flag.Parse()

	return *configPath
}
