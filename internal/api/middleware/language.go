package middleware

import (
	"strings"

	"kube-tide/internal/utils/i18n"

	"github.com/gin-gonic/gin"
)

const (
	// DefaultLanguage is the default language to use if no language is specified
	DefaultLanguage = "en"
	// LanguageKey is the key used to store language in the context
	LanguageKey = "language"
)

// DetectLanguage middleware detects the user's preferred language
// It checks the Accept-Language header, query parameter and cookies
// and sets the language in the context
func DetectLanguage() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Default language
		language := DefaultLanguage

		// Try to get language from query parameter
		if lang := c.Query("lang"); lang != "" {
			language = lang
		}

		// Try to get language from cookie
		if cookie, err := c.Cookie("lang"); err == nil && cookie != "" {
			language = cookie
		}

		// Try to get language from Accept-Language header
		if acceptLang := c.GetHeader("Accept-Language"); acceptLang != "" {
			// Parse Accept-Language header (e.g. "zh-CN,zh;q=0.9,en;q=0.8")
			langs := strings.Split(acceptLang, ",")
			if len(langs) > 0 {
				// Get the first language
				firstLang := strings.Split(langs[0], ";")[0]
				// Extract the language code (e.g. "zh-CN" -> "zh")
				langCode := strings.Split(firstLang, "-")[0]
				if langCode != "" {
					language = langCode
				}
			}
		}

		// Set language in context
		c.Set(i18n.ContextKey, language)

		c.Next()
	}
}

// GetLanguage returns the language from the context, defaults to "en" if not found
func GetLanguage(c *gin.Context) string {
	lang, exists := c.Get(LanguageKey)
	if !exists {
		return "en"
	}
	return lang.(string)
}

// parseAcceptLanguage parses the Accept-Language header and returns the preferred language
func parseAcceptLanguage(header string) string {
	if header == "" {
		return "en"
	}

	// Split by comma to get ordered list of language preferences
	langs := strings.Split(header, ",")
	if len(langs) == 0 {
		return "en"
	}

	// Extract the first language (highest priority)
	firstLang := langs[0]

	// Check for quality value and strip it
	if idx := strings.Index(firstLang, ";"); idx != -1 {
		firstLang = firstLang[:idx]
	}

	// Normalize and return the language code
	return strings.ToLower(strings.TrimSpace(firstLang))
}
