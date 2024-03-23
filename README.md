# Healthcheck Proxy

This project represents an extremely simple application with the specific purpose to translate traditional underlay 
HTTP requests to overlay requests. The intended purpose of this project is to support legacy health checking by 
hostname. Where you would traditionally execute an HTTP healthcheck to an endpoint such as: "http://localhost:80" or 
"https://some.server/some/path", these underlay-based healthchecks become useless when adapting to use 
application-embedded zero trust as there are no longer listening ports on the underlay network. 

This project allows you to inspect the incoming hostname from the underlay network, convert that hostname to an 
OpenZiti service, then dial/connect to that service via OpenZiti to perform the healthcheck.

At this point there will be a few options:
* adapt the application performing the healthcheck to OpenZiti and application embedded zero trust
* inject a tunneler on the host or near the host and allow all traffic to be proxied in the normal ways
* use this project to allow more refined, HTTP-based healthchecks

The project has rudimentary controls to allow controlling which paths, which hosts, and which HTTP verbs will be 
supported.

## Configuration

The following options are available by setting a corresponding environment variable

| Environment Variable                      | Purpose                                                                                                       |
|-------------------------------------------|---------------------------------------------------------------------------------------------------------------|
| `OPENZITI_HEALTHCHECK_IDENTITY`           | The path to the identity file, used to proxy healthchecks. Default: `/opt/openziti/healthcheck/identity.json` |
| `OPENZITI_HEALTHCHECK_PROXY_PORT`         | This port the process will listen on for incoming HTTP requests. Default: `2171`                              |
| `OPENZITI_HEALTHCHECK_ALLOWED_PATH`       | The allowed path to proxy. Must be specified else the default `^[.*]$` is used which matches nothing.         |
| `OPENZITI_HEALTHCHECK_ALLOWED_VERB_REGEX` | A regex that controls the allowed HTTP verb(s). Default: `GET`                                            |
| `OPENZITI_HEALTHCHECK_SEARCH_REGEX`       | The part of the host to match. Default: `(.*)`                                                                |
| `OPENZITI_HEALTHCHECK_REPLACE_REGEX`      | The replacement pattern for the host. Default: `$1`                                                           |
| `OPENZITI_HEALTHCHECK_DEBUG`              | Use "debug" to turn on debug mode, showing the proxied requests. Default: "info"                              |
| `OPENZITI_HEALTHCHECK_CERT`               | A certificate to use when serving HTTPS. Default: `""`                                                        |
| `OPENZITI_HEALTHCHECK_KEY`                | The key to use with the certificate specified in `OPENZITI_HEALTHCHECK_CERT`. Default: `""`                   |

## Example of Running as Docker

This example demonstrates using this program with something like [EdgeX Foundry](https://www.edgexfoundry.org/). By 
default, EdgeX Foundry operates with numerous microservices, all of which are probed via HTTP using 
[Consul](https://www.consul.io/). EdgeX Foundry adopted OpenZiti into the core services and doing so prevents these 
traditional-style underlay health checks from succeeding. In order to allow Consul to probe the EdgeX services, a 
container can be brought online that participates in the same docker network, numerous aliases can be assigned to 
the container, and a single instance of this program is able to act as a healthcheck proxy to the core services 
which are now secured via OpenZiti. 

An example of running this process in a docker container helps explain what's going on:
```
docker run \
    --rm \
    -v $(pwd)/health.json:/opt/openziti/underlay-host-proxy/identity.json \
    -e OPENZITI_HEALTHCHECK_ALLOWED_PATH='^.*/ping$' \
    -e OPENZITI_HEALTHCHECK_SEARCH_REGEX='(.*).edgex.ziti' \
    -e OPENZITI_HEALTHCHECK_REPLACE_REGEX='edgex.$1' \
    -e OPENZITI_HEALTHCHECK_PROXY_PORT=80 \
    --network edgex_edgex-network \
    --network-alias app-rules-engine.edgex.ziti \
    --network-alias core-command.edgex.ziti \
    --network-alias core-data.edgex.ziti \
    --network-alias core-metadata.edgex.ziti \
    --network-alias device-rest.edgex.ziti \
    --network-alias device-virtual.edgex.ziti \
    --network-alias support-notifications.edgex.ziti \
    --network-alias support-scheduler.edgex.ziti \
    openziti/healthcheck-proxy
```

Breaking down each section we see:
* `docker run --rm` - Here I'm starting this container directly using the `docker` command. This could obviously be 
  adapted to a `compose` file. This command instructs `docker` to run and then remove itself when done.
* `-v $(pwd)/health.json:/opt/openziti/underlay-host-proxy/identity.json`. I created and enrolled an identity and 
  authorized it to contact all the EdgeX Foundry services that needed to be checked for health.
* `-e OPENZITI_HEALTHCHECK_ALLOWED_PATH='^.*/ping$'`. This allows only requests that end with `/ping` to be proxied 
  through this service.
* `-e OPENZITI_HEALTHCHECK_SEARCH_REGEX='(.*).edgex.ziti'`. This converts the configured request, which comes to 
  this server as `<service-name>.edgex.ziti` and captures just the service name so that this service can dial the 
  service with the specified name.
* `-e OPENZITI_HEALTHCHECK_REPLACE_REGEX='edgex.$1'`. This injects `edgex.` into the discovered service name. The 
  services are all grouped by `edgex.` within the controller, this allows me to inject the prefix into the service 
  name dynamically.
* `-e OPENZITI_HEALTHCHECK_PROXY_PORT=80`. All OpenZiti secured services are expected to be sent to port 80, this 
  instructs the server to listen on port 80.
* `--network edgex_edgex-network`. This joins the docker container to the expected docker network, allowing docker 
  to handle the network aliases specified below and allowing the consul container to resolve and send traffic 
  (entirely internal to docker) to this process.
* `--network-alias app-rules-engine.edgex.ziti`. This and the following `network-alias` entries instruct docker to 
  use this as an alias for this container. Notice all the services needing a healthcheck are listed here.
* `--network-alias core-command.edgex.ziti`. See above.
* `--network-alias core-data.edgex.ziti`. See above.
* `--network-alias core-metadata.edgex.ziti`. See above.
* `--network-alias device-rest.edgex.ziti`. See above.
* `--network-alias device-virtual.edgex.ziti`. See above.
* `--network-alias support-notifications.edgex.ziti`. See above.
* `--network-alias support-scheduler.edgex.ziti`. See above.
* `openziti/healthcheck-proxy`. This is the container for docker to execute

## Example of Creating/Authorizing an Identity
This is just an example for how the identity for EdgeX Foundry was generated:
```
ziti edge create identity \
  health -o health.jwt \
  -a 'edgex.app-rules-engine-clients,edgex.device-rest-clients,edgex.core-command-clients,'\
'edgex.core-data-clients,edgex.core-metadata-clients,edgex.device-virtual-clients,'\
'edgex.rules-engine-clients,edgex.support-notifications-clients,edgex.support-scheduler-clients,'\
'edgex.sys-mgmt-agent-clients'
```