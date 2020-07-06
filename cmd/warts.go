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

// wartsCmd represents the warts command
var wartsCmd = &cobra.Command{
	Use:   "warts <command> <command args>",
	Short: "Wart controls",
	Long: `Controls for warts themselves.
	Commands:
	default - lists warts
	stop <wart name> - stops named wart`,
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
			//Print out warts
			warts := client.Keys(ctx, cluster+":Warts:*").Val()
			for i := range warts {
				s := strings.Split(warts[i], ":")
				if s[len(s)-1] != "Health" {
					fmt.Println(warts[i])
					fmt.Println("-", "Status", client.HGet(ctx, warts[i], "Status").Val())
					fmt.Println("-", "State", client.HGet(ctx, warts[i], "State").Val())
					fmt.Println("-", "Heartbeat", client.HGet(ctx, warts[i], "Heartbeat").Val())
					fmt.Println("-", "Health")
					fmt.Println("-", "-", "CPU", client.HGet(ctx, warts[i]+":Health", "cpu").Val())
					fmt.Println("-", "-", "Mem", client.HGet(ctx, warts[i]+":Health", "memory").Val())
				}
			}
		case "stop":
			if len(os.Args) < 4 {
				fmt.Println("invalid")
			} else {
				stopWart(client, cluster, os.Args[3])
			}

		}
	},
}

func init() {
	rootCmd.AddCommand(wartsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// wartsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// wartsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
