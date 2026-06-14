package main

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func runEnvVar(name string) (string, *function.FuncError) {
	f := NewEnvVarFunction()
	resp := &function.RunResponse{
		Result: function.NewResultData(types.StringNull()),
	}
	req := function.RunRequest{
		Arguments: function.NewArgumentsData([]attr.Value{types.StringValue(name)}),
	}
	f.Run(context.Background(), req, resp)
	if resp.Error != nil {
		return "", resp.Error
	}
	val := resp.Result.Value()
	if val.IsNull() {
		return "", nil
	}
	return val.(types.String).ValueString(), nil
}

func TestEnvVarFunction(t *testing.T) {
	const envKey = "OUTPOST_PROVIDER_ENV_VAR_TEST"
	t.Setenv(envKey, "hello")

	got, err := runEnvVar(envKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello" {
		t.Fatalf("expected hello, got %q", got)
	}
}

func TestEnvVarFunctionUnset(t *testing.T) {
	const envKey = "OUTPOST_PROVIDER_ENV_VAR_MISSING"
	os.Unsetenv(envKey)

	got, err := runEnvVar(envKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestEnvVarFunctionEmptyName(t *testing.T) {
	_, err := runEnvVar("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}
