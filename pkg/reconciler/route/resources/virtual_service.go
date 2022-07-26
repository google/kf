// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"errors"
	"fmt"
	"net/http"
	"path"
	"sort"
	"strings"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	kfistio "github.com/google/kf/v2/pkg/apis/networking/v1alpha3"
	serviceinstance "github.com/google/kf/v2/pkg/reconciler/serviceinstance/resources"
	istio "istio.io/api/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (

	// KfAppMatchHeader is an HTTP header supplied with the name of an app
	// so a route will direct all traffic directly to that app.
	KfAppMatchHeader = "X-Kf-App"

	// CfForwardedURLHeader contains the originally requested URL.
	// A route service may choose to forward the request to this URL or to another.
	// Route Services in CF are required to forward this header.
	CfForwardedURLHeader = "X-CF-Forwarded-Url"

	// CfProxySignatureHeader is an encrypted value that signals that the request has already gone through a route service.
	// Route Services in CF are required to forward this header.
	CfProxySignatureHeader = "X-CF-Proxy-Signature"

	// CfProxyMetadataHeader aids in the encryption and description of X-CF-Proxy-Signature.
	// Route Services in CF are required to forward this header.
	CfProxyMetadataHeader = "X-CF-Proxy-Metadata"

	// TODO: replace this with logical values and review security (b/158039001)
	// Header values for route services are currently stubbed out to test other aspects of route reconciliation.
	noopCFHeaderValue = "noopRouteServiceHeaderValue"

	// DomainAnnotation is the annotation key that holds the domain.
	DomainAnnotation = "kf.dev/domain"

	// KfExternalIngressGateway holds the gateway for Kf's external HTTP ingress.
	KfExternalIngressGateway = "kf/external-gateway"

	// KfInternalIngressGateway is used as a flag to specify the internal routing.
	// With ASM 1.7 all the internal east-west traffic can use side car proxy and we do not need any additional gateway.
	KfInternalIngressGateway = "kf/internal-gateway"
)

// RouteBindingSlice is a sortable list of v1alpha1.RouteDestination.
type RouteBindingSlice []v1alpha1.RouteDestination

var _ sort.Interface = (RouteBindingSlice)(nil)

// Len implements sort.Interface::Len
func (a RouteBindingSlice) Len() int {
	return len(a)
}

// Less implements sort.Interface::Less
func (a RouteBindingSlice) Less(i, j int) bool {
	if a[i].ServiceName == a[j].ServiceName {
		return a[i].Port < a[j].Port
	}
	return a[i].ServiceName < a[j].ServiceName
}

// Swap implements sort.Interface::Swap
func (a RouteBindingSlice) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// MakeVirtualServiceName creates the name of the VirtualService with the given
// domain.
func MakeVirtualServiceName(domain string) string {
	return v1alpha1.GenerateName(domain)
}

// MakeVirtualService creates a VirtualService from a Route object.
func MakeVirtualService(routes []*v1alpha1.Route, bindings map[string]RouteBindingSlice, routeServiceBindings map[string][]v1alpha1.RouteServiceDestination, spaceDomain *v1alpha1.SpaceDomain) (*kfistio.VirtualService, error) {
	if len(routes) == 0 {
		return nil, errors.New("routes must not be empty")
	}

	namespace := routes[0].Namespace
	domain := routes[0].Spec.RouteSpecFields.Domain

	// Create sorted list of RouteSpecFields from the list of Routes.
	// The VS rules will be populated based on the order of these RouteSpecFields.
	rsfs := v1alpha1.RouteSpecFieldsSlice{}
	for _, r := range routes {
		rsfs = append(rsfs, r.Spec.RouteSpecFields)
	}
	sort.Sort(rsfs)

	httpRoutes, err := buildHTTPRoutes(rsfs, bindings, routeServiceBindings)
	if err != nil {
		return nil, err
	}

	gatewayName, err := buildGatewayName(spaceDomain.GatewayName)
	if err != nil {
		return nil, err
	}

	// Mark all of the Routes as owners so the VS gets deleted if
	// they are all deleted.
	//
	// NOTE that this is NOT marking them as controllers, just as equal owners.
	var owners []metav1.OwnerReference
	for _, route := range routes {
		gvk := route.GetGroupVersionKind()
		owners = append(owners, metav1.OwnerReference{
			APIVersion: gvk.GroupVersion().String(),
			Kind:       gvk.Kind,
			UID:        route.GetUID(),
			Name:       route.GetName(),
		})
	}

	sort.Slice(owners, func(i, j int) bool {
		return owners[i].Name < owners[j].Name
	})

	istioVirtualService := istio.VirtualService{}
	if gatewayName == "" {
		istioVirtualService = istio.VirtualService{
			Hosts: []string{"*." + domain, domain}, // b/176970436: Value of Hosts can be hostname.example.com or example.com since hostname is optional.
			Http:  httpRoutes,
		}
	} else {
		istioVirtualService = istio.VirtualService{
			Gateways: []string{gatewayName},
			Hosts:    []string{"*." + domain, domain},
			Http:     httpRoutes,
		}
	}

	// Configuring the VirtualService based on the spec definition: https://istio.io/latest/docs/reference/config/networking/virtual-service/
	return &kfistio.VirtualService{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "VirtualService",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      MakeVirtualServiceName(domain),
			Namespace: namespace,
			Labels: map[string]string{
				v1alpha1.ManagedByLabel: "kf",
				v1alpha1.ComponentLabel: "virtualservice",
			},
			Annotations: map[string]string{
				DomainAnnotation: domain,
			},
			OwnerReferences: owners,
		},
		Spec: istioVirtualService,
	}, nil
}

// Checks the gatewayName specified.
// For gatewayName=kf/internal-gateway use the internal service mesh and do not use any gateway.
func buildGatewayName(gatewayName string) (string, error) {
	if KfInternalIngressGateway == gatewayName {
		return "", nil
	}
	return gatewayName, nil
}

// Create HTTP route rules for all routes with the same domain.
// Paths that do not have an app bound to them will return a 404 when a request is sent to that path.
func buildHTTPRoutes(routes v1alpha1.RouteSpecFieldsSlice, appBindings map[string]RouteBindingSlice, routeServiceBindings map[string][]v1alpha1.RouteServiceDestination) ([]*istio.HTTPRoute, error) {
	var httpRoutes []*istio.HTTPRoute

	for _, r := range routes {
		var rsfHTTPRoutes []*istio.HTTPRoute
		appDestinations := appBindings[r.String()]

		// Get regex path matchers for route path
		pathMatchers, err := buildPathMatchers(r)
		if err != nil {
			return nil, err
		}

		if len(appDestinations) == 0 {
			// no apps bound to this path, return http route with fault for path
			rsfHTTPRoutes = append(rsfHTTPRoutes, buildDefaultHTTPRoute(*pathMatchers))
		} else {
			// Build HTTP Routes with `x-kf-app` header for matching specific app requests.
			appHeaderHTTPRoutes, err := buildAppHeaderHTTPRoutes(r, appDestinations)
			if err != nil {
				return nil, err
			}

			// Build HTTP Route without the `x-kf-app` header.
			// If only one app is bound to the route, this HTTP Route looks the same as the app header HTTP route, just without the app header match.
			// If the app is bound but stopped, the HTTP Route returns a 404.
			// If there are multiple apps bound, the HTTP Route splits traffic to the app destinations according to the weights on the bindings.
			normalizedHTTPRoute := buildNormalizedHTTPRoute(pathMatchers, r, appDestinations)

			rsfHTTPRoutes = append(rsfHTTPRoutes, appHeaderHTTPRoutes...)
			rsfHTTPRoutes = append(rsfHTTPRoutes, normalizedHTTPRoute)
		}

		// If there is a route service bound to this route, add header match rules to each HTTP route.
		// Then add an HTTP Route that directs to the route service and adds the CF route service headers.
		// Note: There should only be one route service per route, but we handle the case where multiple are bound.
		// The last (most recent) route service is used for the VS definition, and the RouteServiceReady condition for the Route is set to False in the reconciler.
		routeServices := routeServiceBindings[r.String()]
		if len(routeServices) > 0 {
			routeService := routeServices[len(routeServices)-1]
			// Copy original matchers (without the route service header matchers) to use in final HTTP route
			origPathMatchers := *pathMatchers
			for _, httpRoute := range rsfHTTPRoutes {
				var newMatchRules []*istio.HTTPMatchRequest
				// There should only be one match rule defined per HTTP route, but iterate through the list just in case.
				for _, matchRule := range httpRoute.Match {
					newMatchRules = append(newMatchRules, addRouteServiceHeaderMatchers(r, matchRule))
				}
				httpRoute.Match = newMatchRules
			}

			// Add HTTP route for directing request to route service
			routeServiceHTTPRoute := buildRouteServiceHTTPRoute(r, &origPathMatchers, routeService)
			rsfHTTPRoutes = append(rsfHTTPRoutes, routeServiceHTTPRoute)
		}
		httpRoutes = append(httpRoutes, rsfHTTPRoutes...)
	}

	// HTTPRoutes should be grouped by RouteSpecFields, where RouteSpecFields are sorted alphabetically by hostname, then
	// within routes with the same hostname + domain, longest paths come first.
	// Routing rules are evaluated in order from first to last, where the first rule is given highest priority.
	return httpRoutes, nil
}

// buildAppHeaderHTTPRoutes creates a list of HTTPRoutes, where each HTTPRoute contains a header matching rule for an app destination.
// Even when there are multiple apps mapped to a route,
// Kf always directs to the requested app if the request contains the header "x-kf-app": [appname]
// If a route service is bound to the route, then request should first be processed by the route service (indicated by the headers), then be directed to the app.
func buildAppHeaderHTTPRoutes(rsf v1alpha1.RouteSpecFields, bindings RouteBindingSlice) ([]*istio.HTTPRoute, error) {
	appHeaderRoutes := []*istio.HTTPRoute{}
	for _, binding := range bindings {
		httpRoute := &istio.HTTPRoute{}
		pathAppMatchers, err := buildPathAppMatchers(rsf, binding.ServiceName)
		if err != nil {
			return nil, err
		}
		if binding.Weight == 0 {
			// If an app is stopped, return a 404 when a request is sent to that app
			httpRoute = buildDefaultHTTPRoute(*pathAppMatchers)
		} else {
			httpRoute = &istio.HTTPRoute{
				Match: []*istio.HTTPMatchRequest{pathAppMatchers},
				Route: []*istio.HTTPRouteDestination{
					{
						Destination: &istio.Destination{
							Host: binding.ServiceName,
							Port: &istio.PortSelector{
								Number: uint32(binding.Port),
							},
						},
						Weight: 100,
					},
				},
			}
		}
		appHeaderRoutes = append(appHeaderRoutes, httpRoute)
	}
	return appHeaderRoutes, nil
}

// buildNormalizedHTTPRoute creates an HTTP Route for the app binding(s) on the route, without the `x-kf-app` header match rule.
// It normalizes the weights defined on the app route binding(s) to percentages and splits traffic among multiple apps.
func buildNormalizedHTTPRoute(pathMatchers *istio.HTTPMatchRequest, rsf v1alpha1.RouteSpecFields, bindings RouteBindingSlice) *istio.HTTPRoute {
	// If an app is stopped, exclude them from the route destinations
	httpRoute := &istio.HTTPRoute{}
	normalizedBindings := normalizeRouteWeights(bindings)
	if len(normalizedBindings) == 0 {
		httpRoute = buildDefaultHTTPRoute(*pathMatchers)
	} else {
		httpRoute = &istio.HTTPRoute{
			Match: []*istio.HTTPMatchRequest{pathMatchers},
			Route: buildRouteDestinations(normalizedBindings),
		}
	}
	return httpRoute
}

// buildRouteServiceHTTPRoute creates an HTTP Route that handles directing a new request to a route service (if the request has not been processed).
// The HTTP Route directs the request to the Kf proxy service (created for each route service), which adds the `X-CF-Forwarded-URL` header on the request
// and forwards the request to the route service.
// This function adds two other headers to the request that match the CF Route Service headers.
func buildRouteServiceHTTPRoute(rsf v1alpha1.RouteSpecFields, pathMatchers *istio.HTTPMatchRequest, routeService v1alpha1.RouteServiceDestination) *istio.HTTPRoute {
	return &istio.HTTPRoute{
		Match: []*istio.HTTPMatchRequest{pathMatchers},
		Route: []*istio.HTTPRouteDestination{
			{
				Destination: &istio.Destination{
					Host: serviceinstance.ServiceNameForRouteServiceName(routeService.Name),
				},
				Weight: 100, // only one route service can be bound to a route
			},
		},
		Headers: &istio.Headers{
			Request: &istio.Headers_HeaderOperations{
				Add: buildRouteServiceHeaderValues(rsf),
			},
		},
	}
}

// Hostname + domain + path combos with bound app(s) have a custom route destination for each path.
// The request is sent directly to the Service for that app.
// If there are multiple apps bound to a route, the traffic is split uniformly across the apps.
func buildRouteDestinations(normalizedRouteBindings RouteBindingSlice) []*istio.HTTPRouteDestination {
	routeDestinations := []*istio.HTTPRouteDestination{}

	// the bindings should be sorted at this point, but rather than panic
	// if they're not, we'll just sort again for safety
	sort.Sort(normalizedRouteBindings)
	for _, binding := range normalizedRouteBindings {
		routeDestination := &istio.HTTPRouteDestination{
			Destination: &istio.Destination{
				Host: binding.ServiceName,
				Port: &istio.PortSelector{
					Number: uint32(binding.Port),
				},
			},
			Weight: int32(binding.Weight),
		}
		routeDestinations = append(routeDestinations, routeDestination)
	}

	return routeDestinations
}

// normalizeRouteWeights generates integer percentages for route weights that sum to 100, and returns
// a copy of the bindings with their weights normalized to add to 100. Apps that are stopped are not included.
// If the weight proportions do not evenly divide 100, the weights are calculated as follows:
// Round all the weights down, find the difference between that sum and 100, then distribute
// the difference among the weights.
//
// e.g. if number of routes = 6 and each app has weight 1, then 100/6 = 16.666, which rounds down to 16,
// with a remainder of 100 % 6 = 4.
// The final percentages would be [17, 17, 17, 17, 16, 16].
// If there are 3 apps and one app has weight 2 and two apps have weight 1, then the app with weight 2
// will have a weight percentage double the other apps.
// The final percentages would be [50, 25, 25].
func normalizeRouteWeights(appWeights RouteBindingSlice) RouteBindingSlice {
	// Sort the incoming weights so the algorithm is deterministic
	// and make a copy so we don't modify the original
	var sortedWeights RouteBindingSlice
	for _, copy := range appWeights {
		sortedWeights = append(sortedWeights, copy)
	}
	sort.Sort(sortedWeights)

	var totalWeight int32
	for _, appWeight := range sortedWeights {
		totalWeight += appWeight.Weight
	}

	if totalWeight == 0 {
		return nil
	}

	var remainder int32 = 100

	for idx := range sortedWeights {
		if sortedWeights[idx].Weight != 0 {
			routeWeight := (100 * sortedWeights[idx].Weight) / totalWeight // round down
			remainder -= routeWeight
			sortedWeights[idx].Weight = routeWeight
		}
	}

	// Distribute remainder among the route weight percentages, so that they sum to 100
	for idx := range sortedWeights {
		if remainder == 0 {
			break
		}
		sortedWeights[idx].Weight++
		remainder--
	}

	return sortedWeights
}

// buildPathMatchers creates a regex matcher for a given route path.
// These matcher is used in the virtual service to determine which path a request was sent to
func buildPathMatchers(rsf v1alpha1.RouteSpecFields) (*istio.HTTPMatchRequest, error) {
	path := path.Join("/", rsf.Path, "/")
	regexpPath, err := v1alpha1.BuildPathRegexp(path)
	if err != nil {
		return nil, fmt.Errorf("failed to convert path to regexp: %s", err)
	}

	var authorityMatch *istio.StringMatch

	if !strings.HasPrefix(rsf.Host(), "*") {
		authorityMatch = &istio.StringMatch{
			MatchType: &istio.StringMatch_Exact{
				Exact: rsf.Host(),
			},
		}
	}

	return &istio.HTTPMatchRequest{
		Uri: &istio.StringMatch{
			MatchType: &istio.StringMatch_Regex{
				Regex: regexpPath,
			},
		},
		Authority: authorityMatch,
	}, nil
}

// buildPathAppMatchers creates regex matchers for a route path
// and a header request matcher for a given app name.
// These matchers are used in the virtual service to determine which app to direct a request to
func buildPathAppMatchers(rsf v1alpha1.RouteSpecFields, appName string) (*istio.HTTPMatchRequest, error) {
	matchers, err := buildPathMatchers(rsf)
	if err != nil {
		return nil, err
	}

	// matchers.Headers takes lower-case HTTP headers
	matchers.Headers = map[string]*istio.StringMatch{
		strings.ToLower(KfAppMatchHeader): {
			MatchType: &istio.StringMatch_Exact{
				Exact: appName,
			},
		},
	}

	return matchers, nil
}

// addRouteServiceHeaderMatchers adds header match rules required to determine if the request has already
// been processed by a bound route service.
// It adds match rules for the headers `x-cf-proxy-signature`, and `x-cf-proxy-metadata`.
// Eventually the headers will contain information about the request, but for now, they are placeholder values.
func addRouteServiceHeaderMatchers(rsf v1alpha1.RouteSpecFields, matchers *istio.HTTPMatchRequest) *istio.HTTPMatchRequest {
	headerMatchers := matchers.Headers
	if len(headerMatchers) == 0 {
		headerMatchers = make(map[string]*istio.StringMatch)
	}
	routeServiceHeaders := make(map[string]*istio.StringMatch)
	for header, val := range buildRouteServiceHeaderMatchers(rsf) {
		routeServiceHeaders[header] = &istio.StringMatch{
			MatchType: &istio.StringMatch_Exact{
				Exact: val,
			},
		}
	}

	// Merge route service header rules with existing matcher rules
	for header, match := range routeServiceHeaders {
		headerMatchers[header] = match
	}

	matchers.Headers = headerMatchers
	return matchers
}

// buildRouteServiceHeaderMatchers returns the route service headers, where the header names are lowercase.
// Istio header key matchers are required to be lowercase.
func buildRouteServiceHeaderMatchers(rsf v1alpha1.RouteSpecFields) map[string]string {
	headerMatchers := make(map[string]string)
	for header, val := range buildRouteServiceHeaderValues(rsf) {
		headerMatchers[strings.ToLower(header)] = val
	}
	return headerMatchers
}

// buildRouteServiceHeaderValues creates route service headers that are added to requests going to a route service.
// The headers are kept in their original case to support route services that are case sensitive when forwarding headers.
// The "X-CF-Forwarded-Url" header is not set in the VS rule since the value is dependent on the original request URL (including the full path and query params).
// Instead, the "X-CF-Forwarded-Url" header is set in the Kf proxy service.
func buildRouteServiceHeaderValues(rsf v1alpha1.RouteSpecFields) map[string]string {
	return map[string]string{
		CfProxySignatureHeader: noopCFHeaderValue,
		CfProxyMetadataHeader:  noopCFHeaderValue,
	}
}

// buildDefaultHTTPRoute creates a default route that returns a 404 for a given matcher.
func buildDefaultHTTPRoute(matchers istio.HTTPMatchRequest) *istio.HTTPRoute {
	return &istio.HTTPRoute{
		Match: []*istio.HTTPMatchRequest{&matchers},
		Fault: &istio.HTTPFaultInjection{
			Abort: &istio.HTTPFaultInjection_Abort{
				Percentage: &istio.Percent{Value: 100},
				ErrorType: &istio.HTTPFaultInjection_Abort_HttpStatus{
					HttpStatus: http.StatusNotFound,
				},
			},
		},
		// Istio requires a destination route even if the abort is at 100%
		Route: []*istio.HTTPRouteDestination{
			{
				Destination: &istio.Destination{
					Host: "null.invalid",
				},
				Weight: 100,
			},
		},
	}
}
