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
	"path"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

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

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "wart-control",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.wart-control.yaml)")
	rootCmd.PersistentFlags().StringP("redis-address", "a", "localhost:6379", "<redis address>:<port>")
	rootCmd.PersistentFlags().StringP("redis-password", "p", "", "redis password")
	rootCmd.PersistentFlags().StringP("cluster", "c", "default", "cluster the code runs in")
	viper.BindPFlag("redis-address", rootCmd.PersistentFlags().Lookup("redis-address"))
	viper.BindPFlag("redis-password", rootCmd.PersistentFlags().Lookup("redis-password"))
	viper.BindPFlag("cluster", rootCmd.PersistentFlags().Lookup("cluster"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".wart-control" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".wart-control")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Println(err)
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
	client.HSet(ctx, key, "Status", "disabled")
	client.HSet(ctx, key, "State", "stopped")
	time.Sleep(time.Second)
	client.HSet(ctx, key, "Source", string(fBytes))
	client.HSet(ctx, key, "Heartbeat", 0)
	client.HSet(ctx, key, "Owner", "")
	client.HSet(ctx, key, "Error", "")
	client.HSet(ctx, key, "ErrorTime", "")
	client.HSet(ctx, key, "Hang", script.Hang)
	client.HSet(ctx, key, "DeadSeconds", script.DeadSeconds)

	client.HSet(ctx, key, "Status", "enabled")
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
