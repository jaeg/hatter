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
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/cobra"
)

var ctx = context.Background()
var pClient *redis.Client
var cluster = "default"

// proxyCmd represents the proxy command
var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		cluster = "default"
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
				cluster = m["cluster"].(string)
			}
		}

		pClient = redis.NewClient(&redis.Options{
			Addr:     redisAddress,
			Password: redisPassword, // no password set
			DB:       0,             // use default DB
		})

		http.HandleFunc("/warts", wartsHandler)
		http.HandleFunc("/threads", threadsHandler)
		http.HandleFunc("/endpoints", endpointsHandler)
		log.Fatal(http.ListenAndServe(":9898", nil))
	},
}

func init() {
	rootCmd.AddCommand(proxyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// proxyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// proxyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func wartsHandler(w http.ResponseWriter, r *http.Request) {
	addCorsHeader(w)
	if r.Method == http.MethodGet {
		if len(r.URL.Query().Get("name")) > 0 {
			name := r.URL.Query().Get("name")
			wartMeta := &WartMeta{}
			wartMeta.Name = name
			wartMeta.Status = pClient.HGet(ctx, name, "Status").Val()
			wartMeta.State = pClient.HGet(ctx, name, "State").Val()
			wartMeta.Heartbeat = pClient.HGet(ctx, name, "Heartbeat").Val()
			wartMeta.CPU = pClient.HGet(ctx, name+":Health", "cpu").Val()
			wartMeta.Mem = pClient.HGet(ctx, name+":Health", "memory").Val()

			out, err := json.Marshal(wartMeta)
			if err == nil {
				fmt.Fprintf(w, string(out))
			} else {
				fmt.Fprintf(w, "{'error':'"+err.Error()+"'}")
			}
		} else {
			wartMetas := make([]*WartMeta, 0)
			warts := pClient.Keys(ctx, cluster+":Warts:*").Val()
			for i := range warts {
				s := strings.Split(warts[i], ":")
				if s[len(s)-1] != "Health" {
					wartMeta := &WartMeta{}
					wartMeta.Name = warts[i]
					wartMeta.Status = pClient.HGet(ctx, warts[i], "Status").Val()
					wartMeta.State = pClient.HGet(ctx, warts[i], "State").Val()
					wartMeta.Heartbeat = pClient.HGet(ctx, warts[i], "Heartbeat").Val()
					wartMeta.CPU = pClient.HGet(ctx, warts[i]+":Health", "cpu").Val()
					wartMeta.Mem = pClient.HGet(ctx, warts[i]+":Health", "memory").Val()
					wartMetas = append(wartMetas, wartMeta)
				}
			}
			out, err := json.Marshal(wartMetas)
			if err == nil {
				fmt.Fprintf(w, string(out))
			} else {
				fmt.Fprintf(w, "{'error':'"+err.Error()+"'}")
			}
		}
	}
	if r.Method == http.MethodPut {
		decoder := json.NewDecoder(r.Body)
		var wart WartMeta
		err := decoder.Decode(&wart)
		if err != nil {
			panic(err)
		}

		pClient.HSet(ctx, wart.Name, "Status", wart.Status)
		pClient.HSet(ctx, wart.Name, "State", wart.State)
	}
}
func threadsHandler(w http.ResponseWriter, r *http.Request) {
	addCorsHeader(w)
	if r.Method == http.MethodGet {
		if len(r.URL.Query().Get("name")) > 0 {
			name := r.URL.Query().Get("name")
			threadMeta := &ThreadMeta{}
			threadMeta.Name = name
			threadMeta.Status = pClient.HGet(ctx, name, "Status").Val()
			threadMeta.State = pClient.HGet(ctx, name, "State").Val()
			threadMeta.Heartbeat = pClient.HGet(ctx, name, "Heartbeat").Val()
			threadMeta.Owner = pClient.HGet(ctx, name, "Owner").Val()

			out, err := json.Marshal(threadMeta)
			if err == nil {
				fmt.Fprintf(w, string(out))
			} else {
				fmt.Fprintf(w, "{'error':'"+err.Error()+"'}")
			}
		} else {
			threadMetas := make([]*ThreadMeta, 0)
			threads := pClient.Keys(ctx, cluster+":Threads:*").Val()
			for i := range threads {
				threadMeta := &ThreadMeta{}
				threadMeta.Name = threads[i]
				threadMeta.Status = pClient.HGet(ctx, threads[i], "Status").Val()
				threadMeta.State = pClient.HGet(ctx, threads[i], "State").Val()
				threadMeta.Heartbeat = pClient.HGet(ctx, threads[i], "Heartbeat").Val()
				threadMeta.Owner = pClient.HGet(ctx, threads[i], "Owner").Val()
				threadMetas = append(threadMetas, threadMeta)

			}
			out, err := json.Marshal(threadMetas)
			if err == nil {
				fmt.Fprintf(w, string(out))
			} else {
				fmt.Fprintf(w, "{'error':'"+err.Error()+"'}")
			}
		}
	}
	if r.Method == http.MethodPut {
		decoder := json.NewDecoder(r.Body)
		var thread ThreadMeta
		err := decoder.Decode(&thread)
		if err != nil {
			panic(err)
		}
		pClient.HSet(ctx, thread.Name, "Status", thread.Status)
		pClient.HSet(ctx, thread.Name, "State", thread.State)
	}
}
func endpointsHandler(w http.ResponseWriter, r *http.Request) {
	addCorsHeader(w)
	if r.Method == http.MethodGet {
		if len(r.URL.Query().Get("name")) > 0 {
			name := r.URL.Query().Get("name")
			threadMeta := &EndpointMeta{}
			threadMeta.Name = name
			threadMeta.Status = pClient.HGet(ctx, name, "Status").Val()

			out, err := json.Marshal(threadMeta)
			if err == nil {
				fmt.Fprintf(w, string(out))
			} else {
				fmt.Fprintf(w, "{'error':'"+err.Error()+"'}")
			}
		} else {
			threadMetas := make([]*EndpointMeta, 0)
			threads := pClient.Keys(ctx, cluster+":Endpoints:*").Val()
			for i := range threads {
				threadMeta := &EndpointMeta{}
				threadMeta.Name = threads[i]
				threadMeta.Status = pClient.HGet(ctx, threads[i], "Status").Val()
				threadMetas = append(threadMetas, threadMeta)

			}
			out, err := json.Marshal(threadMetas)
			if err == nil {
				fmt.Fprintf(w, string(out))
			} else {
				fmt.Fprintf(w, "{'error':'"+err.Error()+"'}")
			}
		}
	}
	if r.Method == http.MethodPut {
		decoder := json.NewDecoder(r.Body)
		var thread EndpointMeta
		err := decoder.Decode(&thread)
		if err != nil {
			panic(err)
		}
		pClient.HSet(ctx, thread.Name, "Status", thread.Status)
	}
}

func addCorsHeader(res http.ResponseWriter) {
	headers := res.Header()
	headers.Add("Access-Control-Allow-Origin", "*")
	headers.Add("Vary", "Origin")
	headers.Add("Vary", "Access-Control-Request-Method")
	headers.Add("Vary", "Access-Control-Request-Headers")
	headers.Add("Access-Control-Allow-Headers", "Content-Type, Origin, Accept, token")
	headers.Add("Access-Control-Allow-Methods", "GET, POST, PUT,OPTIONS")
}
