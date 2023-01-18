package utils

import (
	"encoding/json"
	"log"

	"github.com/cbroglie/mustache"
)

func Remarshal(in interface{}, out interface{}) error {
	b, err := json.Marshal(in)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, out)
	return err

}

func MergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

func Template(content string, params map[string]interface{}) (string, error) {
	data, err := mustache.Render(content, params)
	if err != nil {
		return "", err
	}

	return data, nil
}

func JsonPretty(in interface{}) error {
	data, err := json.MarshalIndent(in, "", "  ")
	if err != nil {
		return err
	}

	log.Println(string(data))
	return nil
}
