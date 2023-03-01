package synth

import (
	"log"

	"github.com/metaconflux/backend/internal/chains"
	"github.com/metaconflux/backend/internal/transformers"
	"github.com/metaconflux/backend/internal/transformers/core/v1alpha/container"
	"github.com/metaconflux/backend/internal/transformers/core/v1alpha/contract"
	"github.com/metaconflux/backend/internal/transformers/core/v1alpha/local"
	"github.com/metaconflux/backend/internal/transformers/core/v1alpha/print"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var version = "v0.0.0"

var rootCmd = &cobra.Command{
	Use:     "synth",
	Version: version,
}

const (
	manifest_flag    = "manifest"
	manifest_short   = "m"
	manifest_default = "manifest.json"
)

var TransformerMananager *transformers.Transformers

func Execute() {
	var err error
	TransformerMananager, err = PrepTransformers()
	if err != nil {
		logrus.Fatal(err)
	}

	err = rootCmd.Execute()
	if err != nil {
		log.Fatalln(err)
	}
}

func init() {
	initConfig()
	//rootCmd.PersistentFlags().StringP("config", "c", "config.json", "Path to a config file")
	//rootCmd.PersistentFlags().String("config-env", "", "Name of the environment variable which contains base64 encoded config file")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose logging mode")
	rootCmd.PersistentFlags().StringP(manifest_flag, manifest_short, manifest_default, "Input manifest file")

}

func PrepTransformers() (*transformers.Transformers, error) {
	clients := chains.NewClients(nil)
	defer clients.Close()

	contractT := contract.NewTransformer(clients.Clients())

	pritnT := print.NewTransformer()

	localT := local.NewTransformer()

	containerT, err := container.NewTransformer()
	if err != nil {
		return nil, err
	}

	tm, err := transformers.NewTransformerManager()
	if err != nil {
		return nil, err
	}

	err = tm.Register(print.GVK, pritnT.WithSpec, print.NewSpecFromPrompt)
	if err != nil {
		return nil, err
	}

	err = tm.Register(contract.GVK, contractT.WithSpec, contract.NewSpecFromPrompt)
	if err != nil {
		return nil, err
	}

	err = tm.Register(local.GVK, localT.WithSpec, local.NewSpecFromPrompt)
	if err != nil {
		return nil, err
	}

	err = tm.Register(container.GVK, containerT.WithSpec, nil)
	if err != nil {
		return nil, err
	}

	return tm, nil
}
