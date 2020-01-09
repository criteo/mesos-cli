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
	"io"

	mesos "github.com/mesos/mesos-go/api/v1/lib"
	"github.com/mesos/mesos-go/api/v1/lib/master/calls"

	"github.com/mesos/mesos-go/api/v1/lib/master"

	"github.com/spf13/cobra"
)

type masterEventsOptions struct {
	json bool
}

var masterEventsOpts = masterEventsOptions{}

var masterEventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Watch master events",
	Long:  "Watch master events",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := masterCli.Send(context.Background(), calls.NonStreaming(calls.Subscribe()))
		defer func() {
			if resp != nil {
				resp.Close()
			}
		}()
		for err == nil {
			var e master.Event
			if err := resp.Decode(&e); err != nil {
				if err == io.EOF {
					err = nil
					break
				}
				fmt.Println("Error", err)
				continue
			}
			switch t := e.GetType(); t {
			case master.Event_TASK_ADDED:
				task := e.GetTaskAdded().Task
				fmt.Println(t.String(), task.GetFrameworkID(), task.GetTaskID(), task.GetState(), task.GetLabels().Format(), mesos.Resources(task.GetResources()))
			case master.Event_TASK_UPDATED:
				task := e.GetTaskUpdated().GetStatus()
				fmt.Println(t.String(), task.GetTaskID(), task.GetState(), task.GetLabels().Format())
			case master.Event_AGENT_ADDED:
				fmt.Println(t.String(), e.GetAgentAdded().String())
			case master.Event_AGENT_REMOVED:
				fmt.Println(t.String(), e.GetAgentRemoved().String())
			case master.Event_FRAMEWORK_ADDED:
				fmt.Println(t.String(), e.GetFrameworkAdded().String())
			case master.Event_FRAMEWORK_UPDATED:
				fmt.Println(t.String(), e.GetFrameworkUpdated().String())
			case master.Event_FRAMEWORK_REMOVED:
				fmt.Println(t.String(), e.GetFrameworkRemoved().String())
			default:
				fmt.Println(t.String(), e)
			}
		}
		return err
	},
}

func init() {
	masterCmd.AddCommand(masterEventsCmd)
	//masterEventsCmd.Flags().BoolVarP(&masterEventsOpts.json, "json", "j", false, "json output")
}
