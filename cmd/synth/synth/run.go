package synth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/metaconflux/backend/internal/api/v1alpha"
	"github.com/metaconflux/backend/internal/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const ()

var runCmd = &cobra.Command{
	Use:   "run [tokenId]",
	Short: "Execute a manifest locally",
	Run: func(cmd *cobra.Command, args []string) {
		manifest, err := cmd.Flags().GetString(manifest_flag)
		if err != nil {
			logrus.Fatal(err)
		}
		logrus.Infof("Executing manifest '%s' and tokenId %s", manifest, args[0])

		b, err := ioutil.ReadFile(manifest)
		if err != nil {
			logrus.Fatal(err)
		}

		var m v1alpha.Manifest
		err = json.Unmarshal(b, &m)
		if err != nil {
			logrus.Fatal(err)
		}

		params := map[string]interface{}{
			"id":       args[0],
			"contract": m.Contract,
		}

		if TransformerMananager == nil {
			logrus.Fatal(fmt.Errorf("Failed to load Transformer Manager"))
		}
		result, err := TransformerMananager.Execute(m.Transformers, params)
		if err != nil {
			logrus.Fatal(err)
		}

		utils.JsonPretty(result)
	},
}

func init() {

	rootCmd.AddCommand(runCmd)
}
