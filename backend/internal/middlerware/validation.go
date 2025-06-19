package middlerware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"go-chat/pkg"
)

var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	nameRegex     = regexp.MustCompile(`^[a-zA-Z\s\-']{2,50}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,30}$`)
	passwordRegex = regexp.MustCompile(`^.{8,28}$`)
)

type ValidationRules struct {
	MaxBodySize    int64
	RequiredFields []string
	EmailFields    []string
	NameFields     []string
	UsernameFields []string
	PasswordFields []string
	Sanitize       bool
}

var validationRules = map[string]ValidationRules{
	"signup": {
		MaxBodySize:    1024,
		RequiredFields: []string{"email", "password", "name"},
		EmailFields:    []string{"email"},
		NameFields:     []string{"name"},
		PasswordFields: []string{"password"},
		Sanitize:       true,
	},
	"login": {
		MaxBodySize:    512,
		RequiredFields: []string{"email", "password"},
		EmailFields:    []string{"email"},
		PasswordFields: []string{"password"},
		Sanitize:       true,
	},
	"message": {
		MaxBodySize:    10240,
		RequiredFields: []string{"receiver_id", "content"},
		Sanitize:       true,
	},
	"profile": {
		MaxBodySize:    2048,
		NameFields:     []string{"name"},
		Sanitize:       true,
	},
	"friend_request": {
		MaxBodySize:    256,
		RequiredFields: []string{"receiver_id"},
		Sanitize:       true,
	},
}

func ValidateRequest(endpoint string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rules, exists := validationRules[endpoint]
			if !exists {
				rules = ValidationRules{
					MaxBodySize: 5120,
					Sanitize:    true,
				}
			}

			if r.ContentLength > rules.MaxBodySize {
				pkg.WriteErrorResponse(w, http.StatusRequestEntityTooLarge, 
					fmt.Sprintf("Request body too large (max %d bytes)", rules.MaxBodySize))
				return
			}

			if r.Method == "POST" || r.Method == "PUT" {
				if err := validateJSONBody(r, rules); err != nil {
					pkg.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
					return
				}
			}

			if err := validateQueryParams(r); err != nil {
				pkg.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func validateJSONBody(r *http.Request, rules ValidationRules) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body")
	}
	r.Body.Close()

	if len(body) == 0 && len(rules.RequiredFields) > 0 {
		return fmt.Errorf("request body is required")
	}

	if len(body) == 0 {
		return nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return fmt.Errorf("invalid JSON format")
	}

	for _, field := range rules.RequiredFields {
		if _, exists := data[field]; !exists {
			return fmt.Errorf("missing required field: %s", field)
		}
		if data[field] == nil || data[field] == "" {
			return fmt.Errorf("field '%s' cannot be empty", field)
		}
	}

	for _, field := range rules.EmailFields {
		if value, exists := data[field]; exists && value != nil {
			if str, ok := value.(string); ok {
				if !emailRegex.MatchString(str) {
					return fmt.Errorf("invalid email format for field '%s'", field)
				}
			}
		}
	}

	for _, field := range rules.NameFields {
		if value, exists := data[field]; exists && value != nil {
			if str, ok := value.(string); ok {
				if !nameRegex.MatchString(str) {
					return fmt.Errorf("invalid name format for field '%s' (2-50 chars, letters/spaces/hyphens only)", field)
				}
			}
		}
	}

	for _, field := range rules.UsernameFields {
		if value, exists := data[field]; exists && value != nil {
			if str, ok := value.(string); ok {
				if !usernameRegex.MatchString(str) {
					return fmt.Errorf("invalid username format for field '%s' (3-30 chars, alphanumeric/underscore/hyphen only)", field)
				}
			}
		}
	}

	for _, field := range rules.PasswordFields {
		if value, exists := data[field]; exists && value != nil {
			if str, ok := value.(string); ok {
				if !passwordRegex.MatchString(str) {
					return fmt.Errorf("password must be between 8-128 characters")
				}
			}
		}
	}

	if rules.Sanitize {
		sanitizeMap(data)
	}

	sanitizedBody, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to process request data")
	}

	r.Body = io.NopCloser(bytes.NewBuffer(sanitizedBody))
	r.ContentLength = int64(len(sanitizedBody))

	return nil
}

func validateQueryParams(r *http.Request) error {
	query := r.URL.Query()

	if limit := query.Get("limit"); limit != "" {
		if !regexp.MustCompile(`^\d+$`).MatchString(limit) {
			return fmt.Errorf("invalid limit parameter (must be a number)")
		}
	}

	if offset := query.Get("offset"); offset != "" {
		if !regexp.MustCompile(`^\d+$`).MatchString(offset) {
			return fmt.Errorf("invalid offset parameter (must be a number)")
		}
	}

	if q := query.Get("q"); q != "" {
		if len(q) > 100 {
			return fmt.Errorf("search query too long (max 100 characters)")
		}
		if containsSQLInjectionPatterns(q) {
			return fmt.Errorf("invalid search query")
		}
	}

	return nil
}

func sanitizeMap(data map[string]interface{}) {
	for key, value := range data {
		switch v := value.(type) {
		case string:
			data[key] = sanitizeString(v)
		case map[string]interface{}:
			sanitizeMap(v)
		}
	}
}

func sanitizeString(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\x00", "")
	
	var cleaned strings.Builder
	for _, r := range s {
		if r >= 32 || r == '\n' || r == '\t' {
			cleaned.WriteRune(r)
		}
	}
	
	return cleaned.String()
}

func containsSQLInjectionPatterns(s string) bool {
	lower := strings.ToLower(s)
	patterns := []string{
		"'", "\"", ";", "--", "/*", "*/", "union", "select", "insert", 
		"update", "delete", "drop", "create", "alter", "exec", "execute",
	}
	
	for _, pattern := range patterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		
		next.ServeHTTP(w, r)
	})
} 