# `JSESSIONID` Load Balancer

This is an example application acts as a sticky load balancer for HTTP sessions
with a `JSESSIONID` cookie attached.

It works similarly to Cloud Foundry's gorouter package. If the application returns the
cookie `JSESSIONID` then the load balancer appends a `__VCAP_ID__` cookie with the same
`expiry`, `sameSite`, `secure`, `maxAge` and `path` attributes as the `JSESSIONID` cookie.

The client MUST return both cookies in subsequent requests to establish a sticky session.

The load balancer will use the `__VCAP_ID__` cookie to route to the same application instance
every time. The `__VCAP_ID__` cookie is forwarded to the application to enable session continuity.

If the instance referenced by the `__VCAP_ID__` disappears, a different backend instance will be
chosen.

If no instances are healthy, an error is returned.

## Parameters

### Environment variables

| Name             | Default       | Description                             |
| ---------------- | ------------- | --------------------------------------- |
| `PORT`           | N/A           | **Required** Port to listen on.         |
| `SESSION_COOKIE` | `JSESSIONID`  | Cookie that triggers a sticky session.  |
| `STICKY_COOKIE`  | `__VCAP_ID__` | Cookie that contains the ticky session. |
| `PROXY_SERVICE`  | N/A           | **Required** Service to load balance.   |
| `PROXY_PORT`     | `8080`        | Port on the service to load balance.    |
| `PROXY_SCHEME`   | `http`        | Scheme to use for proxied requests.     |

## Kf set up

First, create a headless Kubernetes service that matches the application
runtime Pods for a Kf App. In the example below the App name is `my-app`
and the Space name is `my-space`:

```yaml
apiVersion: v1
kind: Service
metadata:
  # Name for this service `my-app` will already be taken by the
  # Kf App.
  name: my-app-headless
  # Space name
  namespace: my-space
spec:
  # Must be None in order to route directly to each backend using
  # Kubernetes DNS service discovery.
  clusterIP: None
  internalTrafficPolicy: Cluster
  ipFamilies:
  - IPv4
  ipFamilyPolicy: SingleStack
  # Ports should match those on the K8s service `my-app`
  ports:
  - name: http-user-port
    port: 80
    protocol: TCP
    # PROXY_PORT should be set to the value of targetPort.
    targetPort: 8080
  # Selector should match the one on the K8s service `my-app`
  selector:
    app.kubernetes.io/component: app-server
    app.kubernetes.io/managed-by: kf
    app.kubernetes.io/name: my-app
```

If you want to target multiple applications, add a common label to
the manfiest like the following and replace the `app.kubernetes.io/name`
label with it in the Kubernetes service selector:

```yaml
applications:
- name: my-app
  metadata:
    labels:
      jsessionid-lb-target: some-unique-name
```

Next, push the load balancer app pointing to the internal address
of the service you just created:

```bash
kf push --buildpack go_buildpack \
  --env PROXY_SERVICE=my-app-headless.my-space.svc.cluster.local \
  my-proxy
```

The proxy will log the configuration at startup:

```
proxy 2022/08/08 02:58:36.198758 main.go:74: Proxy Configuration
proxy 2022/08/08 02:58:36.198829 main.go:75:   Session cookie: "JSESSIONID"
proxy 2022/08/08 02:58:36.198837 main.go:76:   Sticky cookie: "__VCAP_ID__"
proxy 2022/08/08 02:58:36.198843 main.go:77:   Proxy service: my-app-headless.my-space.svc.cluster.local:8080
proxy 2022/08/08 02:58:36.198843 main.go:86: Listening on :8080
```

For each request, it will also log the request URL, the number of healthy backends, and the destination:

```
proxy 2022/08/08 03:00:53.407454 main.go:133: GET /
proxy 2022/08/08 03:00:53.422316 main.go:175: - 3 healthy backends, forwarding to "http://10.68.0.29:8080/"
```

## Performance

The proxy has good performance, but you should always test with a representative workload.

Attemting 10000 connections with 10 in flight over 10 seconds to a backend with 3 instances:

```
h2load -n 10000 -c 10 -D 10 <proxy/app url>
```

Direct application results:

```
finished in 10.00s, 207.60 req/s, 24.35KB/s
requests: 2076 total, 2086 started, 2076 done, 2076 succeeded, 0 failed, 0 errored, 0 timeout
status codes: 2076 2xx, 0 3xx, 0 4xx, 0 5xx
traffic: 243.54KB (249383) total, 101.09KB (103513) headers (space savings 75.07%), 87.18KB (89268) data
                     min         max         mean         sd        +/- sd
time for request:    45.49ms     62.57ms     47.78ms      1.37ms    76.49%
time for connect:    45.57ms     48.26ms     46.56ms       916us    50.00%
time to 1st byte:    97.97ms    108.68ms    104.34ms      4.24ms    80.00%
req/s           :      20.10       21.30       20.76        0.44    60.00%
```

Same application behind proxy results:

```
finished in 10.00s, 176.70 req/s, 25.63KB/s
requests: 1767 total, 1777 started, 1767 done, 1767 succeeded, 0 failed, 0 errored, 0 timeout
status codes: 1767 2xx, 0 3xx, 0 4xx, 0 5xx
traffic: 256.36KB (262508) total, 135.03KB (138268) headers (space savings 67.71%), 74.20KB (75981) data
                     min         max         mean         sd        +/- sd
time for request:    50.09ms     71.53ms     56.18ms      2.09ms    80.42%
time for connect:    45.61ms     48.28ms     47.29ms       887us    60.00%
time to 1st byte:   115.35ms    119.84ms    117.57ms      1.35ms    60.00%
req/s           :      17.60       17.70       17.67        0.05    70.00%
```

In this test the proxy introduces about a 20ms slowdown in time to first byte and increases
request time about 10ms total.