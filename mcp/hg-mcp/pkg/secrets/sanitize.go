package secrets

import (
	"regexp"
	"strings"
)

// SensitivePatterns contains patterns for identifying sensitive keys.
var SensitivePatterns = []string{
	"password",
	"passwd",
	"secret",
	"token",
	"key",
	"credential",
	"cred",
	"auth",
	"bearer",
	"api_key",
	"apikey",
	"api-key",
	"private",
	"oauth",
	"jwt",
	"session",
	"cookie",
	"cert",
	"certificate",
	"pem",
	"ssh",
	"rsa",
	"dsa",
	"ecdsa",
	"access",
	"refresh",
	"signing",
	"encryption",
	"decrypt",
	"salt",
	"hash",
	"hmac",
	"pin",
	"otp",
	"2fa",
	"mfa",
	"totp",
	"seed",
}

// SensitiveEnvVars contains known sensitive environment variable names.
var SensitiveEnvVars = []string{
	"AWS_ACCESS_KEY_ID",
	"AWS_SECRET_ACCESS_KEY",
	"AWS_SESSION_TOKEN",
	"GITHUB_TOKEN",
	"GH_TOKEN",
	"GITLAB_TOKEN",
	"ANTHROPIC_API_KEY",
	"OPENAI_API_KEY",
	"DISCORD_BOT_TOKEN",
	"DISCORD_TOKEN",
	"SLACK_TOKEN",
	"SLACK_BOT_TOKEN",
	"DATABASE_URL",
	"DB_PASSWORD",
	"POSTGRES_PASSWORD",
	"MYSQL_PASSWORD",
	"REDIS_PASSWORD",
	"MONGO_PASSWORD",
	"NPM_TOKEN",
	"DOCKER_PASSWORD",
	"DOCKER_TOKEN",
	"OP_SERVICE_ACCOUNT_TOKEN",
	"SUPABASE_KEY",
	"SUPABASE_SERVICE_KEY",
	"STRIPE_SECRET_KEY",
	"SENDGRID_API_KEY",
	"TWILIO_AUTH_TOKEN",
	"SSH_PRIVATE_KEY",
	"GPG_PRIVATE_KEY",
	"ENCRYPTION_KEY",
	"JWT_SECRET",
	"SESSION_SECRET",
	"COOKIE_SECRET",
}

// sensitiveRegex caches compiled regex for performance.
var sensitiveRegex *regexp.Regexp

func init() {
	// Build regex pattern from SensitivePatterns
	patterns := make([]string, len(SensitivePatterns))
	for i, p := range SensitivePatterns {
		patterns[i] = regexp.QuoteMeta(p)
	}
	sensitiveRegex = regexp.MustCompile(`(?i)(` + strings.Join(patterns, "|") + `)`)
}

// IsSensitiveKey returns true if the key name suggests sensitive content.
func IsSensitiveKey(key string) bool {
	keyLower := strings.ToLower(key)

	// Check exact matches first
	for _, envVar := range SensitiveEnvVars {
		if strings.EqualFold(key, envVar) {
			return true
		}
	}

	// Check pattern matches
	return sensitiveRegex.MatchString(keyLower)
}

// MaskValue masks a value for safe logging.
func MaskValue(value string) string {
	if len(value) == 0 {
		return ""
	}
	if len(value) <= 4 {
		return "****"
	}
	if len(value) <= 8 {
		return value[:1] + "****" + value[len(value)-1:]
	}
	return value[:2] + "****" + value[len(value)-2:]
}

// Sanitize removes or masks sensitive values from a map.
func Sanitize(params map[string]any) map[string]any {
	result := make(map[string]any, len(params))
	for key, value := range params {
		result[key] = sanitizeValue(key, value)
	}
	return result
}

// sanitizeValue recursively sanitizes a value based on its key.
func sanitizeValue(key string, value any) any {
	if IsSensitiveKey(key) {
		switch v := value.(type) {
		case string:
			return MaskValue(v)
		case []byte:
			return MaskValue(string(v))
		default:
			return "[REDACTED]"
		}
	}

	// Recurse into maps
	if m, ok := value.(map[string]any); ok {
		return Sanitize(m)
	}

	// Recurse into slices
	if slice, ok := value.([]any); ok {
		result := make([]any, len(slice))
		for i, item := range slice {
			if m, ok := item.(map[string]any); ok {
				result[i] = Sanitize(m)
			} else {
				result[i] = item
			}
		}
		return result
	}

	return value
}

// SanitizeString masks sensitive patterns in a string (like URLs with passwords).
func SanitizeString(s string) string {
	// Mask passwords in URLs: postgres://user:password@host
	urlPattern := regexp.MustCompile(`(://[^:]+:)([^@]+)(@)`)
	s = urlPattern.ReplaceAllString(s, "${1}****${3}")

	// Mask Bearer tokens
	bearerPattern := regexp.MustCompile(`(?i)(bearer\s+)(\S+)`)
	s = bearerPattern.ReplaceAllString(s, "${1}****")

	// Mask API keys in query strings: ?api_key=xxx
	apiKeyPattern := regexp.MustCompile(`(?i)(api[_-]?key=)([^&\s]+)`)
	s = apiKeyPattern.ReplaceAllString(s, "${1}****")

	// Mask tokens in query strings: ?token=xxx
	tokenPattern := regexp.MustCompile(`(?i)(token=)([^&\s]+)`)
	s = tokenPattern.ReplaceAllString(s, "${1}****")

	return s
}

// SanitizeHeaders sanitizes HTTP headers, masking Authorization and other sensitive headers.
func SanitizeHeaders(headers map[string]string) map[string]string {
	sensitiveHeaders := []string{
		"authorization",
		"x-api-key",
		"x-auth-token",
		"cookie",
		"set-cookie",
		"x-csrf-token",
		"x-xsrf-token",
	}

	result := make(map[string]string, len(headers))
	for key, value := range headers {
		keyLower := strings.ToLower(key)
		isSensitive := false
		for _, sh := range sensitiveHeaders {
			if keyLower == sh {
				isSensitive = true
				break
			}
		}
		if isSensitive {
			result[key] = MaskValue(value)
		} else {
			result[key] = value
		}
	}
	return result
}

// RedactedString is a string that masks itself when printed.
type RedactedString string

// String returns the masked value.
func (r RedactedString) String() string {
	return MaskValue(string(r))
}

// Value returns the actual value.
func (r RedactedString) Value() string {
	return string(r)
}

// MarshalJSON returns the masked value for JSON encoding.
func (r RedactedString) MarshalJSON() ([]byte, error) {
	masked := MaskValue(string(r))
	return []byte(`"` + masked + `"`), nil
}

// SecureCompare performs a constant-time comparison of two strings.
// This is useful for comparing secrets to prevent timing attacks.
func SecureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}
