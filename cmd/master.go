/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
	"strings"

	"github.com/mesos/mesos-go/api/v1/lib/httpcli"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli/httpmaster"
	mesosMaster "github.com/mesos/mesos-go/api/v1/lib/master"
	"github.com/mesos/mesos-go/api/v1/lib/master/calls"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var masterCli calls.Sender

var masterCmd = &cobra.Command{
	Use:   "master",
	Short: "Interact with Mesos Master",
	Long:  `Interact with Mesos Master`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var auth httpcli.ConfigOpt
		if viper.IsSet("principal") && viper.GetString("principal") != "" {
			auth = httpcli.BasicAuth(
				viper.GetString("principal"),
				viper.GetString("secret"))
		}
		masterCli = httpmaster.NewSender(
			httpcli.New(
				httpcli.Endpoint(viper.GetString("master.url")+"/api/v1"),
				httpcli.Do(httpcli.With(auth))).Send)
	},
}

func init() {
	rootCmd.AddCommand(masterCmd)

	masterCmd.PersistentFlags().StringP("url", "u", "", "Mesos master URL")
	viper.BindPFlag("master.url", masterCmd.PersistentFlags().Lookup("url"))
	masterCmd.MarkPersistentFlagRequired("url")
}

func presetRequiredFlags() {
	if viper.IsSet("master.url") && viper.GetString("master.url") != "" {
		masterCmd.PersistentFlags().Set("url", viper.GetString("master.url"))
	}
}

type MasterCallDef struct {
	call  func() *mesosMaster.Call
	desc  string
	json  func(r *mesosMaster.Response) ([]byte, error)
	print func(r *mesosMaster.Response) error
}

type MasterCallsDef map[string]MasterCallDef

func (m MasterCallsDef) describeCalls() string {
	desc := "Possible call values:\n"
	if _, ok := m[""]; ok {
		desc += fmt.Sprintf("If no call specified: %s\n", m[""].desc)
		desc += "\n"
	}
	for key := range m {
		if key != "" {
			desc += fmt.Sprintf("- %s\n", key)
			desc += fmt.Sprintf("  %s\n", m[key].desc)
		}
	}

	return desc
}

func (m MasterCallsDef) validateArgs(cmd *cobra.Command, args []string) error {
	var key = strings.Join(args, " ")
	if _, ok := m[key]; ok {
		return nil
	} else {
		return fmt.Errorf("invalid arguments: %s try with --help", key)
	}
}
