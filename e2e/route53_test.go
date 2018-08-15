package service

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"

	"github.com/stretchr/testify/assert"

	"github.com/michaeld/ecs-dns/lib"
)

var r = &lib.Route53{
	Domain:       "sandbox1.ecs",
	HostedZoneID: "Z2CK3YSDYYYI0Z",
}

var region = "us-east-1"
var cluster = "sandbox1"

func TestSync(t *testing.T) {

	s, err := session.NewSession(&aws.Config{Region: aws.String(region)})

	if err != nil {
		t.Fatal(err)
	}

	ecsClient := ecs.New(s)
	ec2Client := ec2.New(s)

	var ecs lib.Backend = &lib.ECSCluster{Region: region, Cluster: cluster, ECSClient: ecsClient, EC2Client: ec2Client}

	b, err := ecs.GetTargets()

	if err != nil {
		t.Error(err)
		t.Fail()
	}

	changes, err := r.Sync(b)

	if err != nil {
		t.Error(err)
		t.Fail()
	}

	t.Logf("records synced %d", changes)
}

func TestRemoveAllManagedRecords(t *testing.T) {

	n, err := r.RemoveAllManagedRecords()

	if err != nil {
		t.Error(err)
	}

	assert.NotNil(t, n)
	t.Logf("records removed %d", n)
}
