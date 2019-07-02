[![CircleCI](https://circleci.com/gh/istio/operator.svg?style=svg)](https://circleci.com/gh/istio/operator)
[![Mergify Status](https://gh.mergify.io/badges/istio/operator.png?style=cut)](https://mergify.io)
[![Go Report Card](https://goreportcard.com/badge/github.com/istio/operator)](https://goreportcard.com/report/github.com/istio/operator)
[![GolangCI](https://golangci.com/badges/github.com/istio/operator.svg)](https://golangci.com/r/github.com/istio/operator)

# Istio Operator

## Introduction

The Istio operator CLI is now stable enough for developers to evaluate and experiment with. We welcome your
contributions - pick an [unassigned open issue](https://github.com/istio/istio/issues?q=is%3Aissue+is%3Aopen+label%3Aarea%2Fenvironments%2Foperator+no%3Aassignee),
create a feature proposal, or come to the weekly Environments Working Group meeting if you have ideas.

-  [Bugs and feature requests](https://github.com/istio/operator/blob/master/BUGS-AND-FEATURE-REQUESTS.md)
-  [Contributing guidelines](https://github.com/istio/operator/blob/master/CONTRIBUTING.md)
-  [Working groups](https://github.com/istio/community/blob/master/WORKING-GROUPS.md)

### Background

We reorganized the current [helm installation parameters](https://istio.io/docs/reference/config/installation-options/) into two groups:

-  The new [platform level installation API](https://github.com/istio/operator/blob/master/pkg/apis/istio/v1alpha2/istiocontrolplane_types.proto), for managing
k8s settings like resources, auto scaling, pod disruption budgets etc. 
-  The configuration API that currently uses the
[helm installation parameters](https://istio.io/docs/reference/config/installation-options/) for backwards
compatibility. This API is for managing This API the Istio control plane configuration settings.

Some parameters will temporarily exist in both APIs - for example, setting k8s resources currently can be done through
either API above. However, the Istio community recommends using the first API as it is more consistent, is validated,
and will naturally follow the graduation process for APIs while the same parameters in the configuration API are planned
for deprecation.

We currently provide pre-configured helm values sets for different scenarios as configuration 
[profiles](https://istio.io/docs/setup/kubernetes/additional-setup/config-profiles/), which act as a starting point for
an Istio install and can be customized by creating customization overlay files or passing parameters when
calling helm. Similarly, the operator API uses the same profiles (expressed internally through the new API), which can be selected
as a starting point for the installation. For comparison, the following example shows the command needed to install
Istio using the SDS configuration profile using Helm:

```bash
helm template install/kubernetes/helm/istio --name istio --namespace istio-system \
    --values install/kubernetes/helm/istio/values-istio-sds-auth.yaml | kubectl apply -f -
```

In the new API, the same profile would be selected through a CustomResource (CR):

```yaml
# sds-install-cr.yaml

apiVersion: install.istio.io/v1alpha1
kind: IstioControlPlane
metadata:
  name: istio-operator-config
  namespace: istio-system
spec:
  profile: sds
```

CRs are used when running the operator as a controller in a pod in the cluster. When using operator CLI mode and passing the 
configuration as a file (see [Select a profile](#Select_a_profile)), only the spec portion is required.

```yaml
# sds-install.yaml

profile: sds
```

If you don't specify a configuration profile, Istio is installed using the `default` configuration profile. All 
profiles listed in istio.io are available compiled in, or `profile:` can point to a local file path to reference a custom
profile base to use as a starting point for customization. See the [API reference](https://github.com/istio/operator/blob/master/pkg/apis/istio/v1alpha2/istiocontrolplane_types.proto)
for details.

## Developer quick start

The quick start describes how to install and use the operator `iop` CLI command. 

#### Installation 

```bash
git clone https://istio.io/operator.git
cd operator
go build -o <your bin path> ./cmd/iop.go
```

#### Flags

The `iop` command supports the following flags:

+   `logtostderr`: log to console (by default logs go to ./iop.go).
+   `dry-run`: console output only, nothing applied to cluster or written to files (default is true for now).
+   `verbose`: display entire manifest contents and other debug info (default is false).

#### Basic default manifest

The following command generates a manifest with the compiled in default profile and charts:

```bash
iop manifest
```

You can see these sources for the compiled in profiles in the repo under data/profiles, while the compiled in helm
charts are under data/charts.

#### Output to dirs

The output of the manifest is concatenated into a single file. To generate a directory hierarchy with subdirectory
levels representing a child dependency, use the following command:

```bash
iop manifest -o istio_manifests
```

Use depth first search to traverse the created directory hierarchy when applying your YAML files. Child manifest
directories must wait for the parent, but not sibling manifest directories.

#### Just apply it for me

The following command generates the manifests and applies them in the correct dependency order, waiting for the
dependencies to have the needed CRDs available: 

```bash
iop install
```

Note: right now the configuration that would be applied is only displayed, since `dry-run` is true by default. Set
`--dry-run=false` to actually apply the generated configuration to the cluster.

#### Review the values of the current configuration profile

The following command shows the values of the current configuration profile:

```bash
iop dump-profile
```

#### Select a specific configuration profile

The simplest customization is to select a profile different to `default` e.g. `sds`. Create the following config file: 

```yaml
# sds-install.yaml

profile: sds
```

Use the Istio operator `iop` binary to apply the new configuration profile:

```bash
iop manifest -f sds-install.yaml
```

After running the command, the Helm charts are rendered using `data/profiles/sds.yaml`. 

#### Install from file path

The compiled in charts and profiles are used by default, but a file path can be specified, e.g.

```yaml
profile: file:///usr/home/bob/go/src/github.com/ostromart/istio-installer/data/profiles/default.yaml
customPackagePath: file:///usr/home/bob/go/src/github.com/ostromart/istio-installer/data/charts/
```

These can be mixed and matched e.g. use a compiled in profile with local filesystem charts. 

#### New API customization

The [new platform level installation API](https://github.com/istio/operator/blob/95e89c4fb838b0b374f70d7c5814329e25a64819/pkg/apis/istio/v1alpha1/istioinstaller_types.proto#L25) 
defines install time parameters like feature/component enablement and namespace, and k8s settings like resources, HPA spec etc. in a structured way.   
The simplest customization is to turn features and components on and off. For example, to turn off all policy:

```yaml
profile: sds
policy:
  enabled: false
```

Note that unlike helm, all configurations are validated against a schema, so the operator detects syntax errors. Another
customization is to define custom namespaces for features:

```yaml
profile: sds

trafficManagement:
  namespace: istio-control-custom 
```

The traffic management feature comprises Pilot, AutoInjection and Proxy components. Each of these components has k8s
settings, and these can be overridden from the defaults using official k8s APIs (rather than Istio defined schemas):

```yaml
trafficManagement:
  components:
    pilot:
      k8s:
        resources:
          requests:
            cpu: 1000m # override from default 500m
            memory: 4096Mi # ... default 2048Mi
        hpaSpec:
          maxReplicas: 10 # ... default 5
          minReplicas: 2  # ... default 1 
```

The k8s settings are defined in detail in the [operator API](https://github.com/istio/operator/blob/95e89c4fb838b0b374f70d7c5814329e25a64819/pkg/apis/istio/v1alpha1/istioinstaller_types.proto#L394). The settings are the same for all components, so a user can configure pilot k8s settings in exactly the same, consistent way as galley settings. Supported k8s settings currently include:

+   [resources](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#resource-requests-and-limits-of-pod-and-container)
+   [readiness probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/)
+   [replica count](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
+   [HoriizontalPodAutoscaler](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/)
+   [PodDisruptionBudget](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#how-disruption-budgets-work)
+   [pod annotations](https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/)
+   [service annotations](https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/)
+   [ImagePullPolicy](https://kubernetes.io/docs/concepts/containers/images/)
+   [priority calss name](https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/#priorityclass)
+   [node selector](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector)
+   [affinity and anti-affinity](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#affinity-and-anti-affinity)

All of these k8s settings use the k8s API definitions, so k8s documentation can be used for reference. All k8s overlay
values are also validated in the operator.

#### Customizing the old values.yaml API 

The new platform install API above deals with k8s level settings. The remaining values.yaml parameters deal with Istio
control plane operation rather than platform installation and for the time being the operator just passes these through
to the helm charts unmodified (but validated through a
[schema](https://github.com/istio/operator/blob/master/pkg/apis/istio/v1alpha2/values_types.go)). Values.yaml settings
are overridden the same way as the new API, though a customized CR overlaid over default values for the selected
profile. Here's an example of overriding some global level default values:

```yaml
profile: sds
values:
  global:
    logging:
      level: "default:warning" # override from info
```

Since from 1.3 helm charts are split up per component, values overrides should be specified under the appropriate component e.g. 

```yaml
trafficManagement:
  components:
    pilot:
      values: 
        traceSampling: 0.1 # override from 1.0
```

#### Advanced k8s resource overlays

Advanced users may occasionally have the need to customize parameters (like container command line flags) which are not
exposed through either of the installation or configuration APIs described in this document. For such cases, it's
possible to overlay the generated k8s resources before they are applied with user-defined overlays. For example, to
override some container level values in the Pilot container:

```yaml
trafficManagement:
  enabled: true
  components:
    proxy:
      common:
        enabled: false
    pilot:
      common:
        k8s:
          overlays:
          - kind: Deployment
            name: istio-pilot
            patches:
            - path: spec.template.spec.containers.[name:discovery].args.[30m]
              value: "60m" # OVERRIDDEN
            - path: spec.template.spec.containers.[name:discovery].ports.[containerPort:8080].containerPort
              value: 8090 # OVERRIDDEN
          - kind: Service
            name: istio-pilot
            patches:
            - path: spec.ports.[name:grpc-xds].port
              value: 15099 # OVERRIDDEN
```

The user-defined overlay uses a path spec that includes the ability to select list items by key. In the example above,
the container with the key-value "name: discovery" is selected from the list of containers, and the command line
parameter with value "30m" is selected to be modified. The advanced overlay capability is described in more detail in
the spec.

#### Try the demo customization

This customization contains overlays at all three levels: the new API, values.yaml legacy API and the k8s output overlay. 

```bash
iop manifest -f samples/customize_pilot.yaml
```

## Architecture 

WIP, coming soon.
