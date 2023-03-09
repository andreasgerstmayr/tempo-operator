package v1alpha1

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/pointer"

	"github.com/os-observability/tempo-operator/apis/config/v1alpha1"
)

func TestDefault(t *testing.T) {
	defaulter := &Defaulter{
		ctrlConfig: v1alpha1.ProjectConfig{
			DefaultImages: v1alpha1.ImagesSpec{
				Tempo:        "docker.io/grafana/tempo:x.y.z",
				TempoQuery:   "docker.io/grafana/tempo-query:x.y.z",
				TempoGateway: "docker.io/observatorium/gateway:1.2.3",
			},
		},
	}
	defaultMaxSearch := 0
	defaultDefaultResultLimit := 20

	tests := []struct {
		input    *TempoStack
		expected *TempoStack
		name     string
	}{
		{
			name: "no action default values are provided",
			input: &TempoStack{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: TempoStackSpec{
					ReplicationFactor: 2,
					Images: v1alpha1.ImagesSpec{
						Tempo:        "docker.io/grafana/tempo:1.2.3",
						TempoQuery:   "docker.io/grafana/tempo-query:1.2.3",
						TempoGateway: "docker.io/observatorium/gateway:1.2.3",
					},
					ServiceAccount: "tempo-test",
					Retention: RetentionSpec{
						Global: RetentionConfig{
							Traces: metav1.Duration{Duration: time.Hour},
						},
					},
					StorageSize: resource.MustParse("1Gi"),
					LimitSpec: LimitSpec{
						Global: RateLimitSpec{
							Query: QueryLimit{
								MaxSearchBytesPerTrace: &defaultMaxSearch,
							},
						},
					},
				},
			},
			expected: &TempoStack{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: TempoStackSpec{
					ReplicationFactor: 2,
					Images: v1alpha1.ImagesSpec{
						Tempo:        "docker.io/grafana/tempo:1.2.3",
						TempoQuery:   "docker.io/grafana/tempo-query:1.2.3",
						TempoGateway: "docker.io/observatorium/gateway:1.2.3",
					},
					ServiceAccount: "tempo-test",
					Retention: RetentionSpec{
						Global: RetentionConfig{
							Traces: metav1.Duration{Duration: time.Hour},
						},
					},
					StorageSize: resource.MustParse("1Gi"),
					LimitSpec: LimitSpec{
						Global: RateLimitSpec{
							Query: QueryLimit{
								MaxSearchBytesPerTrace: &defaultMaxSearch,
							},
						},
					},
					SearchSpec: SearchSpec{
						MaxDuration:        metav1.Duration{Duration: 0},
						DefaultResultLimit: &defaultDefaultResultLimit,
					},
					Components: TempoComponentsSpec{
						Distributor: TempoComponentSpec{
							Replicas: pointer.Int32(1),
						},
						Ingester: TempoComponentSpec{
							Replicas: pointer.Int32(1),
						},
					},
				},
			},
		},
		{
			name: "default values are set in the webhook",
			input: &TempoStack{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
			expected: &TempoStack{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: TempoStackSpec{
					ReplicationFactor: 1,
					Images: v1alpha1.ImagesSpec{
						Tempo:        "docker.io/grafana/tempo:x.y.z",
						TempoQuery:   "docker.io/grafana/tempo-query:x.y.z",
						TempoGateway: "docker.io/observatorium/gateway:1.2.3",
					},
					ServiceAccount: "tempo-test",
					Retention: RetentionSpec{
						Global: RetentionConfig{
							Traces: metav1.Duration{Duration: 48 * time.Hour},
						},
					},
					StorageSize: resource.MustParse("10Gi"),
					LimitSpec: LimitSpec{
						Global: RateLimitSpec{
							Query: QueryLimit{
								MaxSearchBytesPerTrace: &defaultMaxSearch,
							},
						},
					},
					SearchSpec: SearchSpec{
						MaxDuration:        metav1.Duration{Duration: 0},
						DefaultResultLimit: &defaultDefaultResultLimit,
					},
					Components: TempoComponentsSpec{
						Distributor: TempoComponentSpec{
							Replicas: pointer.Int32(1),
						},
						Ingester: TempoComponentSpec{
							Replicas: pointer.Int32(1),
						},
					},
				},
			},
		},
		{
			name: "use Edge TLS termination if unset",
			input: &TempoStack{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: TempoStackSpec{
					Components: TempoComponentsSpec{
						QueryFrontend: TempoQueryFrontendSpec{
							JaegerQuery: JaegerQuerySpec{
								Enabled: true,
								Ingress: JaegerQueryIngressSpec{
									Type: IngressTypeRoute,
								},
							},
						},
					},
				},
			},
			expected: &TempoStack{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: TempoStackSpec{
					ReplicationFactor: 1,
					Images: v1alpha1.ImagesSpec{
						Tempo:        "docker.io/grafana/tempo:x.y.z",
						TempoQuery:   "docker.io/grafana/tempo-query:x.y.z",
						TempoGateway: "docker.io/observatorium/gateway:1.2.3",
					},
					ServiceAccount: "tempo-test",
					Retention: RetentionSpec{
						Global: RetentionConfig{
							Traces: metav1.Duration{Duration: 48 * time.Hour},
						},
					},
					StorageSize: resource.MustParse("10Gi"),
					LimitSpec: LimitSpec{
						Global: RateLimitSpec{
							Query: QueryLimit{
								MaxSearchBytesPerTrace: &defaultMaxSearch,
							},
						},
					},
					SearchSpec: SearchSpec{
						MaxDuration:        metav1.Duration{Duration: 0},
						DefaultResultLimit: &defaultDefaultResultLimit,
					},
					Components: TempoComponentsSpec{
						Distributor: TempoComponentSpec{
							Replicas: pointer.Int32(1),
						},
						Ingester: TempoComponentSpec{
							Replicas: pointer.Int32(1),
						},
						QueryFrontend: TempoQueryFrontendSpec{
							JaegerQuery: JaegerQuerySpec{
								Enabled: true,
								Ingress: JaegerQueryIngressSpec{
									Type: "route",
									Route: JaegerQueryRouteSpec{
										Termination: "edge",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := defaulter.Default(context.Background(), test.input)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, test.input)
		})
	}
}

func TestValidateStorageSecret(t *testing.T) {
	tempo := TempoStack{
		Spec: TempoStackSpec{
			Storage: ObjectStorageSpec{
				Secret: "testsecret",
			},
		},
	}
	path := field.NewPath("spec").Child("storage").Child("secret")

	tests := []struct {
		name     string
		input    corev1.Secret
		expected field.ErrorList
	}{
		{
			name:  "empty secret",
			input: corev1.Secret{},
			expected: field.ErrorList{
				field.Invalid(path, tempo.Spec.Storage.Secret, "storage secret is empty"),
			},
		},
		{
			name: "missing or empty fields",
			input: corev1.Secret{
				Data: map[string][]byte{
					"bucket": []byte(""),
				},
			},
			expected: field.ErrorList{
				field.Invalid(path, tempo.Spec.Storage.Secret, "storage secret must contain \"endpoint\" field"),
				field.Invalid(path, tempo.Spec.Storage.Secret, "storage secret must contain \"bucket\" field"),
				field.Invalid(path, tempo.Spec.Storage.Secret, "storage secret must contain \"access_key_id\" field"),
				field.Invalid(path, tempo.Spec.Storage.Secret, "storage secret must contain \"access_key_secret\" field"),
			},
		},
		{
			name: "invalid endpoint 'invalid'",
			input: corev1.Secret{
				Data: map[string][]byte{
					"endpoint":          []byte("invalid"),
					"bucket":            []byte("bucket"),
					"access_key_id":     []byte("id"),
					"access_key_secret": []byte("secret"),
				},
			},
			expected: field.ErrorList{
				field.Invalid(path, tempo.Spec.Storage.Secret, "\"endpoint\" field of storage secret must be a valid URL"),
			},
		},
		{
			name: "invalid endpoint '/invalid'",
			input: corev1.Secret{
				Data: map[string][]byte{
					"endpoint":          []byte("/invalid"),
					"bucket":            []byte("bucket"),
					"access_key_id":     []byte("id"),
					"access_key_secret": []byte("secret"),
				},
			},
			expected: field.ErrorList{
				field.Invalid(path, tempo.Spec.Storage.Secret, "\"endpoint\" field of storage secret must be a valid URL"),
			},
		},
		{
			name: "valid storage secret",
			input: corev1.Secret{
				Data: map[string][]byte{
					"endpoint":          []byte("http://minio.minio.svc:9000"),
					"bucket":            []byte("bucket"),
					"access_key_id":     []byte("id"),
					"access_key_secret": []byte("secret"),
				},
			},
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errs := ValidateStorageSecret(tempo, test.input)
			assert.Equal(t, test.expected, errs)
		})
	}
}

func TestValidateReplicationFactor(t *testing.T) {
	validator := &validator{}
	path := field.NewPath("spec").Child("ReplicationFactor")

	tests := []struct {
		name     string
		expected field.ErrorList
		input    TempoStack
	}{
		{
			name: "no error replicas equal to floor(replication_factor/2) + 1",
			input: TempoStack{
				Spec: TempoStackSpec{
					ReplicationFactor: 3,
					Components: TempoComponentsSpec{
						Ingester: TempoComponentSpec{
							Replicas: pointer.Int32(2),
						},
					},
				},
			},
			expected: nil,
		},
		{
			name: "no error replicas greater than floor(replication_factor/2) + 1",
			input: TempoStack{
				Spec: TempoStackSpec{
					ReplicationFactor: 3,
					Components: TempoComponentsSpec{
						Ingester: TempoComponentSpec{
							Replicas: pointer.Int32(3),
						},
					},
				},
			},
			expected: nil,
		},
		{
			name: "error replicas less than floor(replication_factor/2) + 1",
			input: TempoStack{
				Spec: TempoStackSpec{
					ReplicationFactor: 3,
					Components: TempoComponentsSpec{
						Ingester: TempoComponentSpec{
							Replicas: pointer.Int32(1),
						},
					},
				},
			},
			expected: field.ErrorList{
				field.Invalid(path, 3,
					fmt.Sprintf("replica factor of %d requires at least %d ingester replicas", 3, 2),
				)},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errs := validator.validateReplicationFactor(test.input)
			assert.Equal(t, test.expected, errs)
		})
	}
}

func TestValidateIngressAndRoute(t *testing.T) {
	path := field.NewPath("spec").Child("template").Child("queryFrontend").Child("jaegerQuery").Child("ingress").Child("type")

	tests := []struct {
		name       string
		input      TempoStack
		ctrlConfig v1alpha1.ProjectConfig
		expected   field.ErrorList
	}{
		{
			name: "valid ingress configuration",
			input: TempoStack{
				Spec: TempoStackSpec{
					ReplicationFactor: 3,
					Components: TempoComponentsSpec{
						QueryFrontend: TempoQueryFrontendSpec{
							JaegerQuery: JaegerQuerySpec{
								Enabled: true,
								Ingress: JaegerQueryIngressSpec{
									Type: "ingress",
								},
							},
						},
					},
				},
			},
			expected: nil,
		},
		{
			name: "valid route configuration",
			input: TempoStack{
				Spec: TempoStackSpec{
					ReplicationFactor: 3,
					Components: TempoComponentsSpec{
						QueryFrontend: TempoQueryFrontendSpec{
							JaegerQuery: JaegerQuerySpec{
								Enabled: true,
								Ingress: JaegerQueryIngressSpec{
									Type: "route",
								},
							},
						},
					},
				},
			},
			ctrlConfig: v1alpha1.ProjectConfig{
				Gates: v1alpha1.FeatureGates{
					OpenShift: v1alpha1.OpenShiftFeatureGates{
						OpenShiftRoute: true,
					},
				},
			},
			expected: nil,
		},
		{
			name: "ingress enabled but queryfrontend disabled",
			input: TempoStack{
				Spec: TempoStackSpec{
					ReplicationFactor: 3,
					Components: TempoComponentsSpec{
						QueryFrontend: TempoQueryFrontendSpec{
							JaegerQuery: JaegerQuerySpec{
								Enabled: false,
								Ingress: JaegerQueryIngressSpec{
									Type: "ingress",
								},
							},
						},
					},
				},
			},
			expected: field.ErrorList{
				field.Invalid(
					path,
					IngressTypeIngress,
					"Ingress cannot be enabled if jaegerQuery is disabled",
				),
			},
		},
		{
			name: "route enabled but route feature gate disabled",
			input: TempoStack{
				Spec: TempoStackSpec{
					ReplicationFactor: 3,
					Components: TempoComponentsSpec{
						QueryFrontend: TempoQueryFrontendSpec{
							JaegerQuery: JaegerQuerySpec{
								Enabled: true,
								Ingress: JaegerQueryIngressSpec{
									Type: "route",
								},
							},
						},
					},
				},
			},
			ctrlConfig: v1alpha1.ProjectConfig{
				Gates: v1alpha1.FeatureGates{
					OpenShift: v1alpha1.OpenShiftFeatureGates{
						OpenShiftRoute: false,
					},
				},
			},
			expected: field.ErrorList{
				field.Invalid(
					path,
					IngressTypeRoute,
					"Please enable the featureGates.openshift.openshiftRoute feature gate to use Routes",
				),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			validator := &validator{ctrlConfig: test.ctrlConfig}
			errs := validator.validateQueryFrontend(test.input)
			assert.Equal(t, test.expected, errs)
		})
	}
}

func TestValidateGatewayAndJaegerQuery(t *testing.T) {
	path := field.NewPath("spec").Child("components").Child("gateway").Child("enabled")

	tests := []struct {
		name     string
		input    TempoStack
		expected field.ErrorList
	}{
		{
			name: "valid configuration enabled both",
			input: TempoStack{
				Spec: TempoStackSpec{
					ReplicationFactor: 3,
					Components: TempoComponentsSpec{
						QueryFrontend: TempoQueryFrontendSpec{
							JaegerQuery: JaegerQuerySpec{
								Enabled: true,
							},
						},
						Gateway: TempoGatewaySpec{
							Enabled: true,
						},
					},
				},
			},
			expected: nil,
		},
		{
			name: "valid config disable gateway and enable jaegerQuery",
			input: TempoStack{
				Spec: TempoStackSpec{
					ReplicationFactor: 3,
					Components: TempoComponentsSpec{
						QueryFrontend: TempoQueryFrontendSpec{
							JaegerQuery: JaegerQuerySpec{
								Enabled: true,
								Ingress: JaegerQueryIngressSpec{
									Type: "route",
								},
							},
						},
						Gateway: TempoGatewaySpec{
							Enabled: false,
						},
					},
				},
			},
			expected: nil,
		},
		{
			name: "valid config disable both",
			input: TempoStack{
				Spec: TempoStackSpec{
					ReplicationFactor: 3,
					Components: TempoComponentsSpec{
						QueryFrontend: TempoQueryFrontendSpec{
							JaegerQuery: JaegerQuerySpec{
								Enabled: false,
								Ingress: JaegerQueryIngressSpec{
									Type: "route",
								},
							},
						},
						Gateway: TempoGatewaySpec{
							Enabled: false,
						},
					},
				},
			},
			expected: nil,
		},
		{
			name: "invalid config disable jaegerQuery",
			input: TempoStack{
				Spec: TempoStackSpec{
					ReplicationFactor: 3,
					Components: TempoComponentsSpec{
						QueryFrontend: TempoQueryFrontendSpec{
							JaegerQuery: JaegerQuerySpec{
								Enabled: false,
								Ingress: JaegerQueryIngressSpec{
									Type: "ingress",
								},
							},
						},
						Gateway: TempoGatewaySpec{
							Enabled: true,
						},
					},
				},
			},
			expected: field.ErrorList{
				field.Invalid(path, true,
					"gateway require enable jaeger query to work.",
				),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			validator := &validator{ctrlConfig: v1alpha1.ProjectConfig{}}
			errs := validator.validateGateway(test.input)
			assert.Equal(t, test.expected, errs)
		})
	}
}
