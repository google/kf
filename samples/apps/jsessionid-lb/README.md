# `JSESSIONID` Load Balancer

This application acts as a sticky load balancer for HTTP sessions with a `JSESSIONID`
cokie attached.

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