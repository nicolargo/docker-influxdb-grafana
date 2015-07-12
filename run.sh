#!/bin/bash

docker-compose up -d

echo "Grafana: http://localhost:3000 - admin/admin"
echo "InfluxDB: http://localhost:8083 - root/root"