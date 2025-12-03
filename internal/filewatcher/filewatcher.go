package filewatcher

import (
	"context"
	"crypto/sha512"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	plsv1 "github.com/Networks-it-uc3m/l2sm-switch/api/v1"
	"github.com/Networks-it-uc3m/l2sm-switch/internal/controller"
	"github.com/Networks-it-uc3m/l2sm-switch/pkg/utils"
)

type FileWatcher struct {
	Ctr        *controller.Controller
	FileType   string
	ConfigPath string
	Interval   time.Duration
}

func StartFileWatcher(ctx context.Context, configPath string, ctr *controller.Controller) {

	// Start listening for events.
	fw := &FileWatcher{Ctr: ctr, FileType: plsv1.NEIGHBOR_FILE, ConfigPath: configPath, Interval: 10 * time.Second}
	err := fw.WatchFile(ctx)
	if err != nil {
		log.Fatal(err)
	}

}

func (fw *FileWatcher) WatchFile(ctx context.Context) error {

	if fw.FileType != plsv1.NEIGHBOR_FILE && fw.FileType != plsv1.TOPOLOGY_FILE && fw.FileType != plsv1.SETTINGS_FILE {
		return fmt.Errorf("specified file type %s is not compatible with internal/filewatcher library", fw.FileType)
	}
	f := filepath.Join(fw.ConfigPath, fw.FileType)

	parsedFile, err := os.ReadFile(f)

	if err != nil {
		return fmt.Errorf("error reading the file. err: %v", err)
	}

	sha512sum := sha512.Sum512(parsedFile)

	go func(ctx context.Context) {
		tick := time.NewTicker(fw.Interval)
		defer tick.Stop()

		for {

			select {
			case <-ctx.Done():
				log.Println("finishing file watcher")
				return
			case <-tick.C:

				parsedFile, err := os.ReadFile(f)

				if err != nil {
					log.Printf("File %s parse failed: %s", f, err)
					continue
				}

				s := sha512.Sum512(parsedFile)

				if s != sha512sum {
					sha512sum = s
					switch fw.FileType {
					case plsv1.NEIGHBOR_FILE:
						var node plsv1.Node

						err = utils.ReadFile(f, &node)
						if err != nil {
							log.Printf("ERROR: could not read the provided file: %v", err)
							break
						}

						err = fw.Ctr.ConnectToNeighbors(node)
						if err != nil {
							log.Printf("ERROR: Could not connect neighbors: %v", err)
							break
						}

						log.Printf("Updated neighbors for node: %s", node.Name)

					case plsv1.SETTINGS_FILE:
						var settings plsv1.Settings

						err = utils.ReadFile(f, &settings)
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
