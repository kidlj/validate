package validate

import (
	"reflect"
	"regexp"
	"strings"
)

// Validator is the exported validator
type Validator struct {
	masterTag             string
	codeTag               string
	fieldCustomValidators map[string]FieldCustomValidatorFunc
}

// New creates a Validator
func New() *Validator {
	return &Validator{
		masterTag:             MasterTag,
		codeTag:               CodeTag,
		fieldCustomValidators: map[string]FieldCustomValidatorFunc{},
	}
}

// FieldCustomValidatorFunc represents field custom validator func type
type FieldCustomValidatorFunc func(interface{}) error

// getValidators gets validators
func (v *Validator) getValidators(tag reflect.StructTag) string {
	return tag.Get(v.masterTag)
}

// getCode gets code tag value
func (v *Validator) getCode(tag reflect.StructTag) string {
	return tag.Get(v.codeTag)
}

// SetMasterTag sets the `validate`(default) master tag
func (v *Validator) SetMasterTag(tag string) {
	v.masterTag = tag
}

// SetCodeTag sets the `code`(default) tag
func (v *Validator) SetCodeTag(tag string) {
	v.codeTag = tag
}

// RegisterFieldValidator registers a custom validator func
func (v *Validator) RegisterFieldValidator(name string, validator FieldCustomValidatorFunc) {
	if len(name) > 0 && validator != nil {
		v.fieldCustomValidators[name] = validator
	}
}

// MasterTag is the main validation tag.
const MasterTag = "validate"

// CodeTag is the validation tag to indicate validation error.
const CodeTag = "code"

// CustomValidator is an interface for a validated struct.
type CustomValidator interface {

	// Validate is a custom validation function.
	// Validate does not work when the receiver is a reference.
	// Validate does not work for nested types obtained from unexported field.
	Validate() error
}

// FieldCustomValidator is an interface for a validated struct.
type FieldCustomValidator interface {

	// Validate is a custom validation function.
	// Validate does not work when the receiver is a reference.
	// Validate does not work for nested types obtained from unexported field.
	Validate(value string) error
}

// Validate validates fields of a struct.
// It accepts a struct or a struct pointer as a parameter.
// It returns an error if a struct does not validate or nil if there are no validation errors.
//
//  err := validate.Validate(struct {
//  	field time.Duration `validate:"gte=0s"`
//  }{
//  	field: -time.Second,
//  })
//
//  // err contains an error
func (v *Validator) Validate(element interface{}) error {
	value := reflect.ValueOf(element)

	return v.validateField(value, "", "", "")
}

// validateField validates a struct field
func (v *Validator) validateField(value reflect.Value, fieldName, validators, code string) error {
	kind := value.Kind()

	// Get validator type Map
	validatorTypeMap := v.getValidatorTypeMap()

	// Get validators
	keyValidators, valueValidators, validators, err := splitValidators(validators)
	if err != nil {
		err.setFieldName(fieldName)
		err.setCode(code)
		return err
	}

	// Call a custom validator
	if err := callCustomValidator(value); err != nil {
		return NewErrorCustom(fieldName, value, code, err.Error())
	}

	// Parse validators
	validatorsOr, err := parseValidators(valueValidators)
	if err != nil {
		err.setFieldName(fieldName)
		err.setCode(code)
		return err
	}

	// Perform validators
	for _, validatorsAnd := range validatorsOr {
		for _, validator := range validatorsAnd {
			if validatorFunc, ok := validatorTypeMap[validator.Type]; ok {
				if err = validatorFunc(value, validator.Value); err != nil {
					err.setFieldName(fieldName)
					err.setCode(code)
					break
				}
			} else {
				return &ErrorSyntax{
					fieldName:  fieldName,
					expression: string(validator.Type),
					near:       valueValidators,
					comment:    "could not find a validator",
				}
			}
		}
		if err == nil {
			break
		}
	}
	if err != nil {
		return err
	}

	// Dive one level deep into arrays and pointers
	switch kind {
	case reflect.Struct:
		if err := v.validateStruct(value); err != nil {
			return err
		}
	case reflect.Map:
		for _, key := range value.MapKeys() {
			if err := v.validateField(key, fieldName, keyValidators, code); err != nil {
				return err
			}
			if err := v.validateField(value.MapIndex(key), fieldName, validators, code); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < value.Len(); i++ {
			if err := v.validateField(value.Index(i), fieldName, validators, code); err != nil {
				return err
			}
		}
	case reflect.Ptr:
		if !value.IsNil() {
			if err := v.validateField(value.Elem(), fieldName, validators, code); err != nil {
				return err
			}
		}
	}

	if kind != reflect.Map {
		if len(keyValidators) > 0 {
			return &ErrorSyntax{
				fieldName:  fieldName,
				expression: validators,
				near:       "",
				comment:    "unexpexted expression",
			}
		}
	}

	if kind != reflect.Map && kind != reflect.Slice && kind != reflect.Array && kind != reflect.Ptr {
		if len(validators) > 0 {
			return &ErrorSyntax{
				fieldName:  fieldName,
				expression: validators,
				near:       "",
				comment:    "unexpexted expression",
			}
		}
	}

	return nil
}

// validateStruct validates a struct
func (v *Validator) validateStruct(value reflect.Value) error {
	typ := value.Type()

	// Iterate over struct fields
	for i := 0; i < typ.NumField(); i++ {
		validators := v.getValidators(typ.Field(i).Tag)
		fieldName := typ.Field(i).Name
		code := v.getCode(typ.Field(i).Tag)
		if err := v.validateField(value.Field(i), fieldName, validators, code); err != nil {
			return err
		}
	}

	return nil
}

// splitValidators splits validators into key validators, value validators and remaning validators of the next level
func splitValidators(validators string) (keyValidators string, valValidators string, remaningValidators string, err ErrorField) {
	gt := 0
	bracket := 0
	bracketStart := 0
	bracketEnd := -1

	i := 0
loop:
	for ; i < len(validators); i++ {
		switch validators[i] {
		case '>':
			if bracket == 0 {
				gt++
				break loop
			}
		case '[':
			if bracket == 0 {
				bracketStart = i
			}
			bracket++
		case ']':
			bracket--
			if bracket == 0 {
				bracketEnd = i
			}
		}
	}

	if bracket > 0 {
		err = &ErrorSyntax{
			expression: "",
			near:       validators,
			comment:    "expected \"]\"",
		}
		return
	} else if bracket < 0 {
		err = &ErrorSyntax{
			expression: "",
			near:       validators,
			comment:    "unexpected \"]\"",
		}
		return
	}

	if bracketStart <= len(validators) {
		valValidators += validators[:bracketStart]
	}
	if bracketEnd+1 <= len(validators) {
		if valValidators != "" {
			valValidators += " "
		}
		valValidators += validators[bracketEnd+1 : i]
	}
	if bracketStart+1 <= len(validators) && bracketEnd >= 0 && bracketStart+1 <= bracketEnd {
		keyValidators = validators[bracketStart+1 : bracketEnd]
	}
	if i+1 <= len(validators) {
		remaningValidators = validators[i+1:]
	}

	if gt > 0 && len(remaningValidators) == 0 {
		err = &ErrorSyntax{
			expression: "",
			near:       validators,
			comment:    "expected expression",
		}
		return
	}

	return
}

// parseValidator parses validators into the slice of slices.
// First slice acts as AND logic, second array acts as OR logic.
func parseValidators(validators string) (validatorsOr [][]validator, err ErrorField) {
	regexpType := regexp.MustCompile(`[[:alnum:]_]+`)
	regexpValue := regexp.MustCompile(`[^=\s]+[^=]*[^=\s]+|[^=\s]+`)

	if len(validators) == 0 {
		return
	}

	entriesOr := strings.Split(validators, "|")
	validatorsOr = make([][]validator, 0, len(entriesOr))
	for _, entryOr := range entriesOr {
		entriesAnd := strings.Split(entryOr, "&")
		validatorsAnd := make([]validator, 0, len(entriesAnd))
		for _, entryOr := range entriesAnd {
			entries := strings.Split(entryOr, "=")
			if len(entries) == 0 || len(entries) > 2 {
				err = &ErrorSyntax{
					expression: validators,
					comment:    "could not parse",
				}
				return
			}
			t := regexpType.FindString(entries[0])
			if len(t) == 0 {
				err = &ErrorSyntax{
					expression: entries[0],
					near:       validators,
					comment:    "could not parse",
				}
				return
			}
			v := ""
			if len(entries) == 2 {
				v = regexpValue.FindString(entries[1])
			}
			validatorsAnd = append(validatorsAnd, validator{ValidatorType(t), v})
		}
		if len(validatorsAnd) > 0 {
			validatorsOr = append(validatorsOr, validatorsAnd)
		}
	}

	return
}

// parseTokens parses tokens into array
func parseTokens(str string) []interface{} {
	tokenStrings := strings.Split(str, ",")
	tokens := make([]interface{}, 0, len(tokenStrings))

	for i := range tokenStrings {
		token := strings.TrimSpace(tokenStrings[i])
		if token != "" {
			tokens = append(tokens, token)
		}
	}

	return tokens
}

// tokenOneOf check if a token is one of tokens
func tokenOneOf(token interface{}, tokens []interface{}) bool {
	for _, t := range tokens {
		if t == token {
			return true
		}
	}

	return false
}

// Call a custom validator
func callCustomValidator(value reflect.Value) error {
	if !value.CanInterface() {
		return nil
	}

	// Following code won't work in case if Validate is implemented by reference and value is passed by value
	if customValidator, ok := value.Interface().(CustomValidator); ok {
		return customValidator.Validate()
	}

	// Following code is a fallbak if value is passed by value
	valueCopyPointer := reflect.New(value.Type())
	valueCopyPointer.Elem().Set(value)
	if customValidator, ok := valueCopyPointer.Interface().(CustomValidator); ok {
		return customValidator.Validate()
	}

	return nil
}
