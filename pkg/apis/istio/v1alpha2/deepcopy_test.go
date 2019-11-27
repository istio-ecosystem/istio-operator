package v1alpha2

import (
	"github.com/gogo/protobuf/types"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"reflect"
	"testing"
	"time"
)

func TestDeepCopy(t *testing.T) {
	now := meta.NewTime(time.Now().Truncate(time.Second))
	icp := &IstioControlPlane{
		ObjectMeta: meta.ObjectMeta{
			Name:                       "name",
			GenerateName:               "generateName",
			Namespace:                  "namespace",
			SelfLink:                   "selfLink",
			UID:                        "uid",
			ResourceVersion:            "resourceVersion",
			Generation:                 1,
			CreationTimestamp:          now,
			DeletionTimestamp:          &now,
			DeletionGracePeriodSeconds: pointer.Int64Ptr(15),
			Labels: map[string]string{
				"label": "value",
			},
			Annotations: map[string]string{
				"annotation": "value",
			},
			OwnerReferences: []meta.OwnerReference{
				{
					APIVersion:         "v1",
					Kind:               "Foo",
					Name:               "foo",
					UID:                "123",
					Controller:         pointer.BoolPtr(true),
					BlockOwnerDeletion: pointer.BoolPtr(true),
				},
			},
			Finalizers:  []string{"finalizer"},
			ClusterName: "cluster",
		},
		Spec: &IstioControlPlaneSpec{
			Cni: &CNIFeatureSpec{
				Enabled: &BoolValueForPB{types.BoolValue{Value: true}},
			},
			Profile: "profile",
			Hub:     "hub",
			Tag:     "tag",
		},
	}

	icp2 := icp.DeepCopy()

	if !reflect.DeepEqual(icp, icp2) {
		t.Fatalf("Expected IstioControlPlanes to be equal, but they weren't.\n"+
			"  Expected: %+v,\n"+
			"       got: %+v", *icp, *icp2)
	}
}
