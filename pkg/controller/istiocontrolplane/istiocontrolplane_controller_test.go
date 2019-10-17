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
	"strconv"
	"testing"
	"time"

	"github.com/onsi/gomega"

	"istio.io/operator/pkg/apis/istio/v1alpha2"
	"istio.io/operator/pkg/helmreconciler"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	timeout        = time.Second * 2
	defaultProfile = "default"
	demoProfile    = "demo"
	minimalProfile = "minimal"
	sdsProfile     = "sds"
)

var (
	c      client.Client
	icpkey = types.NamespacedName{
		Name:      "test-istiocontrolplane",
		Namespace: "test-istio-operator",
	}
	expectedRequest = reconcile.Request{NamespacedName: icpkey}

	minimalStatus = map[string]*v1alpha2.InstallStatus_VersionStatus{
		"Pilot": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
		"crds": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
	}
	defaultStatus = map[string]*v1alpha2.InstallStatus_VersionStatus{
		"crds": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
		"Pilot": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
		"Policy": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
		"Telemetry": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
		"Injector": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
		"Citadel": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
		"Galley": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
		"Prometheus": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
		"IngressGateway": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
	}
	demoStatus = map[string]*v1alpha2.InstallStatus_VersionStatus{
		"crds": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
		"Pilot": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
		"Policy": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
		"Telemetry": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
		"Injector": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
		"Citadel": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
		"Galley": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
		"Prometheus": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
		"IngressGateway": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
		"Grafana": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
		"Kiali": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
		"Tracing": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
		"EgressGateway": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
	}
	sdsStatus = map[string]*v1alpha2.InstallStatus_VersionStatus{
		"crds": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
		"Pilot": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
		"Policy": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
		"Telemetry": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
		"Injector": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
		"Citadel": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
		"Galley": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
		"Prometheus": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
		"IngressGateway": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
		"NodeAgent": {
			Status: v1alpha2.InstallStatus_HEALTHY,
		},
	}
)

type testCase struct {
	description    string
	initialProfile string
	targetProfile  string
}

// TestICPController_SwitchProfile
func TestICPController_SwitchProfile(t *testing.T) {
	cases := []testCase{
		{
			description:    "switch profile from minimal to default",
			initialProfile: minimalProfile,
			targetProfile:  defaultProfile,
		},
		{
			description:    "switch profile from default to minimal",
			initialProfile: defaultProfile,
			targetProfile:  minimalProfile,
		},
		{
			description:    "switch profile from default to demo",
			initialProfile: defaultProfile,
			targetProfile:  demoProfile,
		},
		{
			description:    "switch profile from demo to sds",
			initialProfile: demoProfile,
			targetProfile:  sdsProfile,
		},
		{
			description:    "switch profile from sds to default",
			initialProfile: sdsProfile,
			targetProfile:  defaultProfile,
		},
	}
	for i, c := range cases {
		t.Run(strconv.Itoa(i)+":"+c.description, func(t *testing.T) {
			testSwitchProfile(t, c)
		})
	}
}

func testSwitchProfile(t *testing.T, c testCase) {
	t.Helper()
	icp := &v1alpha2.IstioControlPlane{
		Kind:       "IstioControlPlane",
		ApiVersion: "install.istio.io/v1alpha2",
		ObjectMeta: metav1.ObjectMeta{
			Name:      icpkey.Name,
			Namespace: icpkey.Namespace,
		},
		Spec: &v1alpha2.IstioControlPlaneSpec{
			Profile: c.initialProfile,
		},
	}
	objs := []runtime.Object{
		icp,
	}

	s := scheme.Scheme
	s.AddKnownTypes(v1alpha2.SchemeGroupVersion, icp)
	cl := fake.NewFakeClientWithScheme(s, objs...)
	factory := &helmreconciler.Factory{CustomizerFactory: &IstioRenderingCustomizerFactory{}}
	r := &ReconcileIstioControlPlane{client: cl, scheme: s, factory: factory}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      icpkey.Name,
			Namespace: icpkey.Namespace,
		},
	}
	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	// check ICP status
	succeed, err := checkICPStatus(cl, req.NamespacedName, c.initialProfile)
	if !succeed || err != nil {
		t.Fatalf("failed to get expected IstioControlPlane status: (%v)", err)
	}

	//update IstioControlPlane : switch profile from minimal to default and reconcile
	err = switchIstioControlPlaneProfile(cl, req.NamespacedName, c.targetProfile)
	if err != nil {
		t.Fatalf("failed to update IstioControlPlane: (%v)", err)
	}
	res, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	if res.Requeue {
		t.Error("reconcile requeue which is not expected")
	}
	// check ICP status
	succeed, err = checkICPStatus(cl, req.NamespacedName, c.targetProfile)
	if !succeed || err != nil {
		t.Fatalf("failed to get expected IstioControlPlane status: (%v)", err)
	}
}

func statusExpected(s1, s2 *v1alpha2.InstallStatus_VersionStatus) bool {
	return s1.Status.String() == s2.Status.String()
}

func switchIstioControlPlaneProfile(cl client.Client, key client.ObjectKey, profile string) error {
	instance := &v1alpha2.IstioControlPlane{}
	err := cl.Get(context.TODO(), key, instance)
	if err != nil {
		return err
	}
	instance.Spec.Profile = profile
	err = cl.Update(context.TODO(), instance)
	if err != nil {
		return err
	}
	return nil
}
func checkICPStatus(cl client.Client, key client.ObjectKey, profile string) (bool, error) {

	instance := &v1alpha2.IstioControlPlane{}
	err := cl.Get(context.TODO(), key, instance)
	if err != nil {
		return false, err
	}
	var status map[string]*v1alpha2.InstallStatus_VersionStatus
	switch profile {
	case minimalProfile:
		status = minimalStatus
	case defaultProfile:
		status = defaultStatus
	case sdsProfile:
		status = sdsStatus
	case demoProfile:
		status = demoStatus
	}
	installStatus := instance.GetStatus()
	size := len(installStatus.Status)
	expectedSize := len(status)
	if size != expectedSize {
		return false, fmt.Errorf("status size(%v) is not equal to expected status size (%v)", size, expectedSize)
	}
	for k, v := range installStatus.Status {
		if s, ok := status[k]; ok {
			if !statusExpected(s, v) {
				return false, fmt.Errorf("failed to get Expected IstioControlPlane status: (%s)", k)
			}
		} else {
			return false, fmt.Errorf("failed to find Expected IstioControlPlane status: (%s)", k)
		}
	}
	return true, nil
}

// TestReconcile test the reconciler process with manager from end to end.
func TestReconcile(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// Setup the Manager and Controller. Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())

	c = mgr.GetClient()

	rec := newReconciler(mgr)
	recFn, requests := SetupTestReconcile(rec)

	g.Expect(add(mgr, recFn)).NotTo(gomega.HaveOccurred())

	stopMgr, mgrStopped := StartTestManager(mgr, g)

	defer func() {
		close(stopMgr)
		mgrStopped.Wait()
	}()

	icp := &v1alpha2.IstioControlPlane{
		Kind:       "IstioControlPlane",
		ApiVersion: "install.istio.io/v1alpha2",
		ObjectMeta: metav1.ObjectMeta{
			Name:      icpkey.Name,
			Namespace: icpkey.Namespace,
		},
		Spec: &v1alpha2.IstioControlPlaneSpec{
			Profile: defaultProfile,
		},
	}

	g.Expect(c.Create(context.TODO(), icp)).NotTo(gomega.HaveOccurred())

	defer c.Delete(context.TODO(), icp)

	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
}
