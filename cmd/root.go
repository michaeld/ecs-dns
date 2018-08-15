// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"flag"
	"os"

	"github.com/michaeld/ecs-dns/lib"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		glog.Error(err)
		os.Exit(-1)
	}
}

var configuration *lib.Config

func init() {

	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(".")      // adding current directory
	viper.AutomaticEnv()          // read in environment variables that match

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")

	pflag.String("domain", "", "domain name")
	pflag.String("zone", "", "hosted zone id")
	pflag.String("interval", "10", "poll interval in seconds")
	pflag.String("region", "us-east-1", "ecs cluster region")
	pflag.String("cluster", "", "ecs cluster name")

	viper.SetDefault("interval", 10)

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	flag.CommandLine.Parse([]string{})
	viper.BindPFlags(pflag.CommandLine)

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		glog.Info("Using config file:", viper.ConfigFileUsed())
	}

	glog.Info(viper.AllSettings())

	configuration = &lib.Config{
		Region:   viper.GetString("region"),
		Domain:   viper.GetString("domain"),
		Zone:     viper.GetString("zone"),
		Cluster:  viper.GetString("cluster"),
		Interval: viper.GetInt64("interval"),
	}
}
