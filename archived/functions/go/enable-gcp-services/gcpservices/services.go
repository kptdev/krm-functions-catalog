package gcpservices

import (
	"fmt"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Service represents a valid GCP Service
type Service struct {
	yaml.ResourceMeta `json:",inline" yaml:",inline"`
	Spec              serviceSpec `yaml:"spec"`
}

type serviceSpec struct {
	ResourceID string     `yaml:"resourceID"`
	ProjectRef projectRef `yaml:"projectRef,omitempty"`
}
type projectRef struct {
	External string `yaml:"external,omitempty"`
}

// createServicesList creates a slice of services from service strings
func createServicesList(name string, inputServices []string, projectID string) ([]Service, error) {
	var services []Service
	for _, serviceName := range inputServices {
		if len(strings.Split(serviceName, ".")) < 3 {
			return nil, fmt.Errorf("invalid service specified: %s", serviceName)
		}
		serviceShortName := strings.Split(serviceName, ".")[0]
		service := Service{}
		service.APIVersion = serviceUsageAPIVersion
		service.Kind = serviceUsageKind
		service.Name = fmt.Sprintf("%s-%s", name, serviceShortName)
		service.Spec.ResourceID = serviceName
		if projectID != "" {
			service.Spec.ProjectRef.External = projectID
		}
		services = append(services, service)
	}
	return services, nil
}

// createService generates a Service RNode from Service struct
func createService(s Service) (*yaml.RNode, error) {
	yml, err := yaml.Marshal(s)
	if err != nil {
		return nil, err
	}
	r, err := yaml.Parse(string(yml))
	if err != nil {
		return nil, err
	}
	return r, nil
}
