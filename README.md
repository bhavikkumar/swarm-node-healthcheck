# swarm-node-healthcheck
Simple HTTP server written in Go which uses the Docker SDK to check if the current server is correctly functioning inside the swarm cluster.

## Building the project
This project uses dep so it must be on your path to begin with.
```
dep ensure
go build
docker build -t swarm-node-healthcheck
```

## Running the container
The following will work on Linux and Docker for Windows, it has not been tried on OSX.
```
docker run --restart=always -p 44444:44444 -v /var/run/docker.sock:/var/run/docker.sock depost/swarm-node-healthcheck:latest
```
