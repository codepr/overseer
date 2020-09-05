Overseer
========

Monitor a pool of HTTP servers and gather some stats on a time-basis.
Consists of 2 microservices:
- `agent` probe a list of servers by their URL, generally an healthcheck
  endpoint, forward some stats like response time, status code and content
  to the `aggregator` service, through a messaging layer, currently using
  RabbitMQ as backend
- `aggregator` receive stats from the `agent` producing aggregated stats on
  STDOUT like mean response time, availability % of each server, top status
  code returned as they're completed
