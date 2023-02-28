package synth

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/manifoldco/promptui"
	"github.com/metaconflux/backend/internal/api/v1alpha"
	"github.com/metaconflux/backend/internal/transformers"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use: "add",
}

var transformerCmd = &cobra.Command{
	Use: "transformer",
	Run: func(cmd *cobra.Command, args []string) {
		manifestFile, err := cmd.Flags().GetString(manifest_flag)
		if err != nil {
			logrus.Fatal(err)
		}
		b, err := ioutil.ReadFile(manifestFile)
		if err != nil {
			logrus.Fatal(err)
		}

		var manifest v1alpha.Manifest
		err = json.Unmarshal(b, &manifest)
		if err != nil {
			logrus.Fatal(err)
		}

		if manifest.Transformers == nil {
			manifest.Transformers = make([]transformers.BaseTransformer, 0)
		}

		kinds := TransformerMananager.GetRegistered()

		promptS := promptui.Select{
			Label: "Transformer Kind",
			Items: kinds,
		}

		i, transformerGVK, err := promptS.Run()
		if err != nil {
			logrus.Fatal(err)
		}

		logrus.Infof("Adding %s", transformerGVK)

		var base transformers.BaseTransformer
		gvk := kinds[i]
		ti, err := TransformerMananager.Get(gvk)
		if err != nil {
			log.Fatal(err)
		}

		base, err = ti.Prompt()
		if err != nil {
			log.Fatal(err)
		}

		manifest.Transformers = append(manifest.Transformers, base)

		b, err = json.MarshalIndent(manifest, "", "  ")
		if err != nil {
			logrus.Fatal(err)
		}

		err = ioutil.WriteFile(manifestFile, b, 0644)
		if err != nil {
			logrus.Fatal(err)
		}

		logrus.Infof("Manifest '%s' updated.", manifestFile)
	},
}

func init() {
	addCmd.AddCommand(transformerCmd)
	rootCmd.AddCommand(addCmd)
}
