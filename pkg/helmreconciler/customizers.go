package helmreconciler

import (
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/helm/pkg/manifest"

	"istio.io/operator/pkg/controller/common"
)

// Customizer encapsulates all customization applied during rendering, including input, markings, and listeners.
type Customizer struct {
	input    RenderingInput
	markings ResourceMarkings
	listener RenderingListener
}

// CustomizerFactory creates Customizer objects using the specified factories.
type CustomizerFactory struct {
	// Logger is...
	Logger logr.Logger
	// InputFactory is the factory used to create RenderingInput for the custom resource
	InputFactory RenderingInputFactory
	// MarkingsFactory is the factory used to create ResourceMarkings for the custom resource
	MarkingsFactory ResourceMarkingsFactory
	// ListenerFactory is the factory used to create RenderingListener for the custom resource
	ListenerFactory RenderingListenerFactory
}

var _ RenderingInput = &Customizer{}
var _ ResourceMarkings = &Customizer{}
var _ RenderingListener = &Customizer{}

// NewCustomizer creates a new Customizer object for the specified custom resource.  Effectively, this delegates
// to the factories on this object.  That said, it will inject a LoggingRenderingListener, an OwnerReferenceDecorator,
// and a MarkingsDecorator into a CompositeRenderingListener that includes the listener created by the
// RenderingListenerFactory.  This ensures the HelmReconciler can properly implement pruning, etc.
// instance is the custom resource to be processed by the HelmReconciler
func (f *CustomizerFactory) NewCustomizer(instance runtime.Object) (*Customizer, error) {
	input, err := f.InputFactory.NewRenderingInput(instance)
	if err != nil {
		return nil, err
	}
	markings, err := f.MarkingsFactory.NewResourceMarkings(instance)
	if err != nil {
		return nil, err
	}
	listener, err := f.ListenerFactory.NewRenderingListener(instance)
	if err != nil {
		return nil, err
	}
	ownerReferenceDecorator, err := NewOwnerReferenceDecorator(instance)
	if err != nil {
		return nil, err
	}
	return &Customizer{
		input:    input,
		markings: markings,
		listener: &CompositeRenderingListener{
			Listeners: []RenderingListener{
				&LoggingRenderingListener{Level: 1},
				ownerReferenceDecorator,
				NewMarkingsDecorator(markings),
				listener,
			},
		},
	}, nil
}

// Helpers

// SimpleResourceMarkings is a helper to implement ResourceMarkings from a known set of labels,
// annotations, and resource types.
type SimpleResourceMarkings struct {
	// OwnerLabels to be added to all rendered resources.
	OwnerLabels map[string]string
	// OwnerAnnotations to be added to all rendered resources.
	OwnerAnnotations map[string]string
	// NamespacedResources rendered by these charts
	NamespacedResources []schema.GroupVersionKind
	// NonNamespacedResources rendered by these charts
	NonNamespacedResources []schema.GroupVersionKind
}

var _ ResourceMarkings = &SimpleResourceMarkings{}

// GetOwnerLabels returns this.OwnerLabels
func (m *SimpleResourceMarkings) GetOwnerLabels() map[string]string {
	if m.OwnerLabels == nil {
		return map[string]string{}
	}
	return m.OwnerLabels
}

// GetOwnerAnnotations returns this.OwnerAnnotations
func (m *SimpleResourceMarkings) GetOwnerAnnotations() map[string]string {
	if m.OwnerAnnotations == nil {
		return map[string]string{}
	}
	return m.OwnerAnnotations
}

// GetResourceTypes returns this.NamespacedResources and this.NonNamespacedResources
func (m *SimpleResourceMarkings) GetResourceTypes() (namespaced []schema.GroupVersionKind, nonNamespaced []schema.GroupVersionKind) {
	return m.NamespacedResources, m.NonNamespacedResources
}

// RegisterReconciler registers the HelmReconciler with the RenderingListener
func (c *Customizer) RegisterReconciler(reconciler *HelmReconciler) {
	if reconcilerListener, ok := c.listener.(ReconcilerListener); ok {
		reconcilerListener.RegisterReconciler(reconciler)
	}
}

//
// RenderingInput
//

// GetChartPath simply delegates to input.GetChartPath
func (c *Customizer) GetChartPath() string {
	return c.input.GetChartPath()
}

// GetValues simply delegates to input.GetValues
func (c *Customizer) GetValues() map[string]interface{} {
	return c.input.GetValues()
}

// GetTargetNamespace simply delegates to input.GetTargetNamespace
func (c *Customizer) GetTargetNamespace() string {
	return c.input.GetTargetNamespace()
}

// GetProcessingOrder simply delegates to input.GetProcessingOrder
func (c *Customizer) GetProcessingOrder(manifests map[string][]manifest.Manifest) ([]string, error) {
	return c.input.GetProcessingOrder(manifests)
}

//
// ResourceMarkings
//

// GetOwnerLabels simply delegates to markings.GetOwnerLabels
func (c *Customizer) GetOwnerLabels() map[string]string {
	return c.markings.GetOwnerLabels()
}

// GetOwnerAnnotations simply delegates to markings.GetOwnerAnnotations
func (c *Customizer) GetOwnerAnnotations() map[string]string {
	return c.markings.GetOwnerAnnotations()
}

// GetResourceTypes simply delegates to markings.GetResourceTypes
func (c *Customizer) GetResourceTypes() (namespaced []schema.GroupVersionKind, nonNamespaced []schema.GroupVersionKind) {
	return c.markings.GetResourceTypes()
}

//
// RenderingListener
//

// BeginReconcile simply delegates to listener.BeginReconcile
func (c *Customizer) BeginReconcile(instance runtime.Object) error {
	return c.listener.BeginReconcile(instance)
}

// BeginDelete simply delegates to listener.BeginDelete
func (c *Customizer) BeginDelete(instance runtime.Object) error {
	return c.listener.BeginDelete(instance)
}

// BeginChart simply delegates to listener.BeginChart
func (c *Customizer) BeginChart(chart string, manifests []manifest.Manifest) ([]manifest.Manifest, error) {
	return c.listener.BeginChart(chart, manifests)
}

// BeginResource simply delegates to listener.BeginResource
func (c *Customizer) BeginResource(obj runtime.Object) (runtime.Object, error) {
	return c.listener.BeginResource(obj)
}

// ResourceCreated simply delegates to listener.ResourceCreated
func (c *Customizer) ResourceCreated(created runtime.Object) error {
	return c.listener.ResourceCreated(created)
}

// ResourceUpdated simply delegates to listener.ResourceUpdated
func (c *Customizer) ResourceUpdated(updated runtime.Object, old runtime.Object) error {
	return c.listener.ResourceUpdated(updated, old)
}

// ResourceDeleted simply delegates to listener.ResourceDeleted
func (c *Customizer) ResourceDeleted(deleted runtime.Object) error {
	return c.listener.ResourceDeleted(deleted)
}

// ResourceError simply delegates to listener.ResourceError
func (c *Customizer) ResourceError(obj runtime.Object, err error) error {
	return c.listener.ResourceError(obj, err)
}

// EndResource simply delegates to listener.EndResource
func (c *Customizer) EndResource(obj runtime.Object) error {
	return c.listener.EndResource(obj)
}

// EndChart simply delegates to listener.EndChart
func (c *Customizer) EndChart(chart string) error {
	return c.listener.EndChart(chart)
}

// BeginPrune simply delegates to listener.BeginPrune
func (c *Customizer) BeginPrune(all bool) error {
	return c.listener.BeginPrune(all)
}

// EndPrune simply delegates to listener.EndPrune
func (c *Customizer) EndPrune() error {
	return c.listener.EndPrune()
}

// EndDelete simply delegates to listener.EndDelete
func (c *Customizer) EndDelete(instance runtime.Object, err error) error {
	return c.listener.EndDelete(instance, err)
}

// EndReconcile simply delegates to listener.EndReconcile
func (c *Customizer) EndReconcile(instance runtime.Object, err error) error {
	return c.listener.EndReconcile(instance, err)
}

// DefaultChartCustomizerFactory is a factory for creating DefaultChartCustomizer objects
type DefaultChartCustomizerFactory struct {
	// ChartAnnotationKey is the key used to add an annotation identifying the chart that rendered the resource
	// to the rendered resource.
	ChartAnnotationKey string
}

var _ ChartCustomizerFactory = &DefaultChartCustomizerFactory{}

// NewChartCustomizer returns a new DefaultChartCustomizer for the specified chart.
func (f *DefaultChartCustomizerFactory) NewChartCustomizer(chartName string) ChartCustomizer {
	return NewDefaultChartCustomizer(chartName, f.ChartAnnotationKey)
}

// DefaultChartCustomizerListener manages ChartCustomizer objects for a rendering.
type DefaultChartCustomizerListener struct {
	*DefaultRenderingListener
	// ChartCustomizerFactory is the factory used to create ChartCustomizer objects for each chart
	// encountered during rendering.
	ChartCustomizerFactory ChartCustomizerFactory
	// ChartAnnotationKey represents the annotation key in which the chart name is stored on the rendered resource.
	ChartAnnotationKey string
	reconciler         *HelmReconciler
	customizers        map[string]ChartCustomizer
	customizer         ChartCustomizer
}

var _ RenderingListener = &DefaultChartCustomizerListener{}
var _ ReconcilerListener = &DefaultChartCustomizerListener{}

// NewDefaultChartCustomizerListener creates a new DefaultChartCustomizerListener which creates DefaultChartCustomizer
// objects for each chart (which simply adds a chart owner annotation to each rendered resource).
// The ChartCustomizerFactory may be modified by users to create custom ChartCustomizer objects.
func NewDefaultChartCustomizerListener(chartAnnotationKey string) *DefaultChartCustomizerListener {
	return &DefaultChartCustomizerListener{
		DefaultRenderingListener: &DefaultRenderingListener{},
		ChartCustomizerFactory:   &DefaultChartCustomizerFactory{ChartAnnotationKey: chartAnnotationKey},
		ChartAnnotationKey:       chartAnnotationKey,
		customizers:              map[string]ChartCustomizer{},
	}
}

// RegisterReconciler registers the HelmReconciler with the listener.
func (l *DefaultChartCustomizerListener) RegisterReconciler(reconciler *HelmReconciler) {
	l.reconciler = reconciler
}

// BeginChart creates a new ChartCustomizer for the specified chart and delegates listener calls applying to resources
// (e.g. BeginResource) to the customizer up through EndChart.
func (l *DefaultChartCustomizerListener) BeginChart(chartName string, manifests []manifest.Manifest) ([]manifest.Manifest, error) {
	l.customizer = l.GetOrCreateCustomizer(chartName)
	return l.customizer.BeginChart(chartName, manifests)
}

// BeginResource delegates to the active ChartCustomizer's BeginResource
func (l *DefaultChartCustomizerListener) BeginResource(obj runtime.Object) (runtime.Object, error) {
	if l.customizer == nil {
		// XXX: what went wrong
		// this should actually be a warning
		l.reconciler.GetLogger().Info("no active chart customizer")
		return obj, nil
	}
	return l.customizer.BeginResource(obj)
}

// ResourceCreated delegates to the active ChartCustomizer's ResourceCreated
func (l *DefaultChartCustomizerListener) ResourceCreated(created runtime.Object) error {
	if l.customizer == nil {
		// XXX: what went wrong
		// this should actually be a warning
		l.reconciler.GetLogger().Info("no active chart customizer")
		return nil
	}
	return l.customizer.ResourceCreated(created)
}

// ResourceUpdated delegates to the active ChartCustomizer's ResourceUpdated
func (l *DefaultChartCustomizerListener) ResourceUpdated(updated runtime.Object, old runtime.Object) error {
	if l.customizer == nil {
		// XXX: what went wrong
		// this should actually be a warning
		l.reconciler.GetLogger().Info("no active chart customizer")
		return nil
	}
	return l.customizer.ResourceUpdated(updated, old)
}

// ResourceError delegates to the active ChartCustomizer's ResourceError
func (l *DefaultChartCustomizerListener) ResourceError(obj runtime.Object, err error) error {
	if l.customizer == nil {
		// XXX: what went wrong
		// this should actually be a warning
		l.reconciler.GetLogger().Info("no active chart customizer")
		return nil
	}
	return l.customizer.ResourceError(obj, err)
}

// EndResource delegates to the active ChartCustomizer's EndResource
func (l *DefaultChartCustomizerListener) EndResource(obj runtime.Object) error {
	if l.customizer == nil {
		// XXX: what went wrong
		// this should actually be a warning
		l.reconciler.GetLogger().Info("no active chart customizer")
		return nil
	}
	return l.customizer.EndResource(obj)
}

// EndChart delegates to the active ChartCustomizer's EndChart and resets the active ChartCustomizer to nil.
func (l *DefaultChartCustomizerListener) EndChart(chartName string) error {
	if l.customizer == nil {
		return nil
	}
	err := l.customizer.EndChart(chartName)
	l.customizer = nil
	return err
}

// ResourceDeleted looks up the ChartCustomizer for the object that was deleted and invokes its ResourceDeleted method.
func (l *DefaultChartCustomizerListener) ResourceDeleted(deleted runtime.Object) error {
	if chartName, ok := common.GetAnnotation(deleted, l.ChartAnnotationKey); ok && len(chartName) > 0 {
		customizer := l.GetOrCreateCustomizer(chartName)
		return customizer.ResourceDeleted(deleted)
	}
	return nil
}

// GetOrCreateCustomizer does what it says.
func (l *DefaultChartCustomizerListener) GetOrCreateCustomizer(chartName string) ChartCustomizer {
	var ok bool
	var customizer ChartCustomizer
	if customizer, ok = l.customizers[chartName]; !ok {
		customizer = l.ChartCustomizerFactory.NewChartCustomizer(chartName)
		if reconcilerListener, ok := customizer.(ReconcilerListener); ok {
			reconcilerListener.RegisterReconciler(l.reconciler)
		}
		l.customizers[chartName] = customizer
	}
	return customizer
}

// DefaultChartCustomizer is a ChartCustomizer that collects resources created/deleted during rendering and adds
// a chart annotation to rendered resources.
type DefaultChartCustomizer struct {
	ChartName              string
	ChartAnnotationKey     string
	Reconciler             *HelmReconciler
	NewResourcesByKind     map[string][]runtime.Object
	DeletedResourcesByKind map[string][]runtime.Object
}

var _ ChartCustomizer = &DefaultChartCustomizer{}

// NewDefaultChartCustomizer creates a new DefaultChartCustomizer
func NewDefaultChartCustomizer(chartName, chartAnnotationKey string) *DefaultChartCustomizer {
	return &DefaultChartCustomizer{
		ChartName:              chartName,
		ChartAnnotationKey:     chartAnnotationKey,
		NewResourcesByKind:     map[string][]runtime.Object{},
		DeletedResourcesByKind: map[string][]runtime.Object{},
	}
}

// RegisterReconciler registers the HelmReconciler with this.
func (c *DefaultChartCustomizer) RegisterReconciler(reconciler *HelmReconciler) {
	c.Reconciler = reconciler
}

// BeginChart empty implementation
func (c *DefaultChartCustomizer) BeginChart(chart string, manifests []manifest.Manifest) ([]manifest.Manifest, error) {
	return manifests, nil
}

// BeginResource adds the chart annotation to the resource (ChartAnnotationKey=ChartName)
func (c *DefaultChartCustomizer) BeginResource(obj runtime.Object) (runtime.Object, error) {
	if len(c.ChartName) > 0 && len(c.ChartAnnotationKey) > 0 {
		common.SetAnnotation(obj, c.ChartAnnotationKey, c.ChartName)
	}
	return obj, nil
}

// ResourceCreated adds the created object to NewResourcesByKind
func (c *DefaultChartCustomizer) ResourceCreated(created runtime.Object) error {
	kind := created.GetObjectKind().GroupVersionKind().Kind
	objects, ok := c.NewResourcesByKind[kind]
	if !ok {
		objects = []runtime.Object{}
	}
	c.NewResourcesByKind[kind] = append(objects, created)
	return nil
}

// ResourceUpdated adds the updated object to NewResourcesByKind
func (c *DefaultChartCustomizer) ResourceUpdated(updated, old runtime.Object) error {
	kind := updated.GetObjectKind().GroupVersionKind().Kind
	objects, ok := c.NewResourcesByKind[kind]
	if !ok {
		objects = []runtime.Object{}
	}
	c.NewResourcesByKind[kind] = append(objects, updated)
	return nil
}

// ResourceDeleted adds the deleted object to DeletedResourcesByKind
func (c *DefaultChartCustomizer) ResourceDeleted(deleted runtime.Object) error {
	kind := deleted.GetObjectKind().GroupVersionKind().Kind
	objects, ok := c.DeletedResourcesByKind[kind]
	if !ok {
		objects = []runtime.Object{}
	}
	c.DeletedResourcesByKind[kind] = append(objects, deleted)
	return nil
}

// ResourceError empty implementation
func (c *DefaultChartCustomizer) ResourceError(obj runtime.Object, err error) error {
	return nil
}

// EndResource empty implementation
func (c *DefaultChartCustomizer) EndResource(obj runtime.Object) error {
	return nil
}

// EndChart empty implementation
func (c *DefaultChartCustomizer) EndChart(chart string) error {
	return nil
}
