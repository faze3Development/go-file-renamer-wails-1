package namer

import (
	"bytes"
	"crypto/rand"
	"fmt"
	customErrors "go-file-renamer-wails/file_server/errors"
	"math/big"
	"strings"
	"text/template"
	"time"
)

// copyNamer keeps the original filename.
type copyNamer struct{}

func NewCopyNamer() Namer {
	return &copyNamer{}
}

func (n *copyNamer) Info() Info {
	return Info{ID: "copy", Name: "Copy Original", Description: "Keeps the original filename."}
}
func (n *copyNamer) GenerateName(originalBaseName string, _ uint64) (string, error) {
	return originalBaseName, nil
}

func NewRandomNamer(length int) Namer {
	return &randomNamer{length: length}
}

func (n *randomNamer) Info() Info {
	return Info{ID: "random", Name: "Random", Description: "Generates a random alphanumeric name."}
}
func (n *randomNamer) GenerateName(_ string, _ uint64) (string, error) {
	return RandomString(n.length)
}

func NewDateTimeNamer(format string) Namer {
	return &dateTimeNamer{format: format}
}

func (n *dateTimeNamer) Info() Info {
	return Info{ID: "datetime", Name: "Date/Time", Description: "Names files using the current date and time."}
}
func (n *dateTimeNamer) GenerateName(_ string, _ uint64) (string, error) {
	return time.Now().Format(n.format), nil
}

func NewTemplateNamer(templateString string) Namer {
	return &templateNamer{template: templateString}
}

func (n *templateNamer) Info() Info {
	return Info{ID: "template", Name: "Template", Description: "Names files based on a custom template."}
}
func (n *templateNamer) GenerateName(originalBaseName string, counter uint64) (string, error) {
	tmpl, err := template.New("filename").Parse(n.template)
	if err != nil {
		return "", customErrors.NewValidationError(fmt.Sprintf("Invalid template: %v", err))
	}

	data := struct {
		Original string
		Counter  uint64
	}{
		Original: originalBaseName,
		Counter:  counter,
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	return buf.String(), err
}

func RandomString(length int) (string, error) {
	defaultLength := 12
	if length <= 0 {
		length = defaultLength
	}

	var sb strings.Builder
	sb.Grow(length)
	limit := big.NewInt(int64(len(letters)))
	for i := 0; i < length; i++ {
		idx, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", customErrors.NewProcessingError("Failed to generate random number", err)
		}
		sb.WriteByte(letters[idx.Int64()])
	}
	return sb.String(), nil
}
