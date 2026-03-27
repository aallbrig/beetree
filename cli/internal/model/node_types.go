package model

var coreNodeTypes = map[string]bool{
	"action":    true,
	"condition": true,
	"sequence":  true,
	"selector":  true,
	"parallel":  true,
	"decorator": true,
}

var extensionNodeTypes = map[string]bool{
	"utility_selector": true,
	"active_selector":  true,
	"random_selector":  true,
	"random_sequence":  true,
	"subtree":          true,
}

var compositeTypes = map[string]bool{
	"sequence": true,
	"selector": true,
	"parallel": true,
}

var leafTypes = map[string]bool{
	"action":    true,
	"condition": true,
}

var builtinDecorators = map[string]bool{
	"repeat":         true,
	"negate":         true,
	"always_succeed": true,
	"always_fail":    true,
	"until_fail":     true,
	"until_succeed":  true,
	"timeout":        true,
	"cooldown":       true,
	"retry":          true,
}

// CoreNodeTypes returns the names of the six core node types.
func CoreNodeTypes() map[string]bool {
	result := make(map[string]bool, len(coreNodeTypes))
	for k, v := range coreNodeTypes {
		result[k] = v
	}
	return result
}

// ExtensionNodeTypes returns the names of the extension node types.
func ExtensionNodeTypes() map[string]bool {
	result := make(map[string]bool, len(extensionNodeTypes))
	for k, v := range extensionNodeTypes {
		result[k] = v
	}
	return result
}

// IsValidNodeType checks if a type string is a known core or extension type.
func IsValidNodeType(t string) bool {
	return coreNodeTypes[t] || extensionNodeTypes[t]
}

// IsCompositeType checks if the type can have children.
func IsCompositeType(t string) bool {
	return compositeTypes[t]
}

// IsLeafType checks if the type is a leaf node (no children).
func IsLeafType(t string) bool {
	return leafTypes[t]
}

// BuiltinDecorators returns the set of built-in decorator names.
func BuiltinDecorators() map[string]bool {
	result := make(map[string]bool, len(builtinDecorators))
	for k, v := range builtinDecorators {
		result[k] = v
	}
	return result
}
