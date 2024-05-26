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
	// In operator
	InAnd Operator = "in"
	// inor operator
	In Operator = "inor"
)

// IsValid checks if the operator is valid
func (o Operator) IsValid() error {
	switch o {
	case Equal, NotEqual, GreaterThan, GreaterThanOrEqual, LessThan, LessThanOrEqual:
		return nil
	}
	return fmt.Errorf("Invalid operator: %s", o)
}

// ValidateFilter validates the filter against the given vector's payload.
// It performs the following steps:
// - Loads the payload from the hdd
// - Checks if the field exists in the payload
// - Checks if the field value and the filter value have the same type
// - Compares the field value with the filter value using the operator
// - Returns true if the filter condition is met, otherwise returns false
// - Returns an error if an error occurs during file reading or type comparison, otherwise returns nil
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
		// in is special - it takes a slice of values and checks if the field value (which is also a slice) contains any of the values
		// it returns true if all values in the slice are in the field slice
	case InAnd:
		switch v := f.Value.(type) {
		case []int:
			for _, value := range v {
				// if the value is not in the slice, return false
				for _, fieldValue := range (*payload)[f.Field].([]int) {
					if value == fieldValue {
						continue
					}
				}
				return false, nil
			}
			return true, nil
		case []int64:
			for _, value := range v {
				// if the value is not in the slice, return false
				for _, fieldValue := range (*payload)[f.Field].([]int64) {
					if value == fieldValue {
						continue
					}
				}
				return false, nil
			}
			return true, nil
		case []float32:
			for _, value := range v {
				// if the value is not in the slice, return false
				for _, fieldValue := range (*payload)[f.Field].([]float32) {
					if value == fieldValue {
						continue
					}
				}
				return false, nil
			}
			return true, nil
		case []float64:
			for _, value := range v {
				// if the value is not in the slice, return false
				for _, fieldValue := range (*payload)[f.Field].([]float64) {
					if value == fieldValue {
						continue
					}
				}
				return false, nil
			}
			return true, nil
		case []string:
			for _, value := range v {
				// if the value is not in the slice, return false
				for _, fieldValue := range (*payload)[f.Field].([]string) {
					if value == fieldValue {
						continue
					}
				}
				return false, nil
			}
		}
	case In:
		// inor is special too, there must be only one value in the slice that is in the field slice
		switch v := f.Value.(type) {
		case []int:
			for _, value := range v {
				// if the value is in the slice, return true
				for _, fieldValue := range (*payload)[f.Field].([]int) {
					if value == fieldValue {
						return true, nil
					}
				}
			}
			return false, nil
		case []int64:
			for _, value := range v {
				// if the value is in the slice, return true
				for _, fieldValue := range (*payload)[f.Field].([]int64) {
					if value == fieldValue {
						return true, nil
					}
				}
			}
			return false, nil
		case []float32:
			for _, value := range v {
				// if the value is in the slice, return true
				for _, fieldValue := range (*payload)[f.Field].([]float32) {
					if value == fieldValue {
						return true, nil
					}
				}
			}
			return false, nil
		case []float64:
			for _, value := range v {
				// if the value is in the slice, return true
				for _, fieldValue := range (*payload)[f.Field].([]float64) {
					if value == fieldValue {
						return true, nil
					}
				}
			}
			return false, nil
		case []string:
			for _, value := range v {
				// if the value is in the slice, return true
				for _, fieldValue := range (*payload)[f.Field].([]string) {
					if value == fieldValue {
						return true, nil
					}
				}
			}
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
