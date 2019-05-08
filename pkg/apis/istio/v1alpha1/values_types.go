package v1alpha1

// TODO: create remaining enum types.

import (
	corev1 "k8s.io/api/core/v1"
)

type Values struct {
	CertManager     *CertManagerConfig     `json:"certmanager,omitempty"`
	Galley          *GalleyConfig          `json:"galley,omitempty"`
	Global          *GlobalConfig          `json:"global,omitempty"`
	Grafana         map[string]interface{} `json:"grafana,omitempty"`
	Gateways        *GatewaysConfig        `json:"gateways,omitempty"`
	CNI             *CNIConfig             `json:"gateways,omitempty"`
	CoreDNS         *CoreDNSConfig         `json:"istiocoredns,omitempty"`
	Kiali           *KialiConfig           `json:"kiali,omitempty"`
	Mixer           *MixerConfig           `json:"mixer,omitempty"`
	NodeAgent       *NodeAgentConfig       `json:"nodeagent,omitempty"`
	Pilot           *PilotConfig           `json:"pilot,omitempty"`
	Prometheus      *PrometheusConfig      `json:"prometheus,omitempty"`
	Security        *SecurityConfig        `json:"security,omitempty"`
	ServiceGraph    *ServiceGraphConfig    `json:"servicegraph,omitempty"`
	SidecarInjector *SidecarInjectorConfig `json:"sidecarInjectorWebhook,omitempty"`
	Tracing         *TracingConfig         `json:"tracing,omitempty"`
}

type CertManagerConfig struct {
	Enabled   *bool            `json:"enabled,inline"`
	Hub       *string          `json:"hub,omitempty"`
	Tag       *string          `json:"tag,omitempty"`
	Resources *ResourcesConfig `json:"resources,omitempty"`
}

type GalleyConfig struct {
	Enabled      *bool   `json:"enabled,inline"`
	ReplicaCount *uint8  `json:"replicaCount,omitempty"`
	Image        *string `json:"image,omitempty"`
}

type GatewaysConfig struct {
	Enabled        *bool                 `json:"enabled,inline"`
	IngressGateway *IngressGatewayConfig `json:"istio-ingressgateway,inline"`
	EgressGateway  *EgressGatewayConfig  `json:"istio-egressgateway,inline"`
	ILBGateway     *ILBGatewayConfig     `json:"istio-ilbgateway,inline"`
}

type IngressGatewayConfig struct {
	Enabled                  *bool                       `json:"enabled,inline"`
	SDS                      *IngressGatewaySDSConfig    `json:"sds,omitempty"`
	Labels                   *GatewayLabelsConfig        `json:"labels,omitempty"`
	AutoscaleEnabled         *bool                       `json:"autoscaleEnabled,omitempty"`
	AutoscaleMax             *uint8                      `json:"autoscaleMax,omitempty"`
	AutoscaleMin             *uint8                      `json:"autoscaleMin,omitempty"`
	Resources                map[string]interface{}      `json:"resources,omitempty"`
	Cpu                      *CPUTargetUtilizationConfig `json:"cpu,omitempty"`
	LoadBalancerIP           *string                     `json:"loadBalancerIP,omitempty"`
	LoadBalancerSourceRanges []string                    `json:"loadBalancerSourceRanges,omitempty"`
	ExternalIPs              []string                    `json:"externalIPs,omitempty"`
	ServiceAnnotations       map[string]interface{}      `json:"serviceAnnotations,omitempty"`
	PodAnnotations           map[string]interface{}      `json:"podAnnotations,omitempty"`
	Type                     corev1.ServiceType          `json:"type,omitempty"`
	Ports                    []*PortsConfig              `json:"ports,omitempty"`
	MeshExpansionPorts       []*PortsConfig              `json:"meshExpansionPorts,omitempty"`
	SecretVolumes            []*SecretVolume             `json:"secretVolumes,omitempty"`
	NodeSelector             map[string]interface{}      `json:"nodeSelector,omitempty"`
}

type IngressGatewaySDSConfig struct {
	Enabled *bool   `json:"enabled,inline"`
	Image   *string `json:"image,omitempty"`
}

type EgressGatewayConfig struct {
	Enabled            *bool                       `json:"enabled,inline"`
	Labels             *GatewayLabelsConfig        `json:"labels,omitempty"`
	AutoscaleEnabled   *bool                       `json:"autoscaleEnabled,omitempty"`
	AutoscaleMax       *uint8                      `json:"autoscaleMax,omitempty"`
	AutoscaleMin       *uint8                      `json:"autoscaleMin,omitempty"`
	Cpu                *CPUTargetUtilizationConfig `json:"cpu,omitempty"`
	ServiceAnnotations map[string]interface{}      `json:"serviceAnnotations,omitempty"`
	PodAnnotations     map[string]interface{}      `json:"podAnnotations,omitempty"`
	Type               corev1.ServiceType          `json:"type,omitempty"`
	Ports              []*PortsConfig              `json:"ports,omitempty"`
	SecretVolumes      []*SecretVolume             `json:"secretVolumes,omitempty"`
	NodeSelector       map[string]string           `json:"nodeSelector,omitempty"`
}

type ILBGatewayConfig struct {
	Enabled            *bool                       `json:"enabled,inline"`
	Labels             *GatewayLabelsConfig        `json:"labels,omitempty"`
	AutoscaleEnabled   *bool                       `json:"autoscaleEnabled,omitempty"`
	AutoscaleMax       *uint8                      `json:"autoscaleMax,omitempty"`
	AutoscaleMin       *uint8                      `json:"autoscaleMin,omitempty"`
	Cpu                *CPUTargetUtilizationConfig `json:"cpu,omitempty"`
	Resources          *ResourcesConfig            `json:"resources,omitempty"`
	LoadBalancerIP     *string                     `json:"loadBalancerIP,omitempty"`
	ServiceAnnotations map[string]interface{}      `json:"serviceAnnotations,omitempty"`
	PodAnnotations     map[string]interface{}      `json:"podAnnotations,omitempty"`
	Type               corev1.ServiceType          `json:"type,omitempty"`
	Ports              []*PortsConfig              `json:"ports,omitempty"`
	SecretVolumes      []*SecretVolume             `json:"secretVolumes,omitempty"`
	NodeSelector       map[string]interface{}      `json:"nodeSelector,omitempty"`
}

type GlobalConfig struct {
	Hub                         *string                           `json:"hub,omitempty"`
	Tag                         *string                           `json:"tag,omitempty"`
	MonitoringPort              *uint16                           `json:"monitoringPort,omitempty"`
	KubernetesIngress           *KubernetesIngressConfig          `json:"k8sIngress,omitempty"`
	Proxy                       *ProxyConfig                      `json:"proxy,omitempty"`
	ProxyInit                   *ProxyInitConfig                  `json:"proxy_init,omitempty"`
	ImagePullPolicy             corev1.PullPolicy                 `json:"imagePullPolicy,omitempty"`
	ControlPlaneSecurityEnabled *bool                             `json:"controlPlaneSecurityEnabled,omitempty"`
	DisablePolicyChecks         *bool                             `json:"disablePolicyChecks,omitempty"`
	PolicyCheckFailOpen         *bool                             `json:"policyCheckFailOpen,omitempty"`
	EnableTracing               *bool                             `json:"enableTracing,omitempty"`
	Tracer                      *TracerConfig                     `json:"tracer,omitempty"`
	MTLS                        *MTLSConfig                       `json:"mtls,omitempty"`
	Arch                        *ArchConfig                       `json:"arch,omitempty"`
	OneNamespace                *bool                             `json:"oneNamespace,omitempty"`
	DefaultNodeSelector         map[string]interface{}            `json:"defaultNodeSelector,omitempty"`
	ConfigValidation            *bool                             `json:"configValidation,omitempty"`
	MeshExpansion               *MeshExpansionConfig              `json:"meshExpansion,omitempty"`
	MultiCluster                *MultiClusterConfig               `json:"multiCluster,omitempty"`
	DefaultResources            *DefaultResourcesConfig           `json:"defaultResources,omitempty"`
	DefaultPodDisruptionBudget  *DefaultPodDisruptionBudgetConfig `json:"defaultPodDisruptionBudget,omitempty"`
	PriorityClassName           *string                           `json:"priorityClassName,omitempty"`
	UseMCP                      *bool                             `json:"useMCP,omitempty"`
	TrustDomain                 *string                           `json:"trustDomain,omitempty"`
	OutboundTrafficPolicy       *OutboundTrafficPolicyConfig      `json:"outboundTrafficPolicy,omitempty"`
	SDS                         *SDSConfig                        `json:"sds,omitempty"`
	// TODO: check this
	MeshNetworks map[string]interface{} `json:"meshNetworks,omitempty"`
}

// KubernetesIngressConfig represents the configuration for Kubernetes Ingress.
type KubernetesIngressConfig struct {
	Enabled     *bool   `json:"enabled,inline"`
	GatewayName *string `json:"gatewayName,omitempty"`
	EnableHTTPS *bool   `json:"enableHttps,omitempty"`
}

// ProxyConfig specifies how proxies are configured within Istio.
type ProxyConfig struct {
	Image                        *string                       `json:"image,omitempty"`
	ClusterDomain                *string                       `json:"clusterDomain,omitempty"`
	Resources                    *ResourcesConfig              `json:"resources,omitempty"`
	Concurrency                  *uint8                        `json:"concurrency,omitempty"`
	AccessLogFile                *string                       `json:"accessLogFile,omitempty"`
	AccessLogFormat              *string                       `json:"accessLogFormat,omitempty"`
	AccessLogEncoding            ProxyConfig_AccessLogEncoding `json:"accessLogEncoding,omitempty"`
	DnsRefreshRate               *string                       `json:"dnsRefreshRate,omitempty"`
	Privileged                   *bool                         `json:"privileged,omitempty"`
	EnableCoreDump               *bool                         `json:"enableCoreDump,omitempty"`
	StatusPort                   *uint16                       `json:"statusPort,omitempty"`
	ReadinessInitialDelaySeconds *uint16                       `json:"readinessInitialDelaySeconds,omitempty"`
	ReadinessPeriodSeconds       *uint16                       `json:"readinessPeriodSeconds,omitempty"`
	ReadinessFailureThreshold    *uint16                       `json:"readinessFailureThreshold,omitempty"`
	IncludeIPRanges              *string                       `json:"includeIPRanges,omitempty"`
	ExcludeIPRanges              *string                       `json:"excludeIPRanges,omitempty"`
	KubevirtInterfaces           *string                       `json:"kubevirtInterfaces,omitempty"`
	IncludeInboundPorts          *string                       `json:"includeInboundPorts,omitempty"`
	ExcludeInboundPorts          *string                       `json:"excludeInboundPorts,omitempty"`
	AutoInject                   *bool                         `json:"autoInject,omitempty"`
	EnvoyStatsD                  *EnvoyMetricsConfig           `json:"envoyStatsd,omitempty"`
	EnvoyMetricsService          *EnvoyMetricsConfig           `json:"envoyMetricsService,omitempty"`
	Tracer                       *string                       `json:"tracer,omitempty"`
}

type ProxyConfig_AccessLogEncoding int32

const (
	ProxyConfig_JSON ProxyConfig_AccessLogEncoding = 0
	ProxyConfig_TEXT ProxyConfig_AccessLogEncoding = 1
)

var ProxyConfig_AccessLogEncoding_name = map[int32]string{
	0: "JSON",
	1: "TEXT",
}

var ProxyConfig_AccessLogEncoding_value = map[string]int32{
	"JSON": 0,
	"TEXT": 1,
}

type EnvoyMetricsConfig struct {
	Enabled *bool   `json:"enabled,inline"`
	Host    *string `json:"host,omitempty"`
	Port    *string `json:"port,omitempty"`
}

type ProxyInitConfig struct {
	Image *string `json:"image,omitempty"`
}

type TracerConfig struct {
	LightStep *TracerLightStepConfig `json:"lightstep,omitempty"`
	Zipkin    *TracerZipkinConfig    `json:"zipkin,omitempty"`
}

type TracerLightStepConfig struct {
	Address     *string `json:"address,omitempty"`
	AccessToken *string `json:"accessToken,omitempty"`
	Secure      *bool   `json:"secure,omitempty"`
	CACertPath  *string `json:"cacertPath,omitempty"`
}

type TracerZipkinConfig struct {
	Address *string `json:"address,omitempty"`
}

type MTLSConfig struct {
	Enabled *bool `json:"enabled,inline"`
}

type ArchConfig struct {
	Amd64   *uint8 `json:"amd64,omitempty"`
	S390x   *uint8 `json:"s390x,omitempty"`
	Ppc64le *uint8 `json:"ppc64le,omitempty"`
}

type MeshExpansionConfig struct {
	Enabled *bool `json:"enabled,inline"`
	UseILB  *bool `json:"useILB,omitempty"`
}

type MultiClusterConfig struct {
	Enabled *bool `json:"enabled,inline"`
}

type DefaultResourcesConfig struct {
	Requests *ResourcesRequestsConfig `json:"requests,omitempty"`
}

type DefaultPodDisruptionBudgetConfig struct {
	Enabled *bool `json:"enabled,inline"`
}

type OutboundTrafficPolicyConfig struct {
	Mode string `json:"mode,omitempty"`
}

type SDSConfig struct {
	Enabled           *bool   `json:"enabled,inline"`
	UDSPath           *string `json:"udsPath,omitempty"`
	UseTrustworthyJWT *bool   `json:"useTrustworthyJwt,omitempty"`
	UseNormalJWT      *bool   `json:"useNormalJwt,omitempty"`
}

type CNIConfig struct {
	Enabled *bool `json:"enabled,inline"`
}

type CoreDNSConfig struct {
	Enabled            *bool                  `json:"enabled,inline"`
	CoreDNSImage       *string                `json:"coreDNSImage,omitempty"`
	CoreDNSPluginImage *string                `json:"coreDNSPluginImage,omitempty"`
	ReplicaCount       *uint8                 `json:"replicaCount,omitempty"`
	NodeSelector       map[string]interface{} `json:"nodeSelector,omitempty"`
}

type KialiConfig struct {
	Enabled          *bool                  `json:"enabled,inline"`
	ReplicaCount     *uint8                 `json:"replicaCount,omitempty"`
	Hub              *string                `json:"hub,omitempty"`
	Tag              *string                `json:"tag,omitempty"`
	ContextPath      *string                `json:"contextPath,omitempty"`
	NodeSelector     map[string]interface{} `json:"nodeSelector,omitempty"`
	Ingress          *AddonIngressConfig    `json:"ingress,omitempty"`
	Dashboard        *KialiDashboardConfig  `json:"dashboard,omitempty"`
	PrometheusAddr   *string                `json:"prometheusAddr,omitempty"`
	CreateDemoSecret *bool                  `json:"createDemoSecret,omitempty"`
}

type KialiDashboardConfig struct {
	SecretName       *string `json:"secretName,omitempty"`
	UsernameKey      *string `json:"usernameKey,omitempty"`
	PassphraseKey    *string `json:"passphraseKey,omitempty"`
	GrafanaURL       *string `json:"grafanaURL,omitempty"`
	JaegerURL        *string `json:"jaegerURL,omitempty"`
	PrometheusAddr   *string `json:"prometheusAddr,omitempty"`
	CreateDemoSecret *string `json:"createDemoSecret,omitempty"`
}

type MixerConfig struct {
	Enabled   *bool                 `json:"enabled,inline"`
	Image     *string               `json:"image,omitempty"`
	Policy    *MixerPolicyConfig    `json:"policy,omitempty"`
	Telemetry *MixerTelemetryConfig `json:"telemetry,omitempty"`
	Adapters  *MixerAdaptersConfig  `json:"adapters,omitempty"`
	// TODO: env
}

type MixerPolicyConfig struct {
	Enabled          *bool                       `json:"enabled,inline"`
	ReplicaCount     *uint8                      `json:"replicaCount,omitempty"`
	AutoscaleEnabled *bool                       `json:"autoscaleEnabled,omitempty"`
	AutoscaleMax     *uint8                      `json:"autoscaleMax,omitempty"`
	AutoscaleMin     *uint8                      `json:"autoscaleMin,omitempty"`
	Cpu              *CPUTargetUtilizationConfig `json:"cpu,omitempty"`
}

type MixerTelemetryConfig struct {
	Enabled                *bool                      `json:"enabled,inline"`
	ReplicaCount           *uint8                     `json:"replicaCount,omitempty"`
	AutoscaleEnabled       *bool                      `json:"autoscaleEnabled,omitempty"`
	AutoscaleMax           *uint8                     `json:"autoscaleMax,omitempty"`
	AutoscaleMin           *uint8                     `json:"autoscaleMin,omitempty"`
	Cpu                    CPUTargetUtilizationConfig `json:"cpu,omitempty"`
	SessionAffinityEnabled *bool                      `json:"sessionAffinityEnabled,omitempty"`
	LoadShedding           *LoadSheddingConfig        `json:"loadshedding,omitempty"`
	Resources              *ResourcesConfig           `json:"resources,omitempty"`
	PodAnnotations         map[string]interface{}     `json:"podAnnotations,omitempty"`
	NodeSelector           map[string]interface{}     `json:"nodeSelector,omitempty"`
	Adapters               *MixerAdaptersConfig       `json:"adapters,omitempty"`
}

type LoadSheddingConfig struct {
	Mode             *string `json:"mode,omitempty"`
	LatencyThreshold *string `json:"latencyThreshold,omitempty"`
}

type MixerAdaptersConfig struct {
	KubernetesEnv  *KubernetesEnvMixerAdapterConfig `json:"kubernetesenv,omitempty"`
	Stdio          *StdioMixerAdapterConfig         `json:"stdio,omitempty"`
	Prometheus     *PrometheusMixerAdapterConfig    `json:"prometheus,omitempty"`
	UseAdapterCRDs *bool                            `json:"useAdapterCRDs,omitempty"`
}

type KubernetesEnvMixerAdapterConfig struct {
	Enabled *bool `json:"enabled,inline"`
}

type StdioMixerAdapterConfig struct {
	Enabled      *bool `json:"enabled,inline"`
	OutputAsJSON *bool
}

type PrometheusMixerAdapterConfig struct {
	Enabled              *bool `json:"enabled,inline"`
	MetricExpiryDuration string
}

type NodeAgentConfig struct {
	Enabled      *bool                  `json:"enabled,inline"`
	Image        *string                `json:"image,omitempty"`
	NodeSelector map[string]interface{} `json:"nodeSelector,omitempty"`
	// TODO: env, Plugins
}

type PilotConfig struct {
	Enabled                         *bool                       `json:"enabled,inline"`
	AutoscaleEnabled                *bool                       `json:"autoscaleEnabled,omitempty"`
	AutoscaleMax                    *uint8                      `json:"autoscaleMax,omitempty"`
	AutoscaleMin                    *uint8                      `json:"autoscaleMin,omitempty"`
	Image                           *string                     `json:"image,omitempty"`
	Sidecar                         *bool                       `json:"sidecar,omitempty"`
	TraceSampling                   *float64                    `json:"traceSampling,omitempty"`
	Resources                       *ResourcesConfig            `json:"resources,omitempty"`
	Cpu                             *CPUTargetUtilizationConfig `json:"cpu,omitempty"`
	NodeSelector                    map[string]interface{}      `json:"nodeSelector,omitempty"`
	KeepaliveMaxServerConnectionAge *string                     `json:"keepaliveMaxServerConnectionAge,omitempty"`
	// TODO: env
}

type PrometheusConfig struct {
	Enabled        *bool                     `json:"enabled,inline"`
	ReplicaCount   *uint8                    `json:"replicaCount,omitempty"`
	Hub            *string                   `json:"hub,omitempty"`
	Tag            *string                   `json:"tag,omitempty"`
	Retention      *string                   `json:"retention,omitempty"`
	NodeSelector   map[string]interface{}    `json:"nodeSelector,omitempty"`
	ScrapeInterval *string                   `json:"scrapeInterval,omitempty"`
	ContextPath    *string                   `json:"contextPath,omitempty"`
	Ingress        *AddonIngressConfig       `json:"ingress,omitempty"`
	Service        *PrometheusServiceConfig  `json:"service,omitempty"`
	Security       *PrometheusSecurityConfig `json:"security,omitempty"`
}

type PrometheusServiceConfig struct {
	Annotations map[string]interface{}           `json:"annotations,omitempty"`
	NodePort    *PrometheusServiceNodePortConfig `json:"nodePort,omitempty"`
}

type PrometheusServiceNodePortConfig struct {
	Enabled *bool   `json:"enabled,inline"`
	Port    *uint16 `json:"port,omitempty"`
}

type PrometheusSecurityConfig struct {
	Enabled *bool `json:"enabled,inline"`
}

type SecurityConfig struct {
	Enabled          *bool                  `json:"enabled,inline"`
	ReplicaCount     *uint8                 `json:"replicaCount,omitempty"`
	Image            *string                `json:"image,omitempty"`
	SelfSigned       *bool                  `json:"selfSigned,omitempty"`
	CreateMeshPolicy *bool                  `json:"createMeshPolicy,omitempty"`
	NodeSelector     map[string]interface{} `json:"nodeSelector,omitempty"`
}

type ServiceGraphConfig struct {
	Enabled        *bool                  `json:"enabled,inline"`
	ReplicaCount   *uint8                 `json:"replicaCount,omitempty"`
	Image          *string                `json:"image,omitempty"`
	NodeSelector   map[string]interface{} `json:"nodeSelector,omitempty"`
	Annotations    map[string]interface{} `json:"annotations,omitempty"`
	Service        *ServiceConfig         `json:"service,omitempty"`
	Ingress        *AddonIngressConfig    `json:"ingress,omitempty"`
	PrometheusAddr *string                `json:"prometheusAddr,omitempty"`
}

type SidecarInjectorConfig struct {
	Enabled                   *bool                  `json:"enabled,inline"`
	ReplicaCount              *uint8                 `json:"replicaCount,omitempty"`
	Image                     *string                `json:"image,omitempty"`
	EnableNamespacesByDefault *bool                  `json:"enableNamespacesByDefault,omitempty"`
	NodeSelector              map[string]interface{} `json:"nodeSelector,omitempty"`
	RewriteAppHTTPProbe       *bool                  `json:"rewriteAppHTTPProbe,inline"`
}

type TracingConfig struct {
	Enabled      *bool                  `json:"enabled,inline"`
	Provider     *string                `json:"provider,omitempty"`
	NodeSelector map[string]interface{} `json:"nodeSelector,omitempty"`
	Jaeger       *TracingJaegerConfig   `json:"jaeger,omitempty"`
	Zipkin       *TracingZipkinConfig   `json:"zipkin,omitempty"`
	Service      *ServiceConfig         `json:"service,omitempty"`
	Ingress      *TracingIngressConfig  `json:"ingress,omitempty"`
}

type TracingJaegerConfig struct {
	Hub    *string                    `json:"hub,omitempty"`
	Tag    *string                    `json:"tag,omitempty"`
	Memory *TracingJaegerMemoryConfig `json:"memory,omitempty"`
}

type TracingJaegerMemoryConfig struct {
	MaxTraces *string `json:"max_traces,omitempty"`
}

type TracingZipkinConfig struct {
	Hub               *string                  `json:"hub,omitempty"`
	Tag               *string                  `json:"tag,omitempty"`
	ProbeStartupDelay *uint16                  `json:"probeStartupDelay,omitempty"`
	QueryPort         *uint16                  `json:"queryPort,omitempty"`
	Resources         *ResourcesConfig         `json:"resources,omitempty"`
	JavaOptsHeap      *string                  `json:"javaOptsHeap,omitempty"`
	MaxSpans          *string                  `json:"maxSpans,omitempty"`
	Node              *TracingZipkinNodeConfig `json:"node,omitempty"`
}

type TracingZipkinNodeConfig struct {
	CPUs *uint8 `json:"cpus,omitempty"`
}

type TracingIngressConfig struct {
	Enabled *bool `json:"enabled,inline"`
}

// Shared types

type ResourcesConfig struct {
	Requests *ResourcesRequestsConfig `json:"requests,omitempty"`
	Limits   *ResourcesRequestsConfig `json:"limits,omitempty"`
}

type ResourcesRequestsConfig struct {
	Cpu    *string `json:"cpu,omitempty"`
	Memory *string `json:"memory,omitempty"`
}

type ServiceConfig struct {
	Annotations  map[string]interface{} `json:"annotations,omitempty"`
	Name         *string                `json:"name,omitempty"`
	ExternalPort *uint16                `json:"externalPort,omitempty"`
	Type         corev1.ServiceType     `json:"type,omitempty"`
}

type CPUTargetUtilizationConfig struct {
	TargetAverageUtilization *int32 `json:"targetAverageUtilization,omitempty"`
}

type PortsConfig struct {
	Name       *string `json:"name,omitempty"`
	TargetPort *string `json:"targetPort,omitempty"`
	NodePort   *string `json:"nodePort,omitempty"`
}

type SecretVolume struct {
	MountPath  *string `json:"mountPath,omitempty"`
	SecretName *string `json:"secretName,omitempty"`
}

type GatewayLabelsConfig struct {
	App   *string `json:"app,omitempty"`
	Istio *string `json:"istio,omitempty"`
}

type AddonIngressConfig struct {
	Enabled *bool    `json:"enabled,inline"`
	Hosts   []string `json:"hosts,omitempty"`
}
