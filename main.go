package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/go-redis/redis"
)

var ctx = context.Background()

type WartMeta struct {
	Name      string
	Status    string
	State     string
	Heartbeat string
	CPU       string
	Mem       string
}

type ThreadMeta struct {
	Name      string
	Status    string
	State     string
	Owner     string
	Heartbeat string
}

type EndpointMeta struct {
	Name   string
	Status string
}

type Endpoint struct {
	Route    string
	FilePath string
}

type Script struct {
	FilePath    string
	Hang        int
	DeadSeconds int
}

type Env struct {
	Endpoints []Endpoint
	Scripts   []Script
	Cluster   string
}

var client *redis.Client
var cluster = "default"

func main() {
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

	cmd := ""
	if len(os.Args) > 0 {
		cmd = os.Args[1]
	}

	client = redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: redisPassword, // no password set
		DB:       0,             // use default DB
	})

	switch cmd {
	case "":
	case "proxy":
		http.HandleFunc("/warts", wartsHandler)
		http.HandleFunc("/threads", threadsHandler)
		http.HandleFunc("/endpoints", endpointsHandler)
		log.Fatal(http.ListenAndServe(":9898", nil))
	case "apply":
		if len(os.Args) > 2 {
			fileName := os.Args[2]
			fmt.Println(loadEnvironment(client, fileName))
		} else {
			fmt.Println("No file")
			return
		}

	case "purge":
		fmt.Println("Sure? (Y)")
		reader := bufio.NewReader(os.Stdin)
		a, _, err := reader.ReadRune()
		if err == nil {
			if a == 'Y' {
				keys := client.Keys(ctx, cluster+":*").Val()
				for k := range keys {
					fmt.Println("Removed:", keys[k])
					client.Del(ctx, keys[k])
				}
			}
		}

	case "threads":
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
		}

	case "endpoints":
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
				client.HSet(ctx, cluster+":Warts:"+key, "Status", "disabled")
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

	case "warts":
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
	}
}

func stopWart(client *redis.Client, cluster string, wart string) {
	key := cluster + ":Warts:" + wart
	client.HSet(ctx, key, "Status", "disabled")
}

func loadEndpoint(client *redis.Client, cluster string, scriptName string, scriptPath string) (err error) {
	fBytes, err := ioutil.ReadFile(scriptPath)
	if err != nil {
		return
	}
	key := cluster + ":Endpoints:" + scriptName
	client.HSet(ctx, key, "Source", string(fBytes))
	client.HSet(ctx, key, "Status", "enabled")
	client.HSet(ctx, key, "Error", "")
	client.HSet(ctx, key, "ErrorTime", "")
	return
}

func loadScript(client *redis.Client, cluster string, scriptName string, script Script) (err error) {
	fBytes, err := ioutil.ReadFile(scriptName)
	if err != nil {
		return
	}

	key := cluster + ":Threads:" + scriptName
	client.HSet(ctx, key, "Source", string(fBytes))
	client.HSet(ctx, key, "Status", "enabled")
	client.HSet(ctx, key, "State", "stopped")
	client.HSet(ctx, key, "Heartbeat", 0)
	client.HSet(ctx, key, "Owner", "")
	client.HSet(ctx, key, "Error", "")
	client.HSet(ctx, key, "ErrorTime", "")
	client.HSet(ctx, key, "Hang", script.Hang)
	client.HSet(ctx, key, "DeadSeconds", script.DeadSeconds)
	return
}

func loadEnvironment(client *redis.Client, fileName string) (err error) {
	fBytes, err := ioutil.ReadFile(fileName)

	if err == nil {
		var env Env
		err = json.Unmarshal(fBytes, &env)
		if err == nil {
			fmt.Println(env)
			for i := range env.Endpoints {
				err = loadEndpoint(client, env.Cluster, env.Endpoints[i].Route, path.Dir(fileName)+"/"+env.Endpoints[i].FilePath)
				if err != nil {
					return
				}
			}

			for i := range env.Scripts {
				err = loadScript(client, env.Cluster, path.Dir(fileName)+"/"+env.Scripts[i].FilePath, env.Scripts[i])
				if err != nil {
					return
				}
			}
		}
	}
	return
}

func wartsHandler(w http.ResponseWriter, r *http.Request) {
	addCorsHeader(w)
	if r.Method == http.MethodGet {
		if len(r.URL.Query().Get("name")) > 0 {
			name := r.URL.Query().Get("name")
			wartMeta := &WartMeta{}
			wartMeta.Name = name
			wartMeta.Status = client.HGet(ctx, name, "Status").Val()
			wartMeta.State = client.HGet(ctx, name, "State").Val()
			wartMeta.Heartbeat = client.HGet(ctx, name, "Heartbeat").Val()
			wartMeta.CPU = client.HGet(ctx, name+":Health", "cpu").Val()
			wartMeta.Mem = client.HGet(ctx, name+":Health", "memory").Val()

			out, err := json.Marshal(wartMeta)
			if err == nil {
				fmt.Fprintf(w, string(out))
			} else {
				fmt.Fprintf(w, "{'error':'"+err.Error()+"'}")
			}
		} else {
			wartMetas := make([]*WartMeta, 0)
			warts := client.Keys(ctx, cluster+":Warts:*").Val()
			for i := range warts {
				s := strings.Split(warts[i], ":")
				if s[len(s)-1] != "Health" {
					wartMeta := &WartMeta{}
					wartMeta.Name = warts[i]
					wartMeta.Status = client.HGet(ctx, warts[i], "Status").Val()
					wartMeta.State = client.HGet(ctx, warts[i], "State").Val()
					wartMeta.Heartbeat = client.HGet(ctx, warts[i], "Heartbeat").Val()
					wartMeta.CPU = client.HGet(ctx, warts[i]+":Health", "cpu").Val()
					wartMeta.Mem = client.HGet(ctx, warts[i]+":Health", "memory").Val()
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

		client.HSet(ctx, wart.Name, "Status", wart.Status)
		client.HSet(ctx, wart.Name, "State", wart.State)
	}
}
func threadsHandler(w http.ResponseWriter, r *http.Request) {
	addCorsHeader(w)
	if r.Method == http.MethodGet {
		if len(r.URL.Query().Get("name")) > 0 {
			name := r.URL.Query().Get("name")
			threadMeta := &ThreadMeta{}
			threadMeta.Name = name
			threadMeta.Status = client.HGet(ctx, name, "Status").Val()
			threadMeta.State = client.HGet(ctx, name, "State").Val()
			threadMeta.Heartbeat = client.HGet(ctx, name, "Heartbeat").Val()
			threadMeta.Owner = client.HGet(ctx, name, "Owner").Val()

			out, err := json.Marshal(threadMeta)
			if err == nil {
				fmt.Fprintf(w, string(out))
			} else {
				fmt.Fprintf(w, "{'error':'"+err.Error()+"'}")
			}
		} else {
			threadMetas := make([]*ThreadMeta, 0)
			threads := client.Keys(ctx, cluster+":Threads:*").Val()
			for i := range threads {
				threadMeta := &ThreadMeta{}
				threadMeta.Name = threads[i]
				threadMeta.Status = client.HGet(ctx, threads[i], "Status").Val()
				threadMeta.State = client.HGet(ctx, threads[i], "State").Val()
				threadMeta.Heartbeat = client.HGet(ctx, threads[i], "Heartbeat").Val()
				threadMeta.Owner = client.HGet(ctx, threads[i], "Owner").Val()
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
		client.HSet(ctx, thread.Name, "Status", thread.Status)
		client.HSet(ctx, thread.Name, "State", thread.State)
	}
}
func endpointsHandler(w http.ResponseWriter, r *http.Request) {
	addCorsHeader(w)
	if r.Method == http.MethodGet {
		if len(r.URL.Query().Get("name")) > 0 {
			name := r.URL.Query().Get("name")
			threadMeta := &EndpointMeta{}
			threadMeta.Name = name
			threadMeta.Status = client.HGet(ctx, name, "Status").Val()

			out, err := json.Marshal(threadMeta)
			if err == nil {
				fmt.Fprintf(w, string(out))
			} else {
				fmt.Fprintf(w, "{'error':'"+err.Error()+"'}")
			}
		} else {
			threadMetas := make([]*EndpointMeta, 0)
			threads := client.Keys(ctx, cluster+":Endpoints:*").Val()
			for i := range threads {
				threadMeta := &EndpointMeta{}
				threadMeta.Name = threads[i]
				threadMeta.Status = client.HGet(ctx, threads[i], "Status").Val()
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
		client.HSet(ctx, thread.Name, "Status", thread.Status)
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
