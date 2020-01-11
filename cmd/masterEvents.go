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
			if err = resp.Decode(&e); err != nil {
				if err == io.EOF {
					err = nil
				}
				break
			}
			switch t := e.GetType(); t {
			case master.Event_SUBSCRIBED,
				master.Event_HEARTBEAT:
				fmt.Println(t.String())
			case master.Event_TASK_ADDED:
				task := e.GetTaskAdded().Task
				fmt.Printf("%s: framework %s task %s %s %s\n", t.String(), task.GetFrameworkID().Value, task.GetTaskID().Value, task.GetState(), task.GetLabels().Format())
			case master.Event_TASK_UPDATED:
				tu := e.GetTaskUpdated()
				task := tu.GetStatus()
				fmt.Printf("%s: framework %s task %s %s %s\n", t.String(), tu.GetFrameworkID().Value, task.GetTaskID().Value, task.GetState(), task.GetLabels().Format())
			case master.Event_AGENT_ADDED:
				a := e.GetAgentAdded().GetAgent()
				fmt.Println(t.String(), a.GetAgentInfo().ID, a.GetAgentInfo().Hostname, a.GetAgentInfo().Attributes)
			case master.Event_AGENT_REMOVED:
				a := e.GetAgentRemoved()
				fmt.Println(t.String(), a.GetAgentID())
			case master.Event_FRAMEWORK_ADDED:
				fw := e.GetFrameworkAdded().GetFramework()
				fmt.Println(t.String(), fw.GetFrameworkInfo().ID, fw.GetFrameworkInfo().Name, fw.GetFrameworkInfo().Role)
			case master.Event_FRAMEWORK_UPDATED:
				fw := e.GetFrameworkUpdated().GetFramework()
				fmt.Println(t.String(), fw.GetFrameworkInfo().ID, fw.GetFrameworkInfo().Name, fw.GetFrameworkInfo().Role)
			case master.Event_FRAMEWORK_REMOVED:
				fw := e.GetFrameworkRemoved()
				fmt.Println(t.String(), fw.GetFrameworkInfo().ID, fw.GetFrameworkInfo().Name, fw.GetFrameworkInfo().Role)
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
