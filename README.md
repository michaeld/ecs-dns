Service Discovery with Amazon ECS, Routes53, and Prometheus
---

## Goals

- Automatically register and deregister ECS Services
- Enable Prometheus DNS Service Discovery configuration for ECS Services

## Usage

Run Binary
```sh
ecs-dns daemon \
--domain service.alias \
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
- Go 1.9+
- dep 0.4+

### Build
`make`

