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
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	mesos "github.com/mesos/mesos-go/api/v1/lib"
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
			fmt.Println(r.GetGetHealth().GetHealthy())
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
			fmt.Println(r.GetGetLoggingLevel().GetLevel())
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
			table.SetHeaderLine(false)
			table.SetColumnSeparator("")
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
						fmt.Sprintf("%s: %s (%s)", f.FrameworkID, f.Status, time.Unix(0, f.Timestamp.GetNanoseconds()).String()))
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
			table.SetHeaderLine(false)
			table.SetColumnSeparator("")
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
		print: func(r *master.Response) error {
			table := tablewriter.NewWriter(os.Stdout)
			table.Append([]string{"id:", r.GetGetMaster().GetMasterInfo().GetID()})
			table.Append([]string{"hostname:", r.GetGetMaster().GetMasterInfo().GetHostname()})
			table.Append([]string{"ip:", r.GetGetMaster().GetMasterInfo().GetAddress().GetIP()})
			table.Append([]string{"port:", fmt.Sprintf("%d", r.GetGetMaster().GetMasterInfo().GetAddress().GetPort())})
			table.Append([]string{"version:", r.GetGetMaster().GetMasterInfo().GetVersion()})
			table.SetBorder(false)
			table.SetHeaderLine(false)
			table.SetColumnSeparator("")
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.Render()
			return nil
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
		json: func(r *master.Response) ([]byte, error) {
			return json.MarshalIndent(r.GetGetMetrics(), "", "  ")
		},
		print: func(r *master.Response) error {
			table := tablewriter.NewWriter(os.Stdout)
			for _, m := range r.GetGetMetrics().GetMetrics() {
				table.Append([]string{m.GetName(), fmt.Sprintf("%f", m.GetValue())})
			}
			table.SetBorder(false)
			table.SetHeaderLine(false)
			table.SetColumnSeparator("")
			table.Render()
			return nil
		},
	},
	"operations": MasterCallDef{
		call: calls.GetOperations,
		desc: "list of all offer operations throughout the cluster, not including LAUNCH or LAUNCH_GROUP operations which can be retrieved from tasks",
		json: func(r *master.Response) ([]byte, error) {
			return json.MarshalIndent(r.GetGetOperations(), "", "  ")
		},
		print: func(r *master.Response) error {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"agent", "framework", "type", "status"})
			for _, o := range r.GetGetOperations().GetOperations() {
				table.Append([]string{o.GetAgentID().GetValue(), o.GetFrameworkID().GetValue(), o.GetInfo().Type.String(), o.GetLatestStatus().State.String()})
			}
			table.SetBorder(false)
			table.SetHeaderLine(false)
			table.SetColumnSeparator("")
			table.Render()
			return nil
		},
	},
	"quota": MasterCallDef{
		call: calls.GetQuota,
		desc: "Cluster's configured quotas",
		json: func(r *master.Response) ([]byte, error) {
			return json.MarshalIndent(r.GetGetQuota(), "", "  ")
		},
		print: func(r *master.Response) error {
			table := tablewriter.NewWriter(os.Stdout)
			quotas := map[string]map[string][]float64{}
			resourcesMap := map[string]bool{}
			if len(r.GetGetQuota().GetStatus().Configs) > 0 {
				for _, c := range r.GetGetQuota().GetStatus().Configs {
					r := c.GetRole()
					quotas[r] = map[string][]float64{}
					for name := range c.GetGuarantees() {
						if _, ok := quotas[r][name]; !ok {
							quotas[r][name] = make([]float64, 2)
							resourcesMap[name] = true
						}
						quotas[r][name][0] = c.GetGuarantees()[name].Value
					}
					for name := range c.GetLimits() {
						if _, ok := quotas[r][name]; !ok {
							quotas[r][name] = make([]float64, 2)
							resourcesMap[name] = true
						}
						quotas[r][name][1] = c.GetGuarantees()[name].Value
					}
				}
			} else {
				for _, c := range r.GetGetQuota().GetStatus().Infos {
					r := c.GetRole()
					quotas[r] = map[string][]float64{}
					for _, res := range c.GetGuarantee() {
						if _, ok := quotas[r][res.GetName()]; !ok {
							quotas[r][res.GetName()] = make([]float64, 2)
							resourcesMap[res.GetName()] = true
						}
						quotas[r][res.GetName()][0] = res.GetScalar().GetValue()
					}
				}
			}
			resources := []string{}
			for n := range resourcesMap {
				resources = append(resources, n)
			}
			sort.Strings(resources)
			header := []string{"role"}
			for _, n := range resources {
				header = append(header, n)
			}

			table.SetHeader(header)
			for r := range quotas {
				q := []string{r}
				for _, n := range resources {
					gl, ok := quotas[r][n]
					if ok {
						q = append(q, fmt.Sprintf("%.0f-%.0f", gl[0], gl[1]))
					} else {
						q = append(q, "")
					}
				}
				table.Append(q)
			}
			table.SetBorder(false)
			table.SetHeaderLine(false)
			table.SetColumnSeparator("")
			table.Render()
			return nil
		},
	},
	"roles": MasterCallDef{
		call: calls.GetRoles,
		desc: "Information about roles",
		json: func(r *master.Response) ([]byte, error) {
			return json.MarshalIndent(r.GetGetRoles(), "", "  ")
		},
		print: func(r *master.Response) error {
			table := tablewriter.NewWriter(os.Stdout)
			resourcesMap := map[string]bool{}
			roleResources := map[string]map[string]string{}
			for _, role := range r.GetGetRoles().GetRoles() {
				roleResources[role.GetName()] = map[string]string{}
				for _, res := range role.GetResources() {
					val := ""
					switch res.GetType() {
					case mesos.SCALAR:
						val = fmt.Sprintf("%.2f", res.GetScalar().Value)
					case mesos.RANGES:
						ranges := []string{}
						for _, ra := range res.GetRanges().Range {
							ranges = append(ranges, fmt.Sprintf("%d-%d", ra.GetBegin(), ra.GetEnd()))
						}
						if len(ranges) > 4 {
							ranges = ranges[0:4]
							ranges = append(ranges, "...")
						}
						val = fmt.Sprintf("[%s]", strings.Join(ranges, ","))
					case mesos.SET:
						val = fmt.Sprintf("[%s]", strings.Join(res.GetSet().Item, ","))
					default:
						return fmt.Errorf("Role %s, has unknown resource type for %s: %s", role.GetName(), res.GetName(), res.GetType())
					}
					roleResources[role.GetName()][res.GetName()] = val
					resourcesMap[res.GetName()] = true
				}
			}

			resources := []string{}
			for n := range resourcesMap {
				resources = append(resources, n)
			}
			sort.Strings(resources)
			header := []string{"role"}
			for _, n := range resources {
				header = append(header, n)
			}
			table.SetHeader(header)
			for _, role := range r.GetGetRoles().GetRoles() {
				srole := []string{role.GetName(), fmt.Sprintf("%.1f", role.GetWeight())}
				for _, name := range resources {
					if v, ok := roleResources[role.GetName()][name]; ok {
						srole = append(srole, v)
					} else {
						srole = append(srole, "")
					}
				}
				table.Append(srole)
			}
			table.SetBorder(false)
			table.SetHeaderLine(false)
			table.SetColumnSeparator("")
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.Render()
			return nil
		},
	},
	"state": MasterCallDef{
		call: calls.GetState,
		desc: "Overall cluster state",
		json: func(r *master.Response) ([]byte, error) {
			return json.MarshalIndent(r.GetGetState(), "", "  ")
		},
		//print is customzied in init()
	},
	"tasks": MasterCallDef{
		call: calls.GetTasks,
		desc: "Information about all tasks known to the master",
		json: func(r *master.Response) ([]byte, error) {
			return json.MarshalIndent(r.GetGetTasks(), "", "  ")
		},
		print: func(r *master.Response) error {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"agent", "framework", "task_id", "type", "state"})
			for _, task := range r.GetGetTasks().GetPendingTasks() {
				table.Append([]string{task.GetAgentID().Value, task.GetFrameworkID().Value, task.GetTaskID().Value, "pending", task.GetState().String()})
			}
			for _, task := range r.GetGetTasks().GetTasks() {
				table.Append([]string{task.GetAgentID().Value, task.GetFrameworkID().Value, task.GetTaskID().Value, "launched", task.GetState().String()})
			}
			for _, task := range r.GetGetTasks().GetCompletedTasks() {
				table.Append([]string{task.GetAgentID().Value, task.GetFrameworkID().Value, task.GetTaskID().Value, "completed", task.GetState().String()})
			}
			for _, task := range r.GetGetTasks().GetUnreachableTasks() {
				table.Append([]string{task.GetAgentID().Value, task.GetFrameworkID().Value, task.GetTaskID().Value, "unreachable", task.GetState().String()})
			}
			table.SetBorder(false)
			table.SetHeaderLine(false)
			table.SetColumnSeparator("")
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.Render()
			return nil
		},
	},
	"version": MasterCallDef{
		call: calls.GetVersion,
		desc: "Version information",
		json: func(r *master.Response) ([]byte, error) {
			return json.MarshalIndent(r.GetGetVersion(), "", "  ")
		},
		print: func(r *master.Response) error {
			table := tablewriter.NewWriter(os.Stdout)
			table.Append([]string{"version:", r.GetGetVersion().VersionInfo.GetVersion()})
			table.Append([]string{"build_date:", r.GetGetVersion().VersionInfo.GetBuildDate()})
			table.Append([]string{"build_time:", fmt.Sprintf("%v", int64(r.GetGetVersion().VersionInfo.GetBuildTime()))})
			table.Append([]string{"build_user:", r.GetGetVersion().VersionInfo.GetBuildUser()})
			table.Append([]string{"git_branch:", r.GetGetVersion().VersionInfo.GetGitBranch()})
			table.Append([]string{"git_sha:", r.GetGetVersion().VersionInfo.GetGitSHA()})
			table.Append([]string{"git_tag:", r.GetGetVersion().VersionInfo.GetGitTag()})
			table.SetBorder(false)
			table.SetHeaderLine(false)
			table.SetColumnSeparator("")
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.Render()
			return nil
		},
	},
	"weight": MasterCallDef{
		call: calls.GetWeights,
		desc: "Information about role weights",
		json: func(r *master.Response) ([]byte, error) {
			return json.MarshalIndent(r.GetGetWeights(), "", "  ")
		},
		print: func(r *master.Response) error {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"role", "weight"})
			for _, w := range r.GetGetWeights().GetWeightInfos() {
				table.Append([]string{w.GetRole(), fmt.Sprintf("%.1f", w.GetWeight())})
			}
			table.SetBorder(false)
			table.SetHeaderLine(false)
			table.SetColumnSeparator("")
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.Render()
			return nil
		},
	},
	"agents": MasterCallDef{
		call: calls.GetAgents,
		desc: "Information about all agents known to the master",
		json: func(r *master.Response) ([]byte, error) {
			return json.MarshalIndent(r.GetGetAgents(), "", "  ")
		},
		print: func(r *master.Response) error {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"id", "hostname", "version", "registered"})
			for _, a := range r.GetGetAgents().GetAgents() {
				table.Append([]string{
					a.GetAgentInfo().ID.GetValue(),
					a.GetAgentInfo().Hostname,
					a.GetVersion(),
					time.Unix(0, a.GetRegisteredTime().GetNanoseconds()).String(),
				})
			}
			for _, a := range r.GetGetAgents().GetRecoveredAgents() {
				table.Append([]string{a.GetID().GetValue(), a.GetHostname(), "", "unregistered"})
			}
			table.SetBorder(false)
			table.SetHeaderLine(false)
			table.SetColumnSeparator("")
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.Render()
			return nil
		},
	},
	"executors": MasterCallDef{
		call: calls.GetExecutors,
		desc: "Information about all executors known to the master",
		json: func(r *master.Response) ([]byte, error) {
			return json.MarshalIndent(r.GetGetExecutors(), "", "  ")
		},
		print: func(r *master.Response) error {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"agent", "framework", "id", "name"})
			for _, e := range r.GetGetExecutors().GetExecutors() {
				ei := e.GetExecutorInfo()
				table.Append([]string{
					e.GetAgentID().Value,
					ei.FrameworkID.GetValue(),
					ei.ExecutorID.Value,
					ei.GetName(),
				})
			}
			table.SetBorder(false)
			table.SetHeaderLine(false)
			table.SetColumnSeparator("")
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.Render()
			return nil
		},
	},
	"flags": MasterCallDef{
		call: calls.GetFlags,
		desc: "Overall flags configuration",
		json: func(r *master.Response) ([]byte, error) {
			return json.MarshalIndent(r.GetGetFlags(), "", "  ")
		},
		print: func(r *master.Response) error {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"name", "value"})
			for _, f := range r.GetGetFlags().GetFlags() {
				table.Append([]string{f.GetName(), f.GetValue()})
			}
			table.SetBorder(false)
			table.SetHeaderLine(false)
			table.SetColumnSeparator("")
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.Render()
			return nil
		},
	},
	"frameworks": MasterCallDef{
		call: calls.GetFrameworks,
		desc: "Information about all frameworks known to the master",
		json: func(r *master.Response) ([]byte, error) {
			return json.MarshalIndent(r.GetGetFrameworks(), "", "  ")
		},
		print: func(r *master.Response) error {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"id", "name", "roles", "principal", "active", "connected", "recovered"})
			for _, f := range r.GetGetFrameworks().GetFrameworks() {
				fi := f.GetFrameworkInfo()
				roles := fi.GetRole()
				if len(fi.GetRoles()) > 0 {
					roles = strings.Join(fi.GetRoles(), ",")
				}
				table.Append([]string{
					fi.GetID().GetValue(),
					fi.GetName(),
					roles,
					fi.GetPrincipal(),
					fmt.Sprintf("%v", f.GetActive()),
					fmt.Sprintf("%v", f.GetConnected()),
					fmt.Sprintf("%v", f.GetRecovered()),
				})
			}
			for _, f := range r.GetGetFrameworks().GetCompletedFrameworks() {
				fi := f.GetFrameworkInfo()
				roles := fi.GetRole()
				if len(fi.GetRoles()) > 0 {
					roles = strings.Join(fi.GetRoles(), ",")
				}
				table.Append([]string{
					fi.GetID().GetValue(),
					fi.GetName(),
					roles,
					fi.GetPrincipal(),
					fmt.Sprintf("%v", f.GetActive()),
					fmt.Sprintf("%v", f.GetConnected()),
					fmt.Sprintf("%v", f.GetRecovered()),
				})
			}
			table.SetBorder(false)
			table.SetHeaderLine(false)
			table.SetColumnSeparator("")
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.Render()
			return nil
		},
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
				fmt.Println(string(j))
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

	// GetState calls other actions
	stateCall := masterGetCalls["state"]
	stateCall.print = func(r *master.Response) error {
		fr := master.Response{
			GetAgents:     r.GetGetState().GetGetAgents(),
			GetFrameworks: r.GetGetState().GetGetFrameworks(),
			GetExecutors:  r.GetGetState().GetGetExecutors(),
			GetTasks:      r.GetGetState().GetGetTasks(),
		}
		for _, call := range []string{"agents", "frameworks", "executors", "tasks"} {
			fmt.Printf("\nState of %s:\n", call)
			if err := masterGetCalls[call].print(&fr); err != nil {
				return err
			}
		}
		return nil
	}
	masterGetCalls["state"] = stateCall
}
