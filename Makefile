gen_img := gcr.io/istio-testing/protoc:2019-03-29
pwd := $(shell pwd)
mount_dir := /src
uid := $(shell id -u)
docker_gen := docker run --rm --user $(uid) -v /etc/passwd:/etc/passwd:ro -v $(pwd):$(mount_dir) -w $(mount_dir) $(gen_img) -I.
out_path = .

########################
# protoc_gen_gogo*
########################

#gofast_plugin_prefix := --gofast_out=plugins=grpc,
gogofast_plugin_prefix := --gogofast_out=plugins=grpc,

comma := ,
empty:=
space := $(empty) $(empty)

importmaps := \
	gogoproto/gogo.proto=github.com/gogo/protobuf/gogoproto \
	google/protobuf/any.proto=github.com/gogo/protobuf/types \
	google/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor \
	google/protobuf/duration.proto=github.com/gogo/protobuf/types \
	google/protobuf/struct.proto=github.com/gogo/protobuf/types \
	google/protobuf/timestamp.proto=github.com/gogo/protobuf/types \
	google/protobuf/wrappers.proto=github.com/gogo/protobuf/types \
	google/rpc/status.proto=github.com/gogo/googleapis/google/rpc \
	google/rpc/code.proto=github.com/gogo/googleapis/google/rpc \
	google/rpc/error_details.proto=github.com/gogo/googleapis/google/rpc \

# generate mapping directive with M<proto>:<go pkg>, format for each proto file
mapping_with_spaces := $(foreach map,$(importmaps),M$(map),)
gogo_mapping := $(subst $(space),$(empty),$(mapping_with_spaces))

#gofast_plugin := $(gofast_plugin_prefix)$(gogo_mapping):$(out_path)
gogofast_plugin := $(gogofast_plugin_prefix)$(gogo_mapping):$(out_path)

#####################
# Generation Rules
#####################

api_path := pkg/apis/istio/v1alpha2
api_protos := $(shell find $(api_path) -type f -name '*.proto' | sort)
api_pb_gos := $(api_protos:.proto=.pb.go)

generate-api-go: $(api_pb_gos)
	patch pkg/apis/istio/v1alpha2/istiocontrolplane_types.pb.go < pkg/apis/istio/v1alpha2/fixup_go_structs.patch

$(api_pb_gos): $(api_protos)
	@$(docker_gen) $(gogofast_plugin) $^

clean-proto:
	rm -f $(api_pb_gos)

lint:
	@scripts/check_license.sh
	@scripts/run_golangci.sh

fmt:
	@scripts/run_gofmt.sh

include Makefile.common.mk

# Coverage tests
coverage:
	scripts/codecov.sh

# get imported protos to $GOPATH
get_dep_proto:
	GO111MODULE=off go get k8s.io/api/core/v1 k8s.io/api/autoscaling/v2beta1 k8s.io/apimachinery/pkg/apis/meta/v1/ github.com/gogo/protobuf/...

proto:
	protoc -I=${GOPATH}/src -I./pkg/apis/istio/v1alpha2/ --proto_path=pkg/apis/istio/v1alpha2/ --gofast_out=pkg/apis/istio/v1alpha2/ pkg/apis/istio/v1alpha2/istiocontrolplane_types.proto
	sed -i -e 's|github.com/gogo/protobuf/protobuf/google/protobuf|github.com/gogo/protobuf/types|g' pkg/apis/istio/v1alpha2/istiocontrolplane_types.pb.go
	patch pkg/apis/istio/v1alpha2/istiocontrolplane_types.pb.go < pkg/apis/istio/v1alpha2/fixup_go_structs.patch

proto_with_setup: get_dep_proto proto

gen_patch:
	diff -u pkg/apis/istio/v1alpha2/istiocontrolplane_types.pb.go.orig pkg/apis/istio/v1alpha2/istiocontrolplane_types.pb.go > pkg/apis/istio/v1alpha2/fixup_go_structs.patch || true

proto_orig:
	protoc -I=${GOPATH}/src -I./pkg/apis/istio/v1alpha2/ --proto_path=pkg/apis/istio/v1alpha2/ --gofast_out=pkg/apis/istio/v1alpha2/ pkg/apis/istio/v1alpha2/istiocontrolplane_types.proto
	sed -i -e 's|github.com/gogo/protobuf/protobuf/google/protobuf|github.com/gogo/protobuf/types|g' pkg/apis/istio/v1alpha2/istiocontrolplane_types.pb.go
	cp pkg/apis/istio/v1alpha2/istiocontrolplane_types.proto pkg/apis/istio/v1alpha2/istiocontrolplane_types.proto.orig

proto_orig_with_setup: get_dep_proto proto_orig

vfsgen:
	go generate ./cmd/iop.go
