# NGINX App Protect Support

In this example we deploy the NGINX Plus Ingress controller with [NGINX App Protect](https://www.nginx.com/products/nginx-app-protect/), a simple web application and then configure load balancing and WAF protection for that application using the Ingress resource.

## Running the Example

## 1. Deploy the Ingress Controller

1. Follow the installation [instructions](https://docs.nginx.com/nginx-ingress-controller/installation) to deploy the Ingress controller with NGINX App Protect.

2. Save the public IP address of the Ingress controller into a shell variable:
    ```
    $ IC_IP=XXX.YYY.ZZZ.III
    ```
3. Save the HTTPS port of the Ingress controller into a shell variable:
    ```
    $ IC_HTTPS_PORT=<port number>
    ```

## 2. Deploy the Cafe Application

Create the coffee and the tea deployments and services:
```
$ kubectl create -f cafe.yaml
```

## 3. Configure Load Balancing
1. Create the syslog service and pod for the App Protect security logs:
    ```
    $ kubectl create -f syslog.yaml
    ```
2. Create a secret with an SSL certificate and a key:
    ```
    $ kubectl create -f cafe-secret.yaml
    ```
3. Create the App Protect policy, log configuration and user defined signature:
    ```
    $ kubectl create -f ap-dataguard-alarm-policy.yaml
    $ kubectl create -f ap-logconf.yaml
    $ kubectl create -f ap-apple-uds.yaml
    ```
4. Create an Ingress Resource:
    ```
    $ kubectl create -f cafe-ingress.yaml
    ```
    Note the App Protect annotations in the Ingress resource. They enable WAF protection by configuring App Protect with the policy and log configuration created in the previous step.

## 4. Test the Application

1. To access the application, curl the coffee and the tea services. We'll use `curl`'s --insecure option to turn off certificate verification of our self-signed
certificate and the --resolve option to set the Host header of a request with `cafe.example.com`

    To get coffee:
    ```
    $ curl --resolve cafe.example.com:$IC_HTTPS_PORT:$IC_IP https://cafe.example.com:$IC_HTTPS_PORT/coffee --insecure
    Server address: 10.12.0.18:80
    Server name: coffee-7586895968-r26zn
    ...
    ```
    If your prefer tea:
    ```
    $ curl --resolve cafe.example.com:$IC_HTTPS_PORT:$IC_IP https://cafe.example.com:$IC_HTTPS_PORT/tea --insecure
    Server address: 10.12.0.19:80
    Server name: tea-7cd44fcb4d-xfw2x
    ...
    ```
    Now, let's try to send a request with a suspicious URL:
    ```
    $ curl --resolve cafe.example.com:$IC_HTTPS_PORT:$IC_IP "https://cafe.example.com:$IC_HTTPS_PORT/tea/<script>" --insecure
    <html><head><title>Request Rejected</title></head><body>
    ...
    ```
    Lastly, let's try to send some suspicious data that matches the user defined signature.
    ```
    $ curl --resolve cafe.example.com:$IC_HTTPS_PORT:$IC_IP -X POST -d "apple" "https://cafe.example.com:$IC_HTTPS_PORT/tea/" --insecure
    <html><head><title>Request Rejected</title></head><body>
    ...
    ```
    As you can see, the suspicious requests were blocked by App Protect

1. To check the security logs in the syslog pod:
    ```
    $ kubectl exec -it <SYSLOG_POD> -- cat /var/log/messages
    ```
