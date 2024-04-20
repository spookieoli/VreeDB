package Filter

import "fmt"

type Operator string

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

// Filter is the struct that will be used to filter the data
type Filter struct {
	// Field is the field that will be used to filter
	Field string `json:"field"`
	// Operator is the operator that will be used to filter
	Operator Operator `json:"operator"`
	// Value is the value that will be used to filter
	Value interface{} `json:"value"`
}
