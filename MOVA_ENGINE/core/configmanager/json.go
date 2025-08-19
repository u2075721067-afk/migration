package configmanager

import (
	"encoding/json"
	"fmt"
	"io"
)

// JSONExporter implements ConfigExporter for JSON format
type JSONExporter struct{}

// Export exports a ConfigBundle to JSON format
func (e *JSONExporter) Export(bundle *ConfigBundle) ([]byte, error) {
	data, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return data, nil
}

// ExportToWriter exports a ConfigBundle to JSON format to an io.Writer
func (e *JSONExporter) ExportToWriter(bundle *ConfigBundle, w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(bundle)
}

// GetFormat returns the format this exporter handles
func (e *JSONExporter) GetFormat() ConfigFormat {
	return FormatJSON
}

// JSONImporter implements ConfigImporter for JSON format
type JSONImporter struct{}

// Import imports configuration data from JSON format to a ConfigBundle
func (i *JSONImporter) Import(data []byte) (*ConfigBundle, error) {
	var bundle ConfigBundle
	if err := json.Unmarshal(data, &bundle); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return &bundle, nil
}

// ImportFromReader imports configuration data from JSON format from an io.Reader
func (i *JSONImporter) ImportFromReader(r io.Reader) (*ConfigBundle, error) {
	var bundle ConfigBundle
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&bundle); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}
	return &bundle, nil
}

// GetFormat returns the format this importer handles
func (i *JSONImporter) GetFormat() ConfigFormat {
	return FormatJSON
}
