package utils

import (
	"encoding/json"
	"log"
	"math/rand"

	"github.com/vpavlin/mustache"
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
	mustache.Experimental = true
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

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
