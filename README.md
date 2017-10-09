dp-frontend-filter-dataset-controller
==================

An HTTP service for the controlling of data relevant to a particular dataset.

### Configuration

| Environment variable | Default                 | Description
| -------------------- | ----------------------- | --------------------------------------
| BIND_ADDR            | :20001                  | The host and port to bind to.
| RENDERER_URL         | http://localhost:20010  | The URL of dp-frontend-renderer.
| CODE_LIST_API        | http://localhost:22400  | The URL of the code list api
| FILTER_API_URL       | http://localhost:22100  | The URL of the filter api
| DATASET_API_URL      | http://localhost:22000  | The URL of the dataset api
| HIERARCHY_API_URL    | http://localhost:22600  | The URL of the hierarchy api



### Licence

Copyright ©‎ 2017, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
