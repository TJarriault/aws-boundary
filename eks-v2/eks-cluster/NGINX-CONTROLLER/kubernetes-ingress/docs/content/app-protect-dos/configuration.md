---
title: Configuration

description:
weight: 1900
doctypes: [""]
toc: true
---

This document describes how to configure the NGINX App Protect Dos module
> Check out the complete [NGINX Ingress Controller with App Protect Dos example resources on GitHub](https://github.com/nginxinc/kubernetes-ingress/tree/v2.0.3/examples/appprotect-dos).

## App Protect Dos Configuration

A `DosProtectedResource` is a [Custom Resource](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) that holds the configuration of a collection of protected resources.
An [Ingress](/nginx-ingress-controller/configuration/ingress-resources/basic-configuration), [VirtualServer and VirtualServerRoute](/nginx-ingress-controller/configuration/virtualserver-and-virtualserverroute-resources/) can be protected by specifying a reference to the DosProtectedResource.

1. Create an `DosProtectedResource` Custom resource manifest. As an example:
  ```yaml
apiVersion: appprotectdos.f5.com/v1beta1
kind: DosProtectedResource
metadata:
   name: dos-protected
spec:
   enable: true
   name: "webapp.example.com"
   apDosMonitor:
      uri: "webapp.example.com"
      protocol: "http1"
      timeout: 5
  ```
2. Enable App Protect Dos on an Ingress by adding an annotation on the Ingress. Set the value of the annotation to the qualified identifier(`namespace/name`) of a DosProtectedResource:
  ```yaml
   apiVersion: networking.k8s.io/v1
   kind: Ingress
   metadata:
      name: webapp-ingress
      annotations:
         appprotectdos.f5.com/app-protect-dos-resource: "default/dos-protected"
  ```
3. Enable App Protect Dos on a VirtualServer by setting the `dos` field value to the qualified identifier(`namespace/name`) of a DosProtectedResource:
  ```yaml
apiVersion: k8s.nginx.org/v1
kind: VirtualServer
metadata:
   name: webapp
spec:
   host: webapp.example.com
   upstreams:
      - name: webapp
        service: webapp-svc
        port: 80
   routes:
      - path: /
        dos: dos-protected
        action:
           pass: webapp
  ```

## Dos Policy Configuration

You can configure the policy for Dos by creating an `APDosPolicy` [Custom Resource](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) and specifying the qualified identifier(`namespace/name`) of the `ApDosPolicy` in the `DosProtectedResource`.

For example, say you want to use Dos Policy as shown below:

  ```json
  {
   mitigation_mode: "standard",
   signatures: "on",
   bad_actors: "on",
   automation_tools_detection: "on",
   tls_fingerprint: "on",
}
  ```

You would create an `APDosPolicy` resource with the policy defined in the `spec`, as shown below:

  ```yaml
   apiVersion: appprotectdos.f5.com/v1beta1
   kind: APDosPolicy
   metadata:
      name: dospolicy
   spec:
      mitigation_mode: "standard"
      signatures: "on"
      bad_actors: "on"
      automation_tools_detection: "on"
      tls_fingerprint: "on"
  ```

Then add a reference in the `DosProtectedResrouce` to the `ApDosPolicy`:
  ```yaml
   apiVersion: appprotectdos.f5.com/v1beta1
   kind: DosProtectedResource
   metadata:
      name: dos-protected
   spec:
      enable: true
      name: "my-dos"
      apDosMonitor: 
         uri: "webapp.example.com"
      apDosPolicy: "default/dospolicy"
  ```

## App Protect Dos Logs

You can set the [App Protect Dos Log configuration](/nginx-app-protect-dos/logs-overview/types-of-logs/) by creating an `APDosLogConf` [Custom Resource](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) and specifying the qualified identifier(`namespace/name`) of the `ApDosLogConf` in the `DosProtectedResource`.

For example, say you want to log state changing requests for your Ingress resources using App Protect Dos. The App Protect Dos log configuration looks like this:

```json
{
    "filter": {
        "request_type": "all"
    },
    "content": {
        "format": "default",
        "max_request_size": "any",
        "max_message_size": "64k"
    }
}
```

You would add that config in the `spec` of your `APDosLogConf` resource as follows:

```yaml
apiVersion: appprotectdos.f5.com/v1beta1
kind: APDosLogConf
metadata:
   name: doslogconf
spec:
   content:
      format: splunk
      max_message_size: 64k
   filter:
      traffic-mitigation-stats: all
      bad-actors: top 10
      attack-signatures: top 10
```

Then add a reference in the `DosProtectedResource` to the `APDosLogConf`:
  ```yaml
   apiVersion: appprotectdos.f5.com/v1beta1
   kind: DosProtectedResource
   metadata:
      name: dos-protected
   spec:
      enable: true
      name: "my-dos"
      apDosMonitor: 
         uri: "webapp.example.com"
      dosSecurityLog:
         enable: true
         apDosLogConf: "doslogconf"
         dosLogDest: "syslog-svc.default.svc.cluster.local:514"
  ```
## Global Configuration

The NGINX Ingress Controller has a set of global configuration parameters that align with those available in the NGINX App Protect Dos module. See [ConfigMap keys](/nginx-ingress-controller/configuration/global-configuration/configmap-resource/#modules) for the complete list. The App Protect parameters use the `app-protect-dos*` prefix.
