/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package source

import (
	"strings"

	v1 "k8s.io/api/core/v1"

	"sigs.k8s.io/external-dns/endpoint"
)

const (
	mateAnnotationKey     = "zalando.org/dnsname"
	moleculeAnnotationKey = "domainName"
)

// LegacyEndpointsFromService tries to retrieve Endpoints from Services
// annotated with legacy annotations.
func LegacyEndpointsFromService(svc *v1.Service, compatibility string) []*endpoint.Endpoint {
	switch compatibility {
	case "mate":
		return LegacyEndpointsFromMateService(svc)
	case "molecule":
		return LegacyEndpointsFromMoleculeService(svc)
	}

	return []*endpoint.Endpoint{}
}

// LegacyEndpointsFromMateService tries to retrieve Endpoints from Services
// annotated with Mate's annotation semantics.
func LegacyEndpointsFromMateService(svc *v1.Service) []*endpoint.Endpoint {
	var endpoints []*endpoint.Endpoint

	// Get the desired hostname of the service from the annotation.
	hostname, exists := svc.Annotations[mateAnnotationKey]
	if !exists {
		return nil
	}

	// Create a corresponding endpoint for each configured external entrypoint.
	for _, lb := range svc.Status.LoadBalancer.Ingress {
		if lb.IP != "" {
			endpoints = append(endpoints, endpoint.NewEndpoint(hostname, endpoint.RecordTypeA, lb.IP))
		}
		if lb.Hostname != "" {
			endpoints = append(endpoints, endpoint.NewEndpoint(hostname, endpoint.RecordTypeCNAME, lb.Hostname))
		}
	}

	return endpoints
}

// LegacyEndpointsFromMoleculeService tries to retrieve Endpoints from Services
// annotated with Molecule Software's annotation semantics.
func LegacyEndpointsFromMoleculeService(svc *v1.Service) []*endpoint.Endpoint {
	var endpoints []*endpoint.Endpoint

	// Check that the Service opted-in to being processed.
	if svc.Labels["dns"] != "route53" {
		return nil
	}

	// Get the desired hostname of the service from the annotation.
	hostnameAnnotation, exists := svc.Annotations[moleculeAnnotationKey]
	if !exists {
		return nil
	}

	hostnameList := strings.Split(strings.Replace(hostnameAnnotation, " ", "", -1), ",")

	for _, hostname := range hostnameList {
		// Create a corresponding endpoint for each configured external entrypoint.
		for _, lb := range svc.Status.LoadBalancer.Ingress {
			if lb.IP != "" {
				endpoints = append(endpoints, endpoint.NewEndpoint(hostname, endpoint.RecordTypeA, lb.IP))
			}
			if lb.Hostname != "" {
				endpoints = append(endpoints, endpoint.NewEndpoint(hostname, endpoint.RecordTypeCNAME, lb.Hostname))
			}
		}
	}

	return endpoints
}
