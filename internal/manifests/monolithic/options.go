package monolithic

import (
	configv1alpha1 "github.com/grafana/tempo-operator/api/config/v1alpha1"
	"github.com/grafana/tempo-operator/api/tempo/v1alpha1"
	"github.com/grafana/tempo-operator/internal/manifests/manifestutils"
	"github.com/grafana/tempo-operator/internal/tlsprofile"
)

// Options defines calculated options required to generate all manifests.
type Options struct {
	CtrlConfig                configv1alpha1.ProjectConfig
	Tempo                     v1alpha1.TempoMonolithic
	StorageParams             manifestutils.StorageParams
	ConfigChecksum            string
	CertsHash                 string
	GatewayTenantSecret       []*manifestutils.GatewayTenantOIDCSecret
	GatewayTenantsData        []*manifestutils.GatewayTenantsData
	TLSProfile                tlsprofile.TLSProfileOptions
	useServiceCertsOnReceiver bool
}

// CommonAnnotations returns common annotations for each pod created by the operator.
func CommonAnnotations(opts Options) map[string]string {
	annotations := map[string]string{
		"tempo.grafana.com/config.hash": opts.ConfigChecksum,
		"tempo.grafana.com/certs.hash":  opts.CertsHash,
	}

	if opts.StorageParams.SecretHash != "" {
		annotations["tempo.grafana.com/storage.secret.hash"] = opts.StorageParams.SecretHash
	}
	if opts.StorageParams.CloudCredentials.ContentHash != "" {
		annotations["tempo.grafana.com/token.cco.auth.hash"] = opts.StorageParams.CloudCredentials.ContentHash
	}

	return annotations
}
