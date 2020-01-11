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
	"os"

	"github.com/google/uuid"
	mesos "github.com/mesos/mesos-go/api/v1/lib"
	"github.com/mesos/mesos-go/api/v1/lib/agent"
	"github.com/mesos/mesos-go/api/v1/lib/agent/calls"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

type agentLaunchOptions struct {
	tty              bool
	interactive      bool
	detach           bool
	containerId      string
	parentContinerId string
}

var agentLaunchOpts = agentLaunchOptions{}

var agentLaunchCmd = &cobra.Command{
	Use:     "launch [flags] [command]",
	Example: "launch -p a3cfce28-bcca-46d2-a23c-8c780246b7ae -ti bash",
	Short:   "Launch container on agent",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var call *agent.Call
		var containerId = mesos.ContainerID{
			Value: agentLaunchOpts.containerId,
		}
		if agentLaunchOpts.parentContinerId != "" {
			containerId.Parent = &mesos.ContainerID{Value: agentLaunchOpts.parentContinerId}
		}
		shell := false
		var commandInfo = &mesos.CommandInfo{
			Shell:     &shell,
			Value:     &args[0],
			Arguments: args,
		}
		var containerInfo = &mesos.ContainerInfo{
			Type:  mesos.ContainerInfo_MESOS.Enum(),
			Mesos: &mesos.ContainerInfo_MesosInfo{},
		}
		if agentLaunchOpts.tty {
			containerInfo.TTYInfo = &mesos.TTYInfo{}
		}
		if agentLaunchOpts.parentContinerId != "" && !agentLaunchOpts.detach {
			call = calls.LaunchNestedContainerSession(containerId, commandInfo, containerInfo)
		} else {
			return fmt.Errorf("Only nested containers are supported (no detach, and parent flag must be set)")
			//call = calls.LaunchContainer(containerId, commandInfo, containerInfo, []mesos.Resource{})
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		resp, err := agentCli.Send(context.Background(), calls.NonStreaming(call))

		if agentLaunchOpts.interactive {
			previousTerminalState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
			//TODO trap SIGWINCH to set TTY via calls.AttachContainerInputTTY()
			if err == nil {
				defer func() {
					terminal.Restore(int(os.Stdin.Fd()), previousTerminalState)
				}()
			} else {
				return fmt.Errorf("Failed to get raw TTY: %s", err.Error())
			}
			interactive(ctx, containerId)
		}

		defer func() {
			if resp != nil {
				resp.Close()
			}
		}()

		go func() {
			for err == nil {
				var e agent.ProcessIO
				resp.Decode(&e)
				if err != nil {
					if err == io.EOF {
						cancel()
						return
					}
					fmt.Printf("Error decoding response: %s", err)
					cancel()
					return
				}
				switch e.GetType() {
				case agent.ProcessIO_DATA:
					var fd *os.File
					switch e.GetData().GetType() {
					case agent.ProcessIO_Data_STDIN:
						fmt.Printf("Received STDIN data, this is not normal: %b", e.GetData().GetData())
						cancel()
						return
					case agent.ProcessIO_Data_STDERR:
						fd = os.Stderr
					case agent.ProcessIO_Data_STDOUT:
						fd = os.Stdout
					default:
						fmt.Printf("Received unknown data type: %s with data: %b", e.GetData().GetType(), e.GetData().GetData())
						cancel()
						return

					}
					fd.Write(e.GetData().GetData())
				case agent.ProcessIO_CONTROL:
					if e.GetControl().GetType() != agent.ProcessIO_Control_HEARTBEAT {
						fmt.Printf("Received unknown Control: %s", e.GetControl().GetType())
						cancel()
						return
					}
				default:
					fmt.Printf("Received unknown ProcessIO type: %s", e.GetType())
					cancel()
					return
				}
			}
		}()
		<-ctx.Done()

		return err
	},
}

func init() {
	agentCmd.AddCommand(agentLaunchCmd)
	agentLaunchCmd.Flags().BoolVarP(&agentLaunchOpts.tty, "tty", "t", false, "enable TTY")
	//agentLaunchCmd.Flags().BoolVarP(&agentLaunchOpts.detach, "detach", "d", false, "detach from running container after launch")
	agentLaunchCmd.Flags().BoolVarP(&agentLaunchOpts.interactive, "interactive", "i", false, "interactive run (handle STDIN)")
	agentLaunchCmd.Flags().StringVar(&agentLaunchOpts.containerId, "container-id", uuid.New().String(), "container ID")
	agentLaunchCmd.Flags().StringVarP(&agentLaunchOpts.parentContinerId, "parent", "p", "", "parent container ID")

	agentLaunchCmd.SetUsageTemplate(agentSubCommandUsageTemplate)
}

func interactive(ctx context.Context, containerId mesos.ContainerID) {
	var input = make(chan *agent.Call)
	go func() {
		resp, err := agentCli.Send(context.Background(), calls.FromChan(input))
		defer func() {
			if resp != nil {
				resp.Close()
			}
		}()
		if err != nil {
			fmt.Printf("Error sending STDIN: %s\n", err.Error())
		}
		//TODO send heartbeats
	}()
	go func() {
		input <- calls.AttachContainerInput(containerId)
		inBytes := make([]byte, 1024)
		// escape sequence does not work at the moment
		// TODO
		escapeSequence := "\012~." //<Enter>~.
		escapeIndex := 0
		for true {
			var size int
			var err error
			if size, err = os.Stdin.Read(inBytes); err != nil {
				fmt.Printf("Error sending STDIN: %s\n", err.Error())
				return
			}
			for _, b := range inBytes[0:size] {
				if b == escapeSequence[escapeIndex] {
					escapeIndex++
					if escapeIndex == 2 {
						input <- calls.AttachContainerInputData([]byte("\004")) // EOT
						return
					}
				} else {
					escapeIndex = 0
				}
			}
			input <- calls.AttachContainerInputData(inBytes[0:size])
		}
	}()
}
