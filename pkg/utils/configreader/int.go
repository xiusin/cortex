/*
Copyright 2019 Cortex Labs, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package configreader

import (
	"io/ioutil"

	s "github.com/cortexlabs/cortex/pkg/api/strings"
	"github.com/cortexlabs/cortex/pkg/utils/cast"
	"github.com/cortexlabs/cortex/pkg/utils/errors"
	"github.com/cortexlabs/cortex/pkg/utils/util"
)

type IntValidation struct {
	Required             bool
	Default              int
	AllowedValues        []int
	GreaterThan          *int
	GreaterThanOrEqualTo *int
	LessThan             *int
	LessThanOrEqualTo    *int
	Validator            func(int) (int, error)
}

func Int(inter interface{}, v *IntValidation) (int, error) {
	if inter == nil {
		return 0, errors.New(s.ErrCannotBeNull)
	}
	casted, castOk := cast.InterfaceToInt(inter)
	if !castOk {
		return 0, errors.New(s.ErrInvalidPrimitiveType(inter, s.PrimTypeInt))
	}
	return ValidateInt(casted, v)
}

func IntFromInterfaceMap(key string, iMap map[string]interface{}, v *IntValidation) (int, error) {
	inter, ok := ReadInterfaceMapValue(key, iMap)
	if !ok {
		val, err := ValidateIntMissing(v)
		if err != nil {
			return 0, errors.Wrap(err, key)
		}
		return val, nil
	}
	val, err := Int(inter, v)
	if err != nil {
		return 0, errors.Wrap(err, key)
	}
	return val, nil
}

func IntFromStrMap(key string, sMap map[string]string, v *IntValidation) (int, error) {
	valStr, ok := sMap[key]
	if !ok || valStr == "" {
		val, err := ValidateIntMissing(v)
		if err != nil {
			return 0, errors.Wrap(err, key)
		}
		return val, nil
	}
	val, err := IntFromStr(valStr, v)
	if err != nil {
		return 0, errors.Wrap(err, key)
	}
	return val, nil
}

func IntFromStr(valStr string, v *IntValidation) (int, error) {
	if valStr == "" {
		return ValidateIntMissing(v)
	}
	casted, castOk := s.ParseInt(valStr)
	if !castOk {
		return 0, errors.New(s.ErrInvalidPrimitiveType(valStr, s.PrimTypeInt))
	}
	return ValidateInt(casted, v)
}

func IntFromEnv(envVarName string, v *IntValidation) (int, error) {
	valStr := ReadEnvVar(envVarName)
	if valStr == nil || *valStr == "" {
		val, err := ValidateIntMissing(v)
		if err != nil {
			return 0, errors.Wrap(err, s.EnvVar(envVarName))
		}
		return val, nil
	}
	val, err := IntFromStr(*valStr, v)
	if err != nil {
		return 0, errors.Wrap(err, s.EnvVar(envVarName))
	}
	return val, nil
}

func IntFromFile(filePath string, v *IntValidation) (int, error) {
	valBytes, err := ioutil.ReadFile(filePath)
	if err != nil || len(valBytes) == 0 {
		val, err := ValidateIntMissing(v)
		if err != nil {
			return 0, errors.Wrap(err, filePath)
		}
		return val, nil
	}
	valStr := string(valBytes)
	val, err := IntFromStr(valStr, v)
	if err != nil {
		return 0, errors.Wrap(err, filePath)
	}
	return val, nil
}

func IntFromEnvOrFile(envVarName string, filePath string, v *IntValidation) (int, error) {
	valStr := ReadEnvVar(envVarName)
	if valStr != nil && *valStr != "" {
		return IntFromEnv(envVarName, v)
	}
	return IntFromFile(filePath, v)
}

func IntFromPrompt(promptOpts *PromptOptions, v *IntValidation) (int, error) {
	promptOpts.defaultStr = s.Int(v.Default)
	valStr := prompt(promptOpts)
	if valStr == "" {
		return ValidateIntMissing(v)
	}
	return IntFromStr(valStr, v)
}

func ValidateIntMissing(v *IntValidation) (int, error) {
	if v.Required {
		return 0, errors.New(s.ErrMustBeDefined)
	}
	return ValidateInt(v.Default, v)
}

func ValidateInt(val int, v *IntValidation) (int, error) {
	err := ValidateIntVal(val, v)
	if err != nil {
		return 0, err
	}

	if v.Validator != nil {
		return v.Validator(val)
	}
	return val, nil
}

func ValidateIntVal(val int, v *IntValidation) error {
	if v.GreaterThan != nil {
		if val <= *v.GreaterThan {
			return errors.New(s.ErrMustBeGreaterThan(val, *v.GreaterThan))
		}
	}
	if v.GreaterThanOrEqualTo != nil {
		if val < *v.GreaterThanOrEqualTo {
			return errors.New(s.ErrMustBeGreaterThanOrEqualTo(val, *v.GreaterThanOrEqualTo))
		}
	}
	if v.LessThan != nil {
		if val >= *v.LessThan {
			return errors.New(s.ErrMustBeLessThan(val, *v.LessThan))
		}
	}
	if v.LessThanOrEqualTo != nil {
		if val > *v.LessThanOrEqualTo {
			return errors.New(s.ErrMustBeLessThanOrEqualTo(val, *v.LessThanOrEqualTo))
		}
	}

	if v.AllowedValues != nil {
		if !util.IsIntInSlice(val, v.AllowedValues) {
			return errors.New(s.ErrInvalidInt(val, v.AllowedValues...))
		}
	}

	return nil
}

//
// Musts
//

func MustIntFromEnv(envVarName string, v *IntValidation) int {
	val, err := IntFromEnv(envVarName, v)
	if err != nil {
		errors.Panic(err)
	}
	return val
}

func MustIntFromFile(filePath string, v *IntValidation) int {
	val, err := IntFromFile(filePath, v)
	if err != nil {
		errors.Panic(err)
	}
	return val
}

func MustIntFromEnvOrFile(envVarName string, filePath string, v *IntValidation) int {
	val, err := IntFromEnvOrFile(envVarName, filePath, v)
	if err != nil {
		errors.Panic(err)
	}
	return val
}
