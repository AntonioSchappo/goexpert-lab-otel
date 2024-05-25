# goexpert-lab-otel

This project is a simple implementation of [Open Telemetry](https://opentelemetry.io/docs/languages/go/) observability with Go.

The goal of this application is to visualize the spans present in services A and B while they communicate.

Service A receives a Brazilian ZipCode and starts the first span. 

The zipcode is validated and forwarded to Service B to be processed.

Service B tracks the processing through 3 spans. The first one marks the moment the req is received, the second one marks the moment the Brazilian Zipcode API is called and the final one marks the moment the Weather API is called.

The final response is a JSON with the weather information for the chosen Location.

It is possible to visualize the spans through a [Zipkin](https://zipkin.io/) dashboard.


## Commands to start the application locally

1. Just run the command below to start 

```sh
docker compose up -d
```

2. The server for service A will be available at the url below:

```sh
http://localhost:8080
```

3. The main route is accessed by informing a valid CEP as a path parameter. 
http://localhost:8080/{cep}

```sh
http://localhost:8080/01153000
```

4. In order to visualize the telemetry dashboard after executing some queries, please access the url below and click the button "RUN QUERY"

```sh
http://localhost:9411/zipkin
```

## Test Reqs

On the `/api` folder there are four http files with examples of requests that can be made to the Rest server

If you are using VSCode as your IDE downloading the [Rest Client Plugin](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) is an excellent way execute these requests.
