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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/golang/glog"
	"github.com/michaeld/ecs-dns/lib"
	"github.com/spf13/cobra"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "prune dead backends, readd active",
	Run: func(cmd *cobra.Command, args []string) {

		s, err := session.NewSession(&aws.Config{Region: aws.String(configuration.Region)})

		if err != nil {
			glog.Fatal(err)
		}

		ecsClient := ecs.New(s)
		ec2Client := ec2.New(s)

		t := lib.ECSCluster{Region: configuration.Region, Cluster: configuration.Cluster, ECSClient: ecsClient, EC2Client: ec2Client}
		r53 := lib.Route53{Domain: configuration.Domain, HostedZoneID: configuration.Zone}

		b, err := t.GetTargets()

		if err != nil {
			glog.Fatal(err)
		}

		p, err := r53.Prune(b)

		glog.V(1).Info("Pruning", p)

		if err != nil {
			glog.Error(err)
		}

		i, err := r53.Sync(b)

		if err != nil {
			glog.Error(err)
		}

		glog.Infof("Upserting %d records", i)
	},
}

func init() {
	RootCmd.AddCommand(syncCmd)
}
