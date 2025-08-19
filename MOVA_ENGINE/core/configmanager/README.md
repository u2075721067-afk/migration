# MOVA Config Manager

A comprehensive configuration management system for the MOVA Automation Engine that supports export/import operations in multiple formats with validation and version control.

## Quick Start

```go
package main

import (
    "context"
    "github.com/u2075721067-afk/MOVA/core/configmanager"
)

func main() {
    // Create storage and validator
    storage := NewMockStorage() // or real storage implementation
    validator := configmanager.NewValidator()
    
    // Create config manager
    manager := configmanager.NewManager(storage, validator)
    
    // Export configuration to YAML
    opts := configmanager.ExportOptions{
        Format: configmanager.FormatYAML,
        IncludeDLQ: true,
    }
    
    bundle, err := manager.Export(context.Background(), opts)
    if err != nil {
        panic(err)
    }
    
    // Import configuration from JSON
    importOpts := configmanager.ImportOptions{
        Format: configmanager.FormatJSON,
        Mode:   configmanager.ModeMerge,
    }
    
    result, err := manager.Import(context.Background(), jsonData, importOpts)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Imported %d items\n", result.Imported)
}
```

## Features

- **Multi-format Support**: JSON, YAML, HCL
- **Validation**: Comprehensive configuration validation
- **Import Modes**: Merge, overwrite, validate-only
- **Dry-run**: Preview changes without applying
- **Version Control**: Built-in versioning and checksums
- **Terraform Integration**: HCL export for IaC

## Supported Formats

| Format | Export | Import | Notes |
|--------|--------|--------|-------|
| JSON   | ✅     | ✅     | Internal storage format |
| YAML   | ✅     | ✅     | Human-readable, default |
| HCL    | ✅     | ⚠️     | Export only (Terraform) |

## API Usage

### Export Configuration
```go
opts := configmanager.ExportOptions{
    Format:         configmanager.FormatYAML,
    IncludeDLQ:     true,
    IncludeWorkflows: true,
    Compress:       false,
}

bundle, err := manager.Export(ctx, opts)
```

### Import Configuration
```go
opts := configmanager.ImportOptions{
    Format:      configmanager.FormatYAML,
    Mode:        configmanager.ModeMerge,
    ValidateOnly: false,
    DryRun:      false,
    Overwrite:   false,
}

result, err := manager.Import(ctx, data, opts)
```

### Validate Configuration
```go
errors, err := manager.Validate(ctx, data, configmanager.FormatYAML)
if len(errors) > 0 {
    for _, err := range errors {
        fmt.Printf("Error: %s - %s\n", err.Type, err.Message)
    }
}
```

## CLI Commands

```bash
# Export to YAML
mova config export --out=policies.yaml

# Import with merge
mova config import policies.yaml --mode=merge

# Validate only
mova config validate policies.yaml

# Show info
mova config info
```

## Configuration Objects

### Policies
- Retry policies with conditions
- Budget constraints
- Error handling rules

### Budgets
- Resource limits (CPU, memory, API calls)
- Time-based constraints
- Scope-based enforcement

### Retry Profiles
- Exponential backoff configurations
- Jitter and timeout settings

### DLQ Entries
- Dead letter queue configurations
- Retry count tracking
- Error payload preservation

## Validation Rules

- Required fields validation
- Data type validation
- Logical constraint validation
- Business rule validation

## Error Handling

- Detailed error messages
- Line and column information
- Error categorization
- Partial import support

## Examples

See the `examples/` directory for complete working examples:

- [Basic Export/Import](examples/basic.go)
- [Validation](examples/validation.go)
- [Custom Formats](examples/custom-formats.go)
- [Error Handling](examples/error-handling.go)

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestExport
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

MIT License - see LICENSE file for details.
