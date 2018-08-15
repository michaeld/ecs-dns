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
	"os"
	"os/signal"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"

	"github.com/golang/glog"
	"github.com/michaeld/ecs-dns/lib"
	"github.com/mitchellh/hashstructure"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var daemonCmd = &cobra.Command{
	Use:     "daemon",
	Aliases: []string{"d"},
	Short:   "prune, then run upserts daemonized",
	Run: func(cmd *cobra.Command, args []string) {

		ticker := time.NewTicker(time.Second * time.Duration(configuration.Interval))
		defer ticker.Stop()

		go func(c *lib.Config) {

			s, err := session.NewSession(&aws.Config{Region: aws.String(c.Region)})

			if err != nil {
				glog.Fatal(err)
			}

			ecsClient := ecs.New(s)
			ec2Client := ec2.New(s)

			t := lib.ECSCluster{Region: configuration.Region, Cluster: configuration.Cluster, ECSClient: ecsClient, EC2Client: ec2Client}
			r53 := lib.Route53{Domain: configuration.Domain, HostedZoneID: configuration.Zone}

			e, err := t.GetTargets()
			r53.Prune(e)

			if err != nil {
				glog.Error(err)
			}

			var lastHash uint64

			for range ticker.C {
				b, err := t.GetTargets()

				if err != nil {
					glog.Error(err)
					continue
				}

				h, err := hashstructure.Hash(b, nil)

				if err != nil {
					glog.Error(err)
				}

				if h == lastHash {
					glog.Info("targets haven't changed, hash is the same, continuing")
					continue
				}

				i, err := r53.Sync(b)

				if err != nil {
					glog.Error(err)
				}

				glog.Infof("Records updated %d", i)
				glog.V(1).Infof("lastHash %d, new hash %d", lastHash, h)
				lastHash = h

			}
		}(configuration)

		sC := make(chan os.Signal, 1)
		signal.Notify(sC, os.Interrupt, os.Kill)

		glog.Info("running...")

		<-sC

		glog.Info("exiting")
	},
}

func init() {
	RootCmd.AddCommand(daemonCmd)
}
