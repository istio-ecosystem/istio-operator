lint:
	@scripts/check_license.sh
	@scripts/run_golangci.sh

fmt:
	@scripts/run_gofmt.sh

# TODO: this needs to be cleaned up and possibly moved out to istio/api
proto:
	protoc -I./vendor -I./vendor/github.com/gogo/protobuf/protobuf -I./pkg/apis/istio/v1alpha2/ --proto_path=pkg/apis/istio/v1alpha2/ --gofast_out=pkg/apis/istio/v1alpha2/ pkg/apis/istio/v1alpha2/istiocontrolplane_types.proto
	sed -i -e 's|github.com/gogo/protobuf/protobuf/google/protobuf|github.com/gogo/protobuf/types|g' pkg/apis/istio/v1alpha2/istiocontrolplane_types.pb.go
	go run ~/go/src/k8s.io/code-generator/cmd/deepcopy-gen/main.go -O zz_generated.deepcopy -i ./pkg/apis/istio/v1alpha2/... -i ./vendor/github.com/gogo/protobuf/types/...
	patch pkg/apis/istio/v1alpha2/istiocontrolplane_types.pb.go < pkg/apis/istio/v1alpha2/fixup_go_structs.patch

# Note: must add // +k8s:deepcopy-gen=package to doc.go in ./vendor/github.com/gogo/protobuf/types/ for types package
proto_gogo:
	go run ~/go/src/k8s.io/code-generator/cmd/deepcopy-gen/main.go -v 5 -O zz_generated.deepcopy -i ./vendor/github.com/gogo/protobuf/types/...

gen_patch:
	diff -u pkg/apis/istio/v1alpha2/istiocontrolplane_types.pb.go.orig pkg/apis/istio/v1alpha2/istiocontrolplane_types.pb.go > pkg/apis/istio/v1alpha2/fixup_go_structs.patch || true

include Makefile.common.mk
