package v1alpha1

// TODO: create remaining enum types.

import (
	corev1 "k8s.io/api/core/v1"
)

// Values is described in istio.io documentation.
type Values struct {
	CertManager     *CertManagerConfig     `json:"certmanager,omitempty"`
	Galley          *GalleyConfig          `json:"galley,omitempty"`
	Global          *GlobalConfig          `json:"global,omitempty"`
	Grafana         map[string]interface{} `json:"grafana,omitempty"`
	Gateways        *GatewaysConfig        `json:"gateways,omitempty"`
	CNI             *CNIConfig             `json:"cni,omitempty"`
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

// CertManagerConfig is described in istio.io documentation.
type CertManagerConfig struct {
	Enabled   *bool            `json:"enabled,inline"`
	Hub       *string          `json:"hub,omitempty"`
	Tag       *string          `json:"tag,omitempty"`
	Resources *ResourcesConfig `json:"resources,omitempty"`
}

// GalleyConfig is described in istio.io documentation.
type GalleyConfig struct {
	Enabled      *bool   `json:"enabled,inline"`
	ReplicaCount *uint8  `json:"replicaCount,omitempty"`
	Image        *string `json:"image,omitempty"`
}

// GatewaysConfig is described in istio.io documentation.
type GatewaysConfig struct {
	Enabled  *bool `json:"enabled,inline"`
	Gateways map[string]*GatewaysOneOfConfig
}

// GatewaysOneOfConfig is described in istio.io documentation.
type GatewaysOneOfConfig struct {
	Enabled        *bool                 `json:"enabled,inline"`
	IngressGateway *IngressGatewayConfig `json:"istio-ingressgateway,inline"`
	EgressGateway  *EgressGatewayConfig  `json:"istio-egressgateway,inline"`
	ILBGateway     *ILBGatewayConfig     `json:"istio-ilbgateway,inline"`
	UserGateway    *UserGatewayConfig
}

// IngressGatewayConfig is described in istio.io documentation.
type IngressGatewayConfig struct {
	Enabled                  *bool                       `json:"enabled,inline"`
	SDS                      *IngressGatewaySDSConfig    `json:"sds,omitempty"`
	Labels                   *GatewayLabelsConfig        `json:"labels,omitempty"`
	AutoscaleEnabled         *bool                       `json:"autoscaleEnabled,omitempty"`
	AutoscaleMax             *uint8                      `json:"autoscaleMax,omitempty"`
	AutoscaleMin             *uint8                      `json:"autoscaleMin,omitempty"`
	Resources                map[string]interface{}      `json:"resources,omitempty"`
	CPU                      *CPUTargetUtilizationConfig `json:"cpu,omitempty"`
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

// IngressGatewaySDSConfig is described in istio.io documentation.
type IngressGatewaySDSConfig struct {
	Enabled *bool   `json:"enabled,inline"`
	Image   *string `json:"image,omitempty"`
}

// EgressGatewayConfig is described in istio.io documentation.
type EgressGatewayConfig struct {
	Enabled            *bool                       `json:"enabled,inline"`
	Labels             *GatewayLabelsConfig        `json:"labels,omitempty"`
	AutoscaleEnabled   *bool                       `json:"autoscaleEnabled,omitempty"`
	AutoscaleMax       *uint8                      `json:"autoscaleMax,omitempty"`
	AutoscaleMin       *uint8                      `json:"autoscaleMin,omitempty"`
	CPU                *CPUTargetUtilizationConfig `json:"cpu,omitempty"`
	ServiceAnnotations map[string]interface{}      `json:"serviceAnnotations,omitempty"`
	PodAnnotations     map[string]interface{}      `json:"podAnnotations,omitempty"`
	Type               corev1.ServiceType          `json:"type,omitempty"`
	Ports              []*PortsConfig              `json:"ports,omitempty"`
	SecretVolumes      []*SecretVolume             `json:"secretVolumes,omitempty"`
	NodeSelector       map[string]string           `json:"nodeSelector,omitempty"`
}

// ILBGatewayConfig is described in istio.io documentation.
type ILBGatewayConfig struct {
	Enabled            *bool                       `json:"enabled,inline"`
	Labels             *GatewayLabelsConfig        `json:"labels,omitempty"`
	AutoscaleEnabled   *bool                       `json:"autoscaleEnabled,omitempty"`
	AutoscaleMax       *uint8                      `json:"autoscaleMax,omitempty"`
	AutoscaleMin       *uint8                      `json:"autoscaleMin,omitempty"`
	CPU                *CPUTargetUtilizationConfig `json:"cpu,omitempty"`
	Resources          *ResourcesConfig            `json:"resources,omitempty"`
	LoadBalancerIP     *string                     `json:"loadBalancerIP,omitempty"`
	ServiceAnnotations map[string]interface{}      `json:"serviceAnnotations,omitempty"`
	PodAnnotations     map[string]interface{}      `json:"podAnnotations,omitempty"`
	Type               corev1.ServiceType          `json:"type,omitempty"`
	Ports              []*PortsConfig              `json:"ports,omitempty"`
	SecretVolumes      []*SecretVolume             `json:"secretVolumes,omitempty"`
	NodeSelector       map[string]interface{}      `json:"nodeSelector,omitempty"`
}

// UserGatewayConfig is described in istio.io documentation.
type UserGatewayConfig struct {
	Enabled                  *bool                       `json:"enabled,inline"`
	Labels                   *GatewayLabelsConfig        `json:"labels,omitempty"`
	AutoscaleEnabled         *bool                       `json:"autoscaleEnabled,omitempty"`
	AutoscaleMax             *uint8                      `json:"autoscaleMax,omitempty"`
	AutoscaleMin             *uint8                      `json:"autoscaleMin,omitempty"`
	CPU                      *CPUTargetUtilizationConfig `json:"cpu,omitempty"`
	Resources                *ResourcesConfig            `json:"resources,omitempty"`
	LoadBalancerIP           *string                     `json:"loadBalancerIP,omitempty"`
	LoadBalancerSourceRanges []string                    `json:"loadBalancerSourceRanges,omitempty"`
	ServiceAnnotations       map[string]interface{}      `json:"serviceAnnotations,omitempty"`
	PodAnnotations           map[string]interface{}      `json:"podAnnotations,omitempty"`
	Type                     corev1.ServiceType          `json:"type,omitempty"`
	ExternalTrafficPolicy    string                      `json:"externalTrafficPolicy,omitempty"`
	Ports                    []*PortsConfig              `json:"ports,omitempty"`
	SecretVolumes            []*SecretVolume             `json:"secretVolumes,omitempty"`
	NodeSelector             map[string]string           `json:"nodeSelector,omitempty"`
}

// GlobalConfig is described in istio.io documentation.
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
	DNSRefreshRate               *string                       `json:"dnsRefreshRate,omitempty"`
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

// ProxyConfig_AccessLogEncoding is described in istio.io documentation.
type ProxyConfig_AccessLogEncoding int32

const (
	// ProxyConfig_JSON is described in istio.io documentation.
	ProxyConfig_JSON ProxyConfig_AccessLogEncoding = 0
	// ProxyConfig_TEXT is described in istio.io documentation.
	ProxyConfig_TEXT ProxyConfig_AccessLogEncoding = 1
)

// ProxyConfig_AccessLogEncoding_name is described in istio.io documentation.
var ProxyConfig_AccessLogEncoding_name = map[int32]string{
	0: "JSON",
	1: "TEXT",
}

// ProxyConfig_AccessLogEncoding_value is described in istio.io documentation.
var ProxyConfig_AccessLogEncoding_value = map[string]int32{
	"JSON": 0,
	"TEXT": 1,
}

// EnvoyMetricsConfig is described in istio.io documentation.
type EnvoyMetricsConfig struct {
	Enabled *bool   `json:"enabled,inline"`
	Host    *string `json:"host,omitempty"`
	Port    *string `json:"port,omitempty"`
}

// ProxyInitConfig is described in istio.io documentation.
type ProxyInitConfig struct {
	Image *string `json:"image,omitempty"`
}

// TracerConfig is described in istio.io documentation.
type TracerConfig struct {
	LightStep *TracerLightStepConfig `json:"lightstep,omitempty"`
	Zipkin    *TracerZipkinConfig    `json:"zipkin,omitempty"`
}

// TracerLightStepConfig is described in istio.io documentation.
type TracerLightStepConfig struct {
	Address     *string `json:"address,omitempty"`
	AccessToken *string `json:"accessToken,omitempty"`
	Secure      *bool   `json:"secure,omitempty"`
	CACertPath  *string `json:"cacertPath,omitempty"`
}

// TracerZipkinConfig is described in istio.io documentation.
type TracerZipkinConfig struct {
	Address *string `json:"address,omitempty"`
}

// MTLSConfig is described in istio.io documentation.
type MTLSConfig struct {
	Enabled *bool `json:"enabled,inline"`
}

// ArchConfig is described in istio.io documentation.
type ArchConfig struct {
	Amd64   *uint8 `json:"amd64,omitempty"`
	S390x   *uint8 `json:"s390x,omitempty"`
	Ppc64le *uint8 `json:"ppc64le,omitempty"`
}

// MeshExpansionConfig is described in istio.io documentation.
type MeshExpansionConfig struct {
	Enabled *bool `json:"enabled,inline"`
	UseILB  *bool `json:"useILB,omitempty"`
}

// MultiClusterConfig is described in istio.io documentation.
type MultiClusterConfig struct {
	Enabled *bool `json:"enabled,inline"`
}

// DefaultResourcesConfig is described in istio.io documentation.
type DefaultResourcesConfig struct {
	Requests *ResourcesRequestsConfig `json:"requests,omitempty"`
}

// DefaultPodDisruptionBudgetConfig is described in istio.io documentation.
type DefaultPodDisruptionBudgetConfig struct {
	Enabled *bool `json:"enabled,inline"`
}

// OutboundTrafficPolicyConfig is described in istio.io documentation.
type OutboundTrafficPolicyConfig struct {
	Mode string `json:"mode,omitempty"`
}

// SDSConfig is described in istio.io documentation.
type SDSConfig struct {
	Enabled           *bool   `json:"enabled,inline"`
	UDSPath           *string `json:"udsPath,omitempty"`
	UseTrustworthyJWT *bool   `json:"useTrustworthyJwt,omitempty"`
	UseNormalJWT      *bool   `json:"useNormalJwt,omitempty"`
}

// CNIConfig is described in istio.io documentation.
type CNIConfig struct {
	Enabled *bool `json:"enabled,inline"`
}

// CoreDNSConfig is described in istio.io documentation.
type CoreDNSConfig struct {
	Enabled            *bool                  `json:"enabled,inline"`
	CoreDNSImage       *string                `json:"coreDNSImage,omitempty"`
	CoreDNSPluginImage *string                `json:"coreDNSPluginImage,omitempty"`
	ReplicaCount       *uint8                 `json:"replicaCount,omitempty"`
	NodeSelector       map[string]interface{} `json:"nodeSelector,omitempty"`
}

// KialiConfig is described in istio.io documentation.
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

// KialiDashboardConfig is described in istio.io documentation.
type KialiDashboardConfig struct {
	SecretName       *string `json:"secretName,omitempty"`
	UsernameKey      *string `json:"usernameKey,omitempty"`
	PassphraseKey    *string `json:"passphraseKey,omitempty"`
	GrafanaURL       *string `json:"grafanaURL,omitempty"`
	JaegerURL        *string `json:"jaegerURL,omitempty"`
	PrometheusAddr   *string `json:"prometheusAddr,omitempty"`
	CreateDemoSecret *string `json:"createDemoSecret,omitempty"`
}

// MixerConfig is described in istio.io documentation.
type MixerConfig struct {
	Enabled   *bool                 `json:"enabled,inline"`
	Image     *string               `json:"image,omitempty"`
	Policy    *MixerPolicyConfig    `json:"policy,omitempty"`
	Telemetry *MixerTelemetryConfig `json:"telemetry,omitempty"`
	Adapters  *MixerAdaptersConfig  `json:"adapters,omitempty"`
	// TODO: env
}

// MixerPolicyConfig is described in istio.io documentation.
type MixerPolicyConfig struct {
	Enabled          *bool                       `json:"enabled,inline"`
	ReplicaCount     *uint8                      `json:"replicaCount,omitempty"`
	AutoscaleEnabled *bool                       `json:"autoscaleEnabled,omitempty"`
	AutoscaleMax     *uint8                      `json:"autoscaleMax,omitempty"`
	AutoscaleMin     *uint8                      `json:"autoscaleMin,omitempty"`
	CPU              *CPUTargetUtilizationConfig `json:"cpu,omitempty"`
}

// MixerTelemetryConfig is described in istio.io documentation.
type MixerTelemetryConfig struct {
	Enabled                *bool                      `json:"enabled,inline"`
	ReplicaCount           *uint8                     `json:"replicaCount,omitempty"`
	AutoscaleEnabled       *bool                      `json:"autoscaleEnabled,omitempty"`
	AutoscaleMax           *uint8                     `json:"autoscaleMax,omitempty"`
	AutoscaleMin           *uint8                     `json:"autoscaleMin,omitempty"`
	CPU                    CPUTargetUtilizationConfig `json:"cpu,omitempty"`
	SessionAffinityEnabled *bool                      `json:"sessionAffinityEnabled,omitempty"`
	LoadShedding           *LoadSheddingConfig        `json:"loadshedding,omitempty"`
	Resources              *ResourcesConfig           `json:"resources,omitempty"`
	PodAnnotations         map[string]interface{}     `json:"podAnnotations,omitempty"`
	NodeSelector           map[string]interface{}     `json:"nodeSelector,omitempty"`
	Adapters               *MixerAdaptersConfig       `json:"adapters,omitempty"`
}

// LoadSheddingConfig is described in istio.io documentation.
type LoadSheddingConfig struct {
	Mode             *string `json:"mode,omitempty"`
	LatencyThreshold *string `json:"latencyThreshold,omitempty"`
}

// MixerAdaptersConfig is described in istio.io documentation.
type MixerAdaptersConfig struct {
	KubernetesEnv  *KubernetesEnvMixerAdapterConfig `json:"kubernetesenv,omitempty"`
	Stdio          *StdioMixerAdapterConfig         `json:"stdio,omitempty"`
	Prometheus     *PrometheusMixerAdapterConfig    `json:"prometheus,omitempty"`
	UseAdapterCRDs *bool                            `json:"useAdapterCRDs,omitempty"`
}

// KubernetesEnvMixerAdapterConfig is described in istio.io documentation.
type KubernetesEnvMixerAdapterConfig struct {
	Enabled *bool `json:"enabled,inline"`
}

// StdioMixerAdapterConfig is described in istio.io documentation.
type StdioMixerAdapterConfig struct {
	Enabled      *bool `json:"enabled,inline"`
	OutputAsJSON *bool
}

// PrometheusMixerAdapterConfig is described in istio.io documentation.
type PrometheusMixerAdapterConfig struct {
	Enabled              *bool `json:"enabled,inline"`
	MetricExpiryDuration string
}

// NodeAgentConfig is described in istio.io documentation.
type NodeAgentConfig struct {
	Enabled      *bool                  `json:"enabled,inline"`
	Image        *string                `json:"image,omitempty"`
	NodeSelector map[string]interface{} `json:"nodeSelector,omitempty"`
	// TODO: env, Plugins
}

// PilotConfig is described in istio.io documentation.
type PilotConfig struct {
	Enabled                         *bool                       `json:"enabled,inline"`
	AutoscaleEnabled                *bool                       `json:"autoscaleEnabled,omitempty"`
	AutoscaleMax                    *uint8                      `json:"autoscaleMax,omitempty"`
	AutoscaleMin                    *uint8                      `json:"autoscaleMin,omitempty"`
	Image                           *string                     `json:"image,omitempty"`
	Sidecar                         *bool                       `json:"sidecar,omitempty"`
	TraceSampling                   *float64                    `json:"traceSampling,omitempty"`
	Resources                       *ResourcesConfig            `json:"resources,omitempty"`
	CPU                             *CPUTargetUtilizationConfig `json:"cpu,omitempty"`
	NodeSelector                    map[string]interface{}      `json:"nodeSelector,omitempty"`
	KeepaliveMaxServerConnectionAge *string                     `json:"keepaliveMaxServerConnectionAge,omitempty"`
	// TODO: env
}

// PrometheusConfig is described in istio.io documentation.
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

// PrometheusServiceConfig is described in istio.io documentation.
type PrometheusServiceConfig struct {
	Annotations map[string]interface{}           `json:"annotations,omitempty"`
	NodePort    *PrometheusServiceNodePortConfig `json:"nodePort,omitempty"`
}

// PrometheusServiceNodePortConfig is described in istio.io documentation.
type PrometheusServiceNodePortConfig struct {
	Enabled *bool   `json:"enabled,inline"`
	Port    *uint16 `json:"port,omitempty"`
}

// PrometheusSecurityConfig is described in istio.io documentation.
type PrometheusSecurityConfig struct {
	Enabled *bool `json:"enabled,inline"`
}

// SecurityConfig is described in istio.io documentation.
type SecurityConfig struct {
	Enabled          *bool                  `json:"enabled,inline"`
	ReplicaCount     *uint8                 `json:"replicaCount,omitempty"`
	Image            *string                `json:"image,omitempty"`
	SelfSigned       *bool                  `json:"selfSigned,omitempty"`
	CreateMeshPolicy *bool                  `json:"createMeshPolicy,omitempty"`
	NodeSelector     map[string]interface{} `json:"nodeSelector,omitempty"`
}

// ServiceGraphConfig is described in istio.io documentation.
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

// SidecarInjectorConfig is described in istio.io documentation.
type SidecarInjectorConfig struct {
	Enabled                   *bool                  `json:"enabled,inline"`
	ReplicaCount              *uint8                 `json:"replicaCount,omitempty"`
	Image                     *string                `json:"image,omitempty"`
	EnableNamespacesByDefault *bool                  `json:"enableNamespacesByDefault,omitempty"`
	NodeSelector              map[string]interface{} `json:"nodeSelector,omitempty"`
	RewriteAppHTTPProbe       *bool                  `json:"rewriteAppHTTPProbe,inline"`
}

// TracingConfig is described in istio.io documentation.
type TracingConfig struct {
	Enabled      *bool                  `json:"enabled,inline"`
	Provider     *string                `json:"provider,omitempty"`
	NodeSelector map[string]interface{} `json:"nodeSelector,omitempty"`
	Jaeger       *TracingJaegerConfig   `json:"jaeger,omitempty"`
	Zipkin       *TracingZipkinConfig   `json:"zipkin,omitempty"`
	Service      *ServiceConfig         `json:"service,omitempty"`
	Ingress      *TracingIngressConfig  `json:"ingress,omitempty"`
}

// TracingJaegerConfig is described in istio.io documentation.
type TracingJaegerConfig struct {
	Hub    *string                    `json:"hub,omitempty"`
	Tag    *string                    `json:"tag,omitempty"`
	Memory *TracingJaegerMemoryConfig `json:"memory,omitempty"`
}

// TracingJaegerMemoryConfig is described in istio.io documentation.
type TracingJaegerMemoryConfig struct {
	MaxTraces *string `json:"max_traces,omitempty"`
}

// TracingZipkinConfig is described in istio.io documentation.
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

// TracingZipkinNodeConfig is described in istio.io documentation.
type TracingZipkinNodeConfig struct {
	CPUs *uint8 `json:"cpus,omitempty"`
}

// TracingIngressConfig is described in istio.io documentation.
type TracingIngressConfig struct {
	Enabled *bool `json:"enabled,inline"`
}

// Shared types

// ResourcesConfig is described in istio.io documentation.
type ResourcesConfig struct {
	Requests *ResourcesRequestsConfig `json:"requests,omitempty"`
	Limits   *ResourcesRequestsConfig `json:"limits,omitempty"`
}

// ResourcesRequestsConfig is described in istio.io documentation.
type ResourcesRequestsConfig struct {
	CPU    *string `json:"cpu,omitempty"`
	Memory *string `json:"memory,omitempty"`
}

// ServiceConfig is described in istio.io documentation.
type ServiceConfig struct {
	Annotations  map[string]interface{} `json:"annotations,omitempty"`
	Name         *string                `json:"name,omitempty"`
	ExternalPort *uint16                `json:"externalPort,omitempty"`
	Type         corev1.ServiceType     `json:"type,omitempty"`
}

// CPUTargetUtilizationConfig is described in istio.io documentation.
type CPUTargetUtilizationConfig struct {
	TargetAverageUtilization *int32 `json:"targetAverageUtilization,omitempty"`
}

// PortsConfig is described in istio.io documentation.
type PortsConfig struct {
	Name       *string `json:"name,omitempty"`
	TargetPort *string `json:"targetPort,omitempty"`
	NodePort   *string `json:"nodePort,omitempty"`
}

// SecretVolume is described in istio.io documentation.
type SecretVolume struct {
	MountPath  *string `json:"mountPath,omitempty"`
	SecretName *string `json:"secretName,omitempty"`
}

// GatewayLabelsConfig is described in istio.io documentation.
type GatewayLabelsConfig struct {
	App   *string `json:"app,omitempty"`
	Istio *string `json:"istio,omitempty"`
}

// AddonIngressConfig is described in istio.io documentation.
type AddonIngressConfig struct {
	Enabled *bool    `json:"enabled,inline"`
	Hosts   []string `json:"hosts,omitempty"`
}
