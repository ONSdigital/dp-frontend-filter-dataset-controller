dp-frontend-filter-dataset-controller
==================

An HTTP service for the controlling of data relevant to the filtering of a particular dataset.

### Configuration

| Environment variable          | Default                 | Description
| ----------------------------- | ----------------------- | --------------------------------------
| BIND_ADDR                     | :20001                  | The host and port to bind to.
| RENDERER_URL                  | http://localhost:20010  | The URL of dp-frontend-renderer.
| CODE_LIST_API                 | http://localhost:22400  | The URL of the code list api
| FILTER_API_URL                | http://localhost:22100  | The URL of the filter api
| DATASET_API_URL               | http://localhost:22000  | The URL of the dataset api
| HIERARCHY_API_URL             | http://localhost:22600  | The URL of the hierarchy api
| SEARCH_API_URL                | http://localhost:23100  | The URL of the search api
| DOWNLOAD_SERVICE_URL          | http://localhost:23600  | The URL of the download service
| ZEBEDEE_URL                   | http://localhost:8082   | The url of the zebedee service. Only required if profiler is enabled
| DATASET_API_AUTH_TOKEN        | n/a                     | The token used to access the Dataset API
| FILTER_API_AUTH_TOKEN         | n/a                     | The token used to access the Filter API
| SEARCH_API_AUTH_TOKEN         | n/a                     | The token used to access the Search API
| ENABLE_DATASET_PREVIEW        | false                   | Flag to add preview of dataset to output page
| ENABLE_PROFILER               | false                   | Flag to enable go profiler
| GRACEFUL_SHUTDOWN_TIMEOUT     | 5s                      | The graceful shutdown timeout in seconds
| HEALTHCHECK_INTERVAL          | 30s                     | The time between calling healthcheck endpoints for check subsystems
| HEALTHCHECK_CRITICAL_TIMEOUT  | 90s                     | The time taken for the health changes from warning state to critical due to subsystem check failures 

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
In order to use this endpoint, you will need to enable the flag and provide a valid ZebedeeURL, as it is auth-protected. Please, modify the following configuration variables to do so:
```
export ZEBEDEE_URL=your_zebedee_url
export ENABLE_PROFILER=true
```

Then you can us the profiler as follows:

1- Start service, load test or if on environment wait for a number of requests to be made.

2- Send authenticated request and store response in a file (this can be best done in command line like so: `curl <host>:<port>/debug/pprof/heap > heap.out` - see pprof documentation on other endpoints

3- View profile either using a web ui to navigate data (a) or using pprof on command line to navigate data (b) 
  a) `go tool pprof -http=:8080 heap.out`
  b) `go tool pprof heap.out`, -o flag to see various options

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details

### Licence

Copyright ©‎ 2017, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.

