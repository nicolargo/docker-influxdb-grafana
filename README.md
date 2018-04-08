# Docker-compose for statsd telegraf and grafana

To run the docker stack:
```
docker-compose up --build --force-recreate -d
```

**Grafana** is available at [http://localhost:3000](http://localhost:3000)
**statsd** can be written to using `echo "mycounter:10|c" | nc -C -w 1 -u ${host_ip} 8125`
**influxdb** prompt `docker exec -it influxdb influx`

To add a new grafana dashboard:
1. create the dashboard in grafana
1. copy the dashboard json into a file in the `grafana/dashboards` directory
1. remove the top-level `"id"` json field
1. restart the stack
