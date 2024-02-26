package grafana

import (
	"encoding/json"
	"fmt"

	grafanav1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/ptr"

	"github.com/grafana/tempo-operator/internal/manifests/manifestutils"
	"github.com/grafana/tempo-operator/internal/manifests/naming"
)

// BuildGrafanaDatasource creates a data source for Grafana Tempo.
func BuildGrafanaDatasource(params manifestutils.Params) *grafanav1.GrafanaDatasource {
	tempo := params.Tempo
	labels := manifestutils.CommonLabels(tempo.Name)
	instanceSelector := tempo.Spec.Observability.Grafana.InstanceSelector
	var url string

	if tempo.Spec.Template.Gateway.Enabled {
		url = fmt.Sprintf("http://%s:%d", naming.ServiceFqdn(tempo.Namespace, tempo.Name, manifestutils.GatewayComponentName), manifestutils.PortHTTPServer)
	} else {
		url = fmt.Sprintf("http://%s:%d", naming.ServiceFqdn(tempo.Namespace, tempo.Name, manifestutils.QueryFrontendComponentName), manifestutils.PortHTTPServer)
	}

	return NewGrafanaDatasource(tempo.Namespace, tempo.Name, labels, url, instanceSelector)
}

// NewGrafanaDatasource creates a data source for Grafana Tempo.
func NewGrafanaDatasource(namespace string, name string, labels labels.Set, url string, instanceSelector metav1.LabelSelector) *grafanav1.GrafanaDatasource {
	return &grafanav1.GrafanaDatasource{
		TypeMeta: metav1.TypeMeta{
			APIVersion: grafanav1.GroupVersion.String(),
			Kind:       "GrafanaDatasource",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Labels:    labels,
		},
		Spec: grafanav1.GrafanaDatasourceSpec{
			Datasource: &grafanav1.GrafanaDatasourceInternal{
				Name:   name,
				Type:   "tempo",
				Access: "proxy",
				URL:    url,
			},

			// InstanceSelector is a required field in the spec
			InstanceSelector: &instanceSelector,

			// Allow using this datasource from Grafana instances in other namespaces
			AllowCrossNamespaceImport: ptr.To(true),
		},
	}
}

// reference: https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources
type dataSpec struct {
	TLSAuthWithCACert bool   `json:"tlsAuthWithCACert,omitempty"`
	HttpHeaderName1   string `json:"httpHeaderName1,omitempty"`
}

type secureDataSpec struct {
	TLSCACert        string `json:"tlsCACert,omitempty"`
	HttpHeaderValue1 string `json:"httpHeaderValue1,omitempty"`
}

// NewGrafanaDatasource creates a data source for a specified tenant.
func NewGrafanaDatasourceForTenant(
	namespace string,
	name string,
	labels labels.Set,
	tenantName string,
	url string,
	customCA string,
	instanceSelector metav1.LabelSelector,
) (*grafanav1.GrafanaDatasource, error) {
	valuesFrom := []grafanav1.GrafanaDatasourceValueFrom{}
	data := dataSpec{}
	secureData := secureDataSpec{}

	if customCA != "" {
		data.TLSAuthWithCACert = true
		data.HttpHeaderName1 = "Authorization"
		// using Grafana's variable expansion file provider to mount the service account token of the Grafana service account
		// https://grafana.com/docs/grafana/latest/setup-grafana/configure-grafana/#file-provider
		secureData.HttpHeaderValue1 = fmt.Sprintf("Bearer $__file{%s}", manifestutils.BearerTokenFile)

		secureData.TLSCACert = fmt.Sprintf("${%s}", manifestutils.TLSCAFilename)
		valuesFrom = append(valuesFrom, grafanav1.GrafanaDatasourceValueFrom{
			TargetPath: "secureJsonData.tlsCACert",
			ValueFrom: grafanav1.GrafanaDatasourceValueFromSource{
				ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: customCA,
					},
					Key: manifestutils.TLSCAFilename,
				},
			},
		})
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	secureJsonData, err := json.Marshal(secureData)
	if err != nil {
		return nil, err
	}

	return &grafanav1.GrafanaDatasource{
		TypeMeta: metav1.TypeMeta{
			APIVersion: grafanav1.GroupVersion.String(),
			Kind:       "GrafanaDatasource",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      fmt.Sprintf("%s-%s", name, tenantName),
			Labels:    labels,
		},
		Spec: grafanav1.GrafanaDatasourceSpec{
			Datasource: &grafanav1.GrafanaDatasourceInternal{
				Name:           fmt.Sprintf("%s (%s)", name, tenantName),
				Type:           "tempo",
				Access:         "proxy",
				URL:            url,
				JSONData:       jsonData,
				SecureJSONData: secureJsonData,
			},
			ValuesFrom: valuesFrom,

			// InstanceSelector is a required field in the spec
			InstanceSelector: &instanceSelector,

			// Allow using this datasource from Grafana instances in other namespaces
			AllowCrossNamespaceImport: ptr.To(true),
		},
	}, nil
}
