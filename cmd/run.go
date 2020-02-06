package cmd

import (
	civoController "github.com/gabeduke/civo-controller/pkg/controller"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
		civoController.Run(app)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
