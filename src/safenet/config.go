package safenet

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func LoadConfigInt(filename, paramName string, defaultValue int) int {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading config file", filename)
		fmt.Println("Using default value for", paramName, "=", defaultValue)
		fmt.Println(err)
		return defaultValue
	}
	config := map[string]interface{}{}
	err = json.Unmarshal(content, &config)
	if err != nil {
		fmt.Println("JSON error reading config", filename)
		fmt.Println("Using default value for", paramName, "=", defaultValue)
		fmt.Println(err)
		return defaultValue
	}
	config = extendConfigWithAllDotJson(config)
	value, exists := config[paramName]
	if !exists {
		fmt.Println("Key", paramName, "not found in", filename)
		fmt.Println("Using default value for", paramName, "=", defaultValue)
		return defaultValue
	}
	fmt.Println("Configured to use", paramName, "=", value)
	valueFloat := value.(float64)
	return int(valueFloat)
}

func extendConfigWithAllDotJson(config map[string]interface{}) map[string]interface{} {
	content, err := ioutil.ReadFile("config_all.json")
	if err != nil {
		return config
	}
	allConfig := map[string]interface{}{}
	err = json.Unmarshal(content, &config)
	if err != nil {
		return config
	}
	for key := range allConfig {
		config[key] = allConfig[key]
	}
	return config
}
