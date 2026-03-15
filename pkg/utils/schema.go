package utils

import (
	"fmt"
	"reflect"
)

// ValidateSchema 简单的 Schema 校验（支持 required 和 type 检查）
func ValidateSchema(schema map[string]interface{}, args map[string]interface{}) error {
	if schema == nil || len(schema) == 0 {
		return nil
	}

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		return nil
	}

	required, _ := schema["required"].([]interface{})
	requiredFields := make(map[string]bool)
	for _, field := range required {
		if fieldName, ok := field.(string); ok {
			requiredFields[fieldName] = true
		}
	}

	// 检查必填字段
	for field := range requiredFields {
		if _, exists := args[field]; !exists {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	// 检查字段类型
	for field, value := range args {
		propSchema, exists := properties[field]
		if !exists {
			continue
		}

		propMap, ok := propSchema.(map[string]interface{})
		if !ok {
			continue
		}

		expectedType, ok := propMap["type"].(string)
		if !ok {
			continue
		}

		if err := validateType(field, value, expectedType); err != nil {
			return err
		}
	}

	return nil
}

func validateType(field string, value interface{}, expectedType string) error {
	actualType := getJSONType(value)
	if actualType != expectedType {
		return fmt.Errorf("field %s: expected type %s, got %s", field, expectedType, actualType)
	}
	return nil
}

func getJSONType(value interface{}) string {
	if value == nil {
		return "null"
	}

	switch reflect.TypeOf(value).Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.Slice, reflect.Array:
		return "array"
	case reflect.Map:
		return "object"
	default:
		return "unknown"
	}
}
