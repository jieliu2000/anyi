package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestConfig struct {
	Name    string   `mapstructure:"name" json:"name" yaml:"name"`
	Version string   `mapstructure:"version" json:"version" yaml:"version"`
	Tags    []string `mapstructure:"tags" json:"tags" yaml:"tags"`
}

func TestUnmarshallConfig(t *testing.T) {
	// Create a temporary test config file
	yamlContent := `
name: test-config
version: 1.0.0
tags:
  - tag1
  - tag2
  - tag3
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(yamlContent)
	assert.NoError(t, err)
	err = tmpFile.Close()
	assert.NoError(t, err)

	t.Run("Success: Parse valid config file", func(t *testing.T) {
		config := &TestConfig{}
		result, err := UnmarshallConfig(tmpFile.Name(), config)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test-config", result.Name)
		assert.Equal(t, "1.0.0", result.Version)
		assert.Equal(t, []string{"tag1", "tag2", "tag3"}, result.Tags)
	})

	t.Run("Failure: Config file does not exist", func(t *testing.T) {
		config := &TestConfig{}
		result, err := UnmarshallConfig("non-existent-file.yaml", config)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	// Create an invalid config file
	invalidContent := `
name: invalid
version: 1.0.0
tags: [
  - this is invalid yaml
`
	invalidFile, err := os.CreateTemp("", "invalid-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(invalidFile.Name())

	_, err = invalidFile.WriteString(invalidContent)
	assert.NoError(t, err)
	err = invalidFile.Close()
	assert.NoError(t, err)

	t.Run("Failure: Invalid config file format", func(t *testing.T) {
		config := &TestConfig{}
		result, err := UnmarshallConfig(invalidFile.Name(), config)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestUnmarshallConfigFromString(t *testing.T) {
	t.Run("Success: Parse YAML string with auto-detection", func(t *testing.T) {
		yamlContent := `
name: yaml-config
version: 2.0.0
tags:
  - yaml1
  - yaml2
`
		config := &TestConfig{}
		result, err := UnmarshallConfigFromString(yamlContent, "", config)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "yaml-config", result.Name)
		assert.Equal(t, "2.0.0", result.Version)
		assert.Equal(t, []string{"yaml1", "yaml2"}, result.Tags)
	})

	t.Run("Success: Parse JSON string with auto-detection", func(t *testing.T) {
		jsonContent := `{
  "name": "json-config",
  "version": "3.0.0",
  "tags": ["json1", "json2", "json3"]
}`
		config := &TestConfig{}
		result, err := UnmarshallConfigFromString(jsonContent, "", config)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "json-config", result.Name)
		assert.Equal(t, "3.0.0", result.Version)
		assert.Equal(t, []string{"json1", "json2", "json3"}, result.Tags)
	})

	t.Run("Success: Parse YAML string with specified type", func(t *testing.T) {
		yamlContent := `
name: yaml-specified
version: 4.0.0
tags:
  - spec1
  - spec2
`
		config := &TestConfig{}
		result, err := UnmarshallConfigFromString(yamlContent, "yaml", config)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "yaml-specified", result.Name)
		assert.Equal(t, "4.0.0", result.Version)
		assert.Equal(t, []string{"spec1", "spec2"}, result.Tags)
	})

	t.Run("Success: Parse JSON string with specified type", func(t *testing.T) {
		jsonContent := `{
  "name": "json-specified",
  "version": "5.0.0",
  "tags": ["json-spec1", "json-spec2"]
}`
		config := &TestConfig{}
		result, err := UnmarshallConfigFromString(jsonContent, "json", config)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "json-specified", result.Name)
		assert.Equal(t, "5.0.0", result.Version)
		assert.Equal(t, []string{"json-spec1", "json-spec2"}, result.Tags)
	})

	t.Run("Failure: Parse invalid YAML string", func(t *testing.T) {
		invalidYaml := `
name: invalid
version: broken
tags: [
  this is invalid yaml
`
		config := &TestConfig{}
		result, err := UnmarshallConfigFromString(invalidYaml, "", config)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("Failure: Parse invalid JSON string", func(t *testing.T) {
		invalidJson := `{
  "name": "invalid",
  "version": "broken",
  "tags": ["tag1", "tag2",
}`
		config := &TestConfig{}
		result, err := UnmarshallConfigFromString(invalidJson, "json", config)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("Failure: Wrong format type specified", func(t *testing.T) {
		// This is valid JSON but we're telling it to parse as YAML
		jsonContent := `{
  "name": "wrong-type",
  "version": "6.0.0",
  "tags": ["tag1", "tag2"]
}`
		config := &TestConfig{}
		// Specifying YAML type for JSON content may not fail in all cases due to YAML being a superset of JSON,
		// so we use TOML which has a very different format
		result, err := UnmarshallConfigFromString(jsonContent, "toml", config)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
