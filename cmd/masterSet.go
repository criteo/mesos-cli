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
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	mesos "github.com/mesos/mesos-go/api/v1/lib"
	"github.com/mesos/mesos-go/api/v1/lib/master"
	"github.com/mesos/mesos-go/api/v1/lib/master/calls"
	"github.com/mesos/mesos-go/api/v1/lib/quota"
	"github.com/mesos/mesos-go/api/v1/lib/resources"
	"github.com/spf13/cobra"
)

type masterSetOptions struct {
	logLevel         uint32
	logLevelDuration time.Duration
	force            bool
	role             string
	resources        string
}

var masterSetOpts = &masterSetOptions{}

var masterSetCalls = MasterCallsDef{
	"logging level": MasterCallDef{
		call: func() *master.Call {
			return calls.SetLoggingLevel(masterSetOpts.logLevel, masterSetOpts.logLevelDuration)
		},
		desc: "Sets the logging verbosity level for a specified duration for master. (by default it’s 0, libprocess uses levels 1, 2, and 3).",
	},
	"quota": MasterCallDef{
		call: func() *master.Call {
			if masterSetOpts.role == "" {
				fmt.Println("Missing --role to set quota")
				return nil
			}
			guarantee := []mesos.Resource{}
			for _, rs := range strings.Split(masterSetOpts.resources, ",") {
				if rs == "" {
					break
				}
				r := strings.Split(rs, ":")
				if len(r) != 2 {
					fmt.Printf("Bad resource format for %s expecting <name>:<value\n>", rs)
					return nil
				}
				v, err := strconv.ParseFloat(r[1], 64)
				if err != nil {
					fmt.Printf("Wrong resource scalar value format %s, error: %s\n", r[1], err.Error())
					return nil
				}
				guarantee = append(guarantee, resources.Build().Name(resources.Name(r[0])).Scalar(v).Resource)
			}
			quota := quota.QuotaRequest{
				Force:     &masterSetOpts.force,
				Role:      &masterSetOpts.role,
				Guarantee: guarantee,
			}
			return calls.SetQuota(quota)
		},
		desc: "Sets the quota for resources to be used by a particular role.",
	},
}

// masterSetCmd represents the masterSet command
var masterSetCmd = &cobra.Command{
	Use:   "set [call]",
	Short: "Set on master",
	Long:  masterSetCalls.describeCalls(),
	Args:  masterSetCalls.validateArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		key := strings.Join(args, " ")
		resp, err := masterCli.Send(context.Background(), calls.NonStreaming(masterSetCalls[key].call()))
		defer func() {
			if resp != nil {
				resp.Close()
			}
		}()
		if err != nil {
			return fmt.Errorf("Error sending call: %s", err)
		}
		return nil
	},
}

func init() {
	masterCmd.AddCommand(masterSetCmd)
	masterSetCmd.Flags().DurationVar(&masterSetOpts.logLevelDuration, "duration", 0, "log level duration (used to set log level, see --help)")
	masterSetCmd.Flags().Uint32Var(&masterSetOpts.logLevel, "log-level", 0, "log level to set (see --help)")
	masterSetCmd.Flags().BoolVar(&masterSetOpts.force, "force", false, "force quota and don't check for overcommit (see --help)")
	masterSetCmd.Flags().StringVar(&masterSetOpts.role, "role", "", "quota role (see --help)")
	masterSetCmd.Flags().StringVar(&masterSetOpts.resources, "resources", "", "quota resources in the format 'name:value[,name:value...]' (example: 'cpu:12,mem:192' see --help)")
}
