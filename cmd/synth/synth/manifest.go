package synth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/metaconflux/backend/internal/api/v1alpha"
	"github.com/metaconflux/backend/internal/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var manifestCmd = &cobra.Command{
	Use: "manifest",
}

var pullCmd = &cobra.Command{
	Use: "pull [chainId] [address]",
	Run: func(cmd *cobra.Command, args []string) {
		manifestFile, err := cmd.Flags().GetString(manifest_flag)
		if err != nil {
			logrus.Fatal(err)
		}

		token := viper.GetString("token")
		if token == "" {
			logrus.Fatal("Auth token missing in config")
		}
		gateway := viper.Get("gateway")
		if gateway == nil {
			logrus.Fatalf("Failed to load gateway from config")
		}

		if len(args) < 2 {
			cmd.Help()
			logrus.Fatalf("Missing arguments")
		}

		body, _, err := get(http.MethodGet, gateway.(string), token, args[0], args[1])
		if err != nil {
			log.Fatal(err)
		}

		data, err := ioutil.ReadAll(body)
		if err != nil {
			logrus.Fatal(err)
		}

		var manifest v1alpha.Manifest

		err = json.Unmarshal(data, &manifest)
		if err != nil {
			logrus.Fatal(err)
		}
		utils.JsonPretty(manifest)

		err = ioutil.WriteFile(manifestFile, data, 0644)
		if err != nil {
			logrus.Fatal(err)
		}

	},
}

var pushCmd = &cobra.Command{
	Use: "push",
	Run: func(cmd *cobra.Command, args []string) {
		manifestFile, err := cmd.Flags().GetString(manifest_flag)
		if err != nil {
			logrus.Fatal(err)
		}

		token := viper.GetString("token")
		if token == "" {
			logrus.Fatal("Auth token missing in config")
		}
		gateway := viper.Get("gateway")
		if gateway == nil {
			logrus.Fatalf("Failed to load gateway from config")
		}

		data, err := ioutil.ReadFile(manifestFile)
		if err != nil {
			logrus.Fatal(err)
		}

		var manifest v1alpha.Manifest

		err = json.Unmarshal(data, &manifest)
		if err != nil {
			logrus.Fatal(err)
		}

		method := http.MethodPost
		url := fmt.Sprintf("%s/api/v1alpha/manifest/", gateway.(string))

		_, statusCode, err := get(http.MethodGet, gateway.(string), token, fmt.Sprintf("%d", manifest.ChainID), manifest.Contract)
		if err != nil {
			log.Fatal(err)
		}

		if statusCode == http.StatusOK {
			method = http.MethodPut
			url = fmt.Sprintf("%s%d/%s/", url, manifest.ChainID, manifest.Contract)
		}

		logrus.Infof("http.MethodPost %s", url)
		req, err := http.NewRequest(
			method,
			url,
			bytes.NewBuffer(data),
		)
		if err != nil {
			logrus.Fatal(err)
		}

		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
		req.Header.Add("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			logrus.Fatal(err)
		}

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			var apiErr utils.ApiError
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logrus.Error(err)
			}

			err = json.Unmarshal(body, &apiErr)
			if err != nil {
				logrus.Error(err)
			}

			logrus.Infof("%s", apiErr)

			logrus.Fatalf("Failed with status %s", apiErr.Error)
		}

		created, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logrus.Fatal(err)
		}

		var respData v1alpha.MetadataResult
		err = json.Unmarshal(created, &respData)
		if err != nil {
			logrus.Fatal(err)
		}

		action := "Created"
		if method == http.MethodPut {
			action = "Updated"
		}
		fmt.Printf("%s: \n %s%s\n\n", action, gateway, respData.Url)

		fmt.Printf("Metadata URL: %s/api/v1alpha/metadata/%d/%s/:tokenId", gateway.(string), manifest.ChainID, manifest.Contract)

	},
}

func init() {
	manifestCmd.AddCommand(pullCmd)
	manifestCmd.AddCommand(pushCmd)

	rootCmd.AddCommand(manifestCmd)
}

func get(method string, gateway string, token string, chainId string, contract string) (io.ReadCloser, int, error) {
	url := fmt.Sprintf("%s/api/v1alpha/manifest/%s/%s/", gateway, chainId, contract)
	logrus.Infof("GET %s", url)
	req, err := http.NewRequest(
		method,
		url,
		nil,
	)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("Failed with status %s", resp.Status)
	}

	return resp.Body, resp.StatusCode, nil
}
