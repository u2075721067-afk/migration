package validator

// ValidatorInterface defines the interface for envelope validation
type ValidatorInterface interface {
	ValidateEnvelope(file string) (bool, []error)
}
