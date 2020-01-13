/*
Copyright © 2020 Criteo

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string
var verbose bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mesos-cli",
	Short: "A simple Mesos CLI",
	Long: `mesos-cli is a command line interface (CLI) that can be used
to interact with Apache Mesos clusters`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Special case for agent command
	args := os.Args[1:]
	for index, v := range args {
		if v == "agent" && len(args) > index+1 && !strings.HasPrefix(args[index+1], "-") {
			agentOpts.name = args[index+1]
			os.Args = append(os.Args[0:index+2], os.Args[index+3:]...)
		}
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig, presetRequiredFlags)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mesos-cli.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	rootCmd.PersistentFlags().String("principal", "", "Mesos Principal")
	viper.BindPFlag("principal", rootCmd.PersistentFlags().Lookup("principal"))
	rootCmd.PersistentFlags().String("secret", "", "Mesos Secret")
	viper.BindPFlag("secret", rootCmd.PersistentFlags().Lookup("secret"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".mesos-cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".mesos-cli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
