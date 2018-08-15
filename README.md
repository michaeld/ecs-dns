Service Discovery with Amazon ECS, Routes53, and Prometheus
---

Build Status [![CircleCI](https://circleci.com/gh/michaeld/ecs-dns/tree/master.svg?style=svg)](https://circleci.com/gh/michaeld/ecs-dns/tree/master)

## Goals

- Automatically register and deregister ECS Services
- Enable Prometheus DNS Service Discovery configuration for ECS Services

## Usage

Run Binary
```sh
ecs-dns daemon \
--domain production1.ecs \
--zone XYZABCXYZABCXYZABC \
--interval 10 \ 
--logtostderr
```

Prometheus Configuration
```yaml
- job_name: ecs/production1/metrics
  dns_sd_configs:
  - names:
    - devops-ref-app.devops-ref-app.production1.ecs
    type: SRV
  relabel_configs:
  - source_labels: [__meta_dns_name]
    target_label: srv_record
```


Hosted Zone Result

```
RESOURCERECORDSETS	app.service.alias.	backend-service-app	0	SRV	1
RESOURCERECORDS	1 1 12448 10.1.83.187
RESOURCERECORDS	1 1 12448 10.1.28.149
RESOURCERECORDS	1 1 12448 10.1.17.186

RESOURCERECORDSETS	api.service.alias.	backend-service-api	0	SRV	1
RESOURCERECORDS	1 1 24199 10.1.152.78
RESOURCERECORDS	1 1 24199 10.1.93.53
RESOURCERECORDS	1 1 24199 10.1.129.10
RESOURCERECORDS	1 1 24199 10.1.27.80
RESOURCERECORDS	1 1 24199 10.1.83.32
```

Dig Example
```sh

$ dig devops-ref-app.devops-ref-app.production1.ecs SRV

; <<>> DiG 9.9.5-3ubuntu0.14-Ubuntu <<>> devops-ref-app.devops-ref-app.production1.ecs SRV
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 1623
;; flags: qr rd ra; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 1

;; OPT PSEUDOSECTION:
; EDNS: version: 0, flags:; udp: 4096
;; QUESTION SECTION:
;devops-ref-app.devops-ref-app.production1.ecs. IN SRV

;; ANSWER SECTION:
devops-ref-app.devops-ref-app.production1.ecs. 0 IN SRV 1 1 49648 10.1.22.107.

;; Query time: 2 msec
;; SERVER: 10.1.0.2#53(10.1.0.2)
;; WHEN: Wed Aug 15 18:10:44 UTC 2018
;; MSG SIZE  rcvd: 112

```

## Installation

### AWS Policy
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "route53:*",
            "Resource": "arn:aws:route53:::hostedzone/[HostedZoneID]"
        }
    ]
}
```

## Development

### Requirements
- Go 1.10+
- dep 0.4+

### Build
`make`

