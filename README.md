# Wart Control

## Getting dependencies
Requires a version of go that supports go.mod
- go get


## Get up and running
- Build it
  - `make build`
- Create $HOME/.wart-control.yaml file or use premade one
```
  ---
  redis-address: localhost:6379
  redis-password: ''
  cluster: default

```
  Optionally you can pass in a different config using `--config`.
- Then use
  -  `./bin/wc threads`

## Commands
- proxy
  - provides endpoints for wart control UI
- apply <env file path>
  - applies environment file to cluster
- purge
  - destroys cluster

- threads
  - lists threads
- threads disable <name>
  - stops thread
- thread enable <name>
  - starts thread

- endpoints
  -lists endpoints
- endpoints disabled
  - disables endpoint
- endpoints enable
  - enables endpoint

- warts
  - lists warts
- warts stop <name>
  - stops wart
