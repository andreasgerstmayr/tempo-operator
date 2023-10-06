package v1alpha1

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateProjectConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    ProjectConfig
		expected error
	}{
		{
			name: "valid featureGates.tlsProfile setting",
			input: ProjectConfig{
				Images: ImagesSpec{
					Tempo:           "docker.io/grafana/tempo:latest",
					TempoQuery:      "docker.io/grafana/tempo-query:latest",
					TempoGateway:    "quay.io/observatorium/api:latest",
					TempoGatewayOpa: "quay.io/observatorium/opa-openshift:latest",
				},
				Gates: FeatureGates{
					TLSProfile: string(TLSProfileModernType),
				},
			},
			expected: nil,
		},
		{
			name: "invalid featureGates.tlsProfile setting",
			input: ProjectConfig{
				Gates: FeatureGates{
					TLSProfile: "abc",
				},
			},
			expected: errors.New("invalid value 'abc' for setting featureGates.tlsProfile (valid values: Old, Intermediate and Modern)"),
		},
		{
			name:     "empty featureGates.tlsProfile setting",
			input:    ProjectConfig{},
			expected: errors.New("invalid value '' for setting featureGates.tlsProfile (valid values: Old, Intermediate and Modern)"),
		},
		{
			name: "invalid tempo container image",
			input: ProjectConfig{
				Images: ImagesSpec{
					Tempo:           "abc@def",
					TempoQuery:      "docker.io/grafana/tempo-query:latest",
					TempoGateway:    "quay.io/observatorium/api:latest",
					TempoGatewayOpa: "quay.io/observatorium/opa-openshift:latest",
				},
				Gates: FeatureGates{
					TLSProfile: "Modern",
				},
			},
			expected: errors.New("invalid value 'abc@def': please set the RELATED_IMAGE_TEMPO environment variable to a valid container image"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.input.Validate())
		})
	}
}
