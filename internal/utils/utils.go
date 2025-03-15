package utils

import (
	"strings"

	config "github.com/spf13/viper"
)

// UnmarshallConfig is a generic function that unmarshals configuration data from a file into a target object.
// Parameters:
// - configFile string: The path to the configuration file.
// - target *T: A pointer to the target object that will hold the unmarshaled data.
// Return value:
// - *T: A pointer to the target object populated with the unmarshaled data.
// - error: If an error occurs during the unmarshaling process, the corresponding error message is returned.
func UnmarshallConfig[T any](configFile string, target *T) (*T, error) {

	c := config.New()
	c.SetConfigFile(configFile)

	err := c.ReadInConfig() // Find and read the config file
	if err != nil {         // Handle errors reading the config file
		return nil, err
	}

	err = c.Unmarshal(target)
	if err != nil {

		return nil, err
	}
	return target, nil
}

// UnmarshallConfigFromString is a generic function that unmarshals configuration data from a string into a target object.
// Parameters:
// - configContent string: The configuration content as a string.
// - configType string: The type of configuration (e.g., "yaml", "json", "toml"). If empty, the function will try to auto-detect the format.
// - target *T: A pointer to the target object that will hold the unmarshaled data.
// Return value:
// - *T: A pointer to the target object populated with the unmarshaled data.
// - error: If an error occurs during the unmarshaling process, the corresponding error message is returned.
func UnmarshallConfigFromString[T any](configContent string, configType string, target *T) (*T, error) {
	c := config.New()

	// If the user specifies a configuration type, use the specified type
	if configType != "" {
		c.SetConfigType(configType)
	} else {
		// Auto-detect configuration format
		// If the content looks like JSON, set it to JSON format
		if len(configContent) > 0 && configContent[0] == '{' {
			c.SetConfigType("json")
		} else {
			c.SetConfigType("yaml") // Default to YAML format
		}
	}

	err := c.ReadConfig(strings.NewReader(configContent))
	if err != nil {
		return nil, err
	}

	err = c.Unmarshal(target)
	if err != nil {
		return nil, err
	}

	return target, nil
}
