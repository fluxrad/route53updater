package main

import (
	"flag"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"io/ioutil"
	"log"
	"net/http"
)

var (
	zoneID     string
	hostname   string
	domain     string
	ttl        int64
	rr         string
	recordType string
)

type Route53Change struct {
	Zone        string
	Hostname    string
	DomainName  string
	AWSHostname string
	TTL         int64
	RecordType  string
}

func init() {
	log.SetFlags(0)
	flag.StringVar(&zoneID, "zoneid", "", "The Route53 Zone ID to use")
	flag.StringVar(&hostname, "hostname", "", "The hostname to update")
	flag.Int64Var(&ttl, "ttl", 60, "The TTL to use (default: 60)")
	flag.StringVar(&domain, "domain", "", "The domain name to update")
	flag.StringVar(&rr, "rr", "", "The resource record to use (i.e. don't use AWS hostname")
	flag.StringVar(&recordType, "recordtype", "CNAME", "The record type to use (CNAME or A)")
}

func updateRoute53(c *Route53Change) (string, error) {
	svc := route53.New(nil)

	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{ // Required
			Changes: []*route53.Change{ // Required
				{ // Required
					Action: aws.String("UPSERT"), // Required
					ResourceRecordSet: &route53.ResourceRecordSet{ // Required
						Name: aws.String(c.Hostname),   // Required
						Type: aws.String(c.RecordType), // Required
						ResourceRecords: []*route53.ResourceRecord{
							{ // Required
								Value: aws.String(c.AWSHostname), // Required
							},
							// More values...
						},
						TTL: aws.Long(c.TTL),
					},
				},
			},
		},
		HostedZoneID: aws.String(c.Zone), // Required
	}

	resp, err := svc.ChangeResourceRecordSets(params)
	if err != nil {
		return "", err
	}
	return resp.GoString(), nil
}

func fetchAwsHostname() string {
	out, err := http.Get("http://169.254.169.254/latest/meta-data/public-hostname")
	if err != nil {
		// If we can't get our hostname, no point in continuing.
		log.Fatal(err)
	}
	defer out.Body.Close()

	body, err := ioutil.ReadAll(out.Body)

	// need logic here to bail if we don't get a proper hostname

	return string(body)
}

func main() {
	flag.Parse()

	if zoneID == "" || hostname == "" || domain == "" {
		log.Fatal("Error. You must specify a Zone ID and Hostname")
	}

	var awsHostname string
	if rr == "" {
		awsHostname = fetchAwsHostname()
	} else {
		awsHostname = rr
	}

	change := Route53Change{
		Zone:        zoneID,
		Hostname:    hostname,
		DomainName:  domain,
		AWSHostname: awsHostname,
		TTL:         ttl,
		RecordType:  recordType,
	}

	resp, err := updateRoute53(&change)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Print(resp)
	}
}
