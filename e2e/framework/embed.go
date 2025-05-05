package framework

import (
	"embed"
	"fmt"
	"os"
	"text/template"
)

//go:embed templates
var embeddedTemplates embed.FS

func templateGethGenesis(outputFile string, premineAllocations map[string]string) error {
	keys := make([]string, 0, len(premineAllocations))
	for k := range premineAllocations {
		keys = append(keys, k)
	}

	return templateFile("geth-genesis.json", outputFile, struct {
		Allocs map[string]string
		Keys   []string
	}{Allocs: premineAllocations, Keys: keys})
}

func templateFile(templateName string, outputFile string, data any) error {
	rawTemplate, err := embeddedTemplates.ReadFile(fmt.Sprintf("templates/%s.tmpl", templateName))
	if err != nil {
		return fmt.Errorf("failed to read %s template: %w", templateName, err)
	}

	renderedTemplate, err := template.New(templateName).Parse(string(rawTemplate))
	if err != nil {
		return fmt.Errorf("failed to parse %s emplate: %w", templateName, err)
	}

	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to open %s for writing: %w", outputFile, err)
	}

	err = renderedTemplate.Execute(file, data)
	if err != nil {
		return fmt.Errorf("failed to write %s to file: %w", templateName, err)
	}

	return nil
}
