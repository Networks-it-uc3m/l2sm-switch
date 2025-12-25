/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"net"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	// Adjust the import path based on your module path
	plsv1 "github.com/Networks-it-uc3m/l2sm-switch/api/v1"
	"github.com/Networks-it-uc3m/l2sm-switch/internal/controller"
	"github.com/Networks-it-uc3m/l2sm-switch/pkg/datapath"
	"github.com/Networks-it-uc3m/l2sm-switch/pkg/utils"
)

// nedCmd represents the ned command
var spsCmd = &cobra.Command{
	Use:   "sps-init",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		nodeName, err := cmd.Flags().GetString("node_name")
		if err != nil {
			fmt.Println("Error with the node name variable. Error:", err)
			return
		}

		configDir := filepath.Join(configPath, plsv1.SETTINGS_FILE)
		topologyDir := filepath.Join(configPath, plsv1.TOPOLOGY_FILE)

		var settings plsv1.Settings

		err = utils.ReadFile(configDir, &settings)

		if err != nil {
			fmt.Println("Error with the config file. Error:", err)
			return
		}
		var topology plsv1.Topology

		err = utils.ReadFile(topologyDir, &topology)

		if err != nil {
			fmt.Println("Error with the provided file. Error:", err)
			return
		}

		fmt.Println("Initializing switch, connecting to controller: ", settings.ControllerIP)
		switchName := datapath.GetSwitchName(datapath.DatapathParams{NodeName: nodeName, ProviderName: settings.ProviderName})

		ctr := controller.NewSwitchManager(switchName, nodeName, *sudo)
		vs, err := ctr.ConfigureSwitch(
			settings.ControllerPort,
			settings.ControllerIP,
		)

		if err != nil {

			fmt.Println("Could not initialize switch. Error:", err)
			return
		}

		fmt.Println("Switch initialized and connected to the controller.")

		ctr.AddPorts(settings.InterfacesNumber)
		if err != nil {
			fmt.Println("Error:", err)
		}
		fmt.Printf("\nSwitch initialized, current state: %v", vs)

		if settings.ProbingIpAddress != nil {
			_, ip, err := net.ParseCIDR(*settings.ProbingIpAddress)
			if err != nil {
				fmt.Printf("Error parsing ip address for probing port: %v", err)
			} else {
				ctr.AddProbingPort(*ip)
			}
		}

		time.Sleep(20 * time.Second)

		err = ctr.CreateTopology(topology)

		if err != nil {
			fmt.Println("Error creating the topology. Error:", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(spsCmd)
	spsCmd.PersistentFlags().String("node_name", "", "name of the node the script is executed in. Required.")

}
