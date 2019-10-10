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

package mesh

import (
	"strings"

	"istio.io/operator/pkg/kubernetes"
)

type hook func(kubeClient kubernetes.ExecClient, istioNamespace,
	currentVer, targetVer, currentValues, targetValues string)

func runUpgradeHooks(kubeClient kubernetes.ExecClient, istioNamespace,
	currentVer, targetVer, currentValues, targetValues string) {
	for _, h := range hooks {
		h(kubeClient, istioNamespace, currentVer, targetVer, currentValues, targetValues)
	}
}

var hooks = []hook{checkInitCrdJobs}

func checkInitCrdJobs(kubeClient kubernetes.ExecClient, istioNamespace,
	currentVer, targetVer, currentValues, targetValues string) {
	pl, err := kubeClient.PodsForSelector(istioNamespace, "")
	if err != nil {
		l.logAndFatalf("Abort. Failed to list pods: %v", err)
	}

	for _, p := range pl.Items {
		if strings.Contains(p.Name, "istio-init-crd") {
			l.logAndFatalf("Abort. istio-init-crd pods exist: %v. "+
				"Istio was installed with non-operator methods, "+
				"please migrate to operator installation first.", p.Name)
		}
	}
}
