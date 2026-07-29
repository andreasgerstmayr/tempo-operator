package manifestutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAzureShortLiveTokenAnnotation(t *testing.T) {
	annotations := AzureShortLiveTokenAnnotation(AzureStorage{
		TenantID: "test-tenant",
		ClientID: "test-client",
	})

	assert.Equal(t, "test-client", annotations["azure.workload.identity/client-id"])
	assert.Equal(t, "test-tenant", annotations["azure.workload.identity/tenant-id"])
}

func TestCommonAnnotations(t *testing.T) {
	result := CommonAnnotations(Params{
		ConfigChecksum: "abc123",
		CertsHash:      "2025-11-17T10:00:00Z",
	})

	assert.Equal(t, map[string]string{
		"tempo.grafana.com/config.hash":  "abc123",
		"tempo.grafana.com/certs.hash": "2025-11-17T10:00:00Z",
	}, result)
}
