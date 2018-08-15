package lib

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/golang/glog"
)

//DNS represent a DNS provider that manages the SRV records
type DNS interface {
	Sync(Targets) (int, error)
	Prune(Targets) (int, error)
	RemoveAllManagedRecords() (int, error)
}

//Route53 represents functionality for managing service discovery records within AWS Route53
type Route53 struct {
	Domain       string
	HostedZoneID string
}

func (r *Route53) recordSets() ([]*route53.ResourceRecordSet, error) {
	sess, err := session.NewSession()

	if err != nil {
		glog.Fatal(err)
	}

	r53 := route53.New(sess)

	rrs := []*route53.ResourceRecordSet{}

	paramsList := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(r.HostedZoneID), // Required
		MaxItems:     aws.String("100"),
	}

	r53.ListResourceRecordSetsPages(paramsList, func(output *route53.ListResourceRecordSetsOutput, lastPage bool) bool {

		for _, s := range output.ResourceRecordSets {
			if isManagedResourceRecordSet(s) {
				rrs = append(rrs, s)
			}
		}

		return !lastPage
	})

	return rrs, nil
}

//Prune removes managed records no longer registered with the backend
func (r *Route53) Prune(targets Targets) (int, error) {

	records, err := r.recordSets()

	if err != nil {
		glog.Error(err)
	}

	glog.Infof("record sets found %d", len(records))

	var removes []*route53.ResourceRecordSet

	for _, v := range records {
		i := strings.Split(*v.SetIdentifier, ":")

		if len(i) != 3 {
			continue
		}

		_, found := targets[Group(i[1])][Service(i[2])]

		if !found {
			removes = append(removes, v)
		}
	}

	i := r.markForDelete(removes)

	changes, err := r.submitChanges(i)

	if err != nil {
		glog.Error(err)
	}

	return changes, err
}

//Sync upserts Traefik backends into AWS hosted zone as SVC records
func (r *Route53) Sync(targets Targets) (int, error) {

	s := r.createServiceRecords(targets)

	c := r.markForUpsert(s)

	return r.submitChanges(c)
}

//RemoveAllManagedRecords deletes all managed records from the AWS Hosted Zone
func (r *Route53) RemoveAllManagedRecords() (int, error) {
	removes, err := r.recordSets()

	if err != nil {
		glog.Error(err)
	}

	c := r.markForDelete(removes)

	return r.submitChanges(c)
}

func (r *Route53) markForDelete(records []*route53.ResourceRecordSet) []*route53.Change {
	c := []*route53.Change{}

	for _, s := range records {
		glog.Infof("Removing record %s %s", *s.Name, *s.ResourceRecords[0].Value)
		c = append(c, &route53.Change{
			Action:            aws.String(route53.ChangeActionDelete),
			ResourceRecordSet: s,
		})
	}

	return c
}

func (r *Route53) markForUpsert(records []*route53.ResourceRecordSet) []*route53.Change {
	c := []*route53.Change{}

	for _, s := range records {
		glog.Infof("Upserting record %s", *s.Name)
		c = append(c, &route53.Change{
			Action:            aws.String(route53.ChangeActionUpsert),
			ResourceRecordSet: s,
		})
	}

	return c
}

func (r *Route53) submitChanges(changes []*route53.Change) (int, error) {

	if len(changes) == 0 {
		glog.Info("No changes to be made")
		return 0, nil
	}

	sess, err := session.NewSession()

	if err != nil {
		glog.Error(err)
	}

	r53 := route53.New(sess)

	_, err = r53.ChangeResourceRecordSets(&route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Comment: aws.String("Service Discovery Created Record"),
			Changes: changes,
		},
		HostedZoneId: aws.String(r.HostedZoneID),
	})

	if err != nil {
		glog.Error(err)
		return -1, err
	}

	glog.Infof("Changed %d records", len(changes))

	return len(changes), nil
}

func isManagedResourceRecordSet(rrs *route53.ResourceRecordSet) bool {
	return rrs != nil &&
		rrs.Type != nil &&
		*rrs.Type == route53.RRTypeSrv &&
		rrs.SetIdentifier != nil &&
		strings.HasPrefix(*rrs.SetIdentifier, "managed:")
}

func (r *Route53) createServiceRecords(targets Targets) []*route53.ResourceRecordSet {
	rrs := []*route53.ResourceRecordSet{}

	for group, service := range targets {

		for serviceName, containers := range service {

			s := &route53.ResourceRecordSet{
				Name: aws.String(fmt.Sprintf("%s.%s.%s", serviceName, group, r.Domain)),
				// It creates a SRV record with the name of the service
				Type:          aws.String(route53.RRTypeSrv),
				SetIdentifier: aws.String(fmt.Sprintf("managed:%s:%s", group, serviceName)),
				// TTL=0 to avoid DNS caches
				TTL:    aws.Int64(0),
				Weight: aws.Int64(1),
			}

			for _, r := range containers {
				s.ResourceRecords = append(s.ResourceRecords, &route53.ResourceRecord{Value: aws.String(formatTargetSvcRecord(r))})
			}

			rrs = append(rrs, s)
		}

	}

	return rrs
}

func formatTargetSvcRecord(t *Target) string {
	return fmt.Sprintf("%s %s %d %s", "1", "1", t.Port, t.IPAddress)
}
