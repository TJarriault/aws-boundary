---
title: VirtualServer and VirtualServerRoute Resources
description:
weight: 1600
doctypes: [""]
aliases:
    - /virtualserver-and-virtualserverroute/
toc: true
---


The VirtualServer and VirtualServerRoute resources are new load balancing configuration, introduced in release 1.5 as an alternative to the Ingress resource. The resources enable use cases not supported with the Ingress resource, such as traffic splitting and advanced content-based routing. The resources are implemented as [Custom Resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/).

This document is the reference documentation for the resources. To see additional examples of using the resources for specific use cases, go to the [examples/custom-resources](https://github.com/nginxinc/kubernetes-ingress/tree/v2.0.3/examples/custom-resources) folder in our GitHub repo.

## VirtualServer Specification

The VirtualServer resource defines load balancing configuration for a domain name, such as `example.com`. Below is an example of such configuration:
```yaml
apiVersion: k8s.nginx.org/v1
kind: VirtualServer
metadata:
  name: cafe
spec:
  host: cafe.example.com
  tls:
    secret: cafe-secret
  upstreams:
  - name: tea
    service: tea-svc
    port: 80
  - name: coffee
    service: coffee-svc
    port: 80
  routes:
  - path: /tea
    action:
      pass: tea
  - path: /coffee
    action:
      pass: coffee
  - path: ~ ^/decaf/.*\\.jpg$
    action:
      pass: coffee
  - path: = /green/tea
    action:
      pass: tea
```

{{% table %}}
|Field | Description | Type | Required |
| ---| ---| ---| --- |
|``host`` | The host (domain name) of the server. Must be a valid subdomain as defined in RFC 1123, such as ``my-app`` or ``hello.example.com``. Wildcard domains like ``*.example.com`` are not allowed.  The ``host`` value needs to be unique among all Ingress and VirtualServer resources. See also [Handling Host and Listener Collisions](/nginx-ingress-controller/configuration/handling-host-and-listener-collisions). | ``string`` | Yes |
|``tls`` | The TLS termination configuration. | [tls](#virtualservertls) | No |
|``dos`` | A reference to a DosProtectedResource, setting this enables DOS protection of the VirtualServer. | ``string`` | No |
|``policies`` | A list of policies. | [[]policy](#virtualserverpolicy) | No |
|``upstreams`` | A list of upstreams. | [[]upstream](#upstream) | No |
|``routes`` | A list of routes. | [[]route](#virtualserver-route) | No |
|``ingressClassName`` | Specifies which Ingress controller must handle the VirtualServer resource. | ``string`` | No |
|``http-snippets`` | Sets a custom snippet in the http context. | ``string`` | No |
|``server-snippets`` | Sets a custom snippet in server context. Overrides the ``server-snippets`` ConfigMap key. | ``string`` | No |
{{% /table %}}

### VirtualServer.TLS

The tls field defines TLS configuration for a VirtualServer. For example:
```yaml
secret: cafe-secret
redirect:
  enable: true
```

{{% table %}}
|Field | Description | Type | Required |
| ---| ---| ---| --- |
|``secret`` | The name of a secret with a TLS certificate and key. The secret must belong to the same namespace as the VirtualServer. The secret must be of the type ``kubernetes.io/tls`` and contain keys named ``tls.crt`` and ``tls.key`` that contain the certificate and private key as described [here](https://kubernetes.io/docs/concepts/services-networking/ingress/#tls). If the secret doesn't exist or is invalid, NGINX will break any attempt to establish a TLS connection to the host of the VirtualServer. If the secret is not specified but [wildcard TLS secret](/nginx-ingress-controller/configuration/global-configuration/command-line-arguments#cmdoption-wildcard-tls-secret) is configured, NGINX will use the wildcard secret for TLS termination. | ``string`` | No |
|``redirect`` | The redirect configuration of the TLS for a VirtualServer. | [tls.redirect](#virtualservertlsredirect) | No | ### VirtualServer.TLS.Redirect |
{{% /table %}}

### VirtualServer.TLS.Redirect

The redirect field configures a TLS redirect for a VirtualServer:
```yaml
enable: true
code: 301
basedOn: scheme
```

{{% table %}}
|Field | Description | Type | Required |
| ---| ---| ---| --- |
|``enable`` | Enables a TLS redirect for a VirtualServer. The default is ``False``. | ``boolean`` | No |
|``code`` | The status code of a redirect. The allowed values are: ``301`` , ``302`` , ``307`` , ``308``.  The default is ``301``. | ``int`` | No |
|``basedOn`` | The attribute of a request that NGINX will evaluate to send a redirect. The allowed values are ``scheme`` (the scheme of the request) or ``x-forwarded-proto`` (the ``X-Forwarded-Proto`` header of the request). The default is ``scheme``. | ``string`` | No | ### VirtualServer.Policy |
{{% /table %}}

### VirtualServer.Policy

The policy field references a [Policy resource](/nginx-ingress-controller/configuration/policy-resource/) by its name and optional namespace. For example:
```yaml
name: access-control
```


{{% table %}}
|Field | Description | Type | Required |
| ---| ---| ---| --- |
|``name`` | The name of a policy. If the policy doesn't exist or invalid, NGINX will respond with an error response with the `500` status code. | ``string`` | Yes |
|``namespace`` | The namespace of a policy. If not specified, the namespace of the VirtualServer resource is used. | ``string`` | No |
{{% /table %}}

### VirtualServer.Route

The route defines rules for matching client requests to actions like passing a request to an upstream. For example:
```yaml
  path: /tea
  action:
    pass: tea
```

{{% table %}}
|Field | Description | Type | Required |
| ---| ---| ---| --- |
|``path`` | The path of the route. NGINX will match it against the URI of a request. Possible values are: a prefix ( ``/`` , ``/path`` ), an exact match ( ``=/exact/match`` ), a case insensitive regular expression ( ``~*^/Bar.*\.jpg`` ) or a case sensitive regular expression ( ``~^/foo.*\.jpg`` ). In the case of a prefix (must start with ``/`` ) or an exact match (must start with ``=`` ), the path must not include any whitespace characters, ``{`` , ``}`` or ``;``. In the case of the regex matches, all double quotes ``"`` must be escaped and the match can't end in an unescaped backslash ``\``. The path must be unique among the paths of all routes of the VirtualServer. Check the [location](https://nginx.org/en/docs/http/ngx_http_core_module.html#location) directive for more information. | ``string`` | Yes |
|``policies`` | A list of policies. The policies override the policies of the same type defined in the ``spec`` of the VirtualServer. See [Applying Policies](/nginx-ingress-controller/configuration/policy-resource/#applying-policies) for more details. | [[]policy](#virtualserverpolicy) | No |
|``action`` | The default action to perform for a request. | [action](#action) | No |
|``dos`` | A reference to a DosProtectedResource, setting this enables DOS protection of the VirtualServer route. | ``string`` | No |
|``splits`` | The default splits configuration for traffic splitting. Must include at least 2 splits. | [[]split](#split) | No |
|``matches`` | The matching rules for advanced content-based routing. Requires the default ``action`` or ``splits``.  Unmatched requests will be handled by the default ``action`` or ``splits``. | [matches](#match) | No |
|``route`` | The name of a VirtualServerRoute resource that defines this route. If the VirtualServerRoute belongs to a different namespace than the VirtualServer, you need to include the namespace. For example, ``tea-namespace/tea``. | ``string`` | No |
|``errorPages`` | The custom responses for error codes. NGINX will use those responses instead of returning the error responses from the upstream servers or the default responses generated by NGINX. A custom response can be a redirect or a canned response. For example, a redirect to another URL if an upstream server responded with a 404 status code. | [[]errorPage](#errorpage) | No |
|``location-snippets`` | Sets a custom snippet in the location context. Overrides the ``location-snippets`` ConfigMap key. | ``string`` | No |
{{% /table %}}

\* -- a route must include exactly one of the following: `action`, `splits`, or `route`.

## VirtualServerRoute Specification

The VirtualServerRoute resource defines a route for a VirtualServer. It can consist of one or multiple subroutes. The VirtualServerRoute is an alternative to [Mergeable Ingress types](/nginx-ingress-controller/configuration/ingress-resources/cross-namespace-configuration).

In the example below, the VirtualServer `cafe` from the namespace `cafe-ns` defines a route with the path `/coffee`, which is further defined in the VirtualServerRoute `coffee` from the namespace `coffee-ns`.

VirtualServer:
```yaml
apiVersion: k8s.nginx.org/v1
kind: VirtualServer
metadata:
  name: cafe
  namespace: cafe-ns
spec:
  host: cafe.example.com
  upstreams:
  - name: tea
    service: tea-svc
    port: 80
  routes:
  - path: /tea
    action:
      pass: tea
  - path: /coffee
    route: coffee-ns/coffee
```

VirtualServerRoute:
```yaml
apiVersion: k8s.nginx.org/v1
kind: VirtualServerRoute
metadata:
  name: coffee
  namespace: coffee-ns
spec:
  host: cafe.example.com
  upstreams:
  - name: latte
    service: latte-svc
    port: 80
  - name: espresso
    service: espresso-svc
    port: 80
  subroutes:
  - path: /coffee/latte
    action:
      pass: latte
  - path: /coffee/espresso
    action:
      pass: espresso
```

Note that each subroute must have a `path` that starts with the same prefix (here `/coffee`), which is defined in the route of the VirtualServer. Additionally, the `host` in the VirtualServerRoute must be the same as the `host` of the VirtualServer.

{{% table %}}
|Field | Description | Type | Required |
| ---| ---| ---| --- |
|``host`` | The host (domain name) of the server. Must be a valid subdomain as defined in RFC 1123, such as ``my-app`` or ``hello.example.com``. Wildcard domains like ``*.example.com`` are not allowed. Must be the same as the ``host`` of the VirtualServer that references this resource. | ``string`` | Yes |
|``upstreams`` | A list of upstreams. | [[]upstream](#upstream) | No |
|``subroutes`` | A list of subroutes. | [[]subroute](#virtualserverroutesubroute) | No |
|``ingressClassName`` | Specifies which Ingress controller must handle the VirtualServerRoute resource. Must be the same as the ``ingressClassName`` of the VirtualServer that references this resource. | ``string``_ | No |
{{% /table %}}

### VirtualServerRoute.Subroute

The subroute defines rules for matching client requests to actions like passing a request to an upstream. For example:
```yaml
path: /coffee
action:
  pass: coffee
```

{{% table %}}
|Field | Description | Type | Required |
| ---| ---| ---| --- |
|``path`` | The path of the subroute. NGINX will match it against the URI of a request. Possible values are: a prefix ( ``/`` , ``/path`` ), an exact match ( ``=/exact/match`` ), a case insensitive regular expression ( ``~*^/Bar.*\.jpg`` ) or a case sensitive regular expression ( ``~^/foo.*\.jpg`` ). In the case of a prefix, the path must start with the same path as the path of the route of the VirtualServer that references this resource. In the case of an exact or regex match, the path must be the same as the path of the route of the VirtualServer that references this resource. In the case of a prefix or an exact match, the path must not include any whitespace characters, ``{`` , ``}`` or ``;``.  In the case of the regex matches, all double quotes ``"`` must be escaped and the match can't end in an unescaped backslash ``\``. The path must be unique among the paths of all subroutes of the VirtualServerRoute. | ``string`` | Yes |
|``policies`` | A list of policies. The policies override *all* policies defined in the route of the VirtualServer that references this resource. The policies also override the policies of the same type defined in the ``spec`` of the VirtualServer. See [Applying Policies](/nginx-ingress-controller/configuration/policy-resource/#applying-policies) for more details. | [[]policy](#virtualserverpolicy) | No |
|``action`` | The default action to perform for a request. | [action](#action) | No |
|``dos`` | A reference to a DosProtectedResource, setting this enables DOS protection of the VirtualServerRoute subroute. | ``string`` | No |
|``splits`` | The default splits configuration for traffic splitting. Must include at least 2 splits. | [[]split](#split) | No |
|``matches`` | The matching rules for advanced content-based routing. Requires the default ``action`` or ``splits``.  Unmatched requests will be handled by the default ``action`` or ``splits``. | [matches](#match) | No |
|``errorPages`` | The custom responses for error codes. NGINX will use those responses instead of returning the error responses from the upstream servers or the default responses generated by NGINX. A custom response can be a redirect or a canned response. For example, a redirect to another URL if an upstream server responded with a 404 status code. | [[]errorPage](#errorpage) | No |
|``location-snippets`` | Sets a custom snippet in the location context. Overrides the ``location-snippets`` of the VirtualServer (if set) or the ``location-snippets`` ConfigMap key. | ``string`` | No |
{{% /table %}}

\* -- a subroute must include exactly one of the following: `action` or `splits`.

## Common Parts of the VirtualServer and VirtualServerRoute

### Upstream

The upstream defines a destination for the routing configuration. For example:
```yaml
name: tea
service: tea-svc
subselector:
  version: canary
port: 80
lb-method: round_robin
fail-timeout: 10s
max-fails: 1
max-conns: 32
keepalive: 32
connect-timeout: 30s
read-timeout: 30s
send-timeout: 30s
next-upstream: "error timeout non_idempotent"
next-upstream-timeout: 5s
next-upstream-tries: 10
client-max-body-size: 2m
tls:
  enable: true
```

**Note**: The WebSocket protocol is supported without any additional configuration.

{{% table %}}
|Field | Description | Type | Required |
| ---| ---| ---| --- |
|``name`` | The name of the upstream. Must be a valid DNS label as defined in RFC 1035. For example, ``hello`` and ``upstream-123`` are valid. The name must be unique among all upstreams of the resource. | ``string`` | Yes |
|``service`` | The name of a [service](https://kubernetes.io/docs/concepts/services-networking/service/). The service must belong to the same namespace as the resource. If the service doesn't exist, NGINX will assume the service has zero endpoints and return a ``502`` response for requests for this upstream. For NGINX Plus only, services of type [ExternalName](https://kubernetes.io/docs/concepts/services-networking/service/#externalname) are also supported (check the [prerequisites](https://github.com/nginxinc/kubernetes-ingress/tree/v2.0.3/examples/externalname-services#prerequisites) ). | ``string`` | Yes |
|``subselector`` | Selects the pods within the service using label keys and values. By default, all pods of the service are selected. Note: the specified labels are expected to be present in the pods when they are created. If the pod labels are updated, the Ingress Controller will not see that change until the number of the pods is changed. | ``map[string]string`` | No |
|``use-cluster-ip`` | Enables using the Cluster IP and port of the service instead of the default behavior of using the IP and port of the pods. When this field is enabled, the fields that configure NGINX behavior related to multiple upstream servers (like ``lb-method`` and ``next-upstream``) will have no effect, as the Ingress Controller will configure NGINX with only one upstream server that will match the service Cluster IP. | ``boolean`` | No |
|``port`` | The port of the service. If the service doesn't define that port, NGINX will assume the service has zero endpoints and return a ``502`` response for requests for this upstream. The port must fall into the range ``1..65535``. | ``uint16`` | Yes |
|``lb-method`` | The load [balancing method](https://docs.nginx.com/nginx/admin-guide/load-balancer/http-load-balancer/#choosing-a-load-balancing-method). To use the round-robin method, specify ``round_robin``. The default is specified in the ``lb-method`` ConfigMap key. | ``string`` | No |
|``fail-timeout`` | The time during which the specified number of unsuccessful attempts to communicate with an upstream server should happen to consider the server unavailable. See the [fail_timeout](https://nginx.org/en/docs/http/ngx_http_upstream_module.html#fail_timeout) parameter of the server directive. The default is set in the ``fail-timeout`` ConfigMap key. | ``string`` | No |
|``max-fails`` | The number of unsuccessful attempts to communicate with an upstream server that should happen in the duration set by the ``fail-timeout`` to consider the server unavailable. See the [max_fails](https://nginx.org/en/docs/http/ngx_http_upstream_module.html#max_fails) parameter of the server directive. The default is set in the ``max-fails`` ConfigMap key. | ``int`` | No |
|``max-conns`` | The maximum number of simultaneous active connections to an upstream server. See the [max_conns](https://nginx.org/en/docs/http/ngx_http_upstream_module.html#max_conns) parameter of the server directive. By default there is no limit. Note: if keepalive connections are enabled, the total number of active and idle keepalive connections to an upstream server may exceed the ``max_conns`` value. | ``int`` | No |
|``keepalive`` | Configures the cache for connections to upstream servers. The value ``0`` disables the cache. See the [keepalive](https://nginx.org/en/docs/http/ngx_http_upstream_module.html#keepalive) directive. The default is set in the ``keepalive`` ConfigMap key. | ``int`` | No |
|``connect-timeout`` | The timeout for establishing a connection with an upstream server. See the [proxy_connect_timeout](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_connect_timeout) directive. The default is specified in the ``proxy-connect-timeout`` ConfigMap key. | ``string`` | No |
|``read-timeout`` | The timeout for reading a response from an upstream server. See the [proxy_read_timeout](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_read_timeout) directive.  The default is specified in the ``proxy-read-timeout`` ConfigMap key. | ``string`` | No |
|``send-timeout`` | The timeout for transmitting a request to an upstream server. See the [proxy_send_timeout](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_send_timeout) directive. The default is specified in the ``proxy-send-timeout`` ConfigMap key. | ``string`` | No |
|``next-upstream`` | Specifies in which cases a request should be passed to the next upstream server. See the [proxy_next_upstream](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_next_upstream) directive. The default is ``error timeout``. | ``string`` | No |
|``next-upstream-timeout`` | The time during which a request can be passed to the next upstream server. See the [proxy_next_upstream_timeout](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_next_upstream_timeout) directive. The ``0`` value turns off the time limit. The default is ``0``. | ``string`` | No |
|``next-upstream-tries`` | The number of possible tries for passing a request to the next upstream server. See the [proxy_next_upstream_tries](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_next_upstream_tries) directive. The ``0`` value turns off this limit. The default is ``0``. | ``int`` | No |
|``client-max-body-size`` | Sets the maximum allowed size of the client request body. See the [client_max_body_size](https://nginx.org/en/docs/http/ngx_http_core_module.html#client_max_body_size) directive. The default is set in the ``client-max-body-size`` ConfigMap key. | ``string`` | No |
|``tls`` | The TLS configuration for the Upstream. | [tls](#upstreamtls) | No |
|``healthCheck`` | The health check configuration for the Upstream. See the [health_check](https://nginx.org/en/docs/http/ngx_http_upstream_hc_module.html#health_check) directive. Note: this feature is supported only in NGINX Plus. | [healthcheck](#upstreamhealthcheck) | No |
|``slow-start`` | The slow start allows an upstream server to gradually recover its weight from 0 to its nominal value after it has been recovered or became available or when the server becomes available after a period of time it was considered unavailable. By default, the slow start is disabled. See the [slow_start](https://nginx.org/en/docs/http/ngx_http_upstream_module.html#slow_start) parameter of the server directive. Note: The parameter cannot be used along with the ``random`` , ``hash`` or ``ip_hash`` load balancing methods and will be ignored. | ``string`` | No |
|``queue`` | Configures a queue for an upstream. A client request will be placed into the queue if an upstream server cannot be selected immediately while processing the request. By default, no queue is configured. Note: this feature is supported only in NGINX Plus. | [queue](#upstreamqueue) | No |
|``buffering`` | Enables buffering of responses from the upstream server. See the [proxy_buffering](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_buffering) directive. The default is set in the ``proxy-buffering`` ConfigMap key. | ``boolean`` | No |
|``buffers`` | Configures the buffers used for reading a response from the upstream server for a single connection. | [buffers](#upstreambuffers) | No |
|``buffer-size`` | Sets the size of the buffer used for reading the first part of a response received from the upstream server. See the [proxy_buffer_size](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_buffer_size) directive. The default is set in the ``proxy-buffer-size`` ConfigMap key. | ``string`` | No |
|``ntlm`` | Allows proxying requests with NTLM Authentication. See the [ntlm](https://nginx.org/en/docs/http/ngx_http_upstream_module.html#ntlm) directive. In order for NTLM authentication to work, it is necessary to enable keepalive connections to upstream servers using the ``keepalive`` field. Note: this feature is supported only in NGINX Plus.| ``boolean`` | No |
|``type`` |The type of the upstream. Supported values are ``http`` and ``grpc``. The default is ``http``. For gRPC, it is necessary to enable HTTP/2 in the [ConfigMap](/nginx-ingress-controller/configuration/global-configuration/configmap-resource/#listeners) and configure TLS termination in the VirtualServer. | ``string`` | No |
{{% /table %}}

### Upstream.Buffers
The buffers field configures the buffers used for reading a response from the upstream server for a single connection:

```yaml
number: 4
size: 8K
```
See the [proxy_buffers](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_buffers) directive for additional information.

{{% table %}}
|Field | Description | Type | Required |
| ---| ---| ---| --- |
|``number`` | Configures the number of buffers. The default is set in the ``proxy-buffers`` ConfigMap key. | ``int`` | Yes |
|``size`` | Configures the size of a buffer. The default is set in the ``proxy-buffers`` ConfigMap key. | ``string`` | Yes |
{{% /table %}}

### Upstream.TLS

{{% table %}}
|Field | Description | Type | Required |
| ---| ---| ---| --- |
|``enable`` | Enables HTTPS for requests to upstream servers. The default is ``False`` , meaning that HTTP will be used. Note: by default, NGINX will not verify the upstream server certificate. To enable the verification, configure an [EgressMTLS Policy](/nginx-ingress-controller/configuration/policy-resource/#egressmtls). | ``boolean`` | No |
{{% /table %}}

### Upstream.Queue

The queue field configures a queue. A client request will be placed into the queue if an upstream server cannot be selected immediately while processing the request:

```yaml
size: 10
timeout: 60s
```

See [`queue`](https://nginx.org/en/docs/http/ngx_http_upstream_module.html#queue) directive for additional information.

Note: This feature is supported only in NGINX Plus.

{{% table %}}
|Field | Description | Type | Required |
| ---| ---| ---| --- |
|``size`` | The size of the queue. | ``int`` | Yes |
|``timeout`` | The timeout of the queue. A request cannot be queued for a period longer than the timeout. The default is ``60s``. | ``string`` | No |
{{% /table %}}

### Upstream.Healthcheck

The Healthcheck defines an [active health check](https://docs.nginx.com/nginx/admin-guide/load-balancer/http-health-check/). In the example below we enable a health check for an upstream and configure all the available parameters, including the `slow-start` parameter combined with [`mandatory` and `persistent`](https://docs.nginx.com/nginx/admin-guide/load-balancer/http-health-check/#mandatory-health-checks):

```yaml
name: tea
service: tea-svc
port: 80
slow-start: 30s
healthCheck:
  enable: true
  path: /healthz
  interval: 20s
  jitter: 3s
  fails: 5
  passes: 5
  port: 8080
  tls:
    enable: true
  connect-timeout: 10s
  read-timeout: 10s
  send-timeout: 10s
  headers:
  - name: Host
    value: my.service
  statusMatch: "! 500"
  mandatory: true
  persistent: true
```

Note: This feature is supported only in NGINX Plus.

{{% table %}}
|Field | Description | Type | Required |
| ---| ---| ---| --- |
|``enable`` | Enables a health check for an upstream server. The default is ``false``. | ``boolean`` | No |
|``path`` | The path used for health check requests. The default is ``/``. This not configurable for gRPC type upstreams. | ``string`` | No |
|``interval`` | The interval between two consecutive health checks. The default is ``5s``. | ``string`` | No |
|``jitter`` | The time within which each health check will be randomly delayed. By default, there is no delay. | ``string`` | No |
|``fails`` | The number of consecutive failed health checks of a particular upstream server after which this server will be considered unhealthy. The default is ``1``. | ``integer`` | No |
|``passes`` | The number of consecutive passed health checks of a particular upstream server after which the server will be considered healthy. The default is ``1``. | ``integer`` | No |
|``port`` | The port used for health check requests. By default, the port of the upstream is used. Note: in contrast with the port of the upstream, this port is not a service port, but a port of a pod. | ``integer`` | No |
|``tls`` | The TLS configuration used for health check requests. By default, the ``tls`` field of the upstream is used. | [upstream.tls](#upstreamtls) | No |
|``connect-timeout`` | The timeout for establishing a connection with an upstream server. By default, the ``connect-timeout`` of the upstream is used. | ``string`` | No |
|``read-timeout`` | The timeout for reading a response from an upstream server. By default, the ``read-timeout`` of the upstream is used. | ``string`` | No |
|``send-timeout`` | The timeout for transmitting a request to an upstream server. By default, the ``send-timeout`` of the upstream is used. | ``string`` | No |
|``headers`` | The request headers used for health check requests. NGINX Plus always sets the ``Host`` , ``User-Agent`` and ``Connection`` headers for health check requests. | [[]header](#header) | No |
|``statusMatch`` | The expected response status codes of a health check. By default, the response should have status code 2xx or 3xx. Examples: ``"200"`` , ``"! 500"`` , ``"301-303 307"``. See the documentation of the [match](https://nginx.org/en/docs/http/ngx_http_upstream_hc_module.html?#match) directive. This not supported for gRPC type upstreams. | ``string`` | No |
|``grpcStatus`` | The expected [gRPC status code](https://github.com/grpc/grpc/blob/master/doc/statuscodes.md#status-codes-and-their-use-in-grpc) of the upstream server response to the [Check method](https://github.com/grpc/grpc/blob/master/doc/health-checking.md). Configure this field only if your gRPC services do not implement the gRPC health checking protocol. For example, configure ``12`` if the upstream server responds with `12 (UNIMPLEMENTED)` status code. Only valid on gRPC type upstreams. | ``int`` | No |
|``grpcService`` | The gRPC service to be monitored on the upstream server. Only valid on gRPC type upstreams. | ``string`` | No |
|``mandatory`` | Require every newly added server to pass all configured health checks before NGINX Plus sends traffic to it. If this is not specified, or is set to false, the server will be initially considered healthy. When combined with [slow-start](https://nginx.org/en/docs/http/ngx_http_upstream_module.html#slow_start), it gives a new server more time to connect to databases and “warm up” before being asked to handle their full share of traffic. | ``bool`` | No |
|``persistent`` | Set the initial “up” state for a server after reload if the server was considered healthy before reload. Enabling persistent requires that the mandatory parameter is also set to `true`. | ``bool`` | No |
{{% /table %}} 

### Upstream.SessionCookie

The SessionCookie field configures session persistence which allows requests from the same client to be passed to the same upstream server. The information about the designated upstream server is passed in a session cookie generated by NGINX Plus.

In the example below, we configure session persistence with a session cookie for an upstream and configure all the available parameters:

```yaml
name: tea
service: tea-svc
port: 80
sessionCookie:
  enable: true
  name: srv_id
  path: /
  expires: 1h
  domain: .example.com
  httpOnly: false
  secure: true
```
See the [`sticky`](https://nginx.org/en/docs/http/ngx_http_upstream_module.html?#sticky) directive for additional information. The session cookie corresponds to the `sticky cookie` method.

Note: This feature is supported only in NGINX Plus.

{{% table %}}
|Field | Description | Type | Required |
| ---| ---| ---| --- |
|``enable`` | Enables session persistence with a session cookie for an upstream server. The default is ``false``. | ``boolean`` | No |
|``name`` | The name of the cookie. | ``string`` | Yes |
|``path`` | The path for which the cookie is set. | ``string`` | No |
|``expires`` | The time for which a browser should keep the cookie. Can be set to the special value ``max`` , which will cause the cookie to expire on ``31 Dec 2037 23:55:55 GMT``. | ``string`` | No |
|``domain`` | The domain for which the cookie is set. | ``string`` | No |
|``httpOnly`` | Adds the ``HttpOnly`` attribute to the cookie. | ``boolean`` | No |
|``secure`` | Adds the ``Secure`` attribute to the cookie. | ``boolean`` | No |
{{% /table %}}

### Header

The header defines an HTTP Header:
```yaml
name: Host
value: example.com
```

{{% table %}}
|Field | Description | Type | Required |
| ---| ---| ---| --- |
|``name`` | The name of the header. | ``string`` | Yes |
|``value`` | The value of the header. | ``string`` | No |
{{% /table %}}

### Action

The action defines an action to perform for a request.

In the example below, client requests are passed to an upstream `coffee`:
```yaml
 path: /coffee
 action:
  pass: coffee
```

{{% table %}}
|Field | Description | Type | Required |
| ---| ---| ---| --- |
|``pass`` | Passes requests to an upstream. The upstream with that name must be defined in the resource. | ``string`` | No |
|``redirect`` | Redirects requests to a provided URL. | [action.redirect](#actionredirect) | No |
|``return`` | Returns a preconfigured response. | [action.return](#actionreturn) | No |
|``proxy`` | Passes requests to an upstream with the ability to modify the request/response (for example, rewrite the URI or modify the headers). | [action.proxy](#actionproxy) | No |
{{% /table %}}

\* -- an action must include exactly one of the following: `pass`, `redirect`, `return` or `proxy`.

### Action.Redirect

The redirect action defines a redirect to return for a request.

In the example below, client requests are passed to a url `http://www.nginx.com`:
```yaml
redirect:
  url: http://www.nginx.com
  code: 301
```

{{% table %}}
|Field | Description | Type | Required |
| ---| ---| ---| --- |
|``url`` | The URL to redirect the request to. Supported NGINX variables: ``$scheme`` , ``$http_x_forwarded_proto`` , ``$request_uri`` , ``$host``. Variables must be enclosed in curly braces. For example: ``${host}${request_uri}``. | ``string`` | Yes |
|``code`` | The status code of a redirect. The allowed values are: ``301`` , ``302`` , ``307`` , ``308``. The default is ``301``. | ``int`` | No |
{{% /table %}}

### Action.Return

The return action defines a preconfigured response for a request.

In the example below, NGINX will respond with the preconfigured response for every request:
```yaml
return:
  code: 200
  type: text/plain
  body: "Hello World\n"
```

{{% table %}}
|Field | Description | Type | Required |
| ---| ---| ---| --- |
|``code`` | The status code of the response. The allowed values are: ``2XX``, ``4XX`` or ``5XX``. The default is ``200``. | ``int`` | No |
|``type`` | The MIME type of the response. The default is ``text/plain``. | ``string`` | No |
|``body`` | The body of the response. Supports NGINX variables*. Variables must be enclosed in curly brackets. For example: ``Request is ${request_uri}\n``. | ``string`` | Yes |
{{% /table %}}

\* -- Supported NGINX variables: `$request_uri`, `$request_method`, `$request_body`, `$scheme`, `$http_`, `$args`, `$arg_`, `$cookie_`, `$host`, `$request_time`, `$request_length`, `$nginx_version`, `$pid`, `$connection`, `$remote_addr`, `$remote_port`, `$time_iso8601`, `$time_local`, `$server_addr`, `$server_port`, `$server_name`, `$server_protocol`, `$connections_active`, `$connections_reading`, `$connections_writing` and `$connections_waiting`.

### Action.Proxy

The proxy action passes requests to an upstream with the ability to modify the request/response (for example, rewrite the URI or modify the headers).

In the example below, the request URI is rewritten to `/`, and the request and the response headers are modified:
```yaml
proxy:
  upstream: coffee
  requestHeaders:
    pass: true
    set:
    - name: My-Header
      value: Value
    - name: Client-Cert
      value: ${ssl_client_escaped_cert}
  responseHeaders:
    add:
    - name: My-Header
      value: Value
    - name: IC-Nginx-Version
      value: ${nginx_version}
      always: true
    hide:
    - x-internal-version
    ignore:
    - Expires
    - Set-Cookie
    pass:
    - Server
  rewritePath: /
```

{{% table %}}
|Field | Description | Type | Required |
| ---| ---| ---| --- |
|``upstream`` | The name of the upstream which the requests will be proxied to. The upstream with that name must be defined in the resource. | ``string`` | Yes |
|``requestHeaders`` | The request headers modifications. | [action.Proxy.RequestHeaders](#actionproxyrequestheaders) | No |
|``responseHeaders`` | The response headers modifications. | [action.Proxy.ResponseHeaders](#actionproxyresponseheaders) | No |
|``rewritePath`` | The rewritten URI. If the route path is a regular expression (starts with ~), the rewritePath can include capture groups with ``$1-9``. For example `$1` for the first group, and so on. For more information, check the [rewrite](https://github.com/nginxinc/kubernetes-ingress/tree/v2.0.3/examples/custom-resources/rewrites) example. | ``string`` | No |
{{% /table %}}

### Action.Proxy.RequestHeaders

The RequestHeaders field modifies the headers of the request to the proxied upstream server.

{{% table %}}
|Field | Description | Type | Required |
| ---| ---| ---| --- |
|``pass`` | Passes the original request headers to the proxied upstream server. See the [proxy_pass_request_header](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_pass_request_headers) directive for more information. Default is true. | ``bool`` | No |
|``set`` | Allows redefining or appending fields to present request headers passed to the proxied upstream servers. See the [proxy_set_header](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_set_header) directive for more information. | [[]header](#actionproxyrequestheaderssetheader) | No |
{{% /table %}}

### Action.Proxy.RequestHeaders.Set.Header

The header defines an HTTP Header:
```yaml
name: My-Header
value: My-Value
```

It is possible to override the default value of the `Host` header, which the Ingress Controller sets to [`$host`](https://nginx.org/en/docs/http/ngx_http_core_module.html#var_host):
```yaml
name: Host
value: example.com
```

{{% table %}}
|Field | Description | Type | Required |
| ---| ---| ---| --- |
|``name`` | The name of the header. | ``string`` | Yes |
|``value`` | The value of the header. Supports NGINX variables*. Variables must be enclosed in curly brackets. For example: ``${scheme}``. | ``string`` | No |
{{% /table %}}

\* -- Supported NGINX variables: `$request_uri`, `$request_method`, `$request_body`, `$scheme`, `$http_`, `$args`, `$arg_`, `$cookie_`, `$host`, `$request_time`, `$request_length`, `$nginx_version`, `$pid`, `$connection`, `$remote_addr`, `$remote_port`, `$time_iso8601`, `$time_local`, `$server_addr`, `$server_port`, `$server_name`, `$server_protocol`, `$connections_active`, `$connections_reading`, `$connections_writing`, `$connections_waiting`, `$ssl_cipher`, `$ssl_ciphers`, `$ssl_client_cert`, `$ssl_client_escaped_cert`, `$ssl_client_fingerprint`, `$ssl_client_i_dn`, `$ssl_client_i_dn_legacy`, `$ssl_client_raw_cert`, `$ssl_client_s_dn`, `$ssl_client_s_dn_legacy`, `$ssl_client_serial`, `$ssl_client_v_end`, `$ssl_client_v_remain`, `$ssl_client_v_start`, `$ssl_client_verify`, `$ssl_curves`, `$ssl_early_data`, `$ssl_protocol`, `$ssl_server_name`, `$ssl_session_id`, `$ssl_session_reused`, `$jwt_claim_` (NGINX Plus only) and `$jwt_header_` (NGINX Plus only).

### Action.Proxy.ResponseHeaders

The ResponseHeaders field modifies the headers of the response to the client.

{{% table %}}
|Field | Description | Type | Required |
| ---| ---| ---| --- |
|``hide`` | The headers that will not be passed* in the response to the client from a proxied upstream server. See the [proxy_hide_header](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_hide_header) directive for more information. | ``bool`` | No |
|``pass`` | Allows passing the hidden header fields* to the client from a proxied upstream server. See the [proxy_pass_header](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_pass_header) directive for more information. | ``[]string`` | No |
|``ignore`` | Disables processing of certain headers** to the client from a proxied upstream server. See the [proxy_ignore_headers](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_ignore_headers) directive for more information. | ``[]string`` | No |
|``add`` | Adds headers to the response to the client. | [[]addHeader](#addheader) | No |
{{% /table %}}

\* -- Default hidden headers are: `Date`, `Server`, `X-Pad` and `X-Accel-...`.

\** -- The following fields can be ignored: `X-Accel-Redirect`, `X-Accel-Expires`, `X-Accel-Limit-Rate`, `X-Accel-Buffering`, `X-Accel-Charset`, `Expires`, `Cache-Control`, `Set-Cookie` and `Vary`.

### AddHeader

The addHeader defines an HTTP Header with an optional `always` field:
```yaml
name: My-Header
value: My-Value
always: true
```

{{% table %}}
|Field | Description | Type | Required |
| ---| ---| ---| --- |
|``name`` | The name of the header. | ``string`` | Yes |
|``value`` | The value of the header. Supports NGINX variables*. Variables must be enclosed in curly brackets. For example: ``${scheme}``. | ``string`` | No |
|``always`` | If set to true, add the header regardless of the response status code**. Default is false. See the [add_header](http://nginx.org/en/docs/http/ngx_http_headers_module.html#add_header) directive for more information. | ``bool`` | No |
{{% /table %}}

\* -- Supported NGINX variables: `$request_uri`, `$request_method`, `$request_body`, `$scheme`, `$http_`, `$args`, `$arg_`, `$cookie_`, `$host`, `$request_time`, `$request_length`, `$nginx_version`, `$pid`, `$connection`, `$remote_addr`, `$remote_port`, `$time_iso8601`, `$time_local`, `$server_addr`, `$server_port`, `$server_name`, `$server_protocol`, `$connections_active`, `$connections_reading`, `$connections_writing`, `$connections_waiting`, `$ssl_cipher`, `$ssl_ciphers`, `$ssl_client_cert`, `$ssl_client_escaped_cert`, `$ssl_client_fingerprint`, `$ssl_client_i_dn`, `$ssl_client_i_dn_legacy`, `$ssl_client_raw_cert`, `$ssl_client_s_dn`, `$ssl_client_s_dn_legacy`, `$ssl_client_serial`, `$ssl_client_v_end`, `$ssl_client_v_remain`, `$ssl_client_v_start`, `$ssl_client_verify`, `$ssl_curves`, `$ssl_early_data`, `$ssl_protocol`, `$ssl_server_name`, `$ssl_session_id`, `$ssl_session_reused`, `$jwt_claim_` (NGINX Plus only) and `$jwt_header_` (NGINX Plus only).

\*\* -- If `always` is false, the response header is added only if the response status code is any of `200`, `201`, `204`, `206`, `301`, `302`, `303`, `304`, `307` or `308`.

### Split

The split defines a weight for an action as part of the splits configuration.

In the example below NGINX passes 80% of requests to the upstream `coffee-v1` and the remaining 20% to `coffee-v2`:
```yaml
splits:
- weight: 80
  action:
    pass: coffee-v1
- weight: 20
  action:
    pass: coffee-v2
```

{{% table %}}
|Field | Description | Type | Required |
| ---| ---| ---| --- |
|``weight`` | The weight of an action. Must fall into the range ``1..99``. The sum of the weights of all splits must be equal to ``100``. | ``int`` | Yes |
|``action`` | The action to perform for a request. | [action](#action) | Yes |
{{% /table %}}

### Match

The match defines a match between conditions and an action or splits.

In the example below, NGINX routes requests with the path `/coffee` to different upstreams based on the value of the cookie `user`:
* `user=john` -> `coffee-future`
* `user=bob` -> `coffee-deprecated`
* If the cookie is not set or not equal to either `john` or `bob`, NGINX routes to `coffee-stable`

```yaml
path: /coffee
matches:
- conditions:
  - cookie: user
    value: john
  action:
    pass: coffee-future
- conditions:
  - cookie: user
    value: bob
  action:
    pass: coffee-deprecated
action:
  pass: coffee-stable
```

In the next example, NGINX routes requests based on the value of the built-in [`$request_method` variable](https://nginx.org/en/docs/http/ngx_http_core_module.html#var_request_method), which represents the HTTP method of a request:
* all POST requests -> `coffee-post`
* all non-POST requests -> `coffee`

```yaml
path: /coffee
matches:
- conditions:
  - variable: $request_method
    value: POST
  action:
    pass: coffee-post
action:
  pass: coffee
```

{{% table %}}
|Field | Description | Type | Required |
| ---| ---| ---| --- |
|``conditions`` | A list of conditions. Must include at least 1 condition. | [[]condition](#condition) | Yes |
|``action`` | The action to perform for a request. | [action](#action) | No |
|``splits`` | The splits configuration for traffic splitting. Must include at least 2 splits. | [[]split](#split) | No |
{{% /table %}}

\* -- a match must include exactly one of the following: `action` or `splits`.

### Condition

The condition defines a condition in a match.

{{% table %}}
|Field | Description | Type | Required |
| ---| ---| ---| --- |
|``header`` | The name of a header. Must consist of alphanumeric characters or ``-``. | ``string`` | No |
|``cookie`` | The name of a cookie. Must consist of alphanumeric characters or ``_``. | ``string`` | No |
|``argument`` | The name of an argument. Must consist of alphanumeric characters or ``_``. | ``string`` | No |
|``variable`` | The name of an NGINX variable. Must start with ``$``. See the list of the supported variables below the table. | ``string`` | No |
|``value`` | The value to match the condition against. How to define a value is shown below the table. | ``string`` | Yes |
{{% /table %}}

\* -- a condition must include exactly one of the following: `header`, `cookie`, `argument` or `variable`.

Supported NGINX variables: `$args`, `$http2`, `$https`, `$remote_addr`, `$remote_port`, `$query_string`, `$request`, `$request_body`, `$request_uri`, `$request_method`, `$scheme`. Find the documentation for each variable [here](https://nginx.org/en/docs/varindex.html).

The value supports two kinds of matching:
* *Case-insensitive string comparison*. For example:
  * `john` -- case-insensitive matching that succeeds for strings, such as `john`, `John`, `JOHN`.
  * `!john` -- negation of the case-incentive matching for john that succeeds for strings, such as `bob`, `anything`, `''` (empty string).
* *Matching with a regular expression*. Note that NGINX supports regular expressions compatible with those used by the Perl programming language (PCRE). For example:
  * `~^yes` -- a case-sensitive regular expression that matches any string that starts with `yes`. For example: `yes`, `yes123`.
  * `!~^yes` -- negation of the previous regular expression that succeeds for strings like `YES`, `Yes123`, `noyes`. (The negation mechanism is not part of the PCRE syntax).
  * `~*no$` -- a case-insensitive regular expression that matches any string that ends with `no`. For example: `no`, `123no`, `123NO`.

**Note**: a value must not include any unescaped double quotes (`"`) and must not end with an unescaped backslash (`\`). For example, the following are invalid values: `some"value`, `somevalue\`.


### ErrorPage

The errorPage defines a custom response for a route for the case when either an upstream server responds with (or NGINX generates) an error status code. The custom response can be a redirect or a canned response. See the [error_page](https://nginx.org/en/docs/http/ngx_http_core_module.html#error_page) directive for more information.
```yaml
path: /coffee
errorPages:
- codes: [502, 503]
  redirect:
    code: 301
    url: https://nginx.org
- codes: [404]
  return:
    code: 200
    body: "Original resource not found, but success!"
```

{{% table %}}
|Field | Description | Type | Required |
| ---| ---| ---| --- |
|``codes`` | A list of error status codes. | ``[]int`` | Yes |
|``redirect`` | The redirect action for the given status codes. | [errorPage.Redirect](#errorpageredirect) | No |
|``return`` | The canned response action for the given status codes. | [errorPage.Return](#errorpagereturn) | No |
{{% /table %}}

\* -- an errorPage must include exactly one of the following: `return` or `redirect`.

### ErrorPage.Redirect

The redirect defines a redirect for an errorPage.

In the example below, NGINX responds with a redirect when a response from an upstream server has a 404 status code.

```yaml
codes: [404]
redirect:
  code: 301
  url: ${scheme}://cafe.example.com/error.html
```

{{% table %}}
|Field | Description | Type | Required |
| ---| ---| ---| --- |
|``code`` | The status code of a redirect. The allowed values are: ``301`` , ``302`` , ``307`` , ``308``.  The default is ``301``. | ``int`` | No |
|``url`` | The URL to redirect the request to. Supported NGINX variables: ``$scheme`` and ``$http_x_forwarded_proto``. Variables must be enclosed in curly braces. For example: ``${scheme}``. | ``string`` | Yes |
{{% /table %}}

### ErrorPage.Return

The return defines a canned response for an errorPage.

In the example below, NGINX responds with a canned response when a response from an upstream server has either 401 or 403 status code.

```yaml
codes: [401, 403]
return:
  code: 200
  type: application/json
  body: |
    {\"msg\": \"You don't have permission to do this\"}
  headers:
  - name: x-debug-original-statuses
    value: ${upstream_status}
```

{{% table %}}
|Field | Description | Type | Required |
| ---| ---| ---| --- |
|``code`` | The status code of the response. The default is the status code of the original response. | ``int`` | No |
|``type`` | The MIME type of the response. The default is ``text/html``. | ``string`` | No |
|``body`` | The body of the response. Supported NGINX variable: ``$upstream_status`` . Variables must be enclosed in curly braces. For example: ``${upstream_status}``. | ``string`` | Yes |
|``headers`` | The custom headers of the response. | [errorPage.Return.Header](#errorpagereturnheader) | No |
{{% /table %}}

### ErrorPage.Return.Header

The header defines an HTTP Header for a canned response in an errorPage:

```yaml
name: x-debug-original-statuses
value: ${upstream_status}
```

{{% table %}}
|Field | Description | Type | Required |
| ---| ---| ---| --- |
|``name`` | The name of the header. | ``string`` | Yes |
|``value`` | The value of the header. Supported NGINX variable: ``$upstream_status`` . Variables must be enclosed in curly braces. For example: ``${upstream_status}``. | ``string`` | No |
{{% /table %}}

## Using VirtualServer and VirtualServerRoute

You can use the usual `kubectl` commands to work with VirtualServer and VirtualServerRoute resources, similar to Ingress resources.

For example, the following command creates a VirtualServer resource defined in `cafe-virtual-server.yaml` with the name `cafe`:
```
$ kubectl apply -f cafe-virtual-server.yaml
virtualserver.k8s.nginx.org "cafe" created
```

You can get the resource by running:
```
$ kubectl get virtualserver cafe
NAME   STATE   HOST                   IP            PORTS      AGE
cafe   Valid   cafe.example.com       12.13.23.123  [80,443]   3m
```

In the kubectl get and similar commands, you can also use the short name `vs` instead of `virtualserver`.

Working with VirtualServerRoute resources is analogous. In the kubectl commands, use `virtualserverroute` or the short name `vsr`.

### Using Snippets

Snippets allow you to insert raw NGINX config into different contexts of NGINX configuration. In the example below, we use snippets to configure several NGINX features in a VirtualServer:

```yaml
apiVersion: k8s.nginx.org/v1
kind: VirtualServer
metadata:
  name: cafe
  namespace: cafe
spec:
  http-snippets: |
    limit_req_zone $binary_remote_addr zone=mylimit:10m rate=1r/s;
    proxy_cache_path /tmp keys_zone=one:10m;
  host: cafe.example.com
  tls:
    secret: cafe-secret
  server-snippets: |
    limit_req zone=mylimit burst=20;
  upstreams:
  - name: tea
    service: tea-svc
    port: 80
  - name: coffee
    service: coffee-svc
    port: 80
  routes:
  - path: /tea
    location-snippets: |
      proxy_cache one;
      proxy_cache_valid 200 10m;
    action:
      pass: tea
  - path: /coffee
    action:
      pass: coffee
```

Snippets are intended to be used by advanced NGINX users who need more control over the generated NGINX configuration.

However, because of the disadvantages described below, snippets are disabled by default. To use snippets, set the [`enable-snippets`](/nginx-ingress-controller/configuration/global-configuration/command-line-arguments#cmdoption-enable-snippets) command-line argument.

Disadvantages of using snippets:
* *Complexity*. To use snippets, you will need to:
  * Understand NGINX configuration primitives and implement a correct NGINX configuration.
  * Understand how the IC generates NGINX configuration so that a snippet doesn't interfere with the other features in the configuration.
* *Decreased robustness*. An incorrect snippet makes the NGINX config invalid which will lead to a failed reload. This will prevent any new configuration updates, including updates for the other VirtualServer and VirtualServerRoute resources until the snippet is fixed.
* *Security implications*. Snippets give access to NGINX configuration primitives and those primitives are not validated by the Ingress Controller. For example, a snippet can configure NGINX to serve the TLS certificates and keys used for TLS termination for Ingress and VirtualServer resources.

To help catch errors when using snippets, the Ingress Controller reports config reload errors in the logs as well as in the events and status field of VirtualServer and VirtualServerRoute resources. Additionally, a number of Prometheus metrics show the stats about failed reloads – `controller_nginx_last_reload_status` and `controller_nginx_reload_errors_total`.

> Note that during a period when the NGINX config includes an invalid snippet, NGINX will continue to operate with the latest valid configuration.

### Validation

Two types of validation are available for VirtualServer and VirtualServerRoute resources:
* *Structural validation* by the `kubectl` and Kubernetes API server.
* *Comprehensive validation* by the Ingress Controller.

#### Structural Validation

The custom resource definitions for VirtualServer and VirtualServerRoute include structural OpenAPI schema which describes the type of every field of those resources.

If you try to create (or update) a resource that violates the structural schema (for example, you use a string value for the port field of an upstream), `kubectl` and Kubernetes API server will reject such a resource:
* Example of `kubectl` validation:
    ```
    $ kubectl apply -f cafe-virtual-server.yaml
      error: error validating "cafe-virtual-server.yaml": error validating data: ValidationError(VirtualServer.spec.upstreams[0].port): invalid type for org.nginx.k8s.v1.VirtualServer.spec.upstreams.port: got "string", expected "integer"; if you choose to ignore these errors, turn validation off with --validate=false
    ```
* Example of Kubernetes API server validation:
    ```
    $ kubectl apply -f cafe-virtual-server.yaml --validate=false
      The VirtualServer "cafe" is invalid: []: Invalid value: map[string]interface {}{ ... }: validation failure list:
      spec.upstreams.port in body must be of type integer: "string"
    ```

If a resource is not rejected (it doesn't violate the structural schema), the Ingress Controller will validate it further.

#### Comprehensive Validation

The Ingress Controller validates the fields of the VirtualServer and VirtualServerRoute resources. If a resource is invalid, the Ingress Controller will reject it: the resource will continue to exist in the cluster, but the Ingress Controller will ignore it.

You can check if the Ingress Controller successfully applied the configuration for a VirtualServer. For our example `cafe` VirtualServer, we can run:
```
$ kubectl describe vs cafe
. . .
Events:
  Type    Reason          Age   From                      Message
  ----    ------          ----  ----                      -------
  Normal  AddedOrUpdated  16s   nginx-ingress-controller  Configuration for default/cafe was added or updated
```
Note how the events section includes a Normal event with the AddedOrUpdated reason that informs us that the configuration was successfully applied.

If you create an invalid resource, the Ingress Controller will reject it and emit a Rejected event. For example, if you create a VirtualServer `cafe` with two upstream with the same name `tea`, you will get:
```
$ kubectl describe vs cafe
. . .
Events:
  Type     Reason    Age   From                      Message
  ----     ------    ----  ----                      -------
  Warning  Rejected  12s   nginx-ingress-controller  VirtualServer default/cafe is invalid and was rejected: spec.upstreams[1].name: Duplicate value: "tea"
```
Note how the events section includes a Warning event with the Rejected reason.

Additionally, this information is also available in the `status` field of the VirtualServer resource. Note the Status section of the VirtualServer:

```
$ kubectl describe vs cafe
. . .
Status:
  External Endpoints:
    Ip:        12.13.23.123
    Ports:     [80,443]
  Message:  VirtualServer default/cafe is invalid and was rejected: spec.upstreams[1].name: Duplicate value: "tea"
  Reason:   Rejected
  State:    Invalid
```

The Ingress Controller validates VirtualServerRoute resources in a similar way.

**Note**: If you make an existing resource invalid, the Ingress Controller will reject it and remove the corresponding configuration from NGINX.

## Customization via ConfigMap

You can customize the NGINX configuration for VirtualServer and VirtualServerRoutes resources using the [ConfigMap](/nginx-ingress-controller/configuration/global-configuration/configmap-resource). Most of the ConfigMap keys are supported, with the following exceptions:
* `proxy-hide-headers`
* `proxy-pass-headers`
* `hsts`
* `hsts-max-age`
* `hsts-include-subdomains`
* `hsts-behind-proxy`
* `redirect-to-https`
* `ssl-redirect`
