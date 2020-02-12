package cmd

import (
	"fmt"
	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

const app = "civoctl"

var (
	cfgFile string
	VERSION string
)

var rootCmd = &cobra.Command{
	Use:   app,
	Short: "Civo cluster controller",
	Long: `The civo control loop will watch a given list of cluster names
and create/delete clusters as the list is updated.

If a cluster is removed from the civo web application the controll loop will
recreate the cluster.`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Hugo",
	Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(VERSION)
	},
}

func Execute(version string) {
	VERSION = version
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	flags := rootCmd.PersistentFlags()

	flags.StringVar(&cfgFile, "config", "", fmt.Sprintf("config file (default is $HOME/.%s.yaml)", app))
	flags.String("token", "", "Civo API Token (env variable: CIVO_API_KEY)")
	flags.BoolP("dangerous", "d", false, "Dangerous mode will delete clusters not in the config file")
	flags.StringP("log-level", "l", "info", "Log level (error|INFO|debug|trace)")

	viper.BindPFlags(flags)

	viper.SetEnvPrefix("civo")
	viper.BindEnv("token")
}

func initLogger() {
	level := viper.GetString("log-level")
	log.Infof("Log level: %s", level)

	switch level {
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
		viper.Debug()
	case "trace":
		log.SetLevel(log.TraceLevel)
		viper.Debug()
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
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

		// Search config in home directory with name ".civoctl" (without extension).
		viper.AddConfigPath(fmt.Sprintf("/etc/%s/", app))
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")

		viper.SetConfigType("yaml") // REQUIRED if the config file does not have the extension in the name
		viper.SetConfigName("." + app)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

}
