export GO111MODULE=on
ifeq ($(BUILD_WITH_CONTAINER),0)
override GOBIN := $(GOPATH)/bin
endif

pwd := $(shell pwd)

# make targets
.PHONY: lint test_with_coverage mandiff build fmt vfsgen update-charts

lint:
	@scripts/check_license.sh
	@golangci-lint run -j 8 -v ./...

test:
	@go test -race ./...

test_with_coverage:
	@go test -race -coverprofile=coverage.txt -covermode=atomic ./...
	@curl -s https://codecov.io/bash | bash -s -- -c -F aFlag -f coverage.txt

mandiff: update-charts
	@PATH=${PATH}:${GOPATH}/bin scripts/run_mandiff.sh

build: mesh

fmt:
	@goimports -w -local "istio.io" $(shell find . -type f -name '*.go' ! -name '*.gen.go' ! -name '*.pb.go' )

update-charts: installer.sha
	@scripts/run_update_charts.sh `cat installer.sha`

# make target dependencies
vfsgen: data/ update-charts
	go generate ./...

########################
# protoc_gen_gogo*
########################

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

########################
# protoc_gen_docs
########################

gen_doc_iscp: get_dep_proto
	protoc -I/tmp/src -I$(pwd) -I/usr/include/protobuf --docs_out=warnings=true,emit_yaml=true,mode=html_page:$(pwd) pkg/apis/istio/v1alpha2/istiocontrolplane_types.proto

gen_doc_values: get_dep_proto
	protoc -I/tmp/src -I$(pwd) -I/usr/include/protobuf --docs_out=warnings=true,emit_yaml=true,mode=html_page:$(pwd) pkg/apis/istio/v1alpha2/values/values_types.proto

default: mesh

generate-api-go: $(api_pb_gos)
	patch pkg/apis/istio/v1alpha2/istiocontrolplane_types.pb.go < pkg/apis/istio/v1alpha2/fixup_go_structs.patch

$(api_pb_gos): $(api_protos)
	@protoc $(gogofast_plugin) $^

clean-proto:
	rm -f $(api_pb_gos)

# is all of this needed or is this cruft?

get_dep_proto:
	GO111MODULE=off GOPATH=/tmp go get k8s.io/api/core/v1 k8s.io/api/autoscaling/v2beta1 k8s.io/apimachinery/pkg/apis/meta/v1/

proto_iscp: get_dep_proto
	protoc -I=/tmp/src -I$(pwd) -I=/usr/include/protobuf --gogofast_out=$(pwd) pkg/apis/istio/v1alpha2/istiocontrolplane_types.proto
	sed -i -e 's|github.com/gogo/protobuf/protobuf/google/protobuf|github.com/gogo/protobuf/types|g' pkg/apis/istio/v1alpha2/istiocontrolplane_types.pb.go
	patch pkg/apis/istio/v1alpha2/istiocontrolplane_types.pb.go < pkg/apis/istio/v1alpha2/fixup_go_structs.patch

proto_values: get_dep_proto
	protoc -I=/tmp/src -I$(pwd) -I=/usr/include/protobuf --go_out=$(pwd) pkg/apis/istio/v1alpha2/values/values_types.proto
	sed -i -e 's|github.com/gogo/protobuf/protobuf/google/protobuf|github.com/gogo/protobuf/types|g' pkg/apis/istio/v1alpha2/values/values_types.pb.go
	patch pkg/apis/istio/v1alpha2/values/values_types.pb.go < pkg/apis/istio/v1alpha2/values/fix_values_structs.patch

mesh: vfsgen
	go build -o ${GOBIN}/mesh ./cmd/mesh.go

include Makefile.common.mk
