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

package istiocontrolplane

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"istio.io/operator/pkg/apis/istio/v1alpha2"
	"istio.io/operator/pkg/validate"
)

// +kubebuilder:webhook:path=/validate-v1alpha2-istiocontrolplane,mutating=true,failurePolicy=fail,
// groups="install.istio.io",resources=IstioControlPlane,verbs=create;update,versions=v1alpha2,name=mistiocontrolplane.kb.io

// podAnnotator annotates Pods
type IscpValidator struct {
	client  client.Client
	decoder *admission.Decoder
}

// iscpValidator validates created IstioControlPlane CR.
func (a *IscpValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	icp := &v1alpha2.IstioControlPlane{}
	err := a.decoder.Decode(req, icp)
	if err != nil {
		return admission.Denied(err.Error())
	}
	//TODO: update to full validation including values part after values schema formalized.
	if errs := validate.CheckIstioControlPlaneSpecExcludeValues(icp.Spec, false); len(errs) != 0 {
		fmt.Printf("proceed with validation err: %v", errs.Error())
		// TODO: allow the request now until we fully done the validation logic.
		return admission.Allowed("IstioControlPlane schema validated with err")
	}
	return admission.Allowed("IstioControlPlane schema validated")
}

// InjectClient injects the client.
func (a *IscpValidator) InjectClient(c client.Client) error {
	a.client = c
	return nil
}

// InjectDecoder injects the decoder.
func (a *IscpValidator) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}
