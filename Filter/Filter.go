package Filter

import (
	"VreeDB/FileMapper"
	"VreeDB/Logger"
	"VreeDB/Vector"
	"fmt"
	"unsafe"
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

	// Check if they are of the same type
	if !f.checkSameType((*payload)[f.Field], f.Value) {
		return false, nil
	}

	// If the operator is == or !=, we can compare the values with all variable types, otherwise we need to check if
	// the value is of type float64, float32 or int
	switch f.Op {
	case Equal, NotEqual:
		switch v := f.Value.(type) {
		case int, int64, float32, float64:
			switch f.Op {
			case Equal:
				if (*payload)[f.Field] == v {
					return true, nil
				}
			case NotEqual:
				if (*payload)[f.Field] != v {
					return true, nil
				}
			}
		case string:
			switch f.Op {
			case Equal:
				if (*payload)[f.Field] == f.Value {
					return true, nil
				}
			case NotEqual:
				if (*payload)[f.Field] != f.Value {
					return true, nil
				}
			}
		default:
			return false, nil
		}
	// TBD: GreaterThan, GreaterThanOrEqual, LessThan, LessThanOrEqual
	case GreaterThan:
		switch v := f.Value.(type) {
		case int:
			if (*payload)[f.Field].(int) > v {
				return true, nil
			}
		case int64:
			if (*payload)[f.Field].(int64) > v {
				return true, nil
			}
		case float32:
			if (*payload)[f.Field].(float32) > v {
				return true, nil
			}
		case float64:
			if (*payload)[f.Field].(float64) > v {
				return true, nil
			}
		default:
			return false, nil
		}
	case GreaterThanOrEqual:
		switch v := f.Value.(type) {
		case int:
			if (*payload)[f.Field].(int) >= v {
				return true, nil
			}
		case int64:
			if (*payload)[f.Field].(int64) >= v {
				return true, nil
			}
		case float32:
			if (*payload)[f.Field].(float32) >= v {
				return true, nil
			}
		case float64:
			if (*payload)[f.Field].(float64) >= v {
				return true, nil
			}
		default:
			return false, nil
		}
	case LessThan:
		switch v := f.Value.(type) {
		case int:
			if (*payload)[f.Field].(int) < v {
				return true, nil
			}
		case int64:
			if (*payload)[f.Field].(int64) < v {
				return true, nil
			}
		case float32:
			if (*payload)[f.Field].(float32) < v {
				return true, nil
			}
		case float64:
			if (*payload)[f.Field].(float64) < v {
				return true, nil
			}
		default:
			return false, nil
		}
	case LessThanOrEqual:
		switch v := f.Value.(type) {
		case int:
			if (*payload)[f.Field].(int) <= v {
				return true, nil
			}
		case int64:
			if (*payload)[f.Field].(int64) <= v {
				return true, nil
			}
		case float32:
			if (*payload)[f.Field].(float32) <= v {
				return true, nil
			}
		case float64:
			if (*payload)[f.Field].(float64) <= v {
				return true, nil
			}
		default:
			return false, nil
		}
	default:
		// May never happen
		Logger.Log.Log("invalid operator in filter - this is a bug - please report")
		return false, nil
	}
	return false, nil
}

// CheckSameType checks if the two values are of the same type
func (f *Filter) checkSameType(a, b any) bool {
	typeOfA := *(*uintptr)(unsafe.Pointer(&a))
	typeOfB := *(*uintptr)(unsafe.Pointer(&b))

	// Check if the types are the same
	return typeOfA == typeOfB
}
