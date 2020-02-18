package validate

import (
	"fmt"
	"reflect"
)

// ErrorField is an error interface for field/value error.
type ErrorField interface {
	error
	FieldName() string
	setFieldName(string)
}

// NewErrorValidation creates an ErrorValidation
func NewErrorValidation(fieldName string, fieldValue reflect.Value, code string) *ErrorValidation {
	return &ErrorValidation{
		fieldName:  fieldName,
		fieldValue: fieldValue,
		code:       code,
	}
}

// ErrorValidation occurs when validator does not validate.
type ErrorValidation struct {
	fieldName      string
	fieldValue     reflect.Value
	code           string
	validatorType  ValidatorType
	validatorValue string
}

// FieldName gets a field name.
func (e *ErrorValidation) FieldName() string {
	return e.fieldName
}

// setFieldName sets a field name.
func (e *ErrorValidation) setFieldName(fieldName string) {
	e.fieldName = fieldName
}

// setCode sets a validation error code.
func (e *ErrorValidation) setCode(code string) {
	e.code = code
}

// GetCode gets the validation error code.
func (e *ErrorValidation) GetCode() string {
	return e.code
}

// Error returns an error.
func (e *ErrorValidation) Error() string {
	validator := string(e.validatorType)
	if len(e.validatorValue) > 0 {
		validator += "=" + e.validatorValue
	}

	if len(e.fieldName) > 0 {
		return fmt.Sprintf("Validation error in field \"%v\" of type \"%v\" using validator \"%v\"", e.fieldName, e.fieldValue.Type(), validator)
	}

	return fmt.Sprintf("Validation error in value of type \"%v\" using validator \"%v\"", e.fieldValue.Type(), validator)
}

// ErrorSyntax occurs when there is a syntax error.
type ErrorSyntax struct {
	fieldName  string
	expression string
	near       string
	comment    string
}

// FieldName gets a field name.
func (e *ErrorSyntax) FieldName() string {
	return e.fieldName
}

// setFieldName sets a field name.
func (e *ErrorSyntax) setFieldName(fieldName string) {
	e.fieldName = fieldName
}

// Error returns an error.
func (e *ErrorSyntax) Error() string {
	if len(e.fieldName) > 0 {
		return fmt.Sprintf("Syntax error when validating field \"%v\", expression \"%v\" near \"%v\": %v", e.fieldName, e.expression, e.near, e.comment)
	}

	return fmt.Sprintf("Syntax error when validating value, expression \"%v\" near \"%v\": %v", e.expression, e.near, e.comment)
}
