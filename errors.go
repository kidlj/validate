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
	Code() string
	setCode(string)
}

// NewErrorValidation creates an ErrorValidation
func NewErrorValidation(fieldName string, fieldValue reflect.Value, code string, validatorType ValidatorType, validatorValue string) *ErrorValidation {
	return &ErrorValidation{
		fieldName:      fieldName,
		fieldValue:     fieldValue,
		code:           code,
		validatorType:  validatorType,
		validatorValue: validatorValue,
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

// Code gets the validation error code.
func (e *ErrorValidation) Code() string {
	return e.code
}

// setCode sets a validation error code.
func (e *ErrorValidation) setCode(code string) {
	e.code = code
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
	code       string
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

// Code gets the validation error code.
func (e *ErrorSyntax) Code() string {
	return e.code
}

// setCode sets a validation error code.
func (e *ErrorSyntax) setCode(code string) {
	e.code = code
}

// Error returns an error.
func (e *ErrorSyntax) Error() string {
	if len(e.fieldName) > 0 {
		return fmt.Sprintf("Syntax error when validating field \"%v\", expression \"%v\" near \"%v\": %v", e.fieldName, e.expression, e.near, e.comment)
	}

	return fmt.Sprintf("Syntax error when validating value, expression \"%v\" near \"%v\": %v", e.expression, e.near, e.comment)
}

// ErrorCustom occurs when there is a syntax error.
type ErrorCustom struct {
	fieldName  string
	fieldValue reflect.Value
	code       string
	message    string
}

// NewErrorCustom creates an ErrorCustom
func NewErrorCustom(fieldName string, fieldValue reflect.Value, code string, message string) *ErrorCustom {
	return &ErrorCustom{
		fieldName:  fieldName,
		fieldValue: fieldValue,
		code:       code,
		message:    message,
	}
}

// FieldName gets a field name.
func (e *ErrorCustom) FieldName() string {
	return e.fieldName
}

// setFieldName sets a field name.
func (e *ErrorCustom) setFieldName(fieldName string) {
	e.fieldName = fieldName
}

// Code gets the validation error code.
func (e *ErrorCustom) Code() string {
	return e.code
}

// setCode sets a validation error code.
func (e *ErrorCustom) setCode(code string) {
	e.code = code
}

// Error returns an error.
func (e *ErrorCustom) Error() string {
	if len(e.fieldName) > 0 {
		return fmt.Sprintf("Validation error in field \"%v\" of type \"%v\" with error message: %s", e.fieldName, e.fieldValue.Type(), e.message)
	}

	return fmt.Sprintf("Validation error in value of type \"%v\" with error message: %s", e.fieldValue.Type(), e.message)
}
