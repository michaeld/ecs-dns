package lib

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/stretchr/testify/assert"
)

type stubAWSClient struct{}

func (*stubAWSClient) ListTasksPages(i *ecs.ListTasksInput, f func(*ecs.ListTasksOutput, bool) bool) error {

	f(&ecs.ListTasksOutput{TaskArns: []*string{aws.String("task1")}}, true)

	return nil
}

func (*stubAWSClient) DescribeContainerInstances(*ecs.DescribeContainerInstancesInput) (*ecs.DescribeContainerInstancesOutput, error) {
	return &ecs.DescribeContainerInstancesOutput{
		ContainerInstances: []*ecs.ContainerInstance{
			&ecs.ContainerInstance{Ec2InstanceId: aws.String("i-1"), ContainerInstanceArn: aws.String("ci-arn1")},
			&ecs.ContainerInstance{Ec2InstanceId: aws.String("i-2"), ContainerInstanceArn: aws.String("ci-arn2")},
			&ecs.ContainerInstance{Ec2InstanceId: aws.String("i-3"), ContainerInstanceArn: aws.String("ci-arn3")},
		},
	}, nil
}

func (*stubAWSClient) ListContainerInstancesPages(i *ecs.ListContainerInstancesInput, f func(*ecs.ListContainerInstancesOutput, bool) bool) error {

	f(&ecs.ListContainerInstancesOutput{}, true)

	return nil
}

func (*stubAWSClient) DescribeTasks(*ecs.DescribeTasksInput) (*ecs.DescribeTasksOutput, error) {

	return &ecs.DescribeTasksOutput{
		Tasks: []*ecs.Task{
			&ecs.Task{
				TaskArn:              aws.String("taskarn1"),
				ContainerInstanceArn: aws.String("ci-arn1"),
				Group:                aws.String("family1:group1"),
				Containers: []*ecs.Container{
					&ecs.Container{
						Name: aws.String("container1"),
						NetworkBindings: []*ecs.NetworkBinding{
							&ecs.NetworkBinding{
								HostPort: aws.Int64(1234),
							},
						},
					},
				},
			},
		},
	}, nil
}

func (*stubAWSClient) DescribeInstancesPages(i *ec2.DescribeInstancesInput, f func(*ec2.DescribeInstancesOutput, bool) bool) error {
	f(&ec2.DescribeInstancesOutput{
		Reservations: []*ec2.Reservation{
			&ec2.Reservation{
				Instances: []*ec2.Instance{
					&ec2.Instance{
						InstanceId:       aws.String("i-1"),
						PrivateIpAddress: aws.String("1.2.3.4"),
					},
				},
			},
		},
	}, true)

	return nil
}

var ecsCluster = &ECSCluster{Region: "us-east-1", Cluster: "cluster1", ECSClient: &stubAWSClient{}, EC2Client: &stubAWSClient{}}

func TestGetTasks(t *testing.T) {

	tasks, err := ecsCluster.getTasks()

	if err != nil {
		t.Error(err)
	}

	assert.Len(t, tasks, 1)
	assert.Equal(t, *tasks[0].TaskArn, "taskarn1")
}

func TestGetTargets(t *testing.T) {

	var targets Targets

	targets, err := ecsCluster.GetTargets()

	if err != nil {
		t.Error(err)
		t.Fail()
	}

	assert.Equal(t, targets["group1"]["container1"][0].Name, "container1")
}

func TestGetHosts(t *testing.T) {

	h, err := ecsCluster.getHosts()

	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, len(h), 3)
}
