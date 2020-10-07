package controllers

import (
	nsmv1alpha1 "github.com/networkservicemesh/nsm-operator/apis/nsm/v1alpha1"
	admissionregv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *NSMReconciler) mutatingConfigForWebhook(nsm *nsmv1alpha1.NSM) *admissionregv1.MutatingWebhookConfiguration {

	var path string
	path = "/mutate"

	annotations := map[string]string{}
	if r.isPlatformOpenShift() {
		annotations = map[string]string{"service.beta.openshift.io/inject-cabundle": "true"}
	}

	mutatingConfig := &admissionregv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: webhookMutatingConfigName,
			// TODO: Solve TLS Certs for OCP - This annotation below is specific to OpenShift and needs to be addressed other way
			// Serving-ca operator injects the secret automatically created.
			Annotations: annotations,
		},
		Webhooks: []admissionregv1.MutatingWebhook{
			{
				Name: "admission-webhook.networkservicemesh.io",
				ClientConfig: admissionregv1.WebhookClientConfig{
					Service: &admissionregv1.ServiceReference{
						Name:      webhookServiceName,
						Namespace: nsm.Namespace,
						Path:      &path,
					},
					CABundle: caCert,
				},

				Rules: []admissionregv1.RuleWithOperations{
					{
						Operations: []admissionregv1.OperationType{admissionregv1.Create},
						Rule: admissionregv1.Rule{
							APIGroups:   []string{"apps", "extensions", ""},
							APIVersions: []string{"v1", "v1beta1"},
							Resources:   []string{"pods", "deployments", "services"},
						},
					},
				},
			},
		},
	}
	// Set NSM instance as the owner and controller
	controllerutil.SetControllerReference(nsm, mutatingConfig, r.Scheme)
	return mutatingConfig
}
