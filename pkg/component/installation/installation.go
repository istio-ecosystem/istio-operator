package installation

import (
	"github.com/ostromart/istio-installer/pkg/component/feature"
	"github.com/ostromart/istio-installer/pkg/util"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"strings"

	installerv1alpha1 "github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
)

type Installation interface {
	RenderManifest() (string, util.Errors)
}
type InstallationImpl struct {
	features []feature.Feature
}

func NewInstallation(install *installerv1alpha1.InstallSpec, helmChartDir string) Installation {
	fo := &feature.FeatureOptions{

	}
	return &InstallationImpl{
		features: []feature.Feature{
			feature.NewTrafficManagementFeature(fo),
			feature.NewSecurityFeature(install),
			feature.NewPolicyFeature(install),
			feature.NewTelemetryFeature(install),
			feature.NewConfigManagementFeature(install),
		},
	}
}

func (i InstallationImpl) RenderManifest() (manifest string, errsOut util.Errors) {
	var sb strings.Builder
	for _, f := range i.features {
		s, errs := f.RenderManifest()
		errsOut = util.AppendErrs(errsOut, errs)
		_, err := sb.WriteString(s)
		errsOut = util.AppendErr(errsOut, err)
	}
	if len(errsOut) > 0 {
		return "", errsOut
	}
	return sb.String(), nil
}
