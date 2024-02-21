package monolithic

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/grafana/tempo-operator/apis/tempo/v1alpha1"
	"github.com/grafana/tempo-operator/internal/manifests/manifestutils"
	"github.com/grafana/tempo-operator/internal/manifests/naming"
)

// BuildTempoService creates the service for a monolithic deployment.
func BuildTempoService(opts Options) *corev1.Service {
	tempo := opts.Tempo
	labels := ComponentLabels(manifestutils.TempoMonolithComponentName, tempo.Name)
	annotations := map[string]string{}

	if opts.CtrlConfig.Gates.OpenShift.ServingCertsService {
		annotations["service.beta.openshift.io/serving-cert-secret-name"] = naming.ServingCertName(manifestutils.TempoMonolithComponentName, tempo.Name)
	}

	var ports []corev1.ServicePort
	if tempo.Spec.Multitenancy.IsGatewayEnabled() {
		ports = gatewayPorts(tempo)
	} else {
		ports = tempoPorts(tempo)
	}

	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        naming.Name(manifestutils.TempoMonolithComponentName, tempo.Name),
			Namespace:   tempo.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: corev1.ServiceSpec{
			Ports:    ports,
			Selector: labels,
		},
	}
}

func tempoPorts(tempo v1alpha1.TempoMonolithic) []corev1.ServicePort {
	ports := []corev1.ServicePort{{
		Name:       manifestutils.HttpPortName,
		Protocol:   corev1.ProtocolTCP,
		Port:       manifestutils.PortHTTPServer,
		TargetPort: intstr.FromString(manifestutils.HttpPortName),
	}}

	if tempo.Spec.Ingestion != nil && tempo.Spec.Ingestion.OTLP != nil {
		if tempo.Spec.Ingestion.OTLP.GRPC != nil && tempo.Spec.Ingestion.OTLP.GRPC.Enabled {
			ports = append(ports, corev1.ServicePort{
				Name:       manifestutils.OtlpGrpcPortName,
				Protocol:   corev1.ProtocolTCP,
				Port:       manifestutils.PortOtlpGrpcServer,
				TargetPort: intstr.FromString(manifestutils.OtlpGrpcPortName),
			})
		}
		if tempo.Spec.Ingestion.OTLP.HTTP != nil && tempo.Spec.Ingestion.OTLP.HTTP.Enabled {
			ports = append(ports, corev1.ServicePort{
				Name:       manifestutils.PortOtlpHttpName,
				Protocol:   corev1.ProtocolTCP,
				Port:       manifestutils.PortOtlpHttp,
				TargetPort: intstr.FromString(manifestutils.PortOtlpHttpName),
			})
		}
	}

	if tempo.Spec.JaegerUI != nil && tempo.Spec.JaegerUI.Enabled {
		ports = append(ports, []corev1.ServicePort{
			{
				Name:       manifestutils.JaegerGRPCQuery,
				Protocol:   corev1.ProtocolTCP,
				Port:       manifestutils.PortJaegerGRPCQuery,
				TargetPort: intstr.FromString(manifestutils.JaegerGRPCQuery),
			},
			{
				Name:       manifestutils.JaegerUIPortName,
				Protocol:   corev1.ProtocolTCP,
				Port:       manifestutils.PortJaegerUI,
				TargetPort: intstr.FromString(manifestutils.JaegerUIPortName),
			},
			{
				Name:       manifestutils.JaegerMetricsPortName,
				Protocol:   corev1.ProtocolTCP,
				Port:       manifestutils.PortJaegerMetrics,
				TargetPort: intstr.FromString(manifestutils.JaegerMetricsPortName),
			},
		}...)
	}

	return ports
}

func gatewayPorts(tempo v1alpha1.TempoMonolithic) []corev1.ServicePort {
	// use the same port name and numbers like without gateway, to make the gateway a drop-in feature

	ports := []corev1.ServicePort{{
		Name:     manifestutils.HttpPortName,
		Protocol: corev1.ProtocolTCP,
		Port:     manifestutils.PortHTTPServer,
		// proxies Tempo API and optionally Jaeger UI
		TargetPort: intstr.FromString(manifestutils.GatewayHttpPortName),
	}}

	if tempo.Spec.Ingestion != nil && tempo.Spec.Ingestion.OTLP != nil &&
		tempo.Spec.Ingestion.OTLP.GRPC != nil && tempo.Spec.Ingestion.OTLP.GRPC.Enabled {
		ports = append(ports, corev1.ServicePort{
			Name:       manifestutils.OtlpGrpcPortName,
			Protocol:   corev1.ProtocolTCP,
			Port:       manifestutils.PortOtlpGrpcServer,
			TargetPort: intstr.FromString(manifestutils.GatewayGrpcPortName),
		})
	}

	if tempo.Spec.JaegerUI != nil && tempo.Spec.JaegerUI.Enabled {
		ports = append(ports, []corev1.ServicePort{
			{
				Name:     manifestutils.JaegerUIPortName,
				Protocol: corev1.ProtocolTCP,
				Port:     manifestutils.PortJaegerUI,
				// proxies Tempo API and optionally Jaeger UI
				TargetPort: intstr.FromString(manifestutils.GatewayHttpPortName),
			},
		}...)
	}

	return ports
}
