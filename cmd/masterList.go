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

	"github.com/mesos/mesos-go/api/v1/lib/master"
	"github.com/mesos/mesos-go/api/v1/lib/master/calls"
	"github.com/spf13/cobra"
)

type masterListOptions struct {
	path string
}

var masterListOpts = &masterListOptions{}

var masterListCalls = MasterCallsDef{
	"files": MasterCallDef{
		call: func() *master.Call {
			return calls.ListFiles(masterListOpts.path)
		},
		desc: "Retrieves the file listing for a directory in master.",
	},
}

// masterListCmd represents the masterList command
var masterListCmd = &cobra.Command{
	Use:   "list [call]",
	Short: "List on master",
	Long:  masterListCalls.describeCalls(),
	Args:  masterListCalls.validateArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		key := strings.Join(args, " ")
		resp, err := masterCli.Send(context.Background(), calls.NonStreaming(masterListCalls[key].call()))
		defer func() {
			if resp != nil {
				resp.Close()
			}
		}()
		if err != nil {
			return fmt.Errorf("Error sending call: %s", err)
		}
		var e master.Response
		err = resp.Decode(&e)
		if err != nil {
			return fmt.Errorf("Error decoding response: %s", err)
		}
		return nil
	},
}

func init() {
	masterCmd.AddCommand(masterListCmd)
	masterListCmd.Flags().StringVar(&masterListOpts.path, "path", "", "path to list files")
}
