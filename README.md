# underlay-to-overlay-proxy
An extremely simple application to translate traditional underlay HTTP requests to overlay requests for healthchecks

OPENZITI_HEALTHCHECK_PROXY_PORTa=8000
OPENZITI_HEALTHCHECK_REPLACE_REGEX=edgex.$1
OPENZITI_HEALTHCHECK_IDENTITY=/home/cd/git/github/openziti/test-kitchen/healthcheck-proxy/health.json

docker run \
    --rm \
    -v $(pwd)/health.json:/opt/openziti/underlay-host-proxy/identity.json \
    -e OPENZITI_HEALTHCHECK_SEARCH_REGEX='(.*).edgex.ziti' \
    -e OPENZITI_HEALTHCHECK_REPLACE_REGEX='edgex.$1' \
    -e OPENZITI_HEALTHCHECK_PROXY_PORT=80 \
    -p 2171:2171 \
    --network edgex_edgex-network \
    --network-alias app-rules-engine.edgex.ziti \
    --network-alias core-command.edgex.ziti \
    --network-alias core-data.edgex.ziti \
    --network-alias core-metadata.edgex.ziti \
    --network-alias device-rest.edgex.ziti \
    --network-alias device-virtual.edgex.ziti \
    --network-alias support-notifications.edgex.ziti \
    --network-alias support-scheduler.edgex.ziti \
    openziti/underlay-host-proxy


ziti edge create identity health -o health.jwt -a edgex.app-rules-engine-clients,edgex.device-rest-clients,edgex.core-command-clients,edgex.core-data-clients,edgex.core-metadata-clients,edgex.device-virtual-clients,edgex.rules-engine-clients,edgex.support-notifications-clients,edgex.support-scheduler-clients,edgex.sys-mgmt-agent-clients







