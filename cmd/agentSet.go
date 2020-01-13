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
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mesos/mesos-go/api/v1/lib/agent"
	"github.com/mesos/mesos-go/api/v1/lib/agent/calls"
	"github.com/spf13/cobra"
)

type agentSetOptions struct {
	logLevel         uint32
	logLevelDuration time.Duration
	force            bool
	role             string
	resources        string
}

var agentSetOpts = &agentSetOptions{}

var agentSetCalls = AgentCallsDef{
	"logging level": AgentCallDef{
		call: func() *agent.Call {
			return calls.SetLoggingLevel(agentSetOpts.logLevel, agentSetOpts.logLevelDuration)
		},
		desc: "Sets the logging verbosity level for a specified duration for agent. (by default it’s 0, libprocess uses levels 1, 2, and 3).",
	},
}

// agentSetCmd represents the agentSet command
var agentSetCmd = &cobra.Command{
	Use:   "set [call]",
	Short: "Set on agent",
	Long:  agentSetCalls.describeCalls(),
	Args:  agentSetCalls.validateArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		key := strings.Join(args, " ")
		resp, err := agentCli.Send(context.Background(), calls.NonStreaming(agentSetCalls[key].call()))
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
	agentCmd.AddCommand(agentSetCmd)
	agentSetCmd.Flags().DurationVar(&agentSetOpts.logLevelDuration, "duration", 0, "log level duration (used to set log level, see --help)")
	agentSetCmd.Flags().Uint32Var(&agentSetOpts.logLevel, "log-level", 0, "log level to set (see --help)")

	agentSetCmd.SetUsageTemplate(agentSubCommandUsageTemplate)
}
