package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"gopkg.in/yaml.v3"
)

// Ensure the implementation satisfies the desired interfaces.
var _ function.Function = &HelmValuesEncodeFunction{}

// HelmValuesEncodeFunction implements the helm_values_encode function.
type HelmValuesEncodeFunction struct{}

// Metadata defines the function name.
func (f *HelmValuesEncodeFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "helm_values_encode"
}

// Definition defines the function parameters, return type, and documentation.
func (f *HelmValuesEncodeFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Encode Terraform object to YAML with null omission",
		Description: "Takes any Terraform object and converts it to YAML format, recursively removing null values and empty containers that become empty after null removal. This utility function is provided by the Outpost Terraform Provider for generating clean YAML output from Terraform configurations.",
		Parameters: []function.Parameter{
			function.DynamicParameter{
				Name:        "input",
				Description: "The Terraform object to encode as YAML. Can be any valid Terraform data type including objects, arrays, and primitives.",
			},
		},
		Return: function.StringReturn{},
	}
}

// Run contains the function logic.
func (f *HelmValuesEncodeFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var input types.Dynamic

	// Read the dynamic argument
	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &input))
	if resp.Error != nil {
		return
	}

	// Handle null or unknown input
	if input.IsNull() {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, "Input cannot be null"))
		return
	}
	if input.IsUnknown() {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, "Input cannot be unknown"))
		return
	}

	// Convert the dynamic value to a Go interface{} for processing
	goValue, err := f.dynamicToGoValue(ctx, input)
	if err != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(fmt.Sprintf("Failed to convert input: %s", err.Error())))
		return
	}

	// Recursively remove null values and empty containers
	cleanedValue := f.removeNullsAndEmpty(goValue)

	// If the entire structure becomes empty/nil, return empty YAML
	if cleanedValue == nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, ""))
		return
	}

	// Convert to YAML
	yamlBytes, err := yaml.Marshal(cleanedValue)
	if err != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(fmt.Sprintf("Failed to encode YAML: %s", err.Error())))
		return
	}

	// Set the result
	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, string(yamlBytes)))
}

// dynamicToGoValue converts a Terraform dynamic value to a Go interface{} value
func (f *HelmValuesEncodeFunction) dynamicToGoValue(ctx context.Context, dynamic types.Dynamic) (interface{}, error) {
	underlyingValue := dynamic.UnderlyingValue()

	switch v := underlyingValue.(type) {
	case nil:
		// Handle explicit nil values
		return nil, nil

	case types.String:
		if v.IsNull() {
			return nil, nil
		}
		return v.ValueString(), nil

	case types.Number:
		if v.IsNull() {
			return nil, nil
		}
		// Try to convert to int64 first, then float64
		if bigFloat := v.ValueBigFloat(); bigFloat != nil {
			if bigFloat.IsInt() {
				if int64Val, accuracy := bigFloat.Int64(); accuracy == 0 {
					return int64Val, nil
				}
			}
			if float64Val, accuracy := bigFloat.Float64(); accuracy == 0 {
				return float64Val, nil
			}
		}
		return v.ValueBigFloat(), nil

	case types.Bool:
		if v.IsNull() {
			return nil, nil
		}
		return v.ValueBool(), nil

	case types.List:
		if v.IsNull() {
			return nil, nil
		}
		var result []interface{}
		elements := v.Elements()
		for _, elem := range elements {
			goVal, err := f.attrValueToGoValue(ctx, elem)
			if err != nil {
				return nil, err
			}
			result = append(result, goVal)
		}
		return result, nil

	case types.Tuple:
		if v.IsNull() {
			return nil, nil
		}
		var result []interface{}
		elements := v.Elements()
		for _, elem := range elements {
			goVal, err := f.attrValueToGoValue(ctx, elem)
			if err != nil {
				return nil, err
			}
			result = append(result, goVal)
		}
		return result, nil

	case types.Set:
		if v.IsNull() {
			return nil, nil
		}
		var result []interface{}
		elements := v.Elements()
		for _, elem := range elements {
			goVal, err := f.attrValueToGoValue(ctx, elem)
			if err != nil {
				return nil, err
			}
			result = append(result, goVal)
		}
		return result, nil

	case types.Map:
		if v.IsNull() {
			return nil, nil
		}
		result := make(map[string]interface{})
		elements := v.Elements()
		for key, elem := range elements {
			goVal, err := f.attrValueToGoValue(ctx, elem)
			if err != nil {
				return nil, err
			}
			result[key] = goVal
		}
		return result, nil

	case types.Object:
		if v.IsNull() {
			return nil, nil
		}
		result := make(map[string]interface{})
		attributes := v.Attributes()
		for key, attr := range attributes {
			goVal, err := f.attrValueToGoValue(ctx, attr)
			if err != nil {
				return nil, err
			}
			result[key] = goVal
		}
		return result, nil

	case basetypes.DynamicValue:
		// Handle basetypes.DynamicValue by directly processing its underlying value
		underlyingValue := v.UnderlyingValue()
		// Process the underlying value directly without creating another dynamic wrapper
		if underlyingValue != nil {
			return f.attrValueToGoValue(ctx, underlyingValue)
		}
		return nil, nil

	default:
		return nil, fmt.Errorf("unsupported type: %T", v)
	}
}

// attrValueToGoValue converts an attr.Value to a Go interface{} value
func (f *HelmValuesEncodeFunction) attrValueToGoValue(ctx context.Context, value attr.Value) (interface{}, error) {
	switch v := value.(type) {
	case nil:
		// Handle explicit nil values
		return nil, nil

	case types.String:
		if v.IsNull() {
			return nil, nil
		}
		return v.ValueString(), nil

	case types.Number:
		if v.IsNull() {
			return nil, nil
		}
		// Try to convert to int64 first, then float64
		if bigFloat := v.ValueBigFloat(); bigFloat != nil {
			if bigFloat.IsInt() {
				if int64Val, accuracy := bigFloat.Int64(); accuracy == 0 {
					return int64Val, nil
				}
			}
			if float64Val, accuracy := bigFloat.Float64(); accuracy == 0 {
				return float64Val, nil
			}
		}
		return v.ValueBigFloat(), nil

	case types.Bool:
		if v.IsNull() {
			return nil, nil
		}
		return v.ValueBool(), nil

	case types.List:
		if v.IsNull() {
			return nil, nil
		}
		var result []interface{}
		elements := v.Elements()
		for _, elem := range elements {
			goVal, err := f.attrValueToGoValue(ctx, elem)
			if err != nil {
				return nil, err
			}
			result = append(result, goVal)
		}
		return result, nil

	case types.Tuple:
		if v.IsNull() {
			return nil, nil
		}
		var result []interface{}
		elements := v.Elements()
		for _, elem := range elements {
			goVal, err := f.attrValueToGoValue(ctx, elem)
			if err != nil {
				return nil, err
			}
			result = append(result, goVal)
		}
		return result, nil

	case types.Set:
		if v.IsNull() {
			return nil, nil
		}
		var result []interface{}
		elements := v.Elements()
		for _, elem := range elements {
			goVal, err := f.attrValueToGoValue(ctx, elem)
			if err != nil {
				return nil, err
			}
			result = append(result, goVal)
		}
		return result, nil

	case types.Map:
		if v.IsNull() {
			return nil, nil
		}
		result := make(map[string]interface{})
		elements := v.Elements()
		for key, elem := range elements {
			goVal, err := f.attrValueToGoValue(ctx, elem)
			if err != nil {
				return nil, err
			}
			result[key] = goVal
		}
		return result, nil

	case types.Object:
		if v.IsNull() {
			return nil, nil
		}
		result := make(map[string]interface{})
		attributes := v.Attributes()
		for key, attr := range attributes {
			goVal, err := f.attrValueToGoValue(ctx, attr)
			if err != nil {
				return nil, err
			}
			result[key] = goVal
		}
		return result, nil

	case types.Dynamic:
		// Handle dynamic values by calling the main dynamic conversion method
		return f.dynamicToGoValue(ctx, v)

	default:
		return nil, fmt.Errorf("unsupported attr.Value type: %T", v)
	}
}

// removeNullsAndEmpty recursively removes null values and empty containers
func (f *HelmValuesEncodeFunction) removeNullsAndEmpty(value interface{}) interface{} {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, val := range v {
			cleaned := f.removeNullsAndEmpty(val)
			if cleaned != nil {
				result[key] = cleaned
			}
		}
		// Return nil if map becomes empty
		if len(result) == 0 {
			return nil
		}
		return result

	case []interface{}:
		var result []interface{}
		for _, val := range v {
			cleaned := f.removeNullsAndEmpty(val)
			if cleaned != nil {
				result = append(result, cleaned)
			}
		}
		// Return nil if slice becomes empty
		if len(result) == 0 {
			return nil
		}
		return result

	default:
		// For primitive types, return as-is (they're already non-null if we got here)
		return value
	}
}

// NewHelmValuesEncodeFunction creates a new instance of the function
func NewHelmValuesEncodeFunction() function.Function {
	return &HelmValuesEncodeFunction{}
}

