# Route Service Proxy

This is a Go application that acts as a proxy for route services.

Route services can apply transformations to an HTTP request before the request reaches its target application. Common use cases include authentication, rate limiting, and caching services. Developers can bind an application’s route to a route service instance.

When an HTTP request is sent to one of these routes, the request first hits this proxy service, which adds the `X-CF-Forwarded-URL` header and forwards the request to the route service. After processing the request, the route service is responsible for forwarding the request back to the URL provided in the `X-CF-Forwarded-URL` header.