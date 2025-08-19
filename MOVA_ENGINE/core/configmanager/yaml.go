package configmanager

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// YAMLExporter implements ConfigExporter for YAML format
type YAMLExporter struct{}

// Export exports a ConfigBundle to YAML format
func (e *YAMLExporter) Export(bundle *ConfigBundle) ([]byte, error) {
	data, err := yaml.Marshal(bundle)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal YAML: %w", err)
	}
	return data, nil
}

// ExportToWriter exports a ConfigBundle to YAML format to an io.Writer
func (e *YAMLExporter) ExportToWriter(bundle *ConfigBundle, w io.Writer) error {
	encoder := yaml.NewEncoder(w)
	defer encoder.Close()
	return encoder.Encode(bundle)
}

// GetFormat returns the format this exporter handles
func (e *YAMLExporter) GetFormat() ConfigFormat {
	return FormatYAML
}

// YAMLImporter implements ConfigImporter for YAML format
type YAMLImporter struct{}

// Import imports configuration data from YAML format to a ConfigBundle
func (i *YAMLImporter) Import(data []byte) (*ConfigBundle, error) {
	var bundle ConfigBundle
	if err := yaml.Unmarshal(data, &bundle); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}
	return &bundle, nil
}

// ImportFromReader imports configuration data from YAML format from an io.Reader
func (i *YAMLImporter) ImportFromReader(r io.Reader) (*ConfigBundle, error) {
	var bundle ConfigBundle
	decoder := yaml.NewDecoder(r)
	if err := decoder.Decode(&bundle); err != nil {
		return nil, fmt.Errorf("failed to decode YAML: %w", err)
	}
	return &bundle, nil
}

// GetFormat returns the format this importer handles
func (i *YAMLImporter) GetFormat() ConfigFormat {
	return FormatYAML
}
