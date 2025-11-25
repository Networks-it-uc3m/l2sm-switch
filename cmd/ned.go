/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	// Adjust the import path based on your module path
	plsv1 "github.com/Networks-it-uc3m/l2sm-switch/api/v1"

	"github.com/Networks-it-uc3m/l2sm-switch/internal/inits"
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
		fmt.Println("ned called")
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		inits.StartFileWatcher(ctx, configPath, plsv1.NEIGHBOR_FILE, plsv1.SETTINGS_FILE)

		inits.StartGrpcServer(configPath)
		//chatgpt dice que prefiere que haga esto, yo q se
		//<-make(chan struct{})

	},
}

func init() {
	rootCmd.AddCommand(nedCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// nedCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	//nedCmd.Flags().BoolP("grpc_server", "t", false, "Help message for toggle")
}
