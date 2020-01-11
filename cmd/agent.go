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

	"github.com/mesos/mesos-go/api/v1/lib/agent"
	"github.com/mesos/mesos-go/api/v1/lib/agent/calls"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli/httpagent"
	masterH "github.com/mesos/mesos-go/api/v1/lib/httpcli/httpmaster"
	"github.com/mesos/mesos-go/api/v1/lib/master"
	masterC "github.com/mesos/mesos-go/api/v1/lib/master/calls"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type agentOptions struct {
	name string
}

var agentOpts = agentOptions{}

var agentCli calls.Sender

var agentCmd = &cobra.Command{
	Use:     "agent [agent]",
	Short:   "Interact with Mesos Agent",
	Long:    `Interact with Mesos Agent`,
	Example: "agent agent001 get tasks",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if agentOpts.name == "" {
			return fmt.Errorf("Missing agent argument")
		}

		var err error
		if !strings.Contains(agentOpts.name, ":") {
			agentCli, err = getCli(fmt.Sprintf("%s:%d", agentOpts.name, viper.GetUint32("agent.port")))
			if err != nil {
				agentCli, err = getCli(agentOpts.name)
			}
		} else {
			agentCli, err = getCli(agentOpts.name)
		}

		if err != nil && viper.IsSet("master.url") && viper.GetString("master.url") != "" {
			if verbose {
				fmt.Printf("Trying to contact master at %s\n", viper.GetString("master.url"))
			}
			var auth httpcli.ConfigOpt
			if viper.IsSet("principal") && viper.GetString("principal") != "" {
				auth = httpcli.BasicAuth(
					viper.GetString("principal"),
					viper.GetString("secret"))
			}
			var masterCli = masterH.NewSender(
				httpcli.New(
					httpcli.Endpoint(fmt.Sprintf("%s/api/v1", viper.GetString("master.url"))),
					httpcli.Do(httpcli.With(auth))).Send)
			resp, merr := masterCli.Send(context.Background(), masterC.NonStreaming(masterC.GetAgents()))
			if merr == nil {
				var r master.Response
				merr = resp.Decode(&r)
				if merr == nil {
					for _, a := range r.GetGetAgents().GetAgents() {
						if strings.HasPrefix(a.GetAgentInfo().ID.Value, agentOpts.name) ||
							strings.HasPrefix(a.GetAgentInfo().Hostname, agentOpts.name) {
							agentCli, err = getCli(fmt.Sprintf("%s:%d", a.GetAgentInfo().Hostname, *a.GetAgentInfo().Port))
							if err == nil {
								return nil
							}
						}
					}
					merr = fmt.Errorf("Unable to find agent with id or hostname starting with %s", agentOpts.name)
				}
			}
			if merr != nil {
				return fmt.Errorf("Unable to reach agent: %s and master: %s", err, merr)
			}
		}
		if err != nil {
			return fmt.Errorf("Unable to reach agent: %s", err)
		}
		return nil
	},
}

var agentSubCommandUsageTemplate = `Usage:{{if .Runnable}}
{{.Parent.Parent.CommandPath}} agent [agent] {{.Use}}{{end}}{{if gt (len .Aliases) 0}}

Aliases:
{{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
{{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
{{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

func init() {
	rootCmd.AddCommand(agentCmd)

	agentCmd.PersistentFlags().Uint32("agent-port", 5051, "Mesos agent port if not specified (default 5051)")
	viper.BindPFlag("agent.port", agentCmd.PersistentFlags().Lookup("agent-port"))

	agentCmd.SetUsageTemplate(`Usage:
	{{.CommandPath}} [agent] [command]{{if gt (len .Aliases) 0}}
  
  Aliases:
	{{.NameAndAliases}}{{end}}{{if .HasExample}}
  
  Examples:
  {{.Example}}{{end}}
  
  Agent:
	hostname:port
	hostname            with --agent-port set
	agent id            with master.url configuration
	hostname prefix     with master.url configuration{{if .HasAvailableSubCommands}}
  
  Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
	{{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}
  
  Flags:
  {{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}
  
  Global Flags:
  {{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}
  
  Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
	{{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}
  
  Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
  `)
}

func getCli(url string) (calls.Sender, error) {
	if verbose {
		fmt.Printf("Trying agent %s\n", url)
	}
	var auth httpcli.ConfigOpt
	if viper.IsSet("principal") && viper.GetString("principal") != "" {
		auth = httpcli.BasicAuth(
			viper.GetString("principal"),
			viper.GetString("secret"))
	}
	var cli = httpagent.NewSender(
		httpcli.New(
			httpcli.Endpoint(fmt.Sprintf("http://%s/api/v1", url)),
			httpcli.Do(httpcli.With(auth))).Send)
	// Call GET_HEALTH to make sure agent is reachable
	_, err := cli.Send(context.Background(), calls.NonStreaming(calls.GetHealth()))
	return cli, err
}

type AgentCallDef struct {
	call  func() *agent.Call
	desc  string
	json  func(r *agent.Response) ([]byte, error)
	print func(r *agent.Response) error
}

type AgentCallsDef map[string]AgentCallDef

func (m AgentCallsDef) describeCalls() string {
	padding := 0
	for key := range m {
		if len(key) > padding {
			padding = len(key)
		}
	}
	template := fmt.Sprintf("  %%-%ds %%s\n", padding)
	desc := ""
	if _, ok := m[""]; ok {
		desc += fmt.Sprintf("If no call specified: %s\n", m[""].desc)
		desc += "\n"
	}

	desc += "Possible call values:\n"
	for key := range m {
		if key != "" {
			desc += fmt.Sprintf(template, key, m[key].desc)
		}
	}

	return desc
}

func (m AgentCallsDef) validateArgs(cmd *cobra.Command, args []string) error {
	var key = strings.Join(args, " ")
	if _, ok := m[key]; ok {
		return nil
	} else {
		return fmt.Errorf("invalid arguments: %s try with --help", key)
	}
}
