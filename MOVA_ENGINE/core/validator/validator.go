package validator

import (
	"fmt"
	"path/filepath"

	"github.com/xeipuuv/gojsonschema"
)

// Validator provides JSON Schema validation for MOVA envelopes
type Validator struct {
	schemaDir      string
	envelopeSchema *gojsonschema.Schema
}

// NewValidator creates a new validator instance pointing to a schemas directory
func NewValidator(schemaDir string) (*Validator, error) {
	absSchemaDir, err := filepath.Abs(schemaDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve schema dir: %w", err)
	}

	// Load the root envelope schema with a canonical file:// URL so relative $ref work
	envelopeSchemaPath := filepath.Join(absSchemaDir, "envelope.json")
	schemaLoader := gojsonschema.NewReferenceLoader("file://" + envelopeSchemaPath)
	envelopeSchema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return nil, fmt.Errorf("failed to load envelope schema: %w", err)
	}

	return &Validator{
		schemaDir:      absSchemaDir,
		envelopeSchema: envelopeSchema,
	}, nil
}

// ValidateEnvelope validates a MOVA envelope JSON file path against the schema
func (v *Validator) ValidateEnvelope(file string) (bool, []error) {
	absFile, err := filepath.Abs(file)
	if err != nil {
		return false, []error{fmt.Errorf("failed to resolve envelope file: %w", err)}
	}

	docLoader := gojsonschema.NewReferenceLoader("file://" + absFile)
	res, err := v.envelopeSchema.Validate(docLoader)
	if err != nil {
		return false, []error{fmt.Errorf("validation failed: %w", err)}
	}

	if res.Valid() {
		return true, nil
	}

	errs := make([]error, 0, len(res.Errors()))
	for _, e := range res.Errors() {
		errs = append(errs, fmt.Errorf("%s: %s", e.Field(), e.Description()))
	}
	return false, errs
}
