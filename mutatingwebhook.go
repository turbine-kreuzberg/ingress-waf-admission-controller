package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	networkingv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	AnnotationEnabled           = "ingress-waf/enabled"
	AnnotationRequestBodyLimit  = "ingress-waf/request-body-limit"
	AnnotationResponseBodyLimit = "ingress-waf/response-body-limit"
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

	err = enableWAF(ingress)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// marshal
	marshaledIngress, err := json.Marshal(ingress)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledIngress)
}

func prepare(ingress *networkingv1.Ingress) {
	if ingress.Annotations == nil {
		ingress.Annotations = map[string]string{}
	}

	_, found := ingress.Annotations[AnnotationEnabled]
	if !found {
		ingress.Annotations[AnnotationEnabled] = "true"
	}
}

func enableWAF(ingress *networkingv1.Ingress) error {
	if strings.ToLower(ingress.Annotations[AnnotationEnabled]) != "true" {
		return nil
	}

	ingress.Annotations["nginx.ingress.kubernetes.io/enable-modsecurity"] = "true"
	ingress.Annotations["nginx.ingress.kubernetes.io/enable-owasp-core-rules"] = "true"
	ingress.Annotations["nginx.ingress.kubernetes.io/modsecurity-snippet"] = `
SecRuleEngine On
SecAuditEngine RelevantOnly
SecAuditLog /dev/stdout
SecAuditLogParts ABCEFHJZ
SecAction "id:900110,phase:1,log,pass,t:none,setvar:tx.inbound_anomaly_score_threshold=10"
SecAction "phase:5,auditlog,log,pass,msg:\'Anomaly Score %{TX.anomaly_score} Threshold %{TX.inbound_anomaly_score_threshold}\'"
`

	requestBodyLimit, found := ingress.Annotations[AnnotationRequestBodyLimit]
	if found {
		requestBodyLimit, err := strconv.Atoi(requestBodyLimit)
		if err != nil {
			return fmt.Errorf("%s is not a number (bytes): %v", AnnotationRequestBodyLimit, err)
		}

		ingress.Annotations["nginx.ingress.kubernetes.io/modsecurity-snippet"] +=
			fmt.Sprintf("SecRequestBodyLimit %d\n", requestBodyLimit)
	}

	responseBodyLimit, found := ingress.Annotations[AnnotationResponseBodyLimit]
	if found {
		responseBodyLimit, err := strconv.Atoi(responseBodyLimit)
		if err != nil {
			return fmt.Errorf("%s is not a number (bytes): %v", AnnotationResponseBodyLimit, err)
		}

		ingress.Annotations["nginx.ingress.kubernetes.io/modsecurity-snippet"] +=
			fmt.Sprintf("SecResponseBodyLimit %d\n", responseBodyLimit)
	}

	ingress.Annotations["nginx.ingress.kubernetes.io/modsecurity-transaction-id"] = "$request_id"

	return nil
}
