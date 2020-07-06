# Wart Control

## Getting dependencies
- install deps
  - https://github.com/golang/dep
- install dependencies
  - `deps ensure`

## Get up and running
- Build it
  - `make build`
- Create .config file or use premade one
  ```
  {
  "redis-address":"localhost:6379",
  "redis-password":"",
  "cluster":"default"
  }
```
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
