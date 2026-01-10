package middleware

import (
	"html"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const (
	DefaultMaxBodySize = 10 * 1024 * 1024

	headerContentTypeOptions = "X-Content-Type-Options"
	headerFrameOptions       = "X-Frame-Options"
	headerXSSProtection      = "X-XSS-Protection"
	headerHSTS               = "Strict-Transport-Security"
	headerCSP                = "Content-Security-Policy"
	headerDownloadOptions    = "X-Download-Options"
	headerReferrerPolicy     = "Referrer-Policy"
	headerPermissionsPolicy  = "Permissions-Policy"
)

var (
	sqlPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(union\s+select)`),
		regexp.MustCompile(`(?i)(insert\s+into)`),
		regexp.MustCompile(`(?i)(delete\s+from)`),
		regexp.MustCompile(`(?i)(drop\s+table)`),
		regexp.MustCompile(`(?i)(drop\s+database)`),
		regexp.MustCompile(`(?i)(update\s+.+set)`),
		regexp.MustCompile(`(?i)(exec\s*\()`),
		regexp.MustCompile(`(?i)(execute\s+immediate)`),
		regexp.MustCompile(`(?i)(--)`),
		regexp.MustCompile(`(?i)(\/\*)`),
		regexp.MustCompile(`(?i)(xp_cmdshell)`),
		regexp.MustCompile(`(?i)(;.*--)`),
		regexp.MustCompile(`(?i)('.*or.*'.*=.*')`),
	}

	xssPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`),
		regexp.MustCompile(`(?i)<iframe[^>]*>.*?</iframe>`),
		regexp.MustCompile(`(?i)<object[^>]*>.*?</object>`),
		regexp.MustCompile(`(?i)<embed[^>]*>`),
		regexp.MustCompile(`(?i)on\w+\s*=`),
		regexp.MustCompile(`(?i)javascript:`),
		regexp.MustCompile(`(?i)<img[^>]*src\s*=\s*["']?javascript:`),
	}

	pathTraversalPattern = regexp.MustCompile(`\.\.\/|\.\.\\`)

	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,20}$`)
)

type SecurityMiddleware struct {
	maxBodySize       int64
	allowedMediaTypes []string
}

func NewSecurityMiddleware() *SecurityMiddleware {
	return &SecurityMiddleware{
		maxBodySize: DefaultMaxBodySize,
		allowedMediaTypes: []string{
			"application/json",
			"multipart/form-data",
			"application/x-www-form-urlencoded",
		},
	}
}

func (m *SecurityMiddleware) SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentTypeOptions, "nosniff")
		c.Set(headerFrameOptions, "DENY")
		c.Set(headerXSSProtection, "1; mode=block")
		c.Set(headerHSTS, "max-age=31536000; includeSubDomains")
		c.Set(headerCSP, "default-src 'self'")
		c.Set(headerDownloadOptions, "noopen")
		c.Set(headerReferrerPolicy, "strict-origin-when-cross-origin")
		c.Set(headerPermissionsPolicy, "geolocation=(), microphone=(), camera=()")

		return c.Next()
	}
}

func (m *SecurityMiddleware) ValidateContentType() fiber.Handler {
	return func(c *fiber.Ctx) error {
		method := c.Method()
		if method == fiber.MethodGet || method == fiber.MethodDelete || method == fiber.MethodOptions {
			return c.Next()
		}

		contentType := string(c.Request().Header.ContentType())

		for _, allowed := range m.allowedMediaTypes {
			if strings.HasPrefix(contentType, allowed) {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusUnsupportedMediaType).JSON(fiber.Map{
			"error":   "Unsupported Content-Type",
			"allowed": m.allowedMediaTypes,
		})
	}
}

func (m *SecurityMiddleware) SQLInjectionProtection() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var detected bool
		c.Request().URI().QueryArgs().VisitAll(func(key, value []byte) {
			if detected {
				return
			}
			if containsSQLPattern(string(value)) {
				detected = true
			}
		})

		if detected {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Potential SQL injection detected",
			})
		}

		if c.Method() == fiber.MethodPost || c.Method() == fiber.MethodPut {
			if containsSQLPattern(string(c.Body())) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Potential SQL injection detected in request body",
				})
			}
		}

		return c.Next()
	}
}

func (m *SecurityMiddleware) XSSProtection() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var detected bool

		c.Request().URI().QueryArgs().VisitAll(func(key, value []byte) {
			if detected {
				return
			}
			if containsXSSPattern(string(value)) {
				detected = true
			}
		})

		if detected {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Potential XSS attack detected",
			})
		}

		return c.Next()
	}
}

func (m *SecurityMiddleware) PathTraversalProtection() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if pathTraversalPattern.MatchString(c.Path()) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Path traversal detected",
			})
		}

		return c.Next()
	}
}

func containsSQLPattern(input string) bool {
	for _, pattern := range sqlPatterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

func containsXSSPattern(input string) bool {
	for _, pattern := range xssPatterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

func SanitizeString(input string) string {

	sanitized := html.EscapeString(input)

	sanitized = strings.ReplaceAll(sanitized, "\x00", "")

	return strings.TrimSpace(sanitized)
}

func SanitizeEmail(email string) string {
	email = strings.ToLower(email)
	return strings.TrimSpace(email)
}

func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func IsValidUsername(username string) bool {
	return usernameRegex.MatchString(username)
}
