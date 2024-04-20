package Filter

import (
	"VreeDB/FileMapper"
	"VreeDB/Vector"
	"fmt"
)

type Operator string

type Filter struct {
	Field string      `json:"field"`
	Op    Operator    `json:"operator"`
	Value interface{} `json:"value"`
}

// Operators
const (
	// Equal operator
	Equal Operator = "eq"
	// NotEqual operator
	NotEqual Operator = "ne"
	// GreaterThan operator
	GreaterThan Operator = "gt"
	// GreaterThanOrEqual operator
	GreaterThanOrEqual Operator = "ge"
	// LessThan operator
	LessThan Operator = "lt"
	// LessThanOrEqual operator
	LessThanOrEqual Operator = "le"
)

// IsValid checks if the operator is valid
func (o Operator) IsValid() error {
	switch o {
	case Equal, NotEqual, GreaterThan, GreaterThanOrEqual, LessThan, LessThanOrEqual:
		return nil
	}
	return fmt.Errorf("Invalid operator: %s", o)
}

// ValidateFilter will validate the filters on a given Vector
func (f *Filter) ValidateFilter(vector *Vector.Vector) (bool, error) {
	// Load the Payload from the hdd
	payload, err := FileMapper.Mapper.ReadPayload(vector.PayloadStart, vector.Collection)
	if err != nil {
		return false, err
	}

	// Check if the field exists in the payload
	if _, ok := (*payload)[f.Field]; !ok {
		return false, nil
	}

	// If the operator is == or !=, we can compare the values with all variable types, otherwise we need to check if
	// the value is of type float64, float32 or int
	switch f.Op {
	case Equal, NotEqual:
		if (*payload)[f.Field] == f.Value {
			return true, nil
		}
	case GreaterThan, GreaterThanOrEqual, LessThan, LessThanOrEqual:
		switch v := f.Value.(type) {
		case int64:
		case int:
		case float32:
		case float64:
			switch f.Op {
			case GreaterThan:
				if (*payload)[f.Field].(float64) > v {
					return true, nil
				}
			case GreaterThanOrEqual:
				if (*payload)[f.Field].(float64) >= v {
					return true, nil
				}
			case LessThan:
				if (*payload)[f.Field].(float64) < v {

				}
			case LessThanOrEqual:
				if (*payload)[f.Field].(float64) <= v {
					return true, nil
				}
			}
		default:
			return false, nil
		}
	}
	return false, nil
}
