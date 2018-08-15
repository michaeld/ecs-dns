package lib

import (
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/golang/glog"
	"github.com/mitchellh/hashstructure"
)

//Backend has information about targets
type Backend interface {
	GetTargets() (Targets, error)
}

//Target stores the containers scrapeable endpoint
type Target struct {
	Port      int64
	IPAddress string
	Name      string
	Group     string
}

//Targets stores targets grouped by service and container
type Targets map[string]map[string][]*Target

//GetTargets combines host and container information to produce the scrapeable targets
func (e *ECSCluster) GetTargets() (s Targets, err error) {

	hosts, err := e.getHosts()

	if err != nil {
		glog.Error(err)
		return
	}

	tasks, err := e.getTasks()

	if err != nil {
		glog.Error(err)
		return
	}

	s = make(Targets)

	for _, task := range tasks {

		i, found := hosts[*task.ContainerInstanceArn]

		if !found {
			glog.Errorf("Container Instance not found for task %s", *task.TaskArn)
			continue
		}

		group := strings.Split(*task.Group, ":")[1]

		for _, c := range task.Containers {

			if len(c.NetworkBindings) <= 0 {
				continue
			}
			//TODO uses first network binding but should use a docker label
			target := &Target{
				Port:      *c.NetworkBindings[0].HostPort,
				Name:      *c.Name,
				IPAddress: *i.PrivateIPAddress,
				Group:     string(group),
			}

			if s[group] == nil {
				s[group] = make(map[string][]*Target)
			}

			s[group][*c.Name] = append(s[group][*c.Name], target)
		}
	}

	return s, nil
}

func (e *ECSCluster) getTasks() ([]*ecs.Task, error) {

	tasks := []*ecs.Task{}

	input := &ecs.ListTasksInput{
		Cluster: &e.Cluster,
	}

	err := e.ECSClient.ListTasksPages(input, func(taskArns *ecs.ListTasksOutput, lastPage bool) bool {

		descrTasks, err := e.ECSClient.DescribeTasks(&ecs.DescribeTasksInput{
			Cluster: &e.Cluster,
			Tasks:   taskArns.TaskArns,
		})

		if err != nil {
			glog.Error(err)
			return false
		}

		if len(descrTasks.Failures) != 0 {
			glog.Errorf("Failure describing task: %v - %v", *descrTasks.Failures[0].Arn, *descrTasks.Failures[0].Reason)
			return false
		}

		tasks = append(tasks, descrTasks.Tasks...)

		return !lastPage
	})

	if err != nil {
		glog.Error(err)
		return nil, err
	}

	h, err := hashstructure.Hash(tasks, nil)

	if err != nil {
		glog.Error(err)
	}

	glog.V(1).Infof("last hash %d, new hash %d", e.tasksHash, h)

	if h == e.tasksHash && (*e).tasks != nil {
		glog.Info("tasks haven't changed, hash is the same, returning cache")
		return *e.tasks, nil
	}

	e.tasksHash = h
	e.tasks = &tasks

	return *e.tasks, err
}

//ECSApi contains the functions necessary to interact with ECS
type ECSApi interface {
	DescribeContainerInstances(*ecs.DescribeContainerInstancesInput) (*ecs.DescribeContainerInstancesOutput, error)
	ListTasksPages(*ecs.ListTasksInput, func(*ecs.ListTasksOutput, bool) bool) error
	ListContainerInstancesPages(*ecs.ListContainerInstancesInput, func(*ecs.ListContainerInstancesOutput, bool) bool) error
	DescribeTasks(*ecs.DescribeTasksInput) (*ecs.DescribeTasksOutput, error)
}

//EC2Api contains the function necessary to interact with EC2
type EC2Api interface {
	DescribeInstancesPages(*ec2.DescribeInstancesInput, func(*ec2.DescribeInstancesOutput, bool) bool) error
}

//ECSCluster holds the internal state of an ECS Cluster to retrieve scrape targets
type ECSCluster struct {
	Region, Cluster      string
	ECSClient            ECSApi
	EC2Client            EC2Api
	hosts                *map[string]*ecsHost
	tasks                *[]*ecs.Task
	hostsHash, tasksHash uint64
}

//ecsHost stores metadata about ECS Container Hosts
type ecsHost struct {
	InstanceID           *string
	ContainerInstanceArn *string
	PrivateIPAddress     *string
}

func (e *ECSCluster) getHosts() (map[string]*ecsHost, error) {

	ec2InstanceIds := []*string{}
	ec2InstanceIdsForCache := []string{}
	instances := &map[string]*ecsHost{}

	e.ECSClient.ListContainerInstancesPages(&ecs.ListContainerInstancesInput{Cluster: &e.Cluster},
		func(o *ecs.ListContainerInstancesOutput, lastPage bool) bool {

			i, err := e.ECSClient.DescribeContainerInstances(&ecs.DescribeContainerInstancesInput{Cluster: &e.Cluster, ContainerInstances: o.ContainerInstanceArns})

			if err != nil {
				glog.Error(err)
			}

			for _, c := range i.ContainerInstances {
				(*instances)[*c.Ec2InstanceId] = &ecsHost{InstanceID: c.Ec2InstanceId, ContainerInstanceArn: c.ContainerInstanceArn}
				ec2InstanceIdsForCache = append(ec2InstanceIdsForCache, *c.Ec2InstanceId)
				ec2InstanceIds = append(ec2InstanceIds, c.Ec2InstanceId)
			}

			return !lastPage
		})

	sort.Strings(ec2InstanceIdsForCache)

	h, err := hashstructure.Hash(ec2InstanceIds, nil)

	if err != nil {
		glog.Error(err)
	}

	glog.V(1).Info(h, ec2InstanceIds)

	if h == e.hostsHash && (*e).hosts != nil {
		glog.Info("cluster instances haven't changed, hash is the same, returning cache")
		return *e.hosts, nil
	}

	e.hostsHash = h

	e.EC2Client.DescribeInstancesPages(&ec2.DescribeInstancesInput{InstanceIds: ec2InstanceIds},
		func(o *ec2.DescribeInstancesOutput, lastPage bool) bool {

			for _, r := range o.Reservations {
				for _, i := range r.Instances {
					(*instances)[*i.InstanceId].PrivateIPAddress = i.PrivateIpAddress
				}
			}

			return !lastPage
		})

	if e.hosts == nil {
		e.hosts = &map[string]*ecsHost{}
	}

	for _, i := range *instances {
		(*e.hosts)[*i.ContainerInstanceArn] = i
	}

	return *e.hosts, nil
}
