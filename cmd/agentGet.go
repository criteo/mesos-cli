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
	"strings"
	"time"

	"github.com/mesos/mesos-go/api/v1/lib/agent"
	"github.com/mesos/mesos-go/api/v1/lib/agent/calls"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type agentGetOptions struct {
	timeout time.Duration
	json    bool
}

var agentGetOpts = agentGetOptions{}

var agentGetCalls = AgentCallsDef{
	"health": AgentCallDef{
		call: calls.GetHealth,
		desc: "Health status of agent",
		json: func(r *agent.Response) ([]byte, error) {
			return json.MarshalIndent(r.GetGetHealth(), "", "  ")
		},
		print: func(r *agent.Response) error {
			fmt.Println(r.GetGetHealth().GetHealthy())
			return nil
		},
	},
	"logging level": AgentCallDef{
		call: calls.GetLoggingLevel,
		desc: "Logging level (by default it’s 0, libprocess uses levels 1, 2, and 3)",
		json: func(r *agent.Response) ([]byte, error) {
			return json.MarshalIndent(r.GetGetLoggingLevel(), "", "  ")
		},
		print: func(r *agent.Response) error {
			fmt.Println(r.GetGetLoggingLevel().GetLevel())
			return nil
		},
	},
	"": AgentCallDef{
		call: calls.GetAgent,
		desc: "Information about the agent",
		json: func(r *agent.Response) ([]byte, error) {
			return json.MarshalIndent(r.GetGetAgent(), "", "  ")
		},
		print: func(r *agent.Response) error {
			table := tablewriter.NewWriter(os.Stdout)
			ai := r.GetGetAgent().GetAgentInfo()
			d := r.GetGetAgent().GetDrainConfig()
			table.Append([]string{"id:", ai.GetID().GetValue()})
			table.Append([]string{"hostname:", ai.GetHostname()})
			table.Append([]string{"port:", fmt.Sprintf("%d", ai.GetPort())})
			table.Append([]string{"max_grace_period:", time.Duration(d.GetMaxGracePeriod().GetNanoseconds()).String()})
			table.Append([]string{"mark_gone:", fmt.Sprintf("%v", d.GetMarkGone())})
			table.SetBorder(false)
			table.SetHeaderLine(false)
			table.SetColumnSeparator("")
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.Render()
			return nil
		},
	},
	"metrics": AgentCallDef{
		call: func() *agent.Call {
			if agentGetOpts.timeout == 0 {
				return calls.GetMetrics(nil)
			} else {
				return calls.GetMetrics(&agentGetOpts.timeout)
			}
		},
		desc: `Snapshot of current metrics to the end user.
			If --timeout is set, it will be used to determine the maximum amount of time the API will take to respond.
			If the timeout is exceeded, some metrics may not be included`,
		json: func(r *agent.Response) ([]byte, error) {
			return json.MarshalIndent(r.GetGetMetrics(), "", "  ")
		},
		print: func(r *agent.Response) error {
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
	"operations": AgentCallDef{
		call: calls.GetOperations,
		desc: "list of all offer operations throughout the cluster, not including LAUNCH or LAUNCH_GROUP operations which can be retrieved from tasks",
		json: func(r *agent.Response) ([]byte, error) {
			return json.MarshalIndent(r.GetGetOperations(), "", "  ")
		},
		print: func(r *agent.Response) error {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"framework", "type", "status"})
			for _, o := range r.GetGetOperations().GetOperations() {
				table.Append([]string{o.GetFrameworkID().GetValue(), o.GetInfo().Type.String(), o.GetLatestStatus().State.String()})
			}
			table.SetBorder(false)
			table.SetHeaderLine(false)
			table.SetColumnSeparator("")
			table.Render()
			return nil
		},
	},
	//TODO handle --show-nested and --show-standalone options
	"containers": AgentCallDef{
		call: func() *agent.Call {
			c := calls.GetContainers()
			t := true
			c.GetContainers = &agent.Call_GetContainers{
				ShowNested:     &t,
				ShowStandalone: &t,
			}
			return c
		},
		desc: "Retrieves information about containers running on this agent. It contains ContainerStatus and ResourceStatistics along with some metadata of the containers.",
		json: func(r *agent.Response) ([]byte, error) {
			return json.MarshalIndent(r.GetGetContainers(), "", "  ")
		},
		print: func(r *agent.Response) error {
			table := tablewriter.NewWriter(os.Stdout)
			//TODO show nesting tree
			table.SetHeader([]string{"framework", "id", "executor_id", "executor_name"})
			for _, c := range r.GetGetContainers().GetContainers() {
				name := c.GetExecutorName()
				if len(name) > 25 {
					name = name[0:25]
					name = name + "..."
				}
				table.Append([]string{
					c.GetFrameworkID().GetValue(),
					c.GetContainerID().Value,
					c.GetExecutorID().GetValue(),
					name})
			}
			table.SetBorder(false)
			table.SetHeaderLine(false)
			table.SetColumnSeparator("")
			table.Render()
			return nil
		},
	},
	"state": AgentCallDef{
		call: calls.GetState,
		desc: "Overall cluster state",
		json: func(r *agent.Response) ([]byte, error) {
			return json.MarshalIndent(r.GetGetState(), "", "  ")
		},
		//print is customzied in init()
	},
	"tasks": AgentCallDef{
		call: calls.GetTasks,
		desc: "Information about all tasks known to the agent",
		json: func(r *agent.Response) ([]byte, error) {
			return json.MarshalIndent(r.GetGetTasks(), "", "  ")
		},
		print: func(r *agent.Response) error {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"framework", "task_id", "type", "state"})
			for _, task := range r.GetGetTasks().GetPendingTasks() {
				table.Append([]string{task.GetFrameworkID().Value, task.GetTaskID().Value, "pending", task.GetState().String()})
			}
			for _, task := range r.GetGetTasks().GetQueuedTasks() {
				table.Append([]string{task.GetFrameworkID().Value, task.GetTaskID().Value, "queued", task.GetState().String()})
			}
			for _, task := range r.GetGetTasks().GetLaunchedTasks() {
				table.Append([]string{task.GetFrameworkID().Value, task.GetTaskID().Value, "launched", task.GetState().String()})
			}
			for _, task := range r.GetGetTasks().GetTerminatedTasks() {
				table.Append([]string{task.GetFrameworkID().Value, task.GetTaskID().Value, "terminated", task.GetState().String()})
			}
			for _, task := range r.GetGetTasks().GetCompletedTasks() {
				table.Append([]string{task.GetFrameworkID().Value, task.GetTaskID().Value, "completed", task.GetState().String()})
			}
			table.SetBorder(false)
			table.SetHeaderLine(false)
			table.SetColumnSeparator("")
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.Render()
			return nil
		},
	},
	"version": AgentCallDef{
		call: calls.GetVersion,
		desc: "Version information",
		json: func(r *agent.Response) ([]byte, error) {
			return json.MarshalIndent(r.GetGetVersion(), "", "  ")
		},
		print: func(r *agent.Response) error {
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
	"executors": AgentCallDef{
		call: calls.GetExecutors,
		desc: "Information about all executors known to the agent",
		json: func(r *agent.Response) ([]byte, error) {
			return json.MarshalIndent(r.GetGetExecutors(), "", "  ")
		},
		print: func(r *agent.Response) error {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"framework", "id", "name"})
			for _, e := range r.GetGetExecutors().GetExecutors() {
				ei := e.GetExecutorInfo()
				table.Append([]string{
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
	"flags": AgentCallDef{
		call: calls.GetFlags,
		desc: "Overall flags configuration",
		json: func(r *agent.Response) ([]byte, error) {
			return json.MarshalIndent(r.GetGetFlags(), "", "  ")
		},
		print: func(r *agent.Response) error {
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
	"frameworks": AgentCallDef{
		call: calls.GetFrameworks,
		desc: "Information about all frameworks known to the agent",
		json: func(r *agent.Response) ([]byte, error) {
			return json.MarshalIndent(r.GetGetFrameworks(), "", "  ")
		},
		print: func(r *agent.Response) error {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"id", "name", "roles", "principal"})
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

var agentGetCmd = &cobra.Command{
	Use:   "get [call]",
	Short: "Get informations from agent",
	Long:  agentGetCalls.describeCalls(),
	Args:  agentGetCalls.validateArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		key := strings.Join(args, " ")
		resp, err := agentCli.Send(context.Background(), calls.NonStreaming(agentGetCalls[key].call()))
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
		if agentGetOpts.json || agentGetCalls[key].print == nil {
			decode := agentGetCalls[key].json
			if decode == nil {
				decode = func(r *agent.Response) ([]byte, error) {
					return json.MarshalIndent(r, "", "  ")
				}
			}
			if j, err := decode(&e); err == nil {
				fmt.Println(string(j))
			} else {
				return fmt.Errorf("Error marshalling response as JSON: %s", err)
			}
		} else {
			agentGetCalls[key].print(&e)
		}
		return nil
	},
}

func init() {
	agentCmd.AddCommand(agentGetCmd)
	agentGetCmd.Flags().DurationVar(&agentGetOpts.timeout, "timeout", 0, "timeout duration (used for metrics call see --help)")
	agentGetCmd.Flags().BoolVarP(&agentGetOpts.json, "json", "j", false, "json output")

	agentGetCmd.SetUsageTemplate(agentSubCommandUsageTemplate)

	// GetState calls other actions
	stateCall := agentGetCalls["state"]
	stateCall.print = func(r *agent.Response) error {
		fr := agent.Response{
			GetFrameworks: r.GetGetState().GetGetFrameworks(),
			GetExecutors:  r.GetGetState().GetGetExecutors(),
			GetTasks:      r.GetGetState().GetGetTasks(),
		}
		for _, call := range []string{"frameworks", "executors", "tasks"} {
			fmt.Printf("\nState of %s:\n", call)
			if err := agentGetCalls[call].print(&fr); err != nil {
				return err
			}
		}
		return nil
	}
	agentGetCalls["state"] = stateCall
}
