# glimpse-scan

glimpse-scan polls [vitals-glimpse](https://github.com/IanKulin/vitals-glimpse) endpoints to collect server metrics and saves them to a v2 InfluxDB time-series database.

vitals-glimpse endpoints return JSON in the format

```json
{
  "title": "vitals-glimpse",
  "version": 0.2,
  "mem_status": "mem_okay",
  "mem_percent": 46,
  "disk_status": "disk_okay",
  "disk_percent": 79,
  "cpu_status": "cpu_okay",
  "cpu_percent": 5
}
```

When the data are saved in Influx, it's tagged with the server name and the fields are the metrics. 

Database parameters are loaded from the environment variables:
```
INFLUXDB_ORG=ksd
INFLUXDB_BUCKET=server_metrics
INFLUXDB_ADMIN_TOKEN=some_long_token_defined_in_the_influx_setup
INFLUXDB_URL=http://100.106.90.55:8086
POLLING_INTERVAL_MINUTES=5
```

### Build and run for testing
`docker build --platform linux/amd64 -t ghcr.io/iankulin/glimpse_scan:latest .`
`docker compose up`

### Build and push for production
`docker build --platform linux/amd64 -t ghcr.io/iankulin/glimpse_scan:latest .`
`docker push ghcr.io/iankulin/glimpse_scan:latest`