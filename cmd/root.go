package cmd

import (
	"fmt"
	"github.com/gabeduke/civo-controller/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"os"
)

var (
	cfgFile string
	app     config.App
)

var rootCmd = &cobra.Command{
	Use:   "civo-controller",
	Short: "Civo cluster controller",
	Long: `The civo control loop will watch a given list of cluster names
and create/delete clusters as the list is updated.

If a cluster is removed from the civo web application the controll loop will
recreate the cluster.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.civo-controller.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".civo-controller" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".civo-controller")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	c, cfgCh := config.LoadConfig()
	app := config.New(c)
	go func() {
		for {
			app.SetConfig(<-cfgCh)
			log.Println("New config loaded")
		}
	}()

	//TODO
	log.SetLevel(log.DebugLevel)
}
