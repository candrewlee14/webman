package schema

import (
	_ "embed"
	"fmt"
	"io"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

var (
	//go:embed pkg_schema.json
	recipeSchema       []byte
	recipeSchemaLoader = gojsonschema.NewBytesLoader(recipeSchema)

	//go:embed group_schema.json
	groupSchema       []byte
	groupSchemaLoader = gojsonschema.NewBytesLoader(groupSchema)

	//go:embed config_schema.json
	configSchema       []byte
	configSchemaLoader = gojsonschema.NewBytesLoader(configSchema)
)

// LintRecipe is for linting a recipe against the schema
func LintRecipe(r io.Reader) error {
	return lint(r, recipeSchemaLoader)
}

// LintGroup is for linting a group against the schema
func LintGroup(r io.Reader) error {
	return lint(r, groupSchemaLoader)
}

// LintConfig is for linting a config against the schema
func LintConfig(r io.Reader) error {
	return lint(r, configSchemaLoader)
}

func lint(r io.Reader, loader gojsonschema.JSONLoader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	var m any
	if err := Unmarshal(data, &m); err != nil {
		return err
	}
	sourceLoader := gojsonschema.NewGoLoader(m)
	result, err := gojsonschema.Validate(loader, sourceLoader)
	if err != nil {
		return err
	}
	if len(result.Errors()) > 0 {
		return ResultErrors(result.Errors())
	}
	return nil
}

// ResultErrors is a slice of gojsonschema.ResultError that implements error
type ResultErrors []gojsonschema.ResultError

// Error implements error
func (r ResultErrors) Error() string {
	errs := make([]string, 0, len(r))
	for _, re := range r {
		errs = append(errs, fmt.Sprintf("%s: %s", re.Field(), re.Description()))
	}
	return strings.Join(errs, " | ")
}
