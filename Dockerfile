FROM golang:1.21 AS build

WORKDIR /usr/src/l2sm-switch

COPY . ./

RUN go build -o /usr/local/bin/talpa 

FROM ubuntu:jammy 

WORKDIR /usr/local/bin

COPY ./config/vswitch.ovsschema /tmp/

COPY --from=build /usr/local/bin/ .

RUN apt-get update && \
  apt-get install -y net-tools iproute2 netcat-openbsd dnsutils curl iputils-ping iptables nmap tcpdump openvswitch-switch 
  
RUN mkdir /var/run/openvswitch && mkdir -p /etc/openvswitch && ovsdb-tool create /etc/openvswitch/conf.db /tmp/vswitch.ovsschema 

RUN  mkdir /etc/l2sm/

ENTRYPOINT [ "talpa" ]