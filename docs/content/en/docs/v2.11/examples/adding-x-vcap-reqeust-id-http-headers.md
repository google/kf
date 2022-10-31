---
title: Add X-VCAP-REQUEST-ID HTTP Headers.
description: >
  Learn how to add additional Cloud Foundry style headers to your gateways.
---

Kf limits the headers it returns for security and network cost purposes.

If you have applications that need the `X-VCAP-REQUEST-ID` HTTP header and
can't be upgraded, then you can use Istio to add it to requests and responses
to mimic CLoud Foundry's gorouter.

{{% warning %}}Use this example at your own risk. In the future, Kf may move away from requiring
Istio APIs.
{{% /warning %}}

To mimic this header, we can create an `EnvoyFilter` that does the following:

1. Watches for HTTP traffic coming into the gateway (make sure you target the same ingress gateway Kf is using).
1. Saves Envoy's built-in request ID.
1. Copies that ID to the request.
1. Mutates the response with the same request ID.

You may need to make changes to this filter if:

* You rely on the HTTP/1.1 `Upgrade` header.
* You need these headers for mesh (East-West) traffic.
* You only want to target a subset of applications.

## Example

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: EnvoyFilter
metadata:
  name: vcap-http-header
  # Set the namespace to match the gateway.
  namespace: asm-gateways
spec:
  # Set the workload selector to match the Istio ingress gateway
  # your domain targets and/or your workload
  workloadSelector:
    labels:
      istio: ingressgateway
  configPatches:
  - applyTo: HTTP_FILTER
    match:
      context: GATEWAY
      listener:
        filterChain:
          filter:
            name: "envoy.filters.network.http_connection_manager"
            subFilter:
              name: "envoy.filters.http.router"
    patch:
      operation: INSERT_BEFORE
      value:
       name: envoy.lua
       typed_config:
         "@type": "type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua"
         inlineCode: |
          function envoy_on_request(request)
             local metadata = request:streamInfo():dynamicMetadata()
             -- Get Envoy's internal request ID
             local request_id = request:headers():get("x-request-id")

             if request_id ~= nil then
               -- Save the request ID for later and set it on the request
               -- for the application to conusme.
               metadata:set("envoy.filters.http.lua", "req.x-request-id", request_id)
               request:headers():add("x-vcap-request-id", request_id)
             end
           end

           function envoy_on_response(response)
             local metadata = response:streamInfo():dynamicMetadata():get("envoy.filters.http.lua")
             local request_id = metadata["req.x-request-id"]

             -- Set the value on the outbound response as well.
             if request_id ~= nil then
               response:headers():add("x-vcap-request-id", request_id)
             end
           end
```