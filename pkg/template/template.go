package template

import (
	"errors"
	"os"
	"path"
	"strings"
)

var (
	ErrTemplateNotFound    = errors.New("template not found")
	ErrInvalidTemplateFile = errors.New("invalid template file")
)

// NormalizeLocale defensively tries to normalize the locale
//
// it does not necessarily return a valid locale. in this case the default
// locale will be used anyways
func NormalizeLocale(l string) string {
	if l == "" {
		return "_"
	}
	l = strings.Replace(l, "-", "_", -1)
	parts := strings.Split(l, "_")
	if len(parts) == 2 {
		parts[0] = strings.ToLower(parts[0])
		parts[1] = strings.ToUpper(parts[1])
		return strings.Join(parts, "_")
	}
	return l
}

// TemplateFileName returns the template file name to use for the given parameters
//
// It looks in the tmplDir directory like so:
//
//   tmplDir/locale/baseName
//
// If the file does not exist, it will try
//
//   tmplDir/defaultLocale/baseName
//
// If the default does not exist, it will fail.
func TemplateFileName(tmplDir, locale, defaultLocale, baseName string) (string, error) {
	tmplFile := path.Join(tmplDir, NormalizeLocale(locale), baseName)
	inf, err := os.Stat(tmplFile)
	if err != nil && os.IsNotExist(err) {
		tmplFile = path.Join(tmplDir, defaultLocale, baseName)
		inf, err = os.Stat(tmplFile)
	}
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrTemplateNotFound
		}
		return "", err
	}
	if inf.IsDir() {
		return "", ErrInvalidTemplateFile
	}
	return tmplFile, nil
}
