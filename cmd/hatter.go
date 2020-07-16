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

// hatterCmd represents the workers command
var hatterCmd = &cobra.Command{
	Use:   "workers <command> <command args>",
	Short: "Hatter",
	Long: `Controls for workers themselves.
	Commands:
	default - lists workers
	stop <worker name> - stops named worker`,
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
			//Print out workers
			workers := client.Keys(ctx, cluster+":workers:*").Val()
			for i := range workers {
				s := strings.Split(workers[i], ":")
				if s[len(s)-1] != "Health" {
					fmt.Println(workers[i])
					fmt.Println("-", "Status", client.HGet(ctx, workers[i], "Status").Val())
					fmt.Println("-", "State", client.HGet(ctx, workers[i], "State").Val())
					fmt.Println("-", "Heartbeat", client.HGet(ctx, workers[i], "Heartbeat").Val())
					fmt.Println("-", "Health")
					fmt.Println("-", "-", "CPU", client.HGet(ctx, workers[i]+":Health", "cpu").Val())
					fmt.Println("-", "-", "Mem", client.HGet(ctx, workers[i]+":Health", "memory").Val())
				}
			}
		case "stop":
			if len(os.Args) < 4 {
				fmt.Println("invalid")
			} else {
				stopWorker(client, cluster, os.Args[3])
			}

		}
	},
}

func init() {
	rootCmd.AddCommand(hatterCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// hatterCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// hatterCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
