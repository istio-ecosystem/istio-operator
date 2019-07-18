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

# shellcheck source=prow/lib.sh
source "${ROOT}/prow/lib.sh"
#setup_and_export_git_sha

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
    if [ -f "${chart}/values-istio-${profile}.yaml" ]; then
        cfg="${cfg} -f ${chart}/values-istio-${profile}.yaml"
    fi

    mkdir -p ${OUT}/helm-template/istio-${namespace}-${relase}-${profile}

    helm_render_template ${namespace} ${relase} ${chart}/crds ${cfg} $* > ${OUT}/helm-template/istio-${namespace}-${relase}-${profile}/crds.yaml

    helm_render_template ${namespace} ${relase} ${chart}/istio-control/istio-config ${cfg} $* > ${OUT}/helm-template/istio-${namespace}-${relase}-${profile}/galley.yaml
    helm_render_template ${namespace} ${relase} ${chart}/istio-control/istio-discovery ${cfg} $* > ${OUT}/helm-template/istio-${namespace}-${relase}-${profile}/pilot.yaml
    helm_render_template ${namespace} ${relase} ${chart}/istio-control/istio-autoinject ${cfg} $* > ${OUT}/helm-template/istio-${namespace}-${relase}-${profile}/sidecar-injector.yaml

    helm_render_template ${namespace} ${relase} ${chart}/gateways/istio-ingress ${cfg} $* > ${OUT}/helm-template/istio-${namespace}-${relase}-${profile}/ingress.yaml
    helm_render_template ${namespace} ${relase} ${chart}/gateways/istio-egress ${cfg} $* > ${OUT}/helm-template/istio-${namespace}-${relase}-${profile}/egress.yaml

    helm_render_template ${namespace} ${relase} ${chart}/istio-telemetry/mixer-telemetry ${cfg} $* > ${OUT}/helm-template/istio-${namespace}-${relase}-${profile}/telemetry.yaml

    helm_render_template ${namespace} ${relase} ${chart}/istio-policy ${cfg} $* > ${OUT}/helm-template/istio-${namespace}-${relase}-${profile}/policy.yaml

    helm_render_template ${namespace} ${relase} ${chart}/security/certmanager ${cfg} $* > ${OUT}/helm-template/istio-${namespace}-${relase}-${profile}/certmanager.yaml
    helm_render_template ${namespace} ${relase} ${chart}/security/citadel ${cfg} $* > ${OUT}/helm-template/istio-${namespace}-${relase}-${profile}/citadel.yaml
    helm_render_template ${namespace} ${relase} ${chart}/security/nodeagent ${cfg} $* > ${OUT}/helm-template/istio-${namespace}-${relase}-${profile}/nodeagent.yaml
}

# render all the templates with iop manifest.
function iop_manifest() {
    local profile=${1}
    shift
    mkdir -p ${OUT}/iop-manifest/istio-${profile}
    iop manifest --dry-run=false --output ${OUT}/iop-manifest/istio-${profile} 2>&1
}

function iop_mandiff() {
    helm_manifest ${ISTIO_SYSTEM_NS} ${ISTIO_RELEASE} "${ROOT}/data/charts" ${ISTIO_PROFILE}
    iop_manifest ${ISTIO_PROFILE}

    iop diff-manifest --directory ${OUT}/helm-template/istio-${ISTIO_SYSTEM_NS}-${ISTIO_RELEASE}-${ISTIO_PROFILE} ${OUT}/iop-manifest/istio-${ISTIO_PROFILE}

}

# TODO: handle the case that different components are deployed in different namespaces
ISTIO_SYSTEM_NS=${ISTIO_SYSTEM_NS:-istio-system}
ISTIO_RELEASE=${ISTIO_RELEASE:-istio}
ISTIO_PROFILE=${ISTIO_PROFILE:-default}

iop_mandiff

