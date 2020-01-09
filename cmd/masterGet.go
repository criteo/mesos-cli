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
	"os"
	"strings"
	"time"

	"github.com/mesos/mesos-go/api/v1/lib/master/calls"

	"github.com/mesos/mesos-go/api/v1/lib/master"
	"github.com/olekukonko/tablewriter"

	"github.com/spf13/cobra"
)

type masterGetOptions struct {
	timeout time.Duration
	json    bool
}

var masterGetOpts = masterGetOptions{}

var masterGetCalls = MasterCallsDef{
	"health": MasterCallDef{
		call: calls.GetHealth,
		desc: "Health status of master",
		json: func(r *master.Response) ([]byte, error) {
			return json.MarshalIndent(r.GetGetHealth(), "", "  ")
		},
		print: func(r *master.Response) error {
			println(r.GetGetHealth().GetHealthy())
			return nil
		},
	},
	"logging level": MasterCallDef{
		call: calls.GetLoggingLevel,
		desc: "Logging level (by default it’s 0, libprocess uses levels 1, 2, and 3)",
		json: func(r *master.Response) ([]byte, error) {
			return json.MarshalIndent(r.GetGetLoggingLevel(), "", "  ")
		},
		print: func(r *master.Response) error {
			println(r.GetGetLoggingLevel().GetLevel())
			return nil
		},
	},
	"maintenance schedule": MasterCallDef{
		call: calls.GetMaintenanceSchedule,
		desc: "Cluster's maintenance schedule",
		json: func(r *master.Response) ([]byte, error) {
			return json.MarshalIndent(r.GetGetMaintenanceSchedule(), "", "  ")
		},
		print: func(r *master.Response) error {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"agents", "start", "duration"})
			for _, w := range r.GetMaintenanceSchedule.Schedule.Windows {
				agents := []string{}
				for _, m := range w.MachineIDs {
					agents = append(agents, fmt.Sprintf("%s (%s)", m.GetHostname(), m.GetIP()))
				}
				start := time.Unix(0, w.Unavailability.Start.GetNanoseconds())
				duration := time.Duration(w.Unavailability.Duration.GetNanoseconds())
				table.Append([]string{strings.Join(agents, "\n"), start.String(), duration.String()})
			}
			table.SetBorder(false)
			table.Render()
			return nil
		},
	},
	"maintenance status": MasterCallDef{
		call: calls.GetMaintenanceStatus,
		desc: "Cluster's maintenance status",
		json: func(r *master.Response) ([]byte, error) {
			return json.MarshalIndent(r.GetGetMaintenanceStatus(), "", "  ")
		},
		print: func(r *master.Response) error {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"agent", "status", "frameworks"})
			for _, d := range r.GetMaintenanceStatus.Status.DrainingMachines {
				frameworks := []string{}
				for _, f := range d.Statuses {
					frameworks = append(
						frameworks,
						fmt.Sprintf("%s: %s ($s)", f.FrameworkID, f.Status, time.Unix(0, f.Timestamp.GetNanoseconds()).String()))
				}
				table.Append([]string{
					fmt.Sprintf("%s (%s)", d.ID.GetHostname(), d.ID.GetIP()),
					"draining",
					strings.Join(frameworks, "\n"),
				})
			}
			for _, d := range r.GetMaintenanceStatus.Status.DownMachines {
				table.Append([]string{
					fmt.Sprintf("%s (%s)", d.GetHostname(), d.GetIP()),
					"down",
				})
			}
			table.SetBorder(false)
			table.Render()
			return nil
		},
	},
	"": MasterCallDef{
		call: calls.GetMaster,
		desc: "Information about the master",
		json: func(r *master.Response) ([]byte, error) {
			return json.MarshalIndent(r.GetGetMaster(), "", "  ")
		},
	},
	"metrics": MasterCallDef{
		call: func() *master.Call {
			if masterGetOpts.timeout == 0 {
				return calls.GetMetrics(nil)
			} else {
				return calls.GetMetrics(&masterGetOpts.timeout)
			}
		},
		desc: `Snapshot of current metrics to the end user.
		If --timeout is set, it will be used to determine the maximum amount of time the API will take to respond.
		If the timeout is exceeded, some metrics may not be included`,
	},
	"operations": MasterCallDef{
		call: calls.GetOperations,
		desc: "list of all offer operations throughout the cluster, not including LAUNCH or LAUNCH_GROUP operations which can be retrieved from tasks",
	},
	"quota": MasterCallDef{
		call: calls.GetQuota,
		desc: "Cluster's configured quotas",
	},
	"roles": MasterCallDef{
		call: calls.GetRoles,
		desc: "Information about roles",
	},
	"state": MasterCallDef{
		call: calls.GetState,
		desc: "Overall cluster state",
	},
	"tasks": MasterCallDef{
		call: calls.GetTasks,
		desc: "Information about all tasks known to the master",
	},
	"version": MasterCallDef{
		call: calls.GetVersion,
		desc: "Version information",
	},
	"weight": MasterCallDef{
		call: calls.GetWeights,
		desc: "Information about role weights",
	},
	"agents": MasterCallDef{
		call: calls.GetAgents,
		desc: "Information about all agents known to the master",
	},
	"executors": MasterCallDef{
		call: calls.GetExecutors,
		desc: "Information about all executors known to the master",
	},
	"flags": MasterCallDef{
		call: calls.GetFlags,
		desc: "Overall flags configuration",
	},
	"frameworks": MasterCallDef{
		call: calls.GetFrameworks,
		desc: "Information about all frameworks known to the master",
	},
}

// getCmd represents the get command
var masterGetCmd = &cobra.Command{
	Use:   "get [call]",
	Short: "Get informations from master",
	Long:  masterGetCalls.describeCalls(),
	Args:  masterGetCalls.validateArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		key := strings.Join(args, " ")
		resp, err := masterCli.Send(context.Background(), calls.NonStreaming(masterGetCalls[key].call()))
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
		if masterGetOpts.json || masterGetCalls[key].print == nil {
			decode := masterGetCalls[key].json
			if decode == nil {
				decode = func(r *master.Response) ([]byte, error) {
					return json.MarshalIndent(r, "", "  ")
				}
			}
			if j, err := decode(&e); err == nil {
				println(string(j))
			} else {
				return fmt.Errorf("Error marshalling response as JSON: %s", err)
			}
		} else {
			masterGetCalls[key].print(&e)
		}
		return nil
	},
}

func init() {
	masterCmd.AddCommand(masterGetCmd)
	masterGetCmd.Flags().DurationVar(&masterGetOpts.timeout, "timeout", 0, "timeout duration (used for metrics call see --help)")
	masterGetCmd.Flags().BoolVarP(&masterGetOpts.json, "json", "j", false, "json output")
}
