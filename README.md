[![CircleCI](https://circleci.com/gh/istio/operator.svg?style=svg)](https://circleci.com/gh/istio/operator)
[![Mergify Status](https://gh.mergify.io/badges/istio/operator.png?style=cut)](https://mergify.io)
[![Go Report Card](https://goreportcard.com/badge/github.com/istio/operator)](https://goreportcard.com/report/github.com/istio/operator)
[![GolangCI](https://golangci.com/badges/github.com/istio/operator.svg)](https://golangci.com/r/github.com/istio/operator)

# Istio Operator

Istio operator is not yet ready for users to try out (it will be shortly, stay tuned on discuss.istio.io). 
However, the code is stable enough for developers to evaluate and experiment with. Contributions are very welcome - 
see [open issues](), create a feature proposal, or come to the weekly Environments Working Group meeting if you have ideas. 

### Background

The current [helm installation parameters](https://istio.io/docs/reference/config/installation-options/) have been reorganized into two groups:

1.  A new [platform level installation API](https://github.com/istio/operator/blob/master/pkg/apis/istio/v1alpha2/istiocontrolplane_types.proto), dealing with k8s settings like resources, auto scaling, pod disruption budgets etc. 
1.  A configuration API, dealing with Istio control plane configuration settings. This API currently uses the [helm installation parameters](https://istio.io/docs/reference/config/installation-options/) for backwards compatibility, but will be reorganized and slimmed down in the future.

Some parameters will therefore temporarily be in both APIs - for example, setting k8s resources can be done through either API above. However, it's recommended to use the first API, since this has a more consistent structure and it is the one that will remain going forward.  
Typical helm values parameter sets are currently provided as [profiles](https://istio.io/docs/setup/kubernetes/additional-setup/config-profiles/), which act as a starting point for an Istio install, and can currently be customized by editing the values files or passing parameters when calling helm.  
Similarly, the operator API uses the same profiles, which can be selected as a starting point for the installation.   
For comparison, here is an installation of the SDS profile using helm:

```bash
helm template install/kubernetes/helm/istio --name istio --namespace istio-system \
    --values install/kubernetes/helm/istio/values-istio-sds-auth.yaml | kubectl apply -f -
```

The same installation would be expressed as a CustomResource (CR) through the new API as:

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

The profile can be installed into the cluster by installing the operator into a pod into the cluster and applying the above CR, or applying it via the CLI mode. When using CLI mode, it's not necessary to specify the entire CR, only the spec portion i.e.

```yaml
# sds-install.yaml

profile: sds
```

In the absence of a profile, the default demo profile is installed. All profiles are available compiled in, or the profile can point to a local file path.

## Developer quick start

#### Installation 

The latest, not fully reviewed code is in github.com/ostromart/istio-installer. PRs into the official repo at istio.io/operator are lagging by a few days. 

```bash
git clone https://istio.io/operator.git
cd istio-installer
go build -o <your bin path> ./cmd/iop.go
```

#### Flags

The iop command supports the following flags:

+   logtostderr: logs to console (by default logs go to ./iop.go)
+   dry-run (default is true for now): console output only, nothing applied to cluster or written to files
+   verbose: display entire manifest contents and other debug info

#### Basic default manifest

```bash
iop manifest
```

This generates a manifest with all compiled in defaults (charts and base profile). You can see these sources in the repo under data/[profiles|charts]. 

#### Output to dirs

The above output is concatenated into a single file. To generate a directory hierarchy, with subdirectory levels representing a child dependency, use:

```bash
iop manifest -o istio_manifests
```

This will create a dir hierarchy which should be traversed DFS when applying the yamls. Child manifest directories must wait for the parent, but not sibling manifest directories.

#### Just apply it for me

```bash
iop install
```

This command will generate the manifests and apply them in the correct dependency order, waiting for dependencies to have CRDs available.   
Note: right now the actual "kubectl apply" if only displayed but not run by default. Set --dry-run=false to actually apply to cluster. 

#### What's in the defaults I installed?

To see the values for the profile in use:

```bash
iop dump-profile
```

This also works with different selected profiles, as in the next section.

#### Select a profile

The simplest customization is to select a profile different to default e.g. sds. Create the following config file: 

```yaml
# sds-install.yaml

profile: sds
```

Pass it to iop:

```bash
iop manifest -f sds-install.yaml
```

This will cause the helm charts to be rendered with data/profiles/sds.yaml. 

#### Install from file path

The compiled in charts and profiles are used by default, but a file path (and possibly URL going forward) can be specified, e.g. 

```yaml
profile: file:///usr/home/bob/go/src/github.com/ostromart/istio-installer/data/profiles/default.yaml
customPackagePath: file:///usr/home/bob/go/src/github.com/ostromart/istio-installer/data/charts/
```

These can be mixed and matched e.g. use a compiled in profile with local filesystem charts. 

#### New API customization

The [new platform level installation API](https://github.com/istio/operator/blob/95e89c4fb838b0b374f70d7c5814329e25a64819/pkg/apis/istio/v1alpha1/istioinstaller_types.proto#L25) defines install time parameters like feature/component enablement and namespace, and k8s settings like resources, HPA spec etc. in a structured way.   
The simplest customization is to turn features and components on and off. For example, to turn off all policy:

```yaml
profile: sds
policy:
  enabled: false
```
Note that unlike helm, all configurations are validated against a schema, so the operator detects syntax errors. Another customization is to define custom namespaces for features:

```yaml
profile: sds

trafficManagement:
  namespace: istio-control-custom 
```

The traffic management feature comprises Pilot, AutoInjection and Proxy components. Each of these components has k8s settings, and these can be overridden from the defaults using official k8s APIs (rather than Istio defined schemas):

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

All of these k8s settings use the k8s API definitions, so k8s documentation can be used for reference. All k8s overlay values are also validated in the operator.

#### Customizing the old values.yaml API 

The new platform install API above deals with k8s level settings. The remaining values.yaml parameters deal with Istio control plane operation rather than platform installation and for the time being the operator just passes these through to the helm charts unmodified (but validated through a [schema](https://github.com/istio/operator/blob/master/pkg/apis/istio/v1alpha2/values_types.go)).  
Values.yaml settings are overridden the same way as the new API, though a customized CR overlaid over default values for the selected profile. Here's an example of overriding some global level default values:

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

Advanced users may occasionally have the need to customize parameters (like container command line flags) which are not exposed through either of the installation or configuration APIs described in this document.   
For such cases, it's possible to overlay the generated k8s resources before they are applied with user-defined overlays. For example, to override some container level values in the Pilot container:

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

The user-defined overlay uses a path spec that includes the ability to select list items by key. In the example above, the container with the key-value "name: discovery" is selected from the list of containers, and the command line parameter with value "30m" is selected to be modified.   
The advanced overlay capability is described in more detail in the spec. 

#### Try the demo customization

This customization contains overlays at all three levels: the new API, values.yaml legacy API and the k8s output overlay. 

```bash
iop manifest -f samples/customize_pilot.yaml
```



