/*
Copyright © 2023 vishal tewaita <tewatiavishal3@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vishu42/terraformer/cmd/cli/cmd/impl"
)

var cfgFile string

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	fmt.Println("inside init config")
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".terraformer" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".terraformer")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

type CommandGroup struct{}

func (c *CommandGroup) RootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "terraformer",
		Short: "Server-side terraform execution plus integrated RBAC",
		Long: `terraformer a cli tool that is configured to work with terraformer server. It allows us to run the terraform code
	at the server side and provides integrated RBAC for running the terraform modules`,
	}
	// set debug flag
	rootCmd.PersistentFlags().Bool("debug", false, "enable debug mode")
	return rootCmd
}

func (c *CommandGroup) PlanCommand() (cmds *cobra.Command) {
	o := &impl.PlanOpts{}

	p := &cobra.Command{
		Use:   "plan",
		Short: "plan is equivalent to terraform plan",
		Run: func(cmd *cobra.Command, args []string) {
			impl.RunPlan(cmd, args, o)
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			o.ServerAddr = viper.GetString("server")
		},
	}
	p.PersistentFlags().String("server", "", "the server address")

	return p
}

func (c *CommandGroup) LoginCommand() (cmd *cobra.Command) {
	l := &cobra.Command{
		Use:   "login",
		Short: "login using your microsoft account",
		Run:   impl.RunLogin,
	}

	return l
}

func (c *CommandGroup) VersionCommand() (cmd *cobra.Command) {
	o := &impl.VersionOpts{}
	v := &cobra.Command{
		Use:   "version",
		Short: "displays the version of client and server",
		Run: func(cmd *cobra.Command, args []string) {
			impl.RunVersion(cmd, args, o)
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			o.ServerAddr = viper.GetString("server")
		},
	}

	return v
}

func (c *CommandGroup) AddCommands(root *cobra.Command, cmds ...*cobra.Command) {
	root.AddCommand(cmds...)
}

func (c *CommandGroup) All() *cobra.Command {
	cobra.OnInitialize(initConfig)
	root := c.RootCmd()
	c.AddCommands(root, c.PlanCommand(), c.LoginCommand(), c.VersionCommand())
	return root
}
