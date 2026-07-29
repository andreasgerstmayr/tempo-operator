package manifestutils

// CommonAnnotations returns common annotations for each pod created by the operator.
func CommonAnnotations(params Params) map[string]string {
	annotations := map[string]string{
		"tempo.grafana.com/config.hash": params.ConfigChecksum,
		"tempo.grafana.com/certs.hash":  params.CertsHash,
	}

	if params.StorageParams.SecretHash != "" {
		annotations["tempo.grafana.com/storage.secret.hash"] = params.StorageParams.SecretHash
	}
	if params.StorageParams.CloudCredentials.ContentHash != "" {
		annotations["tempo.grafana.com/token.cco.auth.hash"] = params.StorageParams.CloudCredentials.ContentHash
	}

	return annotations
}

// S3AWSSTSAnnotations returns service account annotations required by AWS STS.
func S3AWSSTSAnnotations(secret S3) map[string]string {
	return map[string]string{
		"eks.amazonaws.com/audience": "sts.amazonaws.com",
		"eks.amazonaws.com/role-arn": secret.RoleARN,
	}
}

// AzureShortLiveTokenAnnotation returns service account annotations required by Azure Short Live Token.
func AzureShortLiveTokenAnnotation(secret AzureStorage) map[string]string {
	return map[string]string{
		"azure.workload.identity/client-id": secret.ClientID,
		"azure.workload.identity/tenant-id": secret.TenantID,
	}
}
