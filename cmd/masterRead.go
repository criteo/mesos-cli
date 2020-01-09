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
	"strings"

	"github.com/mesos/mesos-go/api/v1/lib/master"
	"github.com/mesos/mesos-go/api/v1/lib/master/calls"
	"github.com/spf13/cobra"
)

type masterReadOptions struct {
	path   string
	length uint64
	offset uint64
}

var masterReadOpts = &masterReadOptions{}

var masterReadCalls = MasterCallsDef{
	"file": MasterCallDef{
		call: func() *master.Call {
			if masterReadOpts.length > 0 {
				return calls.ReadFile(masterReadOpts.path, masterReadOpts.offset)
			} else {
				return calls.ReadFileWithLength(masterReadOpts.path, masterReadOpts.offset, masterReadOpts.length)
			}
		},
		desc: "Reads data from a file on the master.",
	},
}

// masterListCmd represents the masterList command
var masterReadCmd = &cobra.Command{
	Use:   "read [call]",
	Short: "Read on master",
	Long:  masterReadCalls.describeCalls(),
	Args:  masterReadCalls.validateArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		key := strings.Join(args, " ")
		resp, err := masterCli.Send(context.Background(), calls.NonStreaming(masterReadCalls[key].call()))
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
	masterCmd.AddCommand(masterReadCmd)
	masterReadCmd.Flags().StringVar(&masterReadOpts.path, "path", "", "path to list files")
	masterReadCmd.Flags().Uint64Var(&masterReadOpts.offset, "offset", 0, "the offset to start reading")
	masterReadCmd.Flags().Uint64Var(&masterReadOpts.length, "length", 0, "the maximum number of bytes to read")
}
