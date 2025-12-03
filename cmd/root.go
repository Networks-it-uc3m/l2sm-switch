/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	plsv1 "github.com/Networks-it-uc3m/l2sm-switch/api/v1"
	"github.com/spf13/cobra"
)

var configPath string
var sudo *bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "talpa",
	Short: "Create and manage l2sm and l2sces virtual switch",
	Long: `VSwitch is an implementation for managing L2S-CES virtual switches. 
It can either act as a Slice Packet Switch (SPS) or as a Network Edge Device (NED), with utilities such as file watching for container managing of the configuration, a grpc server for managing the talpa from another microservice or cli for manual configuration by a user.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&configPath, "config_path", plsv1.DEFAULT_CONFIG_PATH, "configuration path where config.json and topology.json are going to be placed.")
	sudo = rootCmd.PersistentFlags().Bool("sudo", false, "Append sudo to commands (for debugging)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("grpc_server", "", false, "Help message for toggle")
}
