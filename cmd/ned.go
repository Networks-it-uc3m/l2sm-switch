/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"net"
	"path/filepath"

	"github.com/spf13/cobra"

	// Adjust the import path based on your module path
	plsv1 "github.com/Networks-it-uc3m/l2sm-switch/api/v1"
	"github.com/Networks-it-uc3m/l2sm-switch/pkg/utils"

	"github.com/Networks-it-uc3m/l2sm-switch/internal/controller"
	"github.com/Networks-it-uc3m/l2sm-switch/internal/filewatcher"
	"github.com/Networks-it-uc3m/l2sm-switch/internal/server"
)

// nedCmd represents the ned command
var nedCmd = &cobra.Command{
	Use:   "ned",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		sudo, err := cmd.Flags().GetBool("sudo")

		if err != nil {
			fmt.Println("Error with the sudo variable. Error:", err)
			return
		}
		port, err := cmd.Flags().GetString("port")
		if err != nil {
			fmt.Println("Error with the port variable. Error:", err)
			return
		}

		configDir := filepath.Join(configPath, plsv1.SETTINGS_FILE)
		neighDir := filepath.Join(configPath, plsv1.NEIGHBOR_FILE)

		var settings plsv1.Settings

		err = utils.ReadFile(configDir, &settings)

		if err != nil {
			fmt.Println("Error with the config file. Error:", err)
			return
		}
		var node plsv1.Node

		err = utils.ReadFile(neighDir, &node)

		if err != nil {
			fmt.Println("Error reading neighbor file. Error:", err)
			return
		}

		ctr := controller.NewSwitchManager(settings.SwitchName, settings.NodeName, sudo)

		_, err = ctr.ConfigureSwitch(
			settings.ControllerPort,
			settings.ControllerIP,
		)
		if err != nil {
			fmt.Println("Error configuring switch. Error:", err)
			return
		}

		if settings.ProbingIpAddress != nil {
			_, ip, err := net.ParseCIDR(*settings.ProbingIpAddress)
			if err != nil {
				fmt.Printf("Error parsing ip address for probing port: %v", err)
			} else {
				ctr.AddProbingPort(*ip)
			}
		}
		err = ctr.ConnectToNeighbors(node)
		if err != nil {
			fmt.Println("Error connecting to neighbors. Error:", err)
			return
		}
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		filewatcher.StartFileWatcher(ctx, configPath, ctr)

		server.StartGrpcServer(port, ctr)

	},
}

func init() {
	rootCmd.AddCommand(nedCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	nedCmd.PersistentFlags().String("port", "50051", "number of the port the grpc will listen to")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
}
