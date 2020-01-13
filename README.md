Mesos CLI
=========

CLI to interact with Mesos clusters, using the HTTP v1 operator API.

Install
-----

`go get github.com/criteo/mesos-cli`

Example:
-----

```
$ mesos-cli agent mesos-agent123 get state

State of frameworks:
                     ID                            NAME         ROLES      PRINCIPAL      
  20151013-103744-17171722-5050-10359-0001   marathon          default  marathon          
  6b791ac2-aab2-4b26-9c00-5a7f43661c6f-0000  marathon_demo     demo     demo_marathon     
  76778f27-2cbf-47d5-ba70-87c520b47f81-0012  marathon_supply   supply   supply_marathon   
  e8455271-66d3-447e-bd4f-3d9d9a78ba14-0000  Flink             flink    flink             
  6251f5d7-9dbd-4348-8ed0-26664cfbd2e1-0000  marathon_creator  creator  creator_marathon  
  a5e9c8ba-8268-4cfb-95ae-e0c6d786fc71-0000  Aurora            aurora   aurora            

State of executors:
                  FRAMEWORK                            ID                               NAME                                             
[...]              

State of tasks:
                  FRAMEWORK                                               TASK ID                                     TYPE         STATE      
  20151013-103744-17171722-5050-10359-0001   observability_test.dacaab82-3620-11ea-a592-d8c497a5d9d6               queued     TASK_STAGING   
  e8455271-66d3-447e-bd4f-3d9d9a78ba14-0000  taskmanager-09699                                                     launched   TASK_RUNNING   
  6b791ac2-aab2-4b26-9c00-5a7f43661c6f-0000  incubator_slow-start.3e77203d-2d71-11ea-980a-1618f26d585d             launched   TASK_RUNNING   
  a5e9c8ba-8268-4cfb-95ae-e0c6d786fc71-0000  bi-dtest-0-d884980c-9965-4be3-9813-e48cc81e6295                       completed  TASK_FAILED     
  a5e9c8ba-8268-4cfb-95ae-e0c6d786fc71-0000  foo-bar-bazr-0-f8fc00da-69e9-402d-b865-4aa631f16059                   completed  TASK_FINISHED  
  a5e9c8ba-8268-4cfb-95ae-e0c6d786fc71-0000  foo-bar-bazr-uploader-0-12efca64-7157-4459-90eb-f07c11ad5f9f          completed  TASK_FINISHED  
  [...]
```

Features
-----

- [x] Master API
  - [x] Get information (version, frameworks, tasks, state, operations...etc)
  - [x] Watch events
  - [x] List/Read files
  - [x] Get/Set logging level
  - [x] Get/Set Quota
  - [ ] Update Quota (requires Mesos 1.9+)
  - [x] Update weights maintenance schedules
  - [ ] Start/Stop maintenance
  - [ ] Reserve/Unreserve resources
  - [ ] Create/Destroy/Grow/Shrink volumes
  - [ ] Mark agent gone
  - [ ] Drain/Deactivate/Reactivate agent
- [x] Agent API
  - [x] Get information (version, frameworks, tasks, containers...etc)
  - [x] Launch nested containers (with and without interactive/TTY)
  - [ ] Launch detached nested or standalone containers
  - [ ] Wait/Kill/Remove container
  - [x] List/Read files
  - [x] Get/Set logging level
  - [ ] Add/Update/Remove resource providers
  - [ ] Mark resource provider gone
  - [ ] Prune images

Usage
----

```
$ mesos-cli  --help
mesos-cli is a command line interface (CLI) that can be used
to interact with Apache Mesos clusters

Usage:
  mesos-cli [command]

Available Commands:
  agent      Interact with Mesos Agent
  help        Help about any command
  master      Interact with Mesos Master

Flags:
      --config string      config file (default is $HOME/.mesos-cli.yaml)
  -h, --help               help for mesos-cli
  -p, --principal string   Mesos Principal
  -s, --secret string      Mesos Secret
  -v, --verbose            verbose output

Use "mesos-cli [command] --help" for more information about a command.
```
