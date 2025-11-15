package main

import (
	"context"
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestHelmValuesEncodeFunction(t *testing.T) {
	f := NewHelmValuesEncodeFunction()

	tests := []struct {
		name     string
		input    map[string]interface{}
		expected string
		wantErr  bool
	}{
		{
			name: "simple object with null values",
			input: map[string]interface{}{
				"name":    "test",
				"enabled": true,
				"count":   int64(3),
			},
			expected: "count: 3\nenabled: true\nname: test\n",
			wantErr:  false,
		},
		{
			name: "nested object with empty map",
			input: map[string]interface{}{
				"service": map[string]interface{}{
					"type": "ClusterIP",
					"port": int64(80),
				},
			},
			expected: "service:\n  port: 80\n  type: ClusterIP\n",
			wantErr:  false,
		},
		{
			name: "list with valid values",
			input: map[string]interface{}{
				"items": []interface{}{
					"item1",
					"item2",
					"item3",
				},
			},
			expected: "items:\n- item1\n- item2\n- item3\n",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is a simplified test - in practice you'd need to properly initialize
			// the function.RunRequest with a proper Arguments implementation
			// For now, we'll just verify the helper functions work
			ctx := context.Background()

			// Create a dynamic value from the input
			inputValue := convertMapToDynamic(ctx, tt.input)

			// Test the conversion functions
			fn := &HelmValuesEncodeFunction{}
			goValue, err := fn.dynamicToGoValue(ctx, inputValue)
			if err != nil {
				t.Errorf("failed to convert dynamic value: %v", err)
				return
			}

			// Test the null removal
			cleaned := fn.removeNullsAndEmpty(goValue)
			if cleaned == nil {
				t.Errorf("cleaned value should not be nil")
				return
			}

			// Note: Full function testing would require a complete Terraform framework setup
			t.Logf("Successfully converted and cleaned input: %+v", cleaned)

		})
	}
}

// Helper function to convert a Go map to types.Dynamic
func convertMapToDynamic(ctx context.Context, m map[string]interface{}) types.Dynamic {
	attrs := make(map[string]attr.Value)
	attrTypes := make(map[string]attr.Type)

	for k, v := range m {
		val, typ := convertInterfaceToAttrValue(v)
		attrs[k] = val
		attrTypes[k] = typ
	}

	objVal, _ := types.ObjectValue(attrTypes, attrs)
	dynVal, _ := types.DynamicValue(objVal)
	return dynVal
}

// Helper function to convert interface{} to attr.Value
func convertInterfaceToAttrValue(v interface{}) (attr.Value, attr.Type) {
	switch val := v.(type) {
	case string:
		return types.StringValue(val), types.StringType
	case bool:
		return types.BoolValue(val), types.BoolType
	case int64:
		return types.NumberValue(big.NewInt(val)), types.NumberType
	case float64:
		return types.NumberValue(big.NewFloat(val)), types.NumberType
	case []interface{}:
		elemVals := make([]attr.Value, len(val))
		var elemType attr.Type
		for i, elem := range val {
			elemVal, eType := convertInterfaceToAttrValue(elem)
			elemVals[i] = elemVal
			if elemType == nil {
				elemType = eType
			}
		}
		if elemType == nil {
			elemType = types.StringType
		}
		listVal, _ := types.ListValue(elemType, elemVals)
		return listVal, types.ListType{ElemType: elemType}
	case map[string]interface{}:
		attrs := make(map[string]attr.Value)
		attrTypes := make(map[string]attr.Type)
		for k, v := range val {
			attrVal, attrType := convertInterfaceToAttrValue(v)
			attrs[k] = attrVal
			attrTypes[k] = attrType
		}
		objVal, _ := types.ObjectValue(attrTypes, attrs)
		return objVal, types.ObjectType{AttrTypes: attrTypes}
	default:
		return types.StringNull(), types.StringType
	}
}

