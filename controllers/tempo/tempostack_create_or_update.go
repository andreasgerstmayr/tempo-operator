package controllers

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	grafanav1 "github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	routev1 "github.com/openshift/api/route/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/grafana/tempo-operator/apis/tempo/v1alpha1"
	"github.com/grafana/tempo-operator/internal/handlers/gateway"
	"github.com/grafana/tempo-operator/internal/manifests"
	"github.com/grafana/tempo-operator/internal/manifests/manifestutils"
	"github.com/grafana/tempo-operator/internal/status"
	"github.com/grafana/tempo-operator/internal/tlsprofile"
)

func listErrors(fieldErrs field.ErrorList) string {
	msgs := make([]string, len(fieldErrs))
	for i, fieldErr := range fieldErrs {
		msgs[i] = fieldErr.Detail
	}
	return strings.Join(msgs, ", ")
}

func (r *TempoStackReconciler) getStorageConfig(ctx context.Context, tempo v1alpha1.TempoStack) (manifestutils.StorageParams, error) {
	storageSecret := &corev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{Namespace: tempo.Namespace, Name: tempo.Spec.Storage.Secret.Name}, storageSecret)
	if err != nil {
		return manifestutils.StorageParams{}, fmt.Errorf("could not fetch storage secret: %w", err)
	}

	fieldErrs := v1alpha1.ValidateStorageSecret(tempo, *storageSecret)
	if len(fieldErrs) > 0 {
		return manifestutils.StorageParams{}, fmt.Errorf("invalid storage secret: %s", listErrors(fieldErrs))
	}

	if tempo.Spec.Storage.TLS.CA != "" {
		caConfigMap := &corev1.ConfigMap{}
		err := r.Get(ctx, types.NamespacedName{Namespace: tempo.Namespace, Name: tempo.Spec.Storage.TLS.CA}, caConfigMap)
		if err != nil {
			return manifestutils.StorageParams{}, fmt.Errorf("could not fetch CA config map: %w", err)
		}

		fieldErrs := v1alpha1.ValidateStorageCAConfigMap(*caConfigMap)
		if len(fieldErrs) > 0 {
			return manifestutils.StorageParams{}, fmt.Errorf("invalid CA config map: %s", listErrors(fieldErrs))
		}
	}

	params := manifestutils.StorageParams{}
	switch tempo.Spec.Storage.Secret.Type {
	case v1alpha1.ObjectStorageSecretAzure:
		params.AzureStorage = GetAzureParams(tempo, storageSecret)
	case v1alpha1.ObjectStorageSecretGCS:
		params.GCS = GetGCSParams(tempo, storageSecret)
	case v1alpha1.ObjectStorageSecretS3:
		params.S3 = GetS3Params(tempo, storageSecret)
	default:
		return manifestutils.StorageParams{}, fmt.Errorf("storage secret type is not recognized")
	}

	return params, nil
}

func (r *TempoStackReconciler) createOrUpdate(ctx context.Context, log logr.Logger, req ctrl.Request, tempo v1alpha1.TempoStack) error {
	storageConfig, err := r.getStorageConfig(ctx, tempo)
	if err != nil {
		return &status.ConfigurationError{
			Reason:  v1alpha1.ReasonInvalidStorageConfig,
			Message: err.Error(),
		}
	}

	if err = v1alpha1.ValidateTenantConfigs(tempo); err != nil {
		return &status.ConfigurationError{
			Message: fmt.Sprintf("Invalid tenants configuration: %s", err),
			Reason:  v1alpha1.ReasonInvalidTenantsConfiguration,
		}
	}

	if tempo.Spec.Tenants != nil && tempo.Spec.Tenants.Mode == v1alpha1.ModeOpenShift && r.CtrlConfig.Gates.OpenShift.BaseDomain == "" {
		domain, err := gateway.GetOpenShiftBaseDomain(ctx, r.Client)
		if err != nil {
			return err
		}
		log.Info("OpenShift base domain set", "openshift-base-domain", domain)
		r.CtrlConfig.Gates.OpenShift.BaseDomain = domain
	}

	tlsProfile, err := tlsprofile.Get(ctx, r.CtrlConfig.Gates, r.Client, log)
	if err != nil {
		switch err {
		case tlsprofile.ErrGetProfileFromCluster:
		case tlsprofile.ErrGetInvalidProfile:
			return &status.ConfigurationError{
				Message: err.Error(),
				Reason:  v1alpha1.ReasonCouldNotGetOpenShiftTLSPolicy,
			}
		default:
			return err
		}

	}

	var tenantSecrets []*manifestutils.GatewayTenantOIDCSecret
	if tempo.Spec.Tenants != nil && tempo.Spec.Tenants.Mode == v1alpha1.ModeStatic {
		tenantSecrets, err = gateway.GetOIDCTenantSecrets(ctx, r.Client, tempo)
		if err != nil {
			return err
		}
	}

	var gatewayTenantsData []*manifestutils.GatewayTenantsData
	if tempo.Spec.Tenants != nil && tempo.Spec.Tenants.Mode == v1alpha1.ModeOpenShift {
		gatewayTenantsData, err = gateway.GetGatewayTenantsData(ctx, r.Client, tempo)
		if err != nil {
			// just log the error the secret is not created if the loop for an instance runs for the first time.
			log.Info("Failed to get gateway secret and/or tenants.yaml", "error", err)
		}
	}

	managedObjects, err := manifests.BuildAll(manifestutils.Params{
		Tempo:               tempo,
		StorageParams:       storageConfig,
		CtrlConfig:          r.CtrlConfig,
		TLSProfile:          tlsProfile,
		GatewayTenantSecret: tenantSecrets,
		GatewayTenantsData:  gatewayTenantsData,
	})
	// TODO (pavolloffay) check error type and change return appropriately
	if err != nil {
		return fmt.Errorf("error building manifests: %w", err)
	}

	// Collect all objects owned by the operator, to be able to prune objects
	// which exist in the cluster but are not managed by the operator anymore.
	// For example, when the Jaeger Query Ingress is enabled and later disabled,
	// the Ingress object should be removed from the cluster.
	ownedObjects, err := r.findObjectsOwnedByTempoOperator(ctx, tempo)
	if err != nil {
		return err
	}

	err = reconcileManagedObjects(ctx, log, r.Client, &tempo, r.Scheme, managedObjects, ownedObjects)
	if err != nil {
		return err
	}

	return nil
}

func (r *TempoStackReconciler) findObjectsOwnedByTempoOperator(ctx context.Context, tempo v1alpha1.TempoStack) (map[types.UID]client.Object, error) {
	ownedObjects := map[types.UID]client.Object{}
	listOps := &client.ListOptions{
		Namespace:     tempo.GetNamespace(),
		LabelSelector: labels.SelectorFromSet(manifestutils.CommonLabels(tempo.Name)),
	}

	// Add all resources where the operator can conditionally create an object.
	// For example, Ingress and Route can be enabled or disabled in the CR.

	ingressList := &networkingv1.IngressList{}
	err := r.List(ctx, ingressList, listOps)
	if err != nil {
		return nil, fmt.Errorf("error listing ingress: %w", err)
	}
	for i := range ingressList.Items {
		ownedObjects[ingressList.Items[i].GetUID()] = &ingressList.Items[i]
	}

	if r.CtrlConfig.Gates.PrometheusOperator {
		servicemonitorList := &monitoringv1.ServiceMonitorList{}
		err := r.List(ctx, servicemonitorList, listOps)
		if err != nil {
			return nil, fmt.Errorf("error listing service monitors: %w", err)
		}
		for i := range servicemonitorList.Items {
			ownedObjects[servicemonitorList.Items[i].GetUID()] = servicemonitorList.Items[i]
		}

		prometheusRulesList := &monitoringv1.PrometheusRuleList{}
		err = r.List(ctx, prometheusRulesList, listOps)
		if err != nil {
			return nil, fmt.Errorf("error listing prometheus rules: %w", err)
		}
		for i := range prometheusRulesList.Items {
			ownedObjects[prometheusRulesList.Items[i].GetUID()] = prometheusRulesList.Items[i]
		}
	}

	if r.CtrlConfig.Gates.OpenShift.OpenShiftRoute {
		routesList := &routev1.RouteList{}
		err := r.List(ctx, routesList, listOps)
		if err != nil {
			return nil, fmt.Errorf("error listing routes: %w", err)
		}
		for i := range routesList.Items {
			ownedObjects[routesList.Items[i].GetUID()] = &routesList.Items[i]
		}
	}

	if r.CtrlConfig.Gates.GrafanaOperator {
		datasourceList := &grafanav1.GrafanaDatasourceList{}
		err := r.List(ctx, datasourceList, listOps)
		if err != nil {
			return nil, fmt.Errorf("error listing datasources: %w", err)
		}
		for i := range datasourceList.Items {
			ownedObjects[datasourceList.Items[i].GetUID()] = &datasourceList.Items[i]
		}
	}

	return ownedObjects, nil
}
