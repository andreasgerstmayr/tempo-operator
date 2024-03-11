package manifestutils

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/grafana/tempo-operator/apis/tempo/v1alpha1"
)

// MountCAConfigMap mounts the CA ConfigMap in a pod.
func MountCAConfigMap(
	pod *corev1.PodSpec,
	containerName string,
	caConfigMap string,
	caDir string,
) error {
	containerIdx, err := findContainerIndex(pod, containerName)
	if err != nil {
		return err
	}

	pod.Containers[containerIdx].VolumeMounts = append(pod.Containers[containerIdx].VolumeMounts, corev1.VolumeMount{
		Name:      caConfigMap,
		MountPath: caDir,
		ReadOnly:  true,
	})
	pod.Volumes = append(pod.Volumes, corev1.Volume{
		Name: caConfigMap,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: caConfigMap,
				},
			},
		},
	})

	return nil
}

// MountCertSecret mounts the Certificate Secret in a pod.
func MountCertSecret(
	pod *corev1.PodSpec,
	containerName string,
	certSecret string,
	certDir string,
) error {
	containerIdx, err := findContainerIndex(pod, containerName)
	if err != nil {
		return err
	}

	pod.Containers[containerIdx].VolumeMounts = append(pod.Containers[containerIdx].VolumeMounts, corev1.VolumeMount{
		Name:      certSecret,
		MountPath: certDir,
		ReadOnly:  true,
	})
	pod.Volumes = append(pod.Volumes, corev1.Volume{
		Name: certSecret,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: certSecret,
			},
		},
	})

	return nil
}

// MountTLSSpecVolumes mounts the CA ConfigMap and Certificate Secret in a pod.
func MountTLSSpecVolumes(
	pod *corev1.PodSpec,
	containerName string,
	tlsSpec v1alpha1.TLSSpec,
	caDir string,
	certDir string,
) error {
	if tlsSpec.CA != "" {
		err := MountCAConfigMap(pod, containerName, tlsSpec.CA, caDir)
		if err != nil {
			return err
		}
	}

	if tlsSpec.Cert != "" {
		err := MountCertSecret(pod, containerName, tlsSpec.Cert, certDir)
		if err != nil {
			return err
		}
	}

	return nil
}

func findContainerIndex(pod *corev1.PodSpec, containerName string) (int, error) {
	for i, container := range pod.Containers {
		if container.Name == containerName {
			return i, nil
		}
	}

	return -1, fmt.Errorf("cannot find container %s", containerName)
}
