package cmd

import (
	"fmt"
	"strings"

	"github.com/aallbrig/beetree-cli/internal/model"
	"github.com/aallbrig/beetree-cli/internal/spec"
	"github.com/aallbrig/beetree-cli/internal/validator"
)

// checkValid returns an error listing every validator message for the tree,
// or nil when the tree is valid.
func checkValid(tree *model.TreeSpec, path string) error {
	errs := validator.Validate(tree)
	if len(errs) == 0 {
		return nil
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s is invalid (%d error(s)); run `beetree validate %s`:", path, len(errs), path)
	for _, e := range errs {
		fmt.Fprintf(&b, "\n  - %s", e.Error())
	}
	return fmt.Errorf("%s", b.String())
}

// loadValidSpec parses a spec file and refuses to continue when it is invalid.
func loadValidSpec(path string) (*model.TreeSpec, error) {
	tree, err := spec.ParseFile(path)
	if err != nil {
		return nil, fmt.Errorf("parse spec: %w", err)
	}
	if err := checkValid(tree, path); err != nil {
		return nil, err
	}
	return tree, nil
}
