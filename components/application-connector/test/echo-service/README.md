# Echo Service

This is internal tool implemented to test Wormhole connection - a simple HTTP service which echoes REST/HTTP requests back to user.

It responds with data that was sent in a request and any additional information that user requested through special headers.

## Building And Running Echo Service

### Building Echo Service

To build Echo Service you can use:

```bash
go build -o echo echo.go logger.go types.go
```

### Starting Echo Service Locally

The command to run Echo Service is:

```bash
go run echo.go logger.go types.go
```

If you've built Echo Service in previous step you can also execute created binary:

```bash
./echo
```

Echo Service can also be ran by using Makefile, to do so following command needs to be executed in root directory of the project:

```bash
make start-echo-service
```

### Building And Publishing Docker Image

Echo Service Docker image can be built using:
```bash
docker build -t eu.gcr.io/kyma-project/tools/echo-service:0.0.1 -f Dockerfile-echo .
```
The above command should be executed from root project catalog.

In order to publish it to repository following command needs to be executed:

```bash
docker push eu.gcr.io/kyma-project/tools/echo-service:0.0.1
```

### Running Dockerized Echo Service

First of all, docker image needs to be pulled from repository:

```bash
docker pull eu.gcr.io/kyma-project/tools/echo-service:0.0.1
```

And then, it can be started:

```bash
docker run eu.gcr.io/kyma-project/tools/echo-service:0.0.1
```

### Changing Port Number

Default port for Echo Service is 9000, to change it simply add --port='XXXX' flag.
For example:

```bash
go run echo.go types.go logger.go --port="8080"
```


## Using Echo Service

Echo Service accepts special headers to modify it's response:

### Requesting Response Headers

Default headers returned in response are:
- Content-Length
- Content-Type
- Date

For the test purposes, you can request to return the additional headers together with the echo response.

In order to receive any additional headers in response you can apply header to your request:
```
echo-header-yourHeaderName : yourHeaderValue
```

After echo-service receives header of this structure it will append response headers with:
```
yourHeaderName : yourHeaderValue
```

### Requesting Response Status Code

By default Echo Service responds with status code 200, for testing you can request exact status code to be returned with the response.

In order to force Echo Service to respond with given status code you can apply a header to your request:
```
echo-statuscode : requestedValue
```

### Response Structure

Echo Service always responds with a JSON body of following structure:

```
{
  "path":"/path/that/was/requested",
  "method":"GET/POST/PUT/...",
  "headers:" {
    "HeaderName":"HeaderValue",
    "AnotherHeader":"AnotherValue",
  },
  "body":"bodyThatWasSent"
}
```

### Example Request/Response

Executing given request:

```bash
curl -X POST -H "Content-Type: application/json" -d '{"testValue1":"xyz","testValue2":"xyz"}' http://localhost:9000/test
```

Will result in following response:

```json
{
  "path":"/test",
  "method":"POST",
  "headers":{
    "Accept":["*/*"],
    "Content-Length":["39"],
    "Content-Type":["application/json"],
    "User-Agent":["curl/7.54.0"]
  },
  "body":"{\"testValue1\":\"xyz\",\"testValue2\":\"xyz\"}"
}
```
