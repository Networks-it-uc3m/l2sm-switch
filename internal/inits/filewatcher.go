package inits

import (
	"context"
	"crypto/sha512"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	plsv1 "github.com/Networks-it-uc3m/l2sm-switch/api/v1"
)

func StartFileWatcher(ctx context.Context, configPath, neighborsFileName, settingsFileName string) {

	neighborsFileDirectory := filepath.Join(configPath, neighborsFileName)
	//configFileDirectory := filepath.Join(configPath, settingsFileName)
	fmt.Println(neighborsFileDirectory)

	// Start listening for events.
	err := WatchFile(ctx, neighborsFileDirectory, 1*time.Second, plsv1.NEIGHBOR_FILE)
	if err != nil {
		log.Fatal(err)
	}

}

func WatchFile(ctx context.Context, f string, i time.Duration, ft string) error {

	if ft != plsv1.NEIGHBOR_FILE && ft != plsv1.TOPOLOGY_FILE && ft != plsv1.SETTINGS_FILE {
		return fmt.Errorf("specified file type %s is not compatible with internal/filewatcher library", ft)
	}
	parsedFile, err := os.ReadFile(f)

	if err != nil {
		return fmt.Errorf("error reading the file. err: %v", err)
	}

	sha512sum := sha512.Sum512(parsedFile)

	var node plsv1.Node
	var settings plsv1.Settings
	go func(ctx context.Context) {
		tick := time.NewTicker(i)
		defer tick.Stop()

		for {

			select {
			case <-ctx.Done():
				log.Println("finishing file watcher")
				return
			case <-tick.C:

				parsedFile, err := os.ReadFile(f)

				if err != nil {
					log.Printf("File %s parse failed: %s", ft, err)
					continue
				}

				s := sha512.Sum512(parsedFile)

				if s != sha512sum {
					sha512sum = s
					switch ft {
					case plsv1.NEIGHBOR_FILE:
						err = ReadFile(f, &node)
						if err != nil {
							log.Printf("ERROR: could not read the provided file: %v", err)
							break
						}

						err = ConnectToNeighbors(settings.SwitchName, node)
						if err != nil {
							log.Printf("ERROR: Could not connect neighbors: %v", err)
							break
						}

						log.Printf("Updated neighbors for node: %s", node.Name)

					case plsv1.SETTINGS_FILE:
						err = ReadFile(f, &settings)
						if err != nil {
							fmt.Println("Error with the provided file. Error:", err)
							break
						}
						fmt.Println("currently general settings modification is not supported")
						// _, err := ConfigureSwitch(
						// 	settings.NodeName,
						// 	settings.SwitchName,
						// 	settings.ControllerPort,
						// 	settings.ControllerIP,
						// )
						// if err != nil {
						// 	fmt.Println("Could not configure switch. Error:", err)
						// 	break
						// }
					// case plsv1.TOPOLOGY_FILE:
					// err = ReadFile(topologyFileDir, &topology)

					// if err != nil {
					// 	fmt.Println("Error with the provided file. Errorr:", err)
					// 	break
					// }
					// err = CreateTopology(settings.SwitchName, topology, settings.NodeName)
					// if err != nil {
					// 	fmt.Println("Could not update topology. Error:", err)
					// 	break
					// }
					case plsv1.TOPOLOGY_FILE:
					}

				}

			}
		}

	}(ctx)
	return nil

}
