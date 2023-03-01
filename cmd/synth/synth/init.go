package synth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/manifoldco/promptui"
	"github.com/metaconflux/backend/internal/api/v1alpha"
	"github.com/metaconflux/backend/internal/chains"
	"github.com/metaconflux/backend/internal/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	chainid_default       = "80001"
	version_flag          = "version"
	version_default       = "v1alpha"
	addr_flag             = "address"
	manifest_file_default = "manifest.json"
	manifest_file_flag    = "output"
)

var initCmd = &cobra.Command{
	Use: "init",
	Run: func(cmd *cobra.Command, args []string) {
		manifestFile, err := cmd.Flags().GetString(manifest_flag)
		if err != nil {
			logrus.Fatal(err)
		}

		promptS := promptui.Select{
			Label: "Version",
			Items: []string{version_default},
			//Default: version_default,
		}

		_, version, err := promptS.Run()
		if err != nil {
			log.Fatal(err)
		}

		promptS = chains.PromptSelect

		i, _, err := promptS.Run()
		if err != nil {
			log.Fatal(err)
		}

		chainId := chains.GetChainId(i)

		promptS = promptui.Select{
			Label: "Refresh After",
			Items: []string{"24h", "16h", "8h", "1h", "30m", "10m", "1m"},
			//Default: version_default,
		}

		_, refreshAfter, err := promptS.Run()
		if err != nil {
			log.Fatal(err)
		}

		refreshAfterD, err := time.ParseDuration(refreshAfter)
		if err != nil {
			log.Fatal(err)
		}

		promptP := promptui.Prompt{
			Label:   "Contract",
			Default: utils.ZERO_ADDR,
			Validate: func(s string) error {
				if !common.IsHexAddress(s) {
					return fmt.Errorf("Invalid address")
				}

				return nil
			},
		}

		addr, err := promptP.Run()
		if err != nil {
			log.Fatal(err)
		}

		m := v1alpha.Manifest{
			Version:  version,
			ChainID:  chainId,
			Contract: addr,
			Config: v1alpha.Config{
				Freeze:       false,
				RefreshAfter: utils.Duration(refreshAfterD),
			},
		}

		b, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		err = ioutil.WriteFile(manifestFile, b, 0644)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Generated manifest: \n\n%s", string(b))
	},
}

func init() {
	//initCmd.Flags().String(version_flag, "v1alpha", "Manifest version")
	//initCmd.Flags().String(addr_flag, zero_addr, "Contract address")
	//initCmd.Flags().StringP(manifest_file_flag, "o", manifest_file_default, "Name of the manifest file")

	rootCmd.AddCommand(initCmd)
}
