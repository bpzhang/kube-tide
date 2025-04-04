package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"kube-tide/internal/utils/logger"

	"github.com/gin-gonic/gin"
)

// Constants for i18n
const (
	// ContextKey is the key to store/retrieve language in gin context
	ContextKey = "language"

	// DefaultLanguage is the default language to use if no language is found
	DefaultLanguage = "en"

	// Directory where translations are stored
	LocalesDir = "internal/utils/i18n/locales"
)

// I18n represents the internationalization service
type I18n struct {
	translations map[string]map[string]interface{}
	mutex        sync.RWMutex
}

var (
	instance *I18n
	once     sync.Once
)

// GetInstance returns a singleton instance of I18n
func GetInstance() *I18n {
	once.Do(func() {
		instance = &I18n{
			translations: make(map[string]map[string]interface{}),
		}
		if err := instance.loadTranslations(); err != nil {
			logger.Errorf("Failed to load translations: %v", err)
		}
	})
	return instance
}

// loadTranslations loads all translation files for all supported languages
func (i *I18n) loadTranslations() error {
	// Get current working directory
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %v", err)
	}

	// Create the locales directory path
	localesPath := filepath.Join(workDir, LocalesDir)

	// List language directories
	entries, err := os.ReadDir(localesPath)
	if err != nil {
		return fmt.Errorf("failed to read locales directory: %v", err)
	}

	// For each language directory
	for _, entry := range entries {
		if entry.IsDir() {
			lang := entry.Name()

			// Path to translation file
			translationFile := filepath.Join(localesPath, lang, "translation.json")

			// Read and parse the translation file
			data, err := os.ReadFile(translationFile)
			if err != nil {
				logger.Errorf("Failed to read translation file for language %s: %v", lang, err)
				continue
			}

			// Parse JSON
			var translations map[string]interface{}
			if err := json.Unmarshal(data, &translations); err != nil {
				logger.Errorf("Failed to parse translation file for language %s: %v", lang, err)
				continue
			}

			// Store translations
			i.mutex.Lock()
			i.translations[lang] = translations
			i.mutex.Unlock()

			logger.Infof("Loaded translations for language: %s", lang)
		}
	}

	return nil
}

// T translates the given message key according to the language in context
func T(c *gin.Context, key string, args ...interface{}) string {
	// Get language from context
	lang, exists := c.Get(ContextKey)
	if !exists {
		lang = DefaultLanguage
	}

	// Get the translation
	return GetInstance().Translate(lang.(string), key, args...)
}

// Translate retrieves a translated message for the specified language and key
func (i *I18n) Translate(lang string, key string, args ...interface{}) string {
	// Fallback to default language if the requested language is not available
	i.mutex.RLock()
	translations, exists := i.translations[lang]
	if !exists {
		translations = i.translations[DefaultLanguage]
	}
	i.mutex.RUnlock()

	// Get translation by key
	message := i.getTranslationByPath(translations, key)
	if message == "" {
		// If key not found, return the key itself
		return key
	}

	// Format message with arguments
	if len(args) > 0 {
		return i.formatMessage(message, args...)
	}

	return message
}

// getTranslationByPath retrieves a nested translation value using dot notation
// e.g. "common.error" will look for translations["common"]["error"]
func (i *I18n) getTranslationByPath(translations map[string]interface{}, path string) string {
	parts := strings.Split(path, ".")
	current := translations

	// Navigate through the nested structure
	for i := 0; i < len(parts)-1; i++ {
		if next, ok := current[parts[i]].(map[string]interface{}); ok {
			current = next
		} else {
			return ""
		}
	}

	// Get the final value
	lastKey := parts[len(parts)-1]
	if value, ok := current[lastKey]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}

	return ""
}

// formatMessage formats a message with the given arguments
// It replaces placeholders like {0}, {1}, etc. with the corresponding argument
func (i *I18n) formatMessage(message string, args ...interface{}) string {
	result := message
	for i, arg := range args {
		placeholder := fmt.Sprintf("{%d}", i)
		result = strings.Replace(result, placeholder, fmt.Sprintf("%v", arg), -1)
	}
	return result
}

// SupportedLanguages returns a list of supported languages
func (i *I18n) SupportedLanguages() []string {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	languages := make([]string, 0, len(i.translations))
	for lang := range i.translations {
		languages = append(languages, lang)
	}

	return languages
}
