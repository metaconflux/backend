package chains

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
)

var PromptSelect = promptui.Select{
	Label: "ChainId",
	Items: Chains,
	Templates: &promptui.SelectTemplates{
		Label:    "{{ .}}?",
		Active:   "{{ .Name  | underline}} ({{ .ChainId | underline }})",
		Selected: fmt.Sprintf(`{{ "%s" | green}} {{ .Name | faint }} ({{ .ChainId | faint }})`, promptui.IconGood),
	},
	//Default: version_default,
}

func GetChainId(idx int) int64 {
	return Chains[idx].ChainId
}

func GetAbi(idx int, address string, key string) ([]map[string]interface{}, error) {
	chain := Chains[idx]

	if !strings.HasSuffix(chain.ScanAPI, "/") {
		chain.ScanAPI = fmt.Sprintf("%s/", chain.ScanAPI)
	}
	url := fmt.Sprintf("%sapi?module=contract&action=getabi&address=%s&key=%s", chain.ScanAPI, address, key)

	logrus.Infof(url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to get ABI with status: %s", resp.Status)
	}

	var result map[string]interface{}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &result)
	if err != nil {
		return nil, err
	}

	if result["status"] != "1" {
		return nil, fmt.Errorf("Failed to get ABI: %s", result["message"])
	}

	var abi []map[string]interface{}
	resultS := result["result"].(string)
	err = json.Unmarshal([]byte(resultS), &abi)
	if err != nil {
		return nil, err
	}

	var funcList []map[string]interface{}

	for _, item := range abi {
		if typ, ok := item["type"]; ok && typ == "function" {
			funcList = append(funcList, item)
		}
	}

	return funcList, nil
}
