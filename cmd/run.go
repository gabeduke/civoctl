package cmd

import (
	"github.com/gabeduke/civoctl/pkg/civo"
	civoController "github.com/gabeduke/civoctl/pkg/controller"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	metricsListenAddr = ":8081"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the civo control loop",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Beginning Civo control loop")

		c, cfgCh := civo.LoadConfig()
		app := civo.NewCivoCtl(c, viper.GetString("token"))
		go func() {
			for {
				app.SetConfig(<-cfgCh)
				log.Println("NewCivoCtl config loaded")
			}
		}()

		//TODO
		log.SetLevel(log.DebugLevel)

		civoController.Run(app)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
