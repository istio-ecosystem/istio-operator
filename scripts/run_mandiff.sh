#!/bin/bash

# Copyright 2019 Istio Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

WD=$(dirname "$0")
WD=$(cd "$WD"; pwd)
ROOT=$(dirname "$WD")
OUT=${OUT:-$ROOT/out}

# No unset vars, print commands as they're executed, and exit on any non-zero
# return code
set -u
set -x
set -e

cd "${ROOT}"

export GO111MODULE=on
# build the istio operator binary
go build -o $GOPATH/bin/iop ./cmd/iop.go

# download the helm binary
${ROOT}/bin/init_helm.sh

function helm_render_template() {
    local namespace=${1}
    shift
    local relase=${1}
    shift
    local chart=${1}
    shift
    local cfg=${1}
    shift

    # render the chart with helm template
    helm template --namespace $namespace --name $relase $chart $cfg $*
}

# render all the templates with helm template.
function helm_manifest() {
    local namespace=${1}
    shift
    local relase=${1}
    shift
    local chart=${1}
    shift
    local profile=${1}
    shift

    # the global settings are the default for the chart
    local cfg="-f ${chart}/global.yaml"

    # the specified profile will override the gloal settings
    if [ -f "${ROOT}/tests/profiles/helm/values-istio-${profile}.yaml" ]; then
        cfg="${cfg} -f ${ROOT}/tests/profiles/helm/values-istio-${profile}.yaml"
    else
        echo "Please verify the values file for ${profile} path exists."
        exit 1
    fi

    # create parent directory for the manifests rendered by helm template
    mkdir -p ${OUT}/helm-template/istio-${namespace}-${relase}-${profile}

    local charts=${profile_charts_map[${profile}]}
    for c in $(echo $charts | tr " " "\n")
    do
        # create the parent directory if it doesn't exist.
        mkdir -p $(dirname ${OUT}/helm-template/istio-${namespace}-${relase}-${profile}/${c}.yaml)
        helm_render_template ${namespace} ${relase} ${chart}/${c} ${cfg} $* > ${OUT}/helm-template/istio-${namespace}-${relase}-${profile}/${c}.yaml
    done
}

# render all the templates with iop manifest.
function iop_manifest() {
    local profile=${1}
    shift

    # check the specified profile CR
    if [ -f "${ROOT}/tests/profiles/iop/${profile}-profile.yaml" ]; then
        mkdir -p ${OUT}/iop-manifest/istio-${profile}
        iop manifest --filename "${ROOT}/tests/profiles/iop/${profile}-profile.yaml" --dry-run=false --output ${OUT}/iop-manifest/istio-${profile} 2>&1
    else
        echo "Please verify the input IstioIstall CR path exists."
        exit 1
    fi
}

# compare the manifests generated by the helm template and iop manifest
function iop_mandiff_with_profile() {
    local profile=${1}

    helm_manifest ${ISTIO_SYSTEM_NS} ${ISTIO_RELEASE} "${ROOT}/data/charts" ${profile}
    iop_manifest ${profile}

    if [ -d "${OUT}/helm-template/istio-${ISTIO_SYSTEM_NS}-${ISTIO_RELEASE}-${profile}" ] && [ -d "${OUT}/iop-manifest/istio-${profile}" ]; then
        # compare the manifests with iop diff-manifest command
        iop diff-manifest --directory ${OUT}/helm-template/istio-${ISTIO_SYSTEM_NS}-${ISTIO_RELEASE}-${profile} ${OUT}/iop-manifest/istio-${profile}
    else
        echo "Please verify the outpath for the manifests does exist."
        exit 1
    fi
}

# TODO: handle the case that different components are deployed in different namespaces
ISTIO_SYSTEM_NS=${ISTIO_SYSTEM_NS:-istio-system}
ISTIO_RELEASE=${ISTIO_RELEASE:-istio}
ISTIO_DEFAULT_PROFILE=${ISTIO_DEFAULT_PROFILE:-default}
ISTIO_DEMO_PROFILE=${ISTIO_DEMO_PROFILE:-demo}
ISTIO_DEMOAUTH_PROFILE=${ISTIO_DEMOAUTH_PROFILE:-"demo-auth"}
ISTIO_MINIMAL_PROFILE=${ISTIO_MINIMAL_PROFILE:-minimal}
ISTIO_SDS_PROFILE=${ISTIO_SDS_PROFILE:-sds}\

# declare map with profile as key and charts as values
declare -A profile_charts_map
profile_charts_map[${ISTIO_DEFAULT_PROFILE}]="crds istio-control/istio-discovery istio-control/istio-config istio-control/istio-autoinject gateways/istio-ingress istio-telemetry/mixer-telemetry istio-policy security/citadel"
profile_charts_map[${ISTIO_DEMO_PROFILE}]="crds istio-control/istio-discovery istio-control/istio-config istio-control/istio-autoinject gateways/istio-ingress gateways/istio-egress istio-telemetry/mixer-telemetry istio-policy security/citadel"
profile_charts_map[${ISTIO_DEMOAUTH_PROFILE}]="crds istio-control/istio-discovery istio-control/istio-config istio-control/istio-autoinject gateways/istio-ingress gateways/istio-egress istio-telemetry/mixer-telemetry istio-policy security/citadel"
profile_charts_map[${ISTIO_MINIMAL_PROFILE}]="crds istio-control/istio-discovery"
profile_charts_map[${ISTIO_SDS_PROFILE}]="crds istio-control/istio-discovery istio-control/istio-config istio-control/istio-autoinject gateways/istio-ingress istio-telemetry/mixer-telemetry istio-policy security/citadel security/nodeagent"

iop_mandiff_with_profile ${ISTIO_DEFAULT_PROFILE}
iop_mandiff_with_profile ${ISTIO_DEMO_PROFILE}
iop_mandiff_with_profile ${ISTIO_DEMOAUTH_PROFILE}
iop_mandiff_with_profile ${ISTIO_MINIMAL_PROFILE}
iop_mandiff_with_profile ${ISTIO_SDS_PROFILE}

