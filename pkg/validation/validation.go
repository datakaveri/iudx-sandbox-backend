package validation

import (
	"net/mail"
	"regexp"
	"strings"
	"unicode"
)

var (
	// Common validation patterns
	uuidPattern         = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	alphanumericPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	domainPattern       = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	tagPattern          = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

	// Security patterns to detect malicious input
	sqlInjectionPattern = regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|script|javascript|vbscript|onload|onerror)`)
	xssPattern          = regexp.MustCompile(`(?i)(<script|javascript:|vbscript:|onload=|onerror=|<iframe|<object|<embed)`)
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (v ValidationError) Error() string {
	return v.Message
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return ""
	}
	if len(v) == 1 {
		return v[0].Error()
	}
	return "multiple validation errors occurred"
}

// HasErrors returns true if there are validation errors
func (v ValidationErrors) HasErrors() bool {
	return len(v) > 0
}

// Add adds a validation error
func (v *ValidationErrors) Add(field, code, message string) {
	*v = append(*v, ValidationError{
		Field:   field,
		Code:    code,
		Message: message,
	})
}

// Validator provides validation methods
type Validator struct {
	errors ValidationErrors
}

// New creates a new validator
func New() *Validator {
	return &Validator{
		errors: make(ValidationErrors, 0),
	}
}

// HasErrors returns true if there are validation errors
func (v *Validator) HasErrors() bool {
	return v.errors.HasErrors()
}

// GetErrors returns all validation errors
func (v *Validator) GetErrors() ValidationErrors {
	return v.errors
}

// Required validates that a field is not empty
func (v *Validator) Required(field, value, message string) *Validator {
	if strings.TrimSpace(value) == "" {
		if message == "" {
			message = field + " is required"
		}
		v.errors.Add(field, "required", message)
	}
	return v
}

// MaxLength validates maximum string length
func (v *Validator) MaxLength(field, value string, maxLen int, message string) *Validator {
	if len(value) > maxLen {
		if message == "" {
			message = field + " must not exceed " + string(rune(maxLen)) + " characters"
		}
		v.errors.Add(field, "max_length", message)
	}
	return v
}

// MinLength validates minimum string length
func (v *Validator) MinLength(field, value string, minLen int, message string) *Validator {
	if len(strings.TrimSpace(value)) < minLen && value != "" {
		if message == "" {
			message = field + " must be at least " + string(rune(minLen)) + " characters"
		}
		v.errors.Add(field, "min_length", message)
	}
	return v
}

// UUID validates UUID format
func (v *Validator) UUID(field, value, message string) *Validator {
	if value != "" && !uuidPattern.MatchString(strings.ToLower(value)) {
		if message == "" {
			message = field + " must be a valid UUID"
		}
		v.errors.Add(field, "invalid_uuid", message)
	}
	return v
}

// Email validates email format
func (v *Validator) Email(field, value, message string) *Validator {
	if value != "" {
		_, err := mail.ParseAddress(value)
		if err != nil {
			if message == "" {
				message = field + " must be a valid email address"
			}
			v.errors.Add(field, "invalid_email", message)
		}
	}
	return v
}

// Alphanumeric validates alphanumeric characters only
func (v *Validator) Alphanumeric(field, value, message string) *Validator {
	if value != "" && !alphanumericPattern.MatchString(value) {
		if message == "" {
			message = field + " must contain only alphanumeric characters, hyphens, and underscores"
		}
		v.errors.Add(field, "invalid_alphanumeric", message)
	}
	return v
}

// Domain validates domain name format
func (v *Validator) Domain(field, value, message string) *Validator {
	if value != "" && !domainPattern.MatchString(value) {
		if message == "" {
			message = field + " must be a valid domain name"
		}
		v.errors.Add(field, "invalid_domain", message)
	}
	return v
}

// Tag validates tag format
func (v *Validator) Tag(field, value, message string) *Validator {
	if value != "" && !tagPattern.MatchString(value) {
		if message == "" {
			message = field + " must contain only alphanumeric characters, hyphens, and underscores"
		}
		v.errors.Add(field, "invalid_tag", message)
	}
	return v
}

// NoSQLInjection checks for SQL injection patterns
func (v *Validator) NoSQLInjection(field, value, message string) *Validator {
	if value != "" && sqlInjectionPattern.MatchString(value) {
		if message == "" {
			message = field + " contains invalid characters"
		}
		v.errors.Add(field, "security_violation", message)
	}
	return v
}

// NoXSS checks for XSS patterns
func (v *Validator) NoXSS(field, value, message string) *Validator {
	if value != "" && xssPattern.MatchString(value) {
		if message == "" {
			message = field + " contains invalid characters"
		}
		v.errors.Add(field, "security_violation", message)
	}
	return v
}

// SafeString performs comprehensive string validation for security
func (v *Validator) SafeString(field, value, message string) *Validator {
	v.NoSQLInjection(field, value, message)
	v.NoXSS(field, value, message)

	// Check for null bytes
	if strings.Contains(value, "\x00") {
		if message == "" {
			message = field + " contains invalid characters"
		}
		v.errors.Add(field, "security_violation", message)
	}

	// Check for excessive control characters
	controlCharCount := 0
	for _, r := range value {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			controlCharCount++
		}
	}
	if controlCharCount > 0 {
		if message == "" {
			message = field + " contains invalid control characters"
		}
		v.errors.Add(field, "security_violation", message)
	}

	return v
}

// Enum validates that value is one of allowed values
func (v *Validator) Enum(field, value string, allowedValues []string, message string) *Validator {
	if value != "" {
		valid := false
		for _, allowed := range allowedValues {
			if value == allowed {
				valid = true
				break
			}
		}
		if !valid {
			if message == "" {
				message = field + " must be one of: " + strings.Join(allowedValues, ", ")
			}
			v.errors.Add(field, "invalid_enum", message)
		}
	}
	return v
}

// ValidateDatasetInput validates dataset input data
func ValidateDatasetInput(title, description, domain string, tags []string) error {
	v := New()

	v.Required("title", title, "").
		MaxLength("title", title, 200, "").
		SafeString("title", title, "")

	v.MaxLength("description", description, 2000, "").
		SafeString("description", description, "")

	v.Required("domain", domain, "").
		Domain("domain", domain, "")

	// Validate tags
	for i, tag := range tags {
		fieldName := "tags[" + string(rune(i)) + "]"
		v.Tag(fieldName, tag, "").
			MaxLength(fieldName, tag, 50, "")
	}

	if v.HasErrors() {
		return v.GetErrors()
	}
	return nil
}

// ValidateResourceInput validates resource input data
func ValidateResourceInput(resourceID, datasetID, name string) error {
	v := New()

	v.Required("resource_id", resourceID, "").
		UUID("resource_id", resourceID, "")

	v.Required("dataset_id", datasetID, "").
		UUID("dataset_id", datasetID, "")

	v.Required("name", name, "").
		MaxLength("name", name, 200, "").
		SafeString("name", name, "")

	if v.HasErrors() {
		return v.GetErrors()
	}
	return nil
}

// ValidateNotebookInput validates notebook input data
func ValidateNotebookInput(name, image string) error {
	v := New()

	v.Required("name", name, "").
		MaxLength("name", name, 100, "").
		Alphanumeric("name", name, "")

	v.Required("image", image, "").
		MaxLength("image", image, 500, "").
		SafeString("image", image, "")

	if v.HasErrors() {
		return v.GetErrors()
	}
	return nil
}
