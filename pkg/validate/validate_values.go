package validate

import (
	"github.com/ostromart/istio-installer/pkg/util"
)

var (
	// defaultValidations maps a data path to a validation function.
	defaultValuesValidations = map[string]ValidateFunc{
		"global.proxy.includeIpRanges":     validateStringList(validateCIDR),
		"global.proxy.excludeIpRanges":     validateStringList(validateCIDR),
		"global.proxy.includeInboundPorts": validateStringList(validatePortNumberString),
		"global.proxy.excludeInboundPorts": validateStringList(validatePortNumberString),
	}

	// requiredValues lists all the values that must be non-empty.
	requiredSetValues = map[string]bool{
	}
)

// ValidateValues validates the values in the given tree, which follows the Istio values.yaml schema.
func ValidateValues(root util.Tree) util.Errors {
	return validateValues(defaultValuesValidations, root, nil)
}

func validateValues(validations map[string]ValidateFunc, node interface{}, path util.Path) (errs util.Errors) {
	pstr := path.String()
	dbgPrint("validateValues %s", pstr)
	vf := defaultValuesValidations[pstr]
	if vf != nil {
		errs = util.AppendErrs(errs, vf(path, node))
	}

	nn, ok := node.(util.Tree)
	if !ok {
		nn, ok = node.(map[string]interface{})
		if !ok {
			// Leaf, nothing more to recurse.
			return
		}
	}
	for k, v := range nn {
		errs = util.AppendErrs(errs, validateValues(validations, v, append(path, k)))
	}

	return errs
}
