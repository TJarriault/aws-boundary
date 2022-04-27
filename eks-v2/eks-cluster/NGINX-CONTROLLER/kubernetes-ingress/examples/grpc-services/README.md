# gRPC support

To support a gRPC application with NGINX Ingress controllers, you need to add the **nginx.org/grpc-services** annotation to your Ingress resource definition.

## Prerequisites

* HTTP/2 must be enabled. See `http2` ConfigMap key in the [ConfigMap](https://docs.nginx.com/nginx-ingress-controller/configuration/global-configuration/configmap-resource/#listeners)
* Ingress resources for gRPC applications must include TLS termination.

## Syntax

The `nginx.org/grpc-services` specifies which services are gRPC services. The annotation syntax is as follows:
```
nginx.org/grpc-services: "service1[,service2,...]"
```

## Example

In the following example we load balance three applications, one of which is using gRPC:
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: grpc-ingress
  annotations:
    nginx.org/grpc-services: "grpc-svc"
    kubernetes.io/ingress.class: "nginx"
spec:
  tls:
  - hosts:
    - grpc.example.com
    secretName: grpc-secret
  rules:
  - host: grpc.example.com
    http:
      paths:
      - path: /helloworld.Greeter
        pathType: Prefix
        backend:
          service:
            name: grpc-svc
            port:
              number: 50051
      - path: /tea
        pathType: Prefix
        backend:
          service:
            name: tea-svc
            port:
              number: 80
      - path: /coffee
        pathType: Prefix
        backend:
          service:
            name: coffee-svc
            port:
              number: 80
```
*grpc-svc* is a service for the gRPC application. The service becomes available at the `/helloworld.Greeter` path. Note how we used the **nginx.org/grpc-services** annotation.
