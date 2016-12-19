#!/usr/bin/env bash

set -x
set -e

IP_ADDRESS=192.168.1.214
LOGIN=root

env GOOS=linux go build

ssh -p 2202 $LOGIN@$IP_ADDRESS "systemctl stop hamrchecker.service; rm -rf /usr/local/src/resources"
scp -P 2202 ./hamrchecker $LOGIN@$IP_ADDRESS:/usr/local/src/hamrchecker
scp -P 2202 -r ./resources $LOGIN@$IP_ADDRESS:/usr/local/src/resources
scp -P 2202 ./hamrchecker.service $LOGIN@$IP_ADDRESS:/etc/systemd/system/hamrchecker.service
ssh -p 2202 $LOGIN@$IP_ADDRESS "systemctl daemon-reload; systemctl start hamrchecker.service"