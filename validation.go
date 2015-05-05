package validation

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
	//"strconv"
)

// Validation function.
type Rule func(string) error

// A set of rules to be applied on a single variable.
type Constraint struct {
	Rule    Rule
	Message error
}

// A set of validation rules.
type Rules struct {
	Map map[string][]Constraint
}

// Returns a new set of validation rules.
func New() *Rules {
	self := &Rules{}
	self.Map = make(map[string][]Constraint)
	return self
}

// Adds a rule to a set of constraints.
// func (self *Constraint) Add(rule Rule) {
//   self.All = append(self.All, rule)
// }

// Validates a key/value pair. Returns wether it's valid
// and a slice of errors.
func (self *Rules) ValidateKeyValue(key, value string) (bool, []string) {
	errors := []string{}
	passed := true
	if constraints, ok := self.Map[key]; ok == true {
		for _, constraint := range constraints {
			test := constraint.Rule(value)
			if test != nil {
				passed = false
				error := test.Error()
				if constraint.Message != nil {
					error = constraint.Message.Error()
				}
				errors = append(errors, error)
			}
		}
	}

	return passed, errors
}

// Validates input data.
func (self *Rules) Validate(params map[string]string) (bool, map[string][]string) {
	valid := true
	messages := map[string][]string{}

	for key, _ := range params {
		value := params[key]
		passed, errors := self.ValidateKeyValue(key, value)
		if passed == false {
			messages[key] = errors
			valid = false
		}
	}

	return valid, messages
}

// Validates values in a structure.
// If a rule name has
func (self *Rules) ValidateStruct(s interface{}) (bool, map[string][]string) {
	return self.validateStructWithKeyPrefix("", s)
}

// Validates values in a structure with a key prefix
// If a rule name has
func (self *Rules) validateStructWithKeyPrefix(prefix string, s interface{}) (bool, map[string][]string) {
	valid := true
	messages := map[string][]string{}

	valueOfT := reflect.ValueOf(s)
	if valueOfT.Kind() == reflect.Ptr {
		valueOfT = valueOfT.Elem()
	}

	typeOfT := valueOfT.Type()
	for i := 0; i < valueOfT.NumField(); i++ {
		f := valueOfT.Field(i)
		key := prefix + typeOfT.Field(i).Name

		passed := true
		var errors []string
		switch f.Kind() {
		case reflect.String:
			passed, errors = self.ValidateKeyValue(key, f.Interface().(string))
			if passed == false {
				messages[key] = errors
				valid = false
			}
		case reflect.Struct:
			var errormap map[string][]string
			passed, errormap = self.validateStructWithKeyPrefix(key+".", f.Interface())
			if passed == false {
				for key, value := range errormap {
					messages[key] = value
				}
				valid = false
			}
		}
	}

	return valid, messages
}

// Adds a new rule
func (self *Rules) Add(name string, rule Rule, message string) {
	constraint := Constraint{Rule: rule}
	var constraints []Constraint
	var ok bool

	if constraints, ok = self.Map[name]; ok == false {
		constraints = []Constraint{}
	}

	if len(message) > 0 {
		constraint.Message = errors.New(message)
	}
	constraints = append(constraints, constraint)
	self.Map[name] = constraints
}

// Adds a new rule that is required
func (self *Rules) AddRequired(name string, rule Rule, message string) {
	self.Add(name, NotEmpty, "")
	self.Add(name, rule, message)
}

// A rule that returns error if the value is empty.
func NotEmpty(value string) error {
	if value == "" {
		return fmt.Errorf("This value is required")
	}
	return nil
}
