package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/go-redis/redis"
)

func main() {
	cmd := ""
	if len(os.Args) > 0 {
		cmd = os.Args[1]
	}

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	switch cmd {
	case "":
	case "threads":
		cmd2 := ""
		if len(os.Args) > 2 {
			cmd2 = os.Args[2]
		}
		switch cmd2 {
		case "":
			threads := client.Keys("default:Threads:*").Val()
			for i := range threads {
				fmt.Println(threads[i])
				fmt.Println("-", "Status", client.HGet(threads[i], "Status").Val())
				fmt.Println("-", "State", client.HGet(threads[i], "State").Val())
				fmt.Println("-", "Owner", client.HGet(threads[i], "Owner").Val())
				fmt.Println("-", "Heartbeat", client.HGet(threads[i], "Heartbeat").Val())
			}
		case "stop":
			if len(os.Args) < 4 {
				fmt.Println("invalid")
			} else {
				key := os.Args[3]
				client.HSet(key, "Status", "disabled")
			}

		case "start":
			if len(os.Args) < 4 {
				fmt.Println("invalid")
			} else {
				key := os.Args[3]
				client.HSet(key, "Status", "enabled")
			}

		case "load":
			if len(os.Args) < 4 {
				fmt.Println("invalid")
			} else {
				scripts := os.Args[3]

				scriptArray := strings.Split(scripts, ",")
				for i := range scriptArray {
					scriptName := scriptArray[i]
					fBytes, err := ioutil.ReadFile(scriptName)
					if err != nil {
						fmt.Println("File load failed")
						return
					}
					client.HSet("default:Threads:"+scriptName, "Source", string(fBytes))
					client.HSet("default:Threads:"+scriptName, "Status", "enabled")
					client.HSet("default:Threads:"+scriptName, "State", "stopped")
					client.HSet("default:Threads:"+scriptName, "Heartbeat", 0)
					client.HSet("default:Threads:"+scriptName, "Owner", "")
				}

			}
		}

	case "endpoints":
		cmd2 := ""
		if len(os.Args) > 2 {
			cmd2 = os.Args[2]
		}
		switch cmd2 {
		case "":
			threads := client.Keys("default:Endpoints:*").Val()
			for i := range threads {
				fmt.Println(threads[i])
				fmt.Println("-", "Status", client.HGet(threads[i], "Status").Val())
			}
		case "stop":
			if len(os.Args) < 4 {
				fmt.Println("invalid")
			} else {
				key := os.Args[3]
				client.HSet(key, "Status", "disabled")
			}

		case "start":
			if len(os.Args) < 4 {
				fmt.Println("invalid")
			} else {
				key := os.Args[3]
				client.HSet(key, "Status", "enabled")
			}

		case "load":
			if len(os.Args) < 4 {
				fmt.Println("invalid")
			} else {
				scripts := os.Args[3]

				scriptArray := strings.Split(scripts, ",")
				for i := range scriptArray {
					scriptName := scriptArray[i]
					fBytes, err := ioutil.ReadFile(scriptName)
					if err != nil {
						fmt.Println("File load failed")
						return
					}
					client.HSet("default:Endpoints:"+scriptName, "Source", string(fBytes))
					client.HSet("default:Endpoints:"+scriptName, "Status", "enabled")
				}

			}
		}

	case "warts":

		cmd2 := ""
		if len(os.Args) > 2 {
			cmd2 = os.Args[2]
		}
		switch cmd2 {
		case "":
			warts := client.Keys("default:Warts:*").Val()
			for i := range warts {
				s := strings.Split(warts[i], ":")
				if s[len(s)-1] != "Health" {
					fmt.Println(warts[i])
					fmt.Println("-", "Status", client.HGet(warts[i], "Status").Val())
					fmt.Println("-", "State", client.HGet(warts[i], "State").Val())
					fmt.Println("-", "Heartbeat", client.HGet(warts[i], "Heartbeat").Val())
					fmt.Println("-", "Health")
					fmt.Println("-", "-", "CPU", client.HGet(warts[i]+":Health", "cpu").Val())
					fmt.Println("-", "-", "Mem", client.HGet(warts[i]+":Health", "memory").Val())
				}
			}
		case "stop":
			if len(os.Args) < 4 {
				fmt.Println("invalid")
			} else {
				key := "default:Warts:" + os.Args[3]
				client.HSet(key, "Status", "disabled")
			}

		}
	}
}
