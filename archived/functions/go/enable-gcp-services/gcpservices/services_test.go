package gcpservices

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetServicesList(t *testing.T) {
	tests := []struct {
		name      string
		resName   string
		services  []string
		projectID string
		want      []Service
		errMsg    string
	}{
		{
			name:     "simple",
			resName:  "project-services",
			services: []string{"compute.googleapis.com", "bigquery.googleapis.com"},
			want: []Service{
				getService("project-services-compute", "compute.googleapis.com", ""),
				getService("project-services-bigquery", "bigquery.googleapis.com", ""),
			},
		},
		{
			name:      "with projectID",
			resName:   "test",
			services:  []string{"compute.googleapis.com", "bigquery.googleapis.com"},
			projectID: "my-project",
			want: []Service{
				getService("test-compute", "compute.googleapis.com", "my-project"),
				getService("test-bigquery", "bigquery.googleapis.com", "my-project"),
			},
		},
		{
			name:     "invalid",
			resName:  "project-services",
			services: []string{"foogoogleapis.com", "bigquery.googleapis.com"},
			errMsg:   "invalid service specified: foogoogleapis.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			got, err := createServicesList(tt.resName, tt.services, tt.projectID)
			if tt.errMsg != "" {
				require.NotNil(err)
				require.Contains(err.Error(), tt.errMsg)
			} else {
				require.NoError(err)
				require.ElementsMatch(got, tt.want, "Services should match")
			}
		})
	}
}

func getService(name, res, projectID string) Service {
	service := Service{}
	service.APIVersion = serviceUsageAPIVersion
	service.Kind = serviceUsageKind
	service.Name = name
	service.Spec.ResourceID = res
	if projectID != "" {
		service.Spec.ProjectRef.External = projectID
	}
	return service
}
