FROM ubuntu

RUN apt-get update && apt-get install -y ca-certificates

ADD ./bin/ecs-dns-linux-amd64 /ecs-dns-linux-amd64

ENTRYPOINT ["/ecs-dns-linux-amd64"]