package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	networkingv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type networkHealthSidecarInjector struct {
	Client  client.Client
	decoder *admission.Decoder
}

func (a *networkHealthSidecarInjector) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}

func (a *networkHealthSidecarInjector) Handle(ctx context.Context, req admission.Request) admission.Response {
	// unmarshal
	ingress := &networkingv1.Ingress{}
	err := a.decoder.Decode(req, ingress)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// process
	prepare(ingress)

	enableWAF(ingress)

	// marshal
	marshaledIngress, err := json.Marshal(ingress)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledIngress)
}

func prepare(ingress *networkingv1.Ingress) {
	if ingress.Labels == nil {
		ingress.Labels = map[string]string{}
	}

	if ingress.Annotations == nil {
		ingress.Annotations = map[string]string{}
	}

	_, found := ingress.Labels["ingress-waf-enabled"]
	if !found {
		ingress.Labels["ingress-waf-enabled"] = "true"
	}
}

func enableWAF(ingress *networkingv1.Ingress) {
	if strings.ToLower(ingress.Labels["ingress-waf-enabled"]) != "true" {
		return
	}

	ingress.Annotations["nginx.ingress.kubernetes.io/enable-modsecurity"] = "true"
	ingress.Annotations["nginx.ingress.kubernetes.io/enable-owasp-core-rules"] = "true"
	ingress.Annotations["nginx.ingress.kubernetes.io/modsecurity-snippet"] = "SecRuleEngine On"
	ingress.Annotations["nginx.ingress.kubernetes.io/modsecurity-transaction-id"] = "$request_id"
}
