package main

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the desired interfaces.
var _ function.Function = &EnvVarFunction{}

// EnvVarFunction reads a process environment variable at evaluation time.
type EnvVarFunction struct{}

// Metadata defines the function name.
func (f *EnvVarFunction) Metadata(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "env_var"
}

// Definition defines the function parameters, return type, and documentation.
func (f *EnvVarFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Read an environment variable",
		Description: "Returns the value of the named environment variable from the Terraform process environment, or an empty string when unset. Useful for runtime config injected by CI or Terrakube workspace variables (e.g. provider::outpost::env_var(\"VAULT_ADDR\")).",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "name",
				Description: "Environment variable name (non-empty).",
			},
		},
		Return: function.StringReturn{},
	}
}

// Run contains the function logic.
func (f *EnvVarFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var name types.String

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &name))
	if resp.Error != nil {
		return
	}

	if name.IsNull() {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, "name cannot be null"))
		return
	}
	if name.IsUnknown() {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, "name cannot be unknown"))
		return
	}

	key := name.ValueString()
	if key == "" {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, "name cannot be empty"))
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, os.Getenv(key)))
}

// NewEnvVarFunction creates a new instance of the function.
func NewEnvVarFunction() function.Function {
	return &EnvVarFunction{}
}
