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
	"log"
	"net/http"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ctx = context.Background()
var pClient *redis.Client
var pCluster = "default"

// proxyCmd represents the proxy command
var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Starts proxy server for ui",
	Long: `Starts proxy server for ui
	Endpoints
	- /workers 
	- /threads
	- /endpoints`,
	Run: func(cmd *cobra.Command, args []string) {
		redisAddress := viper.GetViper().GetString("redis-address")
		redisPassword := viper.GetViper().GetString("redis-password")
		pCluster = viper.GetViper().GetString("cluster")

		pClient = redis.NewClient(&redis.Options{
			Addr:     redisAddress,
			Password: redisPassword, // no password set
			DB:       0,             // use default DB
		})

		http.HandleFunc("/workers", workersHandler)
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

func workersHandler(w http.ResponseWriter, r *http.Request) {
	addCorsHeader(w)
	if r.Method == http.MethodGet {
		if len(r.URL.Query().Get("name")) > 0 {
			name := r.URL.Query().Get("name")
			workerMeta := &WorkerMeta{}
			workerMeta.Name = name
			workerMeta.Status = pClient.HGet(ctx, name, "Status").Val()
			workerMeta.State = pClient.HGet(ctx, name, "State").Val()
			workerMeta.Heartbeat = pClient.HGet(ctx, name, "Heartbeat").Val()
			workerMeta.CPU = pClient.HGet(ctx, name+":Health", "cpu").Val()
			workerMeta.Mem = pClient.HGet(ctx, name+":Health", "memory").Val()

			out, err := json.Marshal(workerMeta)
			if err == nil {
				fmt.Fprintf(w, string(out))
			} else {
				fmt.Fprintf(w, "{'error':'"+err.Error()+"'}")
			}
		} else {
			workerMetas := make([]*WorkerMeta, 0)
			workers := pClient.Keys(ctx, pCluster+":workers:*").Val()
			for i := range workers {
				s := strings.Split(workers[i], ":")
				if s[len(s)-1] != "Health" {
					workerMeta := &WorkerMeta{}
					workerMeta.Name = workers[i]
					workerMeta.Status = pClient.HGet(ctx, workers[i], "Status").Val()
					workerMeta.State = pClient.HGet(ctx, workers[i], "State").Val()
					workerMeta.Heartbeat = pClient.HGet(ctx, workers[i], "Heartbeat").Val()
					workerMeta.CPU = pClient.HGet(ctx, workers[i]+":Health", "cpu").Val()
					workerMeta.Mem = pClient.HGet(ctx, workers[i]+":Health", "memory").Val()
					workerMetas = append(workerMetas, workerMeta)
				}
			}
			out, err := json.Marshal(workerMetas)
			if err == nil {
				fmt.Fprintf(w, string(out))
			} else {
				fmt.Fprintf(w, "{'error':'"+err.Error()+"'}")
			}
		}
	}
	if r.Method == http.MethodPut {
		decoder := json.NewDecoder(r.Body)
		var worker WorkerMeta
		err := decoder.Decode(&worker)
		if err != nil {
			panic(err)
		}

		pClient.HSet(ctx, worker.Name, "Status", worker.Status)
		pClient.HSet(ctx, worker.Name, "State", worker.State)
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
			threads := pClient.Keys(ctx, pCluster+":Threads:*").Val()
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
			threads := pClient.Keys(ctx, pCluster+":Endpoints:*").Val()
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
