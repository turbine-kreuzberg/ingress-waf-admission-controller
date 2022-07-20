# Web Application Firewall setup admission controller

This admission controller acts as a [MutatingAdmissionWebhook](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#mutatingadmissionwebhook) and adds [modsecurity waf](https://kubernetes.github.io/ingress-nginx/user-guide/third-party-addons/modsecurity/) to ingresses.


## Installation

1. install the [dependencies](#Dependencies)
2. download and verify [setup.yaml](setup.yaml)
3. deploy admission controller `kubectl apply -f setup.yaml`

## Dependencies

- [Cert Manager](https://cert-manager.io/docs/installation/helm/#installing-with-helm) is used to [set up certificates](https://cert-manager.io/docs/concepts/ca-injector/) to validate the webhook against the kubernetes control plane.


## Usage

The Admission controller adds WAF enabling annotations to all ingresses by default.

### Disable for a Ingress

Create an ingress and add the label `ingress-waf-enabled: "false"`.

``` yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  labels:
    ingress-waf-enabled: "false"
  name: some-ingress
spec:
  rules:
    - host: "domain.tld"
```

Ingresses with this label will skip the WAF setup.


## local development

1. install [tilt](https://docs.tilt.dev/install.html), [helm](https://helm.sh/docs/intro/install/#from-script), [helmfile](https://github.com/roboll/helmfile#installation), [helm diff](https://github.com/databus23/helm-diff#using-helm-plugin-manager--23x), and [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
2. setup [kind with local registry](https://github.com/tilt-dev/kind-local#how-to-try-it)
3. deploy dependencies `helmfile sync`
4. start environment `tilt up`
