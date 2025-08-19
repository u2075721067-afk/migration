package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mova-engine/mova-engine/core/configmanager"
)

// ConfigHandler handles configuration export/import operations
type ConfigHandler struct {
	configManager configmanager.ConfigManager
}

// NewConfigHandler creates a new configuration handler
func NewConfigHandler(configManager configmanager.ConfigManager) *ConfigHandler {
	return &ConfigHandler{
		configManager: configManager,
	}
}

// ExportConfig handles configuration export requests
func (h *ConfigHandler) ExportConfig(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse query parameters
	format := c.Query("format")
	if format == "" {
		format = "yaml" // Default to YAML
	}

	includeDLQ := c.Query("include_dlq") == "true"
	includeWorkflows := c.Query("include_workflows") == "true"
	compress := c.Query("compress") == "true"

	// Validate format
	configFormat := configmanager.ConfigFormat(format)
	supportedFormats := h.configManager.GetSupportedFormats()
	formatSupported := false
	for _, f := range supportedFormats {
		if f == configFormat {
			formatSupported = true
			break
		}
	}

	if !formatSupported {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("unsupported format: %s. supported formats: %v", format, supportedFormats),
		})
		return
	}

	// Create export options
	opts := configmanager.ExportOptions{
		Format:           configFormat,
		IncludeDLQ:       includeDLQ,
		IncludeWorkflows: includeWorkflows,
		Compress:         compress,
	}

	// Export configuration
	_, err := h.configManager.Export(ctx, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to export configuration: %v", err),
		})
		return
	}

	// Set response headers
	filename := fmt.Sprintf("mova-config-%s.%s", time.Now().Format("2006-01-02"), format)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", getContentType(format))

	// Export to response writer
	if err := h.configManager.ExportToWriter(ctx, opts, c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to write configuration: %v", err),
		})
		return
	}
}

// ImportConfig handles configuration import requests
func (h *ConfigHandler) ImportConfig(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse query parameters
	format := c.Query("format")
	if format == "" {
		// Try to detect format from Content-Type
		contentType := c.GetHeader("Content-Type")
		format = detectFormatFromContentType(contentType)
	}

	if format == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "format is required. specify format query parameter or set Content-Type header",
		})
		return
	}

	// Validate format
	configFormat := configmanager.ConfigFormat(format)
	supportedFormats := h.configManager.GetSupportedFormats()
	formatSupported := false
	for _, f := range supportedFormats {
		if f == configFormat {
			formatSupported = true
			break
		}
	}

	if !formatSupported {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("unsupported format: %s. supported formats: %v", format, supportedFormats),
		})
		return
	}

	// Parse import options
	mode := c.Query("mode")
	if mode == "" {
		mode = "merge" // Default to merge mode
	}

	validateOnly := c.Query("validate_only") == "true"
	dryRun := c.Query("dry_run") == "true"
	overwrite := c.Query("overwrite") == "true"

	// Validate mode
	importMode := configmanager.ImportMode(mode)
	if importMode != configmanager.ModeOverwrite &&
		importMode != configmanager.ModeMerge &&
		importMode != configmanager.ModeValidate {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("invalid import mode: %s. valid modes: overwrite, merge, validate", mode),
		})
		return
	}

	// Create import options
	opts := configmanager.ImportOptions{
		Format:       configFormat,
		Mode:         importMode,
		ValidateOnly: validateOnly,
		DryRun:       dryRun,
		Overwrite:    overwrite,
	}

	// Read request body
	data, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("failed to read request body: %v", err),
		})
		return
	}

	if len(data) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "request body is empty",
		})
		return
	}

	// Import configuration
	result, err := h.configManager.Import(ctx, data, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to import configuration: %v", err),
		})
		return
	}

	// Return import result
	statusCode := http.StatusOK
	if !result.Success {
		statusCode = http.StatusBadRequest
	}

	c.JSON(statusCode, result)
}

// ValidateConfig handles configuration validation requests
func (h *ConfigHandler) ValidateConfig(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse query parameters
	format := c.Query("format")
	if format == "" {
		// Try to detect format from Content-Type
		contentType := c.GetHeader("Content-Type")
		format = detectFormatFromContentType(contentType)
	}

	if format == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "format is required. specify format query parameter or set Content-Type header",
		})
		return
	}

	// Validate format
	configFormat := configmanager.ConfigFormat(format)
	supportedFormats := h.configManager.GetSupportedFormats()
	formatSupported := false
	for _, f := range supportedFormats {
		if f == configFormat {
			formatSupported = true
			break
		}
	}

	if !formatSupported {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("unsupported format: %s. supported formats: %v", format, supportedFormats),
		})
		return
	}

	// Read request body
	data, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("failed to read request body: %v", err),
		})
		return
	}

	if len(data) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "request body is empty",
		})
		return
	}

	// Validate configuration
	errors, err := h.configManager.Validate(ctx, data, configFormat)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("validation failed: %v", err),
		})
		return
	}

	// Return validation result
	result := gin.H{
		"valid":  len(errors) == 0,
		"errors": errors,
		"count":  len(errors),
	}

	statusCode := http.StatusOK
	if len(errors) > 0 {
		statusCode = http.StatusBadRequest
	}

	c.JSON(statusCode, result)
}

// GetSupportedFormats returns list of supported export/import formats
func (h *ConfigHandler) GetSupportedFormats(c *gin.Context) {
	formats := h.configManager.GetSupportedFormats()

	// Convert to string slice
	formatStrings := make([]string, len(formats))
	for i, format := range formats {
		formatStrings[i] = string(format)
	}

	c.JSON(http.StatusOK, gin.H{
		"formats": formatStrings,
		"default": "yaml",
	})
}

// GetConfigInfo returns configuration information
func (h *ConfigHandler) GetConfigInfo(c *gin.Context) {
	version := h.configManager.GetVersion()
	supportedFormats := h.configManager.GetSupportedFormats()

	// Convert to string slice
	formatStrings := make([]string, len(supportedFormats))
	for i, format := range supportedFormats {
		formatStrings[i] = string(format)
	}

	c.JSON(http.StatusOK, gin.H{
		"version":           version,
		"supported_formats": formatStrings,
		"default_format":    "yaml",
		"features": []string{
			"export_policies",
			"export_budgets",
			"export_retry_profiles",
			"export_dlq_entries",
			"export_workflows",
			"import_merge",
			"import_overwrite",
			"validation",
			"dry_run",
		},
	})
}

// getContentType returns the appropriate Content-Type for the given format
func getContentType(format string) string {
	switch strings.ToLower(format) {
	case "json":
		return "application/json"
	case "yaml", "yml":
		return "text/yaml"
	case "hcl":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}

// detectFormatFromContentType attempts to detect format from Content-Type header
func detectFormatFromContentType(contentType string) string {
	if strings.Contains(contentType, "application/json") {
		return "json"
	}
	if strings.Contains(contentType, "text/yaml") || strings.Contains(contentType, "application/yaml") {
		return "yaml"
	}
	if strings.Contains(contentType, "text/plain") && strings.Contains(contentType, "hcl") {
		return "hcl"
	}
	return ""
}
