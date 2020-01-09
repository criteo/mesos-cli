Mesos CLI
=========

CLI to interact with Mesos clusters, using the HTTP v1 operator API.

Install
-----

`go get github.com/criteo/mesos-cli`

Features
-----

Still early version

- [x] Master API
  - [x] Get calls
  - [ ] Other calls
  - [x] Watch events
- [ ] Agent API

Usage
----

```
$ mesos-cli  --help
mesos-cli is a command line interface (CLI) that can be used
to interact with Apache Mesos clusters

Usage:
  mesos-cli [command]

Available Commands:
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
