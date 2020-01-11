package translate

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"istio.io/operator/pkg/tpath"
	"istio.io/operator/pkg/util"
)

const (
	istioOperatorTreeString = `
apiVersion: operator.istio.io/v1alpha1
kind: IstioOperator
`
)

// ReadTranslations reads a file at filePath with key:value pairs in the format expected by TranslateICPToIOP.
func ReadTranslations(filePath string) (map[string]string, error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string)
	if err := yaml.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// TranslateICPToIOP takes an IstioControlPlane YAML string and a map of translations with key:value format
// souce-path:destination-path (where paths are expressed in pkg/tpath format) and returns an IstioOperator string.
func TranslateICPToIOP(icp string, translations map[string]string) (string, error) {
	icps, err := getSpecSubtree(icp)
	if err != nil {
		return "", err
	}

	// Prefill the output tree with gateways if they are set to ensure we have the correct list types created.
	outTree, err := gatewaysOverlay(icps)
	if err != nil {
		return "", err
	}

	translated, err := TranslateYAMLTree(icps, outTree, translations)
	if err != nil {
		return "", err
	}

	out, err := addSpecRoot(translated)
	if err != nil {
		return "", err
	}
	return util.OverlayYAML(istioOperatorTreeString, out)
}

// getSpecSubtree takes a YAML tree with the root node spec and returns the subtree under this root node.
func getSpecSubtree(tree string) (string, error) {
	icpTree := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(tree), &icpTree); err != nil {
		return "", err
	}
	spec := icpTree["spec"]
	if spec == nil {
		return "", fmt.Errorf("spec field not found in input string: \n%s", tree)
	}
	outTree, err := yaml.Marshal(spec)
	if err != nil {
		return "", err
	}
	return string(outTree), nil
}

// gatewaysOverlay takes a source YAML tree and creates empty output gateways list entries for ingress and egress
// gateways if these are present in the source tree.
// Gateways must be created in the tree separately because they are a list type. Inserting into the tree dynamically
// would result in gateway paths being map types.
func gatewaysOverlay(icps string) (string, error) {
	componentsHeaderStr := `
components:
`
	gatewayStr := map[string]string{
		"ingress": `
  ingressGateways:
  - name: istio-ingressgateway
`,
		"egress": `
  egressGateways:
  - name: istio-egressgateway
`,
	}

	icpsT, err := unmarshalTree(icps)
	if err != nil {
		return "", err
	}

	out := ""
	componentsHeaderSet := false
	for _, gt := range []string{"ingress", "egress"} {
		if _, found, _ := tpath.GetFromTreePath(icpsT, util.PathFromString(fmt.Sprintf("gateways.components.%sGateway", gt))); found {
			if !componentsHeaderSet {
				componentsHeaderSet = true
				out += componentsHeaderStr
			}
			out += gatewayStr[gt]
		}
	}
	return out, nil
}

// addSpecRoot is the reverse of getSpecSubtree: it adds a root node called "spec" to the given tree and returns the
// resulting tree.
func addSpecRoot(tree string) (string, error) {
	t, nt := make(map[string]interface{}), make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(tree), &t); err != nil {
		return "", err
	}
	nt["spec"] = t
	out, err := yaml.Marshal(nt)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func unmarshalTree(tree string) (map[string]interface{}, error) {
	out := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(tree), &out); err != nil {
		return nil, err
	}
	return out, nil
}
