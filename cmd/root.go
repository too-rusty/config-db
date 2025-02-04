package cmd

import (
	"fmt"
	"os"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/confighub/db"
	"github.com/flanksource/confighub/utils/kube"
	"github.com/flanksource/kommons"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var dev bool
var httpPort, metricsPort, devGuiPort int
var disableKubernetes bool
var kommonsClient *kommons.Client
var publicEndpoint = "http://localhost:8080"
var defaultSchedule string
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func readFromEnv(v string) string {
	val := os.Getenv(v)
	if val != "" {
		return val
	}
	return v
}

// Root ...
var Root = &cobra.Command{
	Use: "confighub",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		count, _ := cmd.Flags().GetCount("loglevel")
		// logger.StandardLogger().(logsrusapi.Logger).Out = os.Stderr
		logger.StandardLogger().SetLogLevel(count)
		logger.UseZap(cmd.Flags())
		var err error

		if kommonsClient, err = kube.NewKommonsClient(); err != nil {
			logger.Errorf("failed to get kubernetes client: %v", err)
		}

		db.ConnectionString = readFromEnv(db.ConnectionString)
		if db.ConnectionString == "DB_URL" {
			db.ConnectionString = ""
		}
		db.Schema = readFromEnv(db.Schema)
		db.LogLevel = readFromEnv(db.LogLevel)

	},
}

// ServerFlags ...
func ServerFlags(flags *pflag.FlagSet) {
	flags.IntVar(&httpPort, "httpPort", 8080, "Port to expose a health dashboard ")
	flags.IntVar(&devGuiPort, "devGuiPort", 3004, "Port used by a local npm server in development mode")
	flags.IntVar(&metricsPort, "metricsPort", 8081, "Port to expose a health dashboard ")
	flags.BoolVar(&disableKubernetes, "disable-kubernetes", false, "Disable all functionality that requires a kubernetes connection")
	flags.BoolVar(&dev, "dev", false, "Run in development mode")
	flags.StringVar(&defaultSchedule, "default-schedule", "@every 60m", "Default schedule for configs that don't specfiy one")
	flags.StringVar(&publicEndpoint, "public-endpoint", "http://localhost:8080", "Public endpoint that this instance is exposed under")
}

func init() {
	logger.BindFlags(Root.PersistentFlags())

	if len(commit) > 8 {
		version = fmt.Sprintf("%v, commit %v, built at %v", version, commit[0:8], date)
	}
	Root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version of confighub",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
		},
	})

	db.Flags(Root.PersistentFlags())

	Root.AddCommand(Run, Analyze, Serve, GoOffline)
}
