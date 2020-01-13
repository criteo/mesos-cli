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

type agentReadOptions struct {
	path   string
	length uint64
	offset uint64
}

var agentReadOpts = &agentReadOptions{}

var agentReadCalls = AgentCallsDef{
	"file": AgentCallDef{
		call: func() *agent.Call {
			if agentReadOpts.length > 0 {
				return calls.ReadFile(agentReadOpts.path, agentReadOpts.offset)
			} else {
				return calls.ReadFileWithLength(agentReadOpts.path, agentReadOpts.offset, agentReadOpts.length)
			}
		},
		desc: "Reads data from a file on the agent.",
	},
}

// agentListCmd represents the agentList command
var agentReadCmd = &cobra.Command{
	Use:   "read [call]",
	Short: "Read on agent",
	Long:  agentReadCalls.describeCalls(),
	Args:  agentReadCalls.validateArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		key := strings.Join(args, " ")
		resp, err := agentCli.Send(context.Background(), calls.NonStreaming(agentReadCalls[key].call()))
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
	agentCmd.AddCommand(agentReadCmd)
	agentReadCmd.Flags().StringVar(&agentReadOpts.path, "path", "", "path to list files")
	agentReadCmd.Flags().Uint64Var(&agentReadOpts.offset, "offset", 0, "the offset to start reading")
	agentReadCmd.Flags().Uint64Var(&agentReadOpts.length, "length", 0, "the maximum number of bytes to read")

	agentReadCmd.SetUsageTemplate(agentSubCommandUsageTemplate)

}
