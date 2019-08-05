package validation

import (
	"reflect"
	"testing"

	"istio.io/operator/pkg/apis/istio/v1alpha2"
)

func makeBoolPtr(v bool) *bool {
	return &v
}
func makeStringPtr(v string) *string {
	return &v
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name       string
		toValidate *v1alpha2.Values
	}{
		{
			name:       "Empty struct",
			toValidate: &v1alpha2.Values{},
		},
		{
			name: "With CNI defined",
			toValidate: &v1alpha2.Values{
				CNI: &v1alpha2.CNIConfig{
					Enabled: makeBoolPtr(true),
				},
			},
		},
		{
			name: "With Slice",
			toValidate: &v1alpha2.Values{
				Gateways: &v1alpha2.GatewaysConfig{
					Enabled: makeBoolPtr(true),
					EgressGateway: &v1alpha2.EgressGatewayConfig{
						Ports: []*v1alpha2.PortsConfig{
							{
								Name: makeStringPtr("port1"),
							},
							{
								Name: makeStringPtr("port2"),
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		err := validateSubTypes(reflect.ValueOf(tt.toValidate).Elem(), false, tt.toValidate, nil)
		if len(err) != 0 {
			t.Fatalf("Test %s failed with errors: %+v", tt.name, err)
		}
	}
}
