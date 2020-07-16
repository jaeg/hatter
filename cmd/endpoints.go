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
	"fmt"
	"os"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// endpointsCmd represents the endpoints command
var endpointsCmd = &cobra.Command{
	Use:   "endpoints <command> <command args>",
	Short: "Endpoint controls",
	Long: `Controls for endpoints.
	Commands:
	default - lists endpoints
	enable <endpoint name> - enables endpoint
	disable <endpoint name> - disables endpoint
	load <file path> - loads comma separated list of source files into the cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		redisAddress := viper.GetViper().GetString("redis-address")
		redisPassword := viper.GetViper().GetString("redis-password")
		cluster := viper.GetViper().GetString("cluster")

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
			threads := client.Keys(ctx, cluster+":Endpoints:*").Val()
			for i := range threads {
				fmt.Println(threads[i])
				fmt.Println("-", "Status", client.HGet(ctx, threads[i], "Status").Val())
			}
		case "disable":
			if len(os.Args) < 4 {
				fmt.Println("invalid")
			} else {
				key := os.Args[3]
				client.HSet(ctx, cluster+":workers:"+key, "Status", "disabled")
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
					err := loadEndpoint(client, cluster, scriptArray[i], scriptArray[i])
					if err != nil {
						fmt.Println("Failed to load script: " + scriptArray[i])
					}
				}

			}
		}
	},
}

func init() {
	rootCmd.AddCommand(endpointsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// endpointsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// endpointsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
