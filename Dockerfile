FROM scratch
EXPOSE 44444
ADD swarm-node-healthcheck /swarm-node-healthcheck
ENTRYPOINT ["/swarm-node-healthcheck"]
