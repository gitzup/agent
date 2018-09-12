package pkg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

var actionSchema = loadSchema("schema/action.json")
var resourceSchema = loadSchema("schema/resource.json")
var buildRequestSchema = loadSchema("schema/build.request.json", "schema/resource.json")
var buildResponseSchema = loadSchema("schema/build.response.json")
var initRequestSchema = loadSchema("schema/init.request.json")
var initResponseSchema = loadSchema("schema/init.response.json", "schema/action.json")

func GetActionSchema() *Schema        { return actionSchema }
func GetResourceSchema() *Schema      { return resourceSchema }
func GetBuildRequestSchema() *Schema  { return buildRequestSchema }
func GetBuildResponseSchema() *Schema { return buildResponseSchema }
func GetInitRequestSchema() *Schema   { return initRequestSchema }
func GetInitResponseSchema() *Schema  { return initResponseSchema }

// Compiled JSON schema.
type Schema struct {
	jsonLoader       *gojsonschema.JSONLoader
	underlyingSchema *gojsonschema.Schema
}

// Shortcut for loading a JSON schema, and panic-ing on errors
func loadSchema(mainSchema interface{}, additionalSchemas ...interface{}) *Schema {
	schema, err := NewSchema(mainSchema, additionalSchemas...)
	if err != nil {
		panic(err)
	}
	return schema
}

// Creates a JSON schema JSON loader.
func newJSONLoader(source interface{}) (*gojsonschema.JSONLoader, error) {
	switch source := source.(type) {
	case string:
		schemaBytes, err := Asset(source)
		if err != nil {
			return nil, err
		}
		jsonLoader, err := newJSONLoader(schemaBytes)
		if err != nil {
			return nil, err
		}
		return jsonLoader, nil
	case *string:
		schemaBytes, err := Asset(*source)
		if err != nil {
			return nil, err
		}
		jsonLoader, err := newJSONLoader(schemaBytes)
		if err != nil {
			return nil, err
		}
		return jsonLoader, nil
	case []byte:
		jsonLoader := gojsonschema.NewBytesLoader(source)
		return &jsonLoader, nil
	case *[]byte:
		jsonLoader := gojsonschema.NewBytesLoader(*source)
		return &jsonLoader, nil
	case Schema:
		return source.jsonLoader, nil
	case *Schema:
		return (*source).jsonLoader, nil
	default:
		jsonLoader := gojsonschema.NewGoLoader(source)
		return &jsonLoader, nil
	}
}

// Parse & compile a JSON schema from the given sources. The main schema is provided as the first argument, separated
// from any additional schemas that it may reference, that are provided in the second varargs argument.
//
// Each source may be one of:
//  - *string, string: path to an embedded asset. This is used as Asset(<value>)
//  - *[]byte, []byte: bytes containing the actual JSON schema source code
func NewSchema(mainSchemaSource interface{}, additionalSchemaSources ...interface{}) (*Schema, error) {

	// create the main schema loader
	schemaLoader := gojsonschema.NewSchemaLoader()

	// add the extra schemas (these should not include the entrypoint "main" schema)
	for _, source := range additionalSchemaSources {
		jsonLoader, err := newJSONLoader(source)
		if err != nil {
			return nil, err
		}
		err = schemaLoader.AddSchemas(*jsonLoader)
		if err != nil {
			return nil, err
		}
	}

	// compile the full schema
	jsonLoader, err := newJSONLoader(mainSchemaSource)
	underlyingSchema, err := schemaLoader.Compile(*jsonLoader)
	if err != nil {
		return nil, err
	}
	return &Schema{jsonLoader: jsonLoader, underlyingSchema: underlyingSchema}, nil
}

// Parse the JSON from the given source bytes into the given target object.
func (schema *Schema) ParseAndValidate(target interface{}, inputBytes []byte) (err error) {
	result, err := schema.underlyingSchema.Validate(gojsonschema.NewBytesLoader(inputBytes))
	if err != nil {
		return err
	}

	if !result.Valid() {
		var msg = ""
		for _, e := range result.Errors() {
			msg += fmt.Sprintf("\t- %s\n", e.String())
		}
		return errors.New(fmt.Sprintf("JSON validation failed:\n%s", msg))
	}

	// JSON is valid; translate to BuildRequest instance
	decoder := json.NewDecoder(bytes.NewReader(inputBytes))
	decoder.UseNumber()
	err = decoder.Decode(target)
	if err != nil {
		return err
	}

	return nil
}

// Validate that the given source complies with this schema.
func (schema *Schema) Validate(source interface{}) (err error) {
	var result *gojsonschema.Result

	switch source := source.(type) {
	case string:
		result, err = schema.underlyingSchema.Validate(gojsonschema.NewStringLoader(source))
	case *string:
		result, err = schema.underlyingSchema.Validate(gojsonschema.NewStringLoader(*source))
	case []byte:
		result, err = schema.underlyingSchema.Validate(gojsonschema.NewBytesLoader(source))
	case *[]byte:
		result, err = schema.underlyingSchema.Validate(gojsonschema.NewBytesLoader(*source))
	default:
		result, err = schema.underlyingSchema.Validate(gojsonschema.NewGoLoader(source))
	}

	if err != nil {
		return nil
	} else if !result.Valid() {
		var msg = ""
		for _, e := range result.Errors() {
			msg += fmt.Sprintf("\t- %s\n", e.String())
		}
		return errors.New(fmt.Sprintf("JSON validation failed:\n%s", msg))
	}

	return nil
}
