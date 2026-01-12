/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	plsv1 "github.com/Networks-it-uc3m/l2sm-switch/api/v1"
	"github.com/spf13/cobra"
)

var configPath string
var monitorFile string
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
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Gate: only run for these subcommands
		switch cmd.Name() {
		case "ned", "sps-init":
			// read the flag value in a way that matches your current flag design
			useSudo, _ := cmd.Flags().GetBool("sudo")
			// OR if sudo is persistent on root: cmd.Root().Flags().GetBool("sudo")
			ctx := cmd.Context()
			return initOvs(ctx, useSudo)
		default:
			return nil
		}
	},
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
	// when this action is called directly.
	rootCmd.PersistentFlags().StringVar(&monitorFile, "monitor_file", "", "Path to monitoring config file (enables monitoring sidecar/container)")

	//rootCmd.Flags().BoolP("grpc_server", "", false, "Help message for toggle")
}

func initOvs(ctx context.Context, useSudo bool) error {
	// helper
	run := func(name string, args ...string) error {
		if useSudo {
			args = append([]string{name}, args...)
			name = "sudo"
		}
		c := exec.CommandContext(ctx, name, args...)
		out, err := c.CombinedOutput()
		if err != nil {
			return fmt.Errorf("command failed: %s %s: %w\noutput: %s",
				name, strings.Join(args, " "), err, string(out))
		}
		return nil
	}

	if err := run("ovsdb-server",
		"--remote=punix:/var/run/openvswitch/db.sock",
		"--remote=db:Open_vSwitch,Open_vSwitch,manager_options",
		"--pidfile=/var/run/openvswitch/ovsdb-server.pid",
		"--detach",
	); err != nil {
		return err
	}

	if err := run("ovs-vsctl",
		"--db=unix:/var/run/openvswitch/db.sock",
		"--no-wait",
		"init",
	); err != nil {
		return err
	}

	if err := run("ovs-vswitchd",
		"--pidfile=/var/run/openvswitch/ovs-vswitchd.pid",
		"--detach",
	); err != nil {
		return err
	}

	return nil
}
