// Copyright 2013 sigu-399 ( https://github.com/sigu-399 )
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// author           sigu-399
// author-github    https://github.com/sigu-399
// author-mail      sigu.399@gmail.com
//
// repository-name  gojsonschema
// repository-desc  An implementation of JSON Schema, based on IETF's draft v4 - Go language.
//
// description      Defines schemaDocument, the main entry to every schemas.
//                  Contains the parsing logic and error checking.
//
// created          26-02-2013

package gojsonschema

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"

	"github.com/juju/gojsonreference"
)

func NewJsonSchemaDocument(document interface{}) (*JsonSchemaDocument, error) {

	internalLog("New schema document :")

	var err error

	d := JsonSchemaDocument{}
	d.pool = newSchemaPool()
	d.referencePool = newSchemaReferencePool()

	switch document.(type) {

	// document is a reference, file or http scheme
	case string:

		internalLog(fmt.Sprintf(" From http or file (%s)", document.(string)))

		d.documentReference, err = gojsonreference.NewJsonReference(document.(string))
		spd, err := d.pool.GetDocument(d.documentReference)
		if err != nil {
			return nil, err
		}

		err = d.parse(spd.Document)
		if err != nil {
			return nil, err
		}

	// document is json
	case map[string]interface{}:

		internalLog(" From map")

		d.documentReference, err = gojsonreference.NewJsonReference("#")
		d.pool.SetStandaloneDocument(document)
		if err != nil {
			return nil, err
		}

		err = d.parse(document.(map[string]interface{}))
		if err != nil {
			return nil, err
		}

	default:
		return nil, errors.New("Invalid argument, must be a jsonReference string or Json as map[string]interface{}")
	}

	return &d, nil
}

type JsonSchemaDocument struct {
	documentReference gojsonreference.JsonReference
	rootSchema        *jsonSchema
	pool              *schemaPool
	referencePool     *schemaReferencePool
}

func (d *JsonSchemaDocument) parse(document interface{}) error {
	d.rootSchema = &jsonSchema{property: ROOT_SCHEMA_PROPERTY}
	return d.parseSchema(document, d.rootSchema)
}

func (d *JsonSchemaDocument) SetRootSchemaName(name string) {
	d.rootSchema.property = name
}

// Parses a schema
//
// Pretty long function ( sorry :) )... but pretty straight forward, repetitive and boring
// Not much magic involved here, most of the job is to validate the key names and their values,
// then the values are copied into schema struct
//
func (d *JsonSchemaDocument) parseSchema(documentNode interface{}, currentSchema *jsonSchema) error {

	if internalLogEnabled {
		documentJson, err := marshalToJsonString(documentNode)
		if err == nil && documentJson != nil {
			internalLog(fmt.Sprintf("Parsing schema %s", *documentJson))
		} else {
			internalLog(fmt.Sprintf("Parsing schema %v", documentNode))
		}
	}

	if !isKind(documentNode, reflect.Map) {
		return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, STRING_SCHEMA, STRING_OBJECT))
	}

	m := documentNode.(map[string]interface{})

	if currentSchema == d.rootSchema {
		currentSchema.ref = &d.documentReference
	}

	// $schema
	if existsMapKey(m, KEY_SCHEMA) {
		if !isKind(m[KEY_SCHEMA], reflect.String) {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_SCHEMA, STRING_STRING))
		}
		schemaRef := m[KEY_SCHEMA].(string)
		schemaReference, err := gojsonreference.NewJsonReference(schemaRef)
		currentSchema.schema = &schemaReference
		if err != nil {
			return err
		}
	}

	// $ref
	if existsMapKey(m, KEY_REF) && !isKind(m[KEY_REF], reflect.String) {
		return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_REF, STRING_STRING))
	}
	if k, ok := m[KEY_REF].(string); ok {

		if sch, ok := d.referencePool.GetSchema(currentSchema.ref.String() + k); ok {

			currentSchema.refSchema = sch

		} else {
			var err error
			err = d.parseReference(documentNode, currentSchema, k)
			if err != nil {
				return err
			}

			return nil
		}
	}

	// definitions
	if existsMapKey(m, KEY_DEFINITIONS) {
		if isKind(m[KEY_DEFINITIONS], reflect.Map) {
			currentSchema.definitions = make(map[string]*jsonSchema)
			for dk, dv := range m[KEY_DEFINITIONS].(map[string]interface{}) {
				if isKind(dv, reflect.Map) {
					newSchema := &jsonSchema{property: KEY_DEFINITIONS, parent: currentSchema, ref: currentSchema.ref}
					currentSchema.definitions[dk] = newSchema
					err := d.parseSchema(dv, newSchema)
					if err != nil {
						return errors.New(err.Error())
					}
				} else {
					return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_DEFINITIONS, STRING_ARRAY_OF_SCHEMAS))
				}
			}
		} else {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_DEFINITIONS, STRING_ARRAY_OF_SCHEMAS))
		}

	}

	// id
	if existsMapKey(m, KEY_ID) && !isKind(m[KEY_ID], reflect.String) {
		return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_ID, STRING_STRING))
	}
	if k, ok := m[KEY_ID].(string); ok {
		currentSchema.id = &k
	}

	// title
	if existsMapKey(m, KEY_TITLE) && !isKind(m[KEY_TITLE], reflect.String) {
		return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_TITLE, STRING_STRING))
	}
	if k, ok := m[KEY_TITLE].(string); ok {
		currentSchema.title = &k
	}

	// description
	if existsMapKey(m, KEY_DESCRIPTION) && !isKind(m[KEY_DESCRIPTION], reflect.String) {
		return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_DESCRIPTION, STRING_STRING))
	}
	if k, ok := m[KEY_DESCRIPTION].(string); ok {
		currentSchema.description = &k
	}

	// type
	if existsMapKey(m, KEY_TYPE) {
		if isKind(m[KEY_TYPE], reflect.String) {
			if k, ok := m[KEY_TYPE].(string); ok {
				err := currentSchema.types.Add(k)
				if err != nil {
					return err
				}
			}
		} else {
			if isKind(m[KEY_TYPE], reflect.Slice) {
				arrayOfTypes := m[KEY_TYPE].([]interface{})
				for _, typeInArray := range arrayOfTypes {
					if reflect.ValueOf(typeInArray).Kind() != reflect.String {
						return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_TYPE, STRING_STRING+"/"+STRING_ARRAY_OF_STRINGS))
					} else {
						currentSchema.types.Add(typeInArray.(string))
					}
				}

			} else {
				return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_TYPE, STRING_STRING+"/"+STRING_ARRAY_OF_STRINGS))
			}
		}
	}

	// properties
	if existsMapKey(m, KEY_PROPERTIES) {
		err := d.parseProperties(m[KEY_PROPERTIES], currentSchema)
		if err != nil {
			return err
		}
	}

	// additionalProperties
	if existsMapKey(m, KEY_ADDITIONAL_PROPERTIES) {
		if isKind(m[KEY_ADDITIONAL_PROPERTIES], reflect.Bool) {
			currentSchema.additionalProperties = m[KEY_ADDITIONAL_PROPERTIES].(bool)
		} else if isKind(m[KEY_ADDITIONAL_PROPERTIES], reflect.Map) {
			newSchema := &jsonSchema{property: KEY_ADDITIONAL_PROPERTIES, parent: currentSchema, ref: currentSchema.ref}
			currentSchema.additionalProperties = newSchema
			err := d.parseSchema(m[KEY_ADDITIONAL_PROPERTIES], newSchema)
			if err != nil {
				return errors.New(err.Error())
			}
		} else {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_ADDITIONAL_PROPERTIES, STRING_BOOLEAN+"/"+STRING_SCHEMA))
		}
	}

	// patternProperties
	if existsMapKey(m, KEY_PATTERN_PROPERTIES) {
		if isKind(m[KEY_PATTERN_PROPERTIES], reflect.Map) {
			patternPropertiesMap := m[KEY_PATTERN_PROPERTIES].(map[string]interface{})
			if len(patternPropertiesMap) > 0 {
				currentSchema.patternProperties = make(map[string]*jsonSchema)
				for k, v := range patternPropertiesMap {
					_, err := regexp.MatchString(k, "")
					if err != nil {
						return errors.New(fmt.Sprintf("Invalid regex pattern '%s'", k))
					}
					newSchema := &jsonSchema{property: k, parent: currentSchema, ref: currentSchema.ref}
					err = d.parseSchema(v, newSchema)
					if err != nil {
						return errors.New(err.Error())
					}
					currentSchema.patternProperties[k] = newSchema
				}
			}
		} else {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_PATTERN_PROPERTIES, STRING_SCHEMA))
		}
	}

	// dependencies
	if existsMapKey(m, KEY_DEPENDENCIES) {
		err := d.parseDependencies(m[KEY_DEPENDENCIES], currentSchema)
		if err != nil {
			return err
		}
	}

	// items
	if existsMapKey(m, KEY_ITEMS) {
		if isKind(m[KEY_ITEMS], reflect.Slice) {
			for _, itemElement := range m[KEY_ITEMS].([]interface{}) {
				if isKind(itemElement, reflect.Map) {
					newSchema := &jsonSchema{parent: currentSchema, property: KEY_ITEMS}
					newSchema.ref = currentSchema.ref
					currentSchema.AddItemsChild(newSchema)
					err := d.parseSchema(itemElement, newSchema)
					if err != nil {
						return err
					}
				} else {
					return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_ITEMS, STRING_SCHEMA+"/"+STRING_ARRAY_OF_SCHEMAS))
				}
				currentSchema.itemsChildrenIsSingleSchema = false
			}
		} else if isKind(m[KEY_ITEMS], reflect.Map) {
			newSchema := &jsonSchema{parent: currentSchema, property: KEY_ITEMS}
			newSchema.ref = currentSchema.ref
			currentSchema.AddItemsChild(newSchema)
			err := d.parseSchema(m[KEY_ITEMS], newSchema)
			if err != nil {
				return err
			}
			currentSchema.itemsChildrenIsSingleSchema = true
		} else {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_ITEMS, STRING_SCHEMA+"/"+STRING_ARRAY_OF_SCHEMAS))
		}
	}

	// additionalItems
	if existsMapKey(m, KEY_ADDITIONAL_ITEMS) {
		if isKind(m[KEY_ADDITIONAL_ITEMS], reflect.Bool) {
			currentSchema.additionalItems = m[KEY_ADDITIONAL_ITEMS].(bool)
		} else if isKind(m[KEY_ADDITIONAL_ITEMS], reflect.Map) {
			newSchema := &jsonSchema{property: KEY_ADDITIONAL_ITEMS, parent: currentSchema, ref: currentSchema.ref}
			currentSchema.additionalItems = newSchema
			err := d.parseSchema(m[KEY_ADDITIONAL_ITEMS], newSchema)
			if err != nil {
				return errors.New(err.Error())
			}
		} else {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_ADDITIONAL_ITEMS, STRING_BOOLEAN+"/"+STRING_SCHEMA))
		}
	}

	// validation : number / integer

	if existsMapKey(m, KEY_MULTIPLE_OF) {
		if isKind(m[KEY_MULTIPLE_OF], reflect.Float64) {
			multipleOfValue := m[KEY_MULTIPLE_OF].(float64)
			if multipleOfValue <= 0 {
				return errors.New("multipleOf must be strictly greater than 0")
			}
			currentSchema.multipleOf = &multipleOfValue
		} else {
			return errors.New("multipleOf must be a number")
		}
	}

	if existsMapKey(m, KEY_MINIMUM) {
		if isKind(m[KEY_MINIMUM], reflect.Float64) {
			minimumValue := m[KEY_MINIMUM].(float64)
			currentSchema.minimum = &minimumValue
		} else if isKind(m[KEY_MINIMUM], reflect.Int) {
			minimumValue := float64(m[KEY_MINIMUM].(int))
			currentSchema.minimum = &minimumValue
		} else {
			return errors.New("minimum must be a number")
		}
	}

	if existsMapKey(m, KEY_EXCLUSIVE_MINIMUM) {
		if isKind(m[KEY_EXCLUSIVE_MINIMUM], reflect.Bool) {
			if currentSchema.minimum == nil {
				return errors.New("exclusiveMinimum cannot exist without minimum")
			}
			exclusiveMinimumValue := m[KEY_EXCLUSIVE_MINIMUM].(bool)
			currentSchema.exclusiveMinimum = exclusiveMinimumValue
		} else {
			return errors.New("exclusiveMinimum must be a boolean")
		}
	}

	if existsMapKey(m, KEY_MAXIMUM) {
		if isKind(m[KEY_MAXIMUM], reflect.Float64) {
			maximumValue := m[KEY_MAXIMUM].(float64)
			currentSchema.maximum = &maximumValue
		} else if isKind(m[KEY_MAXIMUM], reflect.Int) {
			maximumValue := float64(m[KEY_MAXIMUM].(int))
			currentSchema.maximum = &maximumValue
		} else {
			return errors.New("maximum must be a number")
		}
	}

	if existsMapKey(m, KEY_EXCLUSIVE_MAXIMUM) {
		if isKind(m[KEY_EXCLUSIVE_MAXIMUM], reflect.Bool) {
			if currentSchema.maximum == nil {
				return errors.New("exclusiveMaximum cannot exist without maximum")
			}
			exclusiveMaximumValue := m[KEY_EXCLUSIVE_MAXIMUM].(bool)
			currentSchema.exclusiveMaximum = exclusiveMaximumValue
		} else {
			return errors.New("exclusiveMaximum must be a boolean")
		}
	}

	if currentSchema.minimum != nil && currentSchema.maximum != nil {
		if *currentSchema.minimum > *currentSchema.maximum {
			return errors.New("minimum cannot be greater than maximum")
		}
	}

	// validation : string

	if existsMapKey(m, KEY_MIN_LENGTH) {
		if isKind(m[KEY_MIN_LENGTH], reflect.Float64) {
			minLengthValue := m[KEY_MIN_LENGTH].(float64)
			if isFloat64AnInteger(minLengthValue) {
				if minLengthValue < 0 {
					return errors.New("minLength must be greater than or equal to 0")
				}
				minLengthIntegerValue := int(minLengthValue)
				currentSchema.minLength = &minLengthIntegerValue
			} else {
				return errors.New("minLength must be an integer")
			}
		} else {
			return errors.New("minLength must be an integer")
		}
	}

	if existsMapKey(m, KEY_MAX_LENGTH) {
		if isKind(m[KEY_MAX_LENGTH], reflect.Float64) {
			maxLengthValue := m[KEY_MAX_LENGTH].(float64)
			if isFloat64AnInteger(maxLengthValue) {
				if maxLengthValue < 0 {
					return errors.New("maxLength must be greater than or equal to 0")
				}
				maxLengthIntegerValue := int(maxLengthValue)
				currentSchema.maxLength = &maxLengthIntegerValue
			} else {
				return errors.New("maxLength must be an integer")
			}
		} else {
			return errors.New("maxLength must be an integer")
		}
	}

	if currentSchema.minLength != nil && currentSchema.maxLength != nil {
		if *currentSchema.minLength > *currentSchema.maxLength {
			return errors.New("minLength cannot be greater than maxLength")
		}
	}

	if existsMapKey(m, KEY_PATTERN) {
		if isKind(m[KEY_PATTERN], reflect.String) {
			regexpObject, err := regexp.Compile(m[KEY_PATTERN].(string))
			if err != nil {
				return errors.New("pattern must be a valid regular expression")
			}
			currentSchema.pattern = regexpObject
		} else {
			return errors.New("pattern must be a string")
		}
	}

	// validation : object

	if existsMapKey(m, KEY_MIN_PROPERTIES) {
		if isKind(m[KEY_MIN_PROPERTIES], reflect.Float64) {
			minPropertiesValue := m[KEY_MIN_PROPERTIES].(float64)
			if isFloat64AnInteger(minPropertiesValue) {
				if minPropertiesValue < 0 {
					return errors.New("minProperties must be greater than or equal to 0")
				}
				minPropertiesntegerValue := int(minPropertiesValue)
				currentSchema.minProperties = &minPropertiesntegerValue
			} else {
				return errors.New("minProperties must be an integer")
			}
		} else {
			return errors.New("minProperties must be an integer")
		}
	}

	if existsMapKey(m, KEY_MAX_PROPERTIES) {
		if isKind(m[KEY_MAX_PROPERTIES], reflect.Float64) {
			maxPropertiesValue := m[KEY_MAX_PROPERTIES].(float64)
			if isFloat64AnInteger(maxPropertiesValue) {
				if maxPropertiesValue < 0 {
					return errors.New("maxProperties must be greater than or equal to 0")
				}
				maxPropertiesntegerValue := int(maxPropertiesValue)
				currentSchema.maxProperties = &maxPropertiesntegerValue
			} else {
				return errors.New("maxProperties must be an integer")
			}
		} else {
			return errors.New("maxProperties must be an integer")
		}
	}

	if currentSchema.minProperties != nil && currentSchema.maxProperties != nil {
		if *currentSchema.minProperties > *currentSchema.maxProperties {
			return errors.New("minProperties cannot be greater than maxProperties")
		}
	}

	if existsMapKey(m, KEY_REQUIRED) {
		if isKind(m[KEY_REQUIRED], reflect.Slice) {
			requiredValues := m[KEY_REQUIRED].([]interface{})
			for _, requiredValue := range requiredValues {
				if isKind(requiredValue, reflect.String) {
					err := currentSchema.AddRequired(requiredValue.(string))
					if err != nil {
						return err
					}
				} else {
					return errors.New("required items must be string")
				}
			}
		} else {
			return errors.New("required must be an array")
		}
	}

	// validation : array

	if existsMapKey(m, KEY_MIN_ITEMS) {
		if isKind(m[KEY_MIN_ITEMS], reflect.Float64) {
			minItemsValue := m[KEY_MIN_ITEMS].(float64)
			if isFloat64AnInteger(minItemsValue) {
				if minItemsValue < 0 {
					return errors.New("minItems must be greater than or equal to 0")
				}
				minItemsIntegerValue := int(minItemsValue)
				currentSchema.minItems = &minItemsIntegerValue
			} else {
				return errors.New("minItems must be an integer")
			}
		} else {
			return errors.New("minItems must be an integer")
		}
	}

	if existsMapKey(m, KEY_MAX_ITEMS) {
		if isKind(m[KEY_MAX_ITEMS], reflect.Float64) {
			maxItemsValue := m[KEY_MAX_ITEMS].(float64)
			if isFloat64AnInteger(maxItemsValue) {
				if maxItemsValue < 0 {
					return errors.New("maxItems must be greater than or equal to 0")
				}
				maxItemsIntegerValue := int(maxItemsValue)
				currentSchema.maxItems = &maxItemsIntegerValue
			} else {
				return errors.New("maxItems must be an integer")
			}
		} else {
			return errors.New("maxItems must be an integer")
		}
	}

	if existsMapKey(m, KEY_UNIQUE_ITEMS) {
		if isKind(m[KEY_UNIQUE_ITEMS], reflect.Bool) {
			currentSchema.uniqueItems = m[KEY_UNIQUE_ITEMS].(bool)
		} else {
			return errors.New("uniqueItems must be an boolean")
		}
	}

	// validation : all

	if existsMapKey(m, KEY_ENUM) {
		if isKind(m[KEY_ENUM], reflect.Slice) {
			for _, v := range m[KEY_ENUM].([]interface{}) {
				err := currentSchema.AddEnum(v)
				if err != nil {
					return err
				}
			}
		} else {
			return errors.New("enum must be an array")
		}
	}

	// validation : schema

	if existsMapKey(m, KEY_ONE_OF) {
		if isKind(m[KEY_ONE_OF], reflect.Slice) {
			for _, v := range m[KEY_ONE_OF].([]interface{}) {
				newSchema := &jsonSchema{property: KEY_ONE_OF, parent: currentSchema, ref: currentSchema.ref}
				currentSchema.AddOneOf(newSchema)
				err := d.parseSchema(v, newSchema)
				if err != nil {
					return err
				}
			}
		} else {
			return errors.New("oneOf must be an array")
		}
	}

	if existsMapKey(m, KEY_ANY_OF) {
		if isKind(m[KEY_ANY_OF], reflect.Slice) {
			for _, v := range m[KEY_ANY_OF].([]interface{}) {
				newSchema := &jsonSchema{property: KEY_ANY_OF, parent: currentSchema, ref: currentSchema.ref}
				currentSchema.AddAnyOf(newSchema)
				err := d.parseSchema(v, newSchema)
				if err != nil {
					return err
				}
			}
		} else {
			return errors.New("anyOf must be an array")
		}
	}

	if existsMapKey(m, KEY_ALL_OF) {
		if isKind(m[KEY_ALL_OF], reflect.Slice) {
			for _, v := range m[KEY_ALL_OF].([]interface{}) {
				newSchema := &jsonSchema{property: KEY_ALL_OF, parent: currentSchema, ref: currentSchema.ref}
				currentSchema.AddAllOf(newSchema)
				err := d.parseSchema(v, newSchema)
				if err != nil {
					return err
				}
			}
		} else {
			return errors.New("anyOf must be an array")
		}
	}

	if existsMapKey(m, KEY_NOT) {
		if isKind(m[KEY_NOT], reflect.Map) {
			newSchema := &jsonSchema{property: KEY_NOT, parent: currentSchema, ref: currentSchema.ref}
			currentSchema.SetNot(newSchema)
			err := d.parseSchema(m[KEY_NOT], newSchema)
			if err != nil {
				return err
			}
		} else {
			return errors.New("not must be an object")
		}
	}

	return nil
}

func (d *JsonSchemaDocument) parseReference(documentNode interface{}, currentSchema *jsonSchema, reference string) (e error) {

	var err error

	jsonReference, err := gojsonreference.NewJsonReference(reference)
	if err != nil {
		return err
	}

	standaloneDocument := d.pool.GetStandaloneDocument()

	if jsonReference.GetUrl().IsAbs() || standaloneDocument != nil {
		currentSchema.ref = &jsonReference
	} else {
		inheritedReference, err := currentSchema.ref.Inherits(jsonReference)
		if err != nil {
			return err
		}
		currentSchema.ref = inheritedReference
	}

	jsonPointer := currentSchema.ref.GetPointer()

	var refdDocumentNode interface{}

	if standaloneDocument != nil {

		var err error
		refdDocumentNode, _, err = jsonPointer.Get(standaloneDocument)
		if err != nil {
			return err
		}

	} else {

		var err error
		dsp, err := d.pool.GetDocument(*currentSchema.ref)
		if err != nil {
			return err
		}

		refdDocumentNode, _, err = jsonPointer.Get(dsp.Document)
		if err != nil {
			return err
		}
	}

	if !isKind(refdDocumentNode, reflect.Map) {
		return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, STRING_SCHEMA, STRING_OBJECT))
	}

	// returns the loaded referenced schema for the caller to update its current schema
	newSchemaDocument := refdDocumentNode.(map[string]interface{})

	newSchema := &jsonSchema{property: KEY_REF, parent: currentSchema, ref: currentSchema.ref}
	d.referencePool.AddSchema(currentSchema.ref.String()+reference, newSchema)

	err = d.parseSchema(newSchemaDocument, newSchema)
	if err != nil {
		return err
	}

	currentSchema.refSchema = newSchema

	return nil

}

func (d *JsonSchemaDocument) parseProperties(documentNode interface{}, currentSchema *jsonSchema) error {

	if !isKind(documentNode, reflect.Map) {
		return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, STRING_PROPERTIES, STRING_OBJECT))
	}

	m := documentNode.(map[string]interface{})
	for k := range m {
		schemaProperty := k
		newSchema := &jsonSchema{property: schemaProperty, parent: currentSchema, ref: currentSchema.ref}
		currentSchema.AddPropertiesChild(newSchema)
		err := d.parseSchema(m[k], newSchema)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *JsonSchemaDocument) parseDependencies(documentNode interface{}, currentSchema *jsonSchema) error {

	if !isKind(documentNode, reflect.Map) {
		return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_DEPENDENCIES, STRING_OBJECT))
	}

	m := documentNode.(map[string]interface{})
	currentSchema.dependencies = make(map[string]interface{})

	for k := range m {
		switch reflect.ValueOf(m[k]).Kind() {

		case reflect.Slice:
			values := m[k].([]interface{})
			var valuesToRegister []string

			for _, value := range values {
				if !isKind(value, reflect.String) {
					return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, STRING_DEPENDENCY, STRING_SCHEMA_OR_ARRAY_OF_STRINGS))
				} else {
					valuesToRegister = append(valuesToRegister, value.(string))
				}
				currentSchema.dependencies[k] = valuesToRegister
			}

		case reflect.Map:
			depSchema := &jsonSchema{property: k, parent: currentSchema, ref: currentSchema.ref}
			err := d.parseSchema(m[k], depSchema)
			if err != nil {
				return err
			}
			currentSchema.dependencies[k] = depSchema

		default:
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, STRING_DEPENDENCY, STRING_SCHEMA_OR_ARRAY_OF_STRINGS))
		}

	}

	return nil
}
