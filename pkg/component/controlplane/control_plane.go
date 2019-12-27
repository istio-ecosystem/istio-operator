// Copyright 2019 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controlplane

import (
	"fmt"

	"istio.io/api/operator/v1alpha1"
	"istio.io/operator/pkg/component/component"
	"istio.io/operator/pkg/name"
	"istio.io/operator/pkg/translate"
	"istio.io/operator/pkg/util"
)

// IstioControlPlane is an installation of an Istio control plane.
type IstioControlPlane struct {
	// installSpec is the installation spec for the control plane.
	installSpec *v1alpha1.IstioOperatorSpec
	// translator is the translator for this feature.
	translator *translate.Translator
	// components is a slice of components that are part of the feature.
	components []component.IstioComponent
	started    bool
}

// NewIstioControlPlane creates a new IstioControlPlane and returns a pointer to it.
func NewIstioControlPlane(installSpec *v1alpha1.IstioOperatorSpec, translator *translate.Translator) (*IstioControlPlane, error) {
	out := &IstioControlPlane{}
	opts := &component.Options{
		InstallSpec: installSpec,
		Translator:  translator,
	}
	for _, c := range name.AllCoreComponentNames {
		o := *opts
		ns, err := name.Namespace(c, installSpec)
		if err != nil {
			return nil, err
		}
		o.Namespace = ns
		out.components = append(out.components, component.NewComponent(c, &o))
	}
	for _, g := range installSpec.Components.IngressGateways {
		o := *opts
		o.Namespace = g.Namespace
		out.components = append(out.components, component.NewIngressComponent(g.Name, &o))
	}
	for _, g := range installSpec.Components.EgressGateways {
		o := *opts
		o.Namespace = g.Namespace
		out.components = append(out.components, component.NewEgressComponent(g.Name, &o))
	}
	for c := range installSpec.AddonComponents {
		rn := ""
		// For well-known addon components like Prometheus, the resource names are included
		// in the translations.
		if cm := translator.ComponentMap(c); cm != nil {
			rn = cm.ResourceName
		}
		out.components = append(out.components, component.NewAddonComponent(c, rn, opts))
	}
	return out, nil
}

// Run starts the Istio control plane.
func (i *IstioControlPlane) Run() error {
	for _, c := range i.components {
		if err := c.Run(); err != nil {
			return err
		}
	}
	i.started = true
	return nil
}

// RenderManifest returns a manifest rendered against
func (i *IstioControlPlane) RenderManifest() (manifests name.ManifestMap, errsOut util.Errors) {
	if !i.started {
		return nil, util.NewErrs(fmt.Errorf("istioControlPlane must be Run before calling RenderManifest"))
	}

	manifests = make(name.ManifestMap)
	for _, c := range i.components {
		ms, err := c.RenderManifest()
		errsOut = util.AppendErr(errsOut, err)
		manifests[c.ComponentName()] = ms
	}
	if len(errsOut) > 0 {
		return nil, errsOut
	}
	return
}
