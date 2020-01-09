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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	mesos "github.com/mesos/mesos-go/api/v1/lib"
	"github.com/mesos/mesos-go/api/v1/lib/maintenance"
	"github.com/mesos/mesos-go/api/v1/lib/master"
	"github.com/mesos/mesos-go/api/v1/lib/master/calls"
	"github.com/spf13/cobra"
)

type masterUpdateOptions struct {
	weights          string
	scheduleFilePath string
}

var masterUpdateOpts = &masterUpdateOptions{}

var masterUpdateCalls = MasterCallsDef{
	"weights": MasterCallDef{
		call: func() *master.Call {
			weightInfos := []mesos.WeightInfo{}
			for _, ws := range strings.Split(masterUpdateOpts.weights, ",") {
				if ws == "" {
					break
				}
				w := strings.Split(ws, ":")
				if len(w) != 2 {
					fmt.Printf("Bad weight format, expected <role>:<weight> got %s\n", ws)
					return nil
				}
				wf, err := strconv.ParseFloat(w[1], 64)
				if err != nil {
					fmt.Printf("Cannot parse weight %s as float: %s\n", w[1], err)
					return nil
				}
				weightInfos = append(weightInfos, mesos.WeightInfo{
					Role:   &w[0],
					Weight: wf,
				})
			}
			return calls.UpdateWeights(weightInfos...)
		},
		desc: "Updates weights for specific role",
	},
	"maintenance schedule": MasterCallDef{
		call: func() *master.Call {
			if masterUpdateOpts.scheduleFilePath == "" {
				fmt.Println("Missing required JSON file path in --schedule")
				return nil
			}
			jsonFile, err := os.Open(masterUpdateOpts.scheduleFilePath)
			if err != nil {
				fmt.Printf("Errot opening file %s: %s", masterUpdateOpts.scheduleFilePath, err)
				return nil
			}
			defer jsonFile.Close()
			jsonBytes, err := ioutil.ReadAll(jsonFile)
			if err != nil {
				fmt.Printf("Errot reading file %s: %s", masterUpdateOpts.scheduleFilePath, err)
				return nil
			}
			schedule := maintenance.Schedule{}
			err = json.Unmarshal(jsonBytes, &schedule)
			if err != nil {
				fmt.Printf("Errot parsing JSON file %s: %s", masterUpdateOpts.scheduleFilePath, err)
				return nil
			}
			return calls.UpdateMaintenanceSchedule(schedule)
		},
		desc: "Updates the cluster’s maintenance schedule.",
	},
	//TODO: update quota
}

var masterUpdateCmd = &cobra.Command{
	Use:   "update [call]",
	Short: "Update on master",
	Long:  masterUpdateCalls.describeCalls(),
	Args:  masterUpdateCalls.validateArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		key := strings.Join(args, " ")
		resp, err := masterCli.Send(context.Background(), calls.NonStreaming(masterUpdateCalls[key].call()))
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
	masterCmd.AddCommand(masterUpdateCmd)
	masterUpdateCmd.Flags().StringVar(&masterUpdateOpts.weights, "weights", "", "weight infos to update 'role:weight[,role:weight...]' (see --help)")
	masterUpdateCmd.Flags().StringVar(&masterUpdateOpts.scheduleFilePath, "schedule", "", "file path of maintenance schedule JSON (see --help)")

}
