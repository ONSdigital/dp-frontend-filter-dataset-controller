# dp-frontend-filter-dataset-controller

==================

An HTTP service for the controlling of data relevant to the filtering of a particular dataset.

## Configuration

| Environment variable         | Default                               | Description                                                                                          |
| ---------------------------- | ------------------------------------- | ---------------------------------------------------------------------------------------------------- |
| API_ROUTER_URL               | <http://localhost:23200/v1>           | The URL of the API Router                                                                            |
| BATCH_MAX_WORKERS            | 100                                   | maximum number of concurrent go-routines requesting items concurrently from APIs with pagination     |
| BATCH_SIZE_LIMIT             | 1000                                  | maximum limit value to get items from APIs in a single call                                          |
| BIND_ADDR                    | <http://localhost:20001>              | The host and port to bind to.                                                                        |
| DEBUG                        | false                                 | Enable local debugging                                                                               |
| DOWNLOAD_SERVICE_URL         | <http://localhost:23600>              | The URL of the download service                                                                      |
| ENABLE_DATASET_PREVIEW       | false                                 | Flag to add preview of dataset to output page                                                        |
| ENABLE_PROFILER              | false                                 | Flag to enable go profiler                                                                           |
| GRACEFUL_SHUTDOWN_TIMEOUT    | 5s                                    | The graceful shutdown timeout in seconds                                                             |
| HEALTHCHECK_CRITICAL_TIMEOUT | 90s                                   | The time taken for the health changes from warning state to critical due to subsystem check failures |
| HEALTHCHECK_INTERVAL         | 30s                                   | The time between calling healthcheck endpoints for check subsystems                                  |
| MAX_DATASET_OPTIONS          | 200                                   | maximum number of IDs that will be requested to dataset API in a single call as query parmeters      |
| PATTERN_LIBRARY_ASSETS_PATH  | ""                                    | Pattern library location                                                                             |
| PPROF_TOKEN                  | ""                                    | The profiling token to access service profiling                                                      |
| SEARCH_API_AUTH_TOKEN        | n/a                                   | The token used to access the Search API                                                              |
| SITE_DOMAIN                  | string                                | Domain taken from environment configs                                                                |
| OTEL_EXPORTER_OTLP_ENDPOINT  | localhost:4317                        | Endpoint for OpenTelemetry service                                                                   |
| OTEL_SERVICE_NAME            | dp-frontend-filter-dataset-controller | Label of service for OpenTelemetry service                                                           |

`HEALTHCHECK_INTERVAL` and `HEALTHCHECK_CRITICAL_TIMEOUT` can use the following formats to represent duration of time:

```
"300ms" = 300 milliseconds
"1.5h" =  1.5 hours
"2h45m" = 2 hours 45 minutes

Valid time units are
"ns" = nanosecond
"us" (or "µs") = microsecond
"ms" = millisecond
"s" = second
"m" = minute
"h" = hour
```

### Profiling

An optional `/debug` endpoint has been added, in order to profile this service via `pprof` go library.
In order to use this endpoint, you will need to enable profiler flag and set a PPROF_TOKEN:

```
export ENABLE_PROFILER=true
export PPROF_TOKEN={generated uuid}
```

Then you can us the profiler as follows:

1- Start service, load test or if on environment wait for a number of requests to be made.

2- Send authenticated request and store response in a file (this can be best done in command line like so: `curl <host>:<port>/debug/pprof/heap -H "Authorization: Bearer {generated uuid} > heap.out` - see pprof documentation on other endpoints

3- View profile either using a web ui to navigate data (a) or using pprof on command line to navigate data (b)

- a = `go tool pprof -http=:8080 heap.out`
- b = `go tool pprof heap.out`, -o flag to see various options

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details

### Licence

Copyright © 2023, Office for National Statistics (<https://www.ons.gov.uk>)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
