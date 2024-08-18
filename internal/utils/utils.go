package utils

import (
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
