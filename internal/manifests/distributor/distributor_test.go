package distributor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"

	configv1alpha1 "github.com/os-observability/tempo-operator/apis/config/v1alpha1"
	"github.com/os-observability/tempo-operator/apis/tempo/v1alpha1"
	"github.com/os-observability/tempo-operator/internal/manifests/manifestutils"
)

func TestBuildDistributor(t *testing.T) {
	objects, err := BuildDistributor(manifestutils.Params{Tempo: v1alpha1.Microservices{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "project1",
		},
		Spec: v1alpha1.MicroservicesSpec{
			Images: configv1alpha1.ImagesSpec{
				Tempo: "docker.io/grafana/tempo:1.5.0",
			},
			ServiceAccount: "tempo-test-serviceaccount",
			Components: v1alpha1.TempoComponentsSpec{
				Distributor: v1alpha1.TempoComponentSpec{
					Replicas:     pointer.Int32(1),
					NodeSelector: map[string]string{"a": "b"},
					Tolerations: []corev1.Toleration{
						{
							Key: "c",
						},
					},
				},
			},
			Resources: v1alpha1.Resources{
				Total: &corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1000m"),
						corev1.ResourceMemory: resource.MustParse("2Gi"),
					},
				},
			},
		},
	}})
	require.NoError(t, err)

	labels := manifestutils.ComponentLabels("distributor", "test")
	annotations := manifestutils.CommonAnnotations("")
	assert.Equal(t, 2, len(objects))
	assert.Equal(t, &v1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tempo-test-distributor",
			Namespace: "project1",
			Labels:    labels,
		},
		Spec: v1.DeploymentSpec{
			Replicas: pointer.Int32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      k8slabels.Merge(labels, map[string]string{"tempo-gossip-member": "true"}),
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "tempo-test-serviceaccount",
					NodeSelector:       map[string]string{"a": "b"},
					Tolerations: []corev1.Toleration{
						{
							Key: "c",
						},
					},
					Affinity: manifestutils.DefaultAffinity(labels),
					Containers: []corev1.Container{
						{
							Name:  "tempo",
							Image: "docker.io/grafana/tempo:1.5.0",
							Args:  []string{"-target=distributor", "-config.file=/conf/tempo.yaml"},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      manifestutils.ConfigVolumeName,
									MountPath: "/conf",
									ReadOnly:  true,
								},
								{
									Name:      manifestutils.TmpStorageVolumeName,
									MountPath: manifestutils.TmpStoragePath,
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          manifestutils.OtlpGrpcPortName,
									ContainerPort: manifestutils.PortOtlpGrpcServer,
									Protocol:      corev1.ProtocolTCP,
								},
								{
									Name:          manifestutils.HttpPortName,
									ContainerPort: manifestutils.PortHTTPServer,
									Protocol:      corev1.ProtocolTCP,
								},
								{
									Name:          manifestutils.HttpMemberlistPortName,
									ContainerPort: manifestutils.PortMemberlist,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: manifestutils.TempoReadinessPath,
										Port: intstr.FromString(manifestutils.HttpPortName),
									},
								},
								InitialDelaySeconds: 15,
								TimeoutSeconds:      1,
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    *resource.NewMilliQuantity(270, resource.BinarySI),
									corev1.ResourceMemory: *resource.NewQuantity(257698032, resource.BinarySI),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    *resource.NewMilliQuantity(81, resource.BinarySI),
									corev1.ResourceMemory: *resource.NewQuantity(77309416, resource.BinarySI),
								},
							},
							SecurityContext: manifestutils.TempoContainerSecurityContext(),
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: manifestutils.ConfigVolumeName,
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "tempo-test",
									},
								},
							},
						},
						{
							Name: manifestutils.TmpStorageVolumeName,
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}, objects[0])

	assert.NoError(t, err)
	assert.Equal(t, &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tempo-test-distributor",
			Namespace: "project1",
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       manifestutils.OtlpGrpcPortName,
					Protocol:   corev1.ProtocolTCP,
					Port:       manifestutils.PortOtlpGrpcServer,
					TargetPort: intstr.FromString(manifestutils.OtlpGrpcPortName),
				},
				{
					Name:       manifestutils.HttpPortName,
					Protocol:   corev1.ProtocolTCP,
					Port:       manifestutils.PortHTTPServer,
					TargetPort: intstr.FromString(manifestutils.HttpPortName),
				},
			},
			Selector: labels,
		},
	}, objects[1])
}
