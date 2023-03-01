package synth

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/metaconflux/backend/internal/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var filename = "synth.yaml"
var synthConfigDir = ""
var configPath = filename

var configCmd = &cobra.Command{
	Use: "config option=value [option2=value2]",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			var configMap map[string]interface{}
			err := viper.Unmarshal(&configMap)
			if err != nil {
				log.Fatal(err)
			}
			err = utils.JsonPretty(configMap)
			if err != nil {
				log.Fatal(err)
			}

			return
		}
		for _, arg := range args {
			split := strings.Split(arg, "=")
			key := split[0]
			val := split[1]

			configKey := viper.Get(key)

			if configKey == nil {
				log.Fatalf("Unknown key %s", key)
			}

			logrus.Infof("%s", reflect.TypeOf(configKey).Kind())

			var configVal interface{}
			var err error

			switch reflect.TypeOf(configKey).Kind() {
			case reflect.Bool:
				configVal, err = strconv.ParseBool(val)
			case reflect.String:
				configVal = val
			case reflect.Int:
				configVal, err = strconv.Atoi(val)
			}

			if err != nil {
				log.Fatal(err)
			}

			logrus.Infof("Val: %s", configVal)

			viper.Set(key, configVal)
		}

		viper.WriteConfig()
	},
}

func init() {

	initConfig()
	rootCmd.AddCommand(configCmd)
}

type Config struct {
	Token      string
	ShareUsage bool
	Gateway    string
}

func initConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	synthConfigDir = fmt.Sprintf("%s/.config/synth", home)

	configPath = filepath.Join(synthConfigDir, filename)

	viper.SetConfigFile(configPath)
	if _, err := os.Stat(configPath); err != nil {
		logrus.Infof("Creating default config in %s", configPath)
		viper.SetDefault("token", "")
		viper.SetDefault("shareUsage", true)
		viper.SetDefault("gateway", "http://localhost:8081")

		err = os.Mkdir(synthConfigDir, 0700)
		if err != nil && !os.IsExist(err) {
			log.Fatal(err)
		}
		err = viper.WriteConfig()
		if err != nil {
			log.Fatal(err)
		}
	}

	viper.ReadInConfig()
}
