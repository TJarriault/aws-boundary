---
title: Command-line Arguments
description: 
weight: 1700
doctypes: [""]
toc: true
---


The Ingress Controller supports several command-line arguments. Setting the arguments depends on how you install the Ingress Controller:

* If you're using *Kubernetes manifests* (Deployment or DaemonSet) to install the Ingress Controller, to set the command-line arguments, modify those manifests accordingly. See the [Installation with Manifests](/nginx-ingress-controller/installation/installation-with-manifests) doc.
* If you're using *Helm* to install the Ingress Controller, modify the parameters of the Helm chart that correspond to the command-line arguments. See the [Installation with Helm](/nginx-ingress-controller/installation/installation-with-helm) doc.

Below we describe the available command-line arguments:  
&nbsp;
<a name="cmdoption-enable-snippets"></a>

### -enable-snippets

Enable custom NGINX configuration snippets in Ingress, VirtualServer, VirtualServerRoute and TransportServer resources.

Default `false`.  
&nbsp;
<a name="cmdoption-default-server-tls-secret"></a>

### -default-server-tls-secret `<string>`

Secret with a TLS certificate and key for TLS termination of the default server.

* If not set, certificate and key in the file `/etc/nginx/secrets/default` are used.
* If `/etc/nginx/secrets/default` doesn't exist, the Ingress Controller will configure NGINX to reject TLS connections to the default server.
* If a secret is set, but the Ingress controller is not able to fetch it from Kubernetes API, or it is not set and the Ingress Controller fails to read the file "/etc/nginx/secrets/default", the Ingress controller will fail to start.

Format: `<namespace>/<name>`  
&nbsp;
<a name="cmdoption-wildcard-tls-secret"></a>

### -wildcard-tls-secret `<string>`

A Secret with a TLS certificate and key for TLS termination of every Ingress/VirtualServer host for which TLS termination is enabled but the Secret is not specified.

* If the argument is not set, for such Ingress/VirtualServer hosts NGINX will break any attempt to establish a TLS connection.

* If the argument is set, but the Ingress controller is not able to fetch the Secret from Kubernetes API, the Ingress controller will fail to start.

Format: `<namespace>/<name>`  
&nbsp;
<a name="cmdoption-enable-custom-resources"></a>

### -enable-custom-resources

Enables custom resources.

Default `true`.  
&nbsp;  
<a name="cmdoption-enable-preview-policies"></a>

### -enable-preview-policies

Enables preview policies.

Default `false`.  
&nbsp;  
<a name="cmdoption-enable-leader-election"></a>
### -enable-leader-election

Enables Leader election to avoid multiple replicas of the controller reporting the status of Ingress, VirtualServer and VirtualServerRoute resources -- only one replica will report status.
Default `true`.

See [-report-ingress-status](#cmdoption-report-ingress-status) flag.  
&nbsp;  
<a name="cmdoption-enable-tls-passthrough"></a>

### -enable-tls-passthrough

Enable TLS Passthrough on port 443.

Requires [-enable-custom-resources](#cmdoption-enable-custom-resources).  
&nbsp;  
<a name="cmdoption-external-service"></a> 

### -external-service `<string>`

Specifies the name of the service with the type LoadBalancer through which the Ingress controller pods are exposed externally. The external address of the service is used when reporting the status of Ingress, VirtualServer and VirtualServerRoute resources.

For Ingress resources only: Requires [-report-ingress-status](#cmdoption-report-ingress-status).  
&nbsp;  
<a name="cmdoption-ingresslink"></a> 

### -ingresslink `<string>`

Specifies the name of the IngressLink resource, which exposes the Ingress Controller pods via a BIG-IP system. The IP of the BIG-IP system is used when reporting the status of Ingress, VirtualServer and VirtualServerRoute resources.

For Ingress resources only: Requires [-report-ingress-status](#cmdoption-report-ingress-status).  
&nbsp;  
<a name="cmdoption-global-configuration"></a> 

### -global-configuration `<string>`

A GlobalConfiguration resource for global configuration of the Ingress Controller.

Format: `<namespace>/<name>`

Requires [-enable-custom-resources](#cmdoption-enable-custom-resources).  
&nbsp;  
<a name="cmdoption-health-status"></a>

### -health-status

Adds a location "/nginx-health" to the default server. The location responds with the 200 status code for any request.

Useful for external health-checking of the Ingress controller.  
&nbsp;  
<a name="cmdoption-health-status-uri"></a>

### -health-status-uri `<string>`

Sets the URI of health status location in the default server. Requires [-health-status](#cmdoption-health-status). (default `/nginx-health`)  
&nbsp;  
<a name="cmdoption-ingress-class"></a>

### -ingress-class `<string>`

A class of the Ingress controller.

A corresponding IngressClass resource with the name equal to the class must be deployed. Otherwise, the Ingress Controller will fail to start.
The Ingress controller only processes resources that belong to its class - i.e. have the `ingressClassName` field resource equal to the class.

The Ingress Controller processes all the resources that do not have the `ingressClassName` field.

Default `nginx`.  
&nbsp;
<a name="cmdoption-ingress-template-path"></a>  

### -ingress-template-path `<string>`

Path to the ingress NGINX configuration template for an ingress resource. Default for NGINX is `nginx.ingress.tmpl`; default for NGINX Plus is `nginx-plus.ingress.tmpl`.  
&nbsp;  
<a name="cmdoption-leader-election-lock-name"></a>

### -leader-election-lock-name `<string>`

Specifies the name of the ConfigMap, within the same namespace as the controller, used as the lock for leader election.

Requires [-enable-leader-election](#cmdoption-enable-leader-election).  
&nbsp;  
<a name="cmdoption-log_backtrace_at"></a>

### -log_backtrace_at `<value>`

When logging hits line `file:N`, emit a stack trace.  
&nbsp;  
<a name="cmdoption-main-template-path"></a>

### -main-template-path `<string>`

Path to the main NGINX configuration template.

* Default for NGINX is `nginx.ingress.tmpl`.
* Default for NGINX Plus is `nginx-plus.ingress.tmpl`.  
&nbsp;
<a name="cmdoption-nginx-configmaps"></a>

### -nginx-configmaps `<string>`

A ConfigMap resource for customizing NGINX configuration. If a ConfigMap is set, but the Ingress controller is not able to fetch it from Kubernetes API, the Ingress controller will fail to start.

Format: `<namespace>/<name>`  
&nbsp;  
<a name="cmdoption-nginx-debug"></a> 

### -nginx-debug

Enable debugging for NGINX. Uses the nginx-debug binary. Requires 'error-log-level: debug' in the ConfigMap.  
&nbsp;  
<a name="cmdoption-nginx-plus"></a>

### -nginx-plus

Enable support for NGINX Plus.  
&nbsp;  
<a name="cmdoption-nginx-reload-timeout"></a>

### -nginx-reload-timeout `<value>`

Timeout in milliseconds which the Ingress Controller will wait for a successful NGINX reload after a change or at the initial start.

Default is 4000. Default is 20000 instead if `enable-app-protect` is true.  
&nbsp;  
<a name="cmdoption-nginx-status"></a>

### -nginx-status

Enable the NGINX stub_status, or the NGINX Plus API.

Default `true`.  
&nbsp;  
<a name="cmdoption-nginx-status-allow-cidrs"></a>

### -nginx-status-allow-cidrs `<string>`

Add IPv4 IP/CIDR blocks to the allow list for NGINX stub_status or the NGINX Plus API.

Separate multiple IP/CIDR by commas. (default `127.0.0.1`)  
&nbsp;  
<a name="cmdoption-nginx-status-port"></a> 

### -nginx-status-port `<int>`

Set the port where the NGINX stub_status or the NGINX Plus API is exposed.

Format: `[1024 - 65535]` (default `8080`)  
&nbsp;  
<a name="cmdoption-proxy"></a> 

### -proxy `<string>`

Use a proxy server to connect to Kubernetes API started by "kubectl proxy" command. **For testing purposes only**.

The Ingress controller does not start NGINX and does not write any generated NGINX configuration files to disk.  
&nbsp;  
<a name="cmdoption-report-ingress-status"></a>

### -report-ingress-status

Updates the address field in the status of Ingress resources.

Requires the [-external-service](#cmdoption-external-service) or [-ingresslink](#cmdoption-ingresslink) flag, or the `external-status-address` key in the ConfigMap.  
&nbsp;
<a name="cmdoption-transportserver-template-path"></a> 

### -transportserver-template-path `<string>`

Path to the TransportServer NGINX configuration template for a TransportServer resource.

* Default for NGINX is `nginx.transportserver.tmpl`.
* Default for NGINX Plus is `nginx-plus.transportserver.tmpl`.  
  
&nbsp;  
<a name="cmdoption-v"></a> 

### -v `<value>`

Log level for V logs.  
&nbsp;  
<a name="cmdoption-version"></a> 

### -version

Print the version, git-commit hash and build date and exit.  
&nbsp;  
<a name="cmdoption-virtualserver-template-path"></a> 

### -virtualserver-template-path `<string>`

Path to the VirtualServer NGINX configuration template for a VirtualServer resource.

* Default for NGINX is `nginx.ingress.tmpl`.
* Default for NGINX Plus is `nginx-plus.ingress.tmpl`.  
  
&nbsp;  
<a name="cmdoption-vmodule"></a> 

### -vmodule `<value>`

A comma-separated list of pattern=N settings for file-filtered logging.  
&nbsp;
<a name="cmdoption-watch-namespace"></a> 

### -watch-namespace `<string>`

Namespace to watch for Ingress resources. By default the Ingress controller watches all namespaces.  
&nbsp;  
<a name="cmdoption-enable-prometheus-metrics"></a> 

### -enable-prometheus-metrics

Enables exposing NGINX or NGINX Plus metrics in the Prometheus format.  
&nbsp;
<a name="cmdoption-prometheus-metrics-listen-port"></a>

### -prometheus-metrics-listen-port `<int>`

Sets the port where the Prometheus metrics are exposed.

Format: `[1024 - 65535]` (default `9113`)  
&nbsp;  
<a name="cmdoption-prometheus-tls-secret"></a> 

### -prometheus-tls-secret `<string>`

A Secret with a TLS certificate and key for TLS termination of the Prometheus metrics endpoint.

* If the argument is not set, the prometheus endpoint will not use a TLS connection.
* If the argument is set, but the Ingress controller is not able to fetch the Secret from Kubernetes API, the Ingress controller will fail to start.

Format: `<namespace>/<name>`  
&nbsp;  
<a name="cmdoption-spire-agent-address"></a>

### -spire-agent-address `<string>`

Specifies the address of a running Spire agent. **For use with NGINX Service Mesh only**.

Requires [-nginx-plus](#cmdoption-nginx-plus).

* If the argument is set, but the Ingress Controller is unable to connect to the Spire Agent, the Ingress Controller will fail to start.  

&nbsp;  
<a name="cmdoption-enable-internal-routes"></a> 

### -enable-internal-routes

Enable support for internal routes with NGINX Service Mesh. **For use with NGINX Service Mesh only**.

Requires [-nginx-plus](#cmdoption-nginx-plus) and [-spire-agent-address](#cmdoption-spire-agent-address).

* If the argument is set, but `nginx-plus` is set to false, or the `spire-agent-address` is not provided, the Ingress Controller will fail to start.  
  
&nbsp;  
<a name="cmdoption-enable-latency-metrics"></a> 

### -enable-latency-metrics

Enable collection of latency metrics for upstreams.
Requires [-enable-prometheus-metrics](#cmdoption-enable-prometheus-metrics).  
&nbsp;
<a name="cmdoption-enable-app-protect"></a> 

### -enable-app-protect

Enables support for App Protect.

Requires [-nginx-plus](#cmdoption-nginx-plus).

* If the argument is set, but `nginx-plus` is set to false, the Ingress Controller will fail to start.  

&nbsp;  
<a name="cmdoption-enable-app-protect-dos"></a> 

### -enable-app-protect-dos

Enables support for App Protect Dos.

Requires [-nginx-plus](#cmdoption-nginx-plus).

* If the argument is set, but `nginx-plus` is set to false, the Ingress Controller will fail to start.  

&nbsp;  
<a name="cmdoption-app-protect-dos-debug"></a> 

### -app-protect-dos-debug

Enable debugging for App Protect Dos.

Requires [-nginx-plus](#cmdoption-nginx-plus) and [-enable-app-protect-dos](#cmdoption-enable-app-protect-dos).

* If the argument is set, but `nginx-plus` and `enable-app-protect-dos` are set to false, the Ingress Controller will fail to start.  

&nbsp;  
<a name="cmdoption-app-protect-dos-max-daemons"></a> 

### -app-protect-dos-max-daemons

Max number of ADMD instances.

Default `1`. 

Requires [-nginx-plus](#cmdoption-nginx-plus) and [-enable-app-protect-dos](#cmdoption-enable-app-protect-dos).

* If the argument is set, but `nginx-plus` and `enable-app-protect-dos` are set to false, the Ingress Controller will fail to start.  

&nbsp;  
<a name="cmdoption-app-protect-dos-max-workers"></a> 

### -app-protect-dos-max-workers

Max number of nginx processes to support.

Default `Number of CPU cores in the machine`. 

Requires [-nginx-plus](#cmdoption-nginx-plus) and [-enable-app-protect-dos](#cmdoption-enable-app-protect-dos).

* If the argument is set, but `nginx-plus` and `enable-app-protect-dos` are set to false, the Ingress Controller will fail to start.  


&nbsp;  
<a name="cmdoption-app-protect-dos-memory"></a> 

### -app-protect-dos-memory

RAM memory size to consume in MB

Default `50% of free RAM in the container or 80MB, the smaller`. 

Requires [-nginx-plus](#cmdoption-nginx-plus) and [-enable-app-protect-dos](#cmdoption-enable-app-protect-dos).

* If the argument is set, but `nginx-plus` and `enable-app-protect-dos` are set to false, the Ingress Controller will fail to start.  


&nbsp;  
<a name="cmdoption-ready-status"></a> 

### -ready-status

Enables the readiness endpoint `/nginx-ready`. The endpoint returns a success code when NGINX has loaded all the config after the startup.

Default `true`.  
&nbsp;  
<a name="cmdoption-ready-status-port"></a> 

### -ready-status-port

The HTTP port for the readiness endpoint.

Format: `[1024 - 65535]` (default `8081`)  
&nbsp; 
