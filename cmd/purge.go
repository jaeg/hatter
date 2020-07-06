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
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/cobra"
)

// purgeCmd represents the purge command
var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Clears all cluster related entries in redis`",
	Long:  `Clears all cluster related entries in redis`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Sure? (Y)")
		reader := bufio.NewReader(os.Stdin)
		a, _, err := reader.ReadRune()
		if err == nil {
			if a == 'Y' {
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
				keys := client.Keys(ctx, cluster+":*").Val()
				for k := range keys {
					fmt.Println("Removed:", keys[k])
					client.Del(ctx, keys[k])
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(purgeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// purgeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// purgeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
