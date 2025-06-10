FROM golang:1.21

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y \
    openvswitch-switch \
    openvswitch-common \
    golang \
    iproute2 \
    net-tools \
    curl \
    git \
    && apt-get clean

RUN mkdir -p /app
WORKDIR /app

COPY . .

COPY ./config/vswitch.ovsschema /tmp/


RUN mkdir /var/run/openvswitch && mkdir -p /etc/openvswitch && ovsdb-tool create /etc/openvswitch/conf.db /tmp/vswitch.ovsschema 

# Start OVS (can be customized if needed)

CMD ["./scripts/integration_tests.sh"]

