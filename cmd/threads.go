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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/cobra"
)

// threadsCmd represents the threads command
var threadsCmd = &cobra.Command{
	Use:   "threads <command> <command args>",
	Short: "A brief description of your command",
	Long: `Controls for threads.
	Commands:
	default - lists threads
	enable <endpoint name> - enables thread
	disable <endpoint name> - disables thread
	load <file path> - loads comma separated list of source files into the cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		redisAddress := "localhost:6379"
		redisPassword := ""

		fBytes, err := ioutil.ReadFile(".config")
		if err == nil {
			var f interface{}
			err2 := json.Unmarshal(fBytes, &f)
			if err2 == nil {
				m := f.(map[string]interface{})
				redisAddress = m["redis-address"].(string)
				redisPassword = m["redis-password"].(string)
			}
		}

		client := redis.NewClient(&redis.Options{
			Addr:     redisAddress,
			Password: redisPassword, // no password set
			DB:       0,             // use default DB
		})
		cmd2 := ""
		if len(os.Args) > 2 {
			cmd2 = os.Args[2]
		}
		switch cmd2 {
		case "":
			threads := client.Keys(ctx, cluster+":Threads:*").Val()
			for i := range threads {
				fmt.Println(threads[i])
				fmt.Println("-", "Status", client.HGet(ctx, threads[i], "Status").Val())
				fmt.Println("-", "State", client.HGet(ctx, threads[i], "State").Val())
				fmt.Println("-", "Owner", client.HGet(ctx, threads[i], "Owner").Val())
				fmt.Println("-", "Heartbeat", client.HGet(ctx, threads[i], "Heartbeat").Val())
				fmt.Println("-", "Error", client.HGet(ctx, threads[i], "Error").Val())
				fmt.Println("-", "ErrorTime", client.HGet(ctx, threads[i], "ErrorTime").Val())
			}
		case "disable":
			if len(os.Args) < 4 {
				fmt.Println("invalid")
			} else {
				key := os.Args[3]
				client.HSet(ctx, key, "Status", "disabled")
			}

		case "enable":
			if len(os.Args) < 4 {
				fmt.Println("invalid")
			} else {
				key := os.Args[3]
				client.HSet(ctx, key, "Status", "enabled")
			}

		case "load":
			if len(os.Args) < 4 {
				fmt.Println("invalid")
			} else {
				scripts := os.Args[3]

				scriptArray := strings.Split(scripts, ",")
				for i := range scriptArray {
					script := Script{FilePath: scriptArray[i], DeadSeconds: 5, Hang: 100}
					err := loadScript(client, cluster, scriptArray[i], script)
					if err != nil {
						fmt.Println("Failed to load script: " + scriptArray[i])
					}
				}

			}
		}
	},
}

func init() {
	rootCmd.AddCommand(threadsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// threadsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// threadsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
