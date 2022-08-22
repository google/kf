---
title: Set up HTTPS ingress
description: Learn to set up a TLS certificate to enable HTTPS.
---

You can secure the ingress gateway with HTTPS by using simple TLS, and enable HTTPS connections to specific webpages. In addition, you can redirect HTTP connections to HTTPS.

HTTPS creates a secure channel over an insecure network, protecting against man-in-the-middle attacks and encrypting traffic between the client and server. To prepare a web server to accept HTTPS connections, an administrator must create a public key certificate for the server. This certificate must be signed by a trusted certificate authority for a web browser to accept it without warning.

{{< note >}} These instructions supplement the Istio instructions to [configure a TLS ingress gateway](https://istio.io/latest/docs/tasks/traffic-management/ingress/secure-ingress), and assume that a valid certificate and private key for the server have already been created.{{< /note >}}

Edit the gateway named external-gateway in the `kf` namespace using the built-in Kubernetes editor:

```sh
kubectl edit gateway -n kf external-gateway
```

1. Assuming you have a certificate and key for your service, create a Kubernetes secret for the ingress gateway. Make sure the secret name does not begin with `istio` or `prometheus`. For this example, the secret is named `myapp-https-credential`.
1. Under `servers:`
  1. Add a section for port 443.
  1. Under `tls:`, set the `credentialName` to the name of the secret you just created.
  1. Under `hosts:`, add the host name of the service you want to secure with HTTPS. This can be set to an entire domain using a wildcard (e.g. `*.example.com`) or scoped to just one hostname (e.g. `myapp.example.com`).
1. There should already be a section under `servers:` for port 80 HTTP. Keep this section in the Gateway definition if you would like all traffic to come in as HTTP.
1. To redirect HTTP to HTTPS, add the value `httpsRedirect: true` under `tls` in the HTTP server section. See the [Istio Gateway documentation](https://istio.io/latest/docs/reference/config/networking/gateway/) for reference. Note that adding this in the section where `hosts` is set to `*` means that **all** traffic is redirected to HTTPS. If you only want to redirect HTTP to HTTPS for a single app/domain, add a separate HTTP section specifying the redirect.

Shown below is an example of a Gateway `spec` that sets up HTTPS for `myapp.example.com` and redirects HTTP to HTTPS for that host:

```yaml
spec:
  selector:
    istio: ingressgateway
  servers:
  - hosts:
    - myapp.example.com
    port:
      name: https
      number: 443
      protocol: HTTPS
    tls:
      credentialName: myapp-https-credential
      mode: SIMPLE
  - hosts:
    - myapp.example.com
    port:
      name: http-my-app
      number: 80
      protocol: HTTP
    tls:
      httpsRedirect: true
  - hosts:
    - '*'
    port:
      name: http
      number: 80
      protocol: HTTP
```
