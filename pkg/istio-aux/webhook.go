package istioaux

import (
	"context"
	"encoding/json"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,failurePolicy=fail,groups="",resources=pods,verbs=create;update,versions=v1,name=mpod.kb.io,sideEffects=None,admissionReviewVersions=v1

type PodMutator struct {
	Client  client.Client
	decoder *admission.Decoder
}

func (a *PodMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	logger := ctrl.Log.WithName("webhook")
	pod := &corev1.Pod{}

	err := (*a.decoder).Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	/* not checking for the target namespace existense and required labels
	as they can be configured via a namespaceSelector in the MutatingWebhookConfiguration:
	  namespaceSelector:
		matchExpressions:
		- key: io.datastrophic/istio-aux
	      operator: In
		  values: ["enabled"]
		- key: istio-injection
	      operator: In
		  values: ["enabled"]
	*/

	logger.Info("processing", "pod-generate-name", pod.GenerateName, "pod-name", pod.ObjectMeta.Name)
	SetMetadata(&pod.ObjectMeta)
	logger.Info("processed", "pod-generate-name", pod.GenerateName, "pod-name", pod.ObjectMeta.Name)

	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

func (a *PodMutator) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}
