#!/bin/bash

./scripts/setup_ovs.sh
go test -v -tags=integration ./pkg/ovs