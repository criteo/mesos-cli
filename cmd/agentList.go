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

	"github.com/mesos/mesos-go/api/v1/lib/agent"
	"github.com/mesos/mesos-go/api/v1/lib/agent/calls"
	"github.com/spf13/cobra"
)

type agentListOptions struct {
	path string
}

var agentListOpts = &agentListOptions{}

var agentListCalls = AgentCallsDef{
	"files": AgentCallDef{
		call: func() *agent.Call {
			return calls.ListFiles(agentListOpts.path)
		},
		desc: "Retrieves the file listing for a directory in agent.",
	},
}

// agentListCmd represents the agentList command
var agentListCmd = &cobra.Command{
	Use:   "list [call]",
	Short: "List on agent",
	Long:  agentListCalls.describeCalls(),
	Args:  agentListCalls.validateArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		key := strings.Join(args, " ")
		resp, err := agentCli.Send(context.Background(), calls.NonStreaming(agentListCalls[key].call()))
		defer func() {
			if resp != nil {
				resp.Close()
			}
		}()
		if err != nil {
			return fmt.Errorf("Error sending call: %s", err)
		}
		var e agent.Response
		err = resp.Decode(&e)
		if err != nil {
			return fmt.Errorf("Error decoding response: %s", err)
		}
		return nil
	},
}

func init() {
	agentCmd.AddCommand(agentListCmd)
	agentListCmd.Flags().StringVar(&agentListOpts.path, "path", "", "path to list files")

	agentListCmd.SetUsageTemplate(agentSubCommandUsageTemplate)
}
