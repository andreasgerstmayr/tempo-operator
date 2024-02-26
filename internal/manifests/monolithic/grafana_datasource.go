package monolithic

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/grafana/tempo-operator/internal/manifests/grafana"
	"github.com/grafana/tempo-operator/internal/manifests/manifestutils"
	"github.com/grafana/tempo-operator/internal/manifests/naming"
)

// BuildGrafanaDatasources creates Grafana data sources.
func BuildGrafanaDatasources(opts Options) ([]client.Object, error) {
	tempo := opts.Tempo
	labels := ComponentLabels(manifestutils.TempoMonolithComponentName, tempo.Name)
	instanceSelector := ptr.Deref(tempo.Spec.Observability.Grafana.DataSource.InstanceSelector, metav1.LabelSelector{})

	if tempo.Spec.Multitenancy.IsGatewayEnabled() {
		scheme := "http"
		caName := ""

		if opts.CtrlConfig.Gates.OpenShift.ServingCertsService {
			scheme = "https"
			caName = naming.ServingCABundleName(tempo.Name)
		}

		datasources := []client.Object{}
		for _, tenant := range tempo.Spec.Multitenancy.TenantsSpec.Authentication {
			url := fmt.Sprintf(
				"%s://%s:%d/api/traces/v1/%s/tempo",
				scheme,
				naming.ServiceFqdn(tempo.Namespace, tempo.Name, manifestutils.TempoMonolithComponentName),
				manifestutils.PortHTTPServer,
				tenant.TenantName,
			)

			ds, err := grafana.NewGrafanaDatasourceForTenant(tempo.Namespace, tempo.Name, labels, tenant.TenantName, url, caName, instanceSelector)
			if err != nil {
				return nil, err
			}

			datasources = append(datasources, ds)
		}
		return datasources, nil
	} else {
		url := fmt.Sprintf("http://%s:%d", naming.ServiceFqdn(tempo.Namespace, tempo.Name, manifestutils.TempoMonolithComponentName), manifestutils.PortHTTPServer)
		datasource := grafana.NewGrafanaDatasource(tempo.Namespace, tempo.Name, labels, url, instanceSelector)
		return []client.Object{datasource}, nil
	}
}
