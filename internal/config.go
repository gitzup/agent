// Config package provides global configuration access for the application.
package internal

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
	"os"
)

const (
	DefaultTimestamps   = false    // Default value for showing timestamps in log output
	DefaultSourceFile   = false    // Default value for showing source code file in log output
	DefaultProject      = "gitzup" // Default name of GCP project to search Pub/Sub subscription in.
	DefaultSubscription = "agents" // Default name of GCP Pub/Sub subscription to search for.
	DefaultWorkspace    = "."      // Default workspace location
	Version             = "0.0.0-alpha"
)

// Configuration type holding all agent configuration flags.
type Configuration struct {
	Timestamps   bool   `mapstructure:"timestamps"`   // Prefix log output with timestamps
	SourceFile   bool   `mapstructure:"source"`       // Prefix log output with source code file
	Project      string `mapstructure:"project"`      // GCP project to pull build requests from (via Pub/Sub)
	Subscription string `mapstructure:"subscription"` // Pub/Sub subscription to pull build requests from
	Workspace    string `mapstructure:"workspace"`    // Workspace directory
}

// Configuration instance.
var Config Configuration

// Prints command-line flags usage and exits with exit code 1.
func Usage(message string) {
	log.Println(message)
	pflag.Usage()
	os.Exit(1)
}

func init() {
	log.SetFlags(0)

	// Setup command-line flags
	pflag.BoolP("timestamps", "t", DefaultTimestamps, "Prefix output with timestamps")
	pflag.BoolP("source", "c", DefaultSourceFile, "Prefix output with source code file")
	pflag.StringP("project", "p", DefaultProject, "GCP project ID")
	pflag.StringP("subscription", "s", DefaultSubscription, "GCP Pub/Sub subscription")
	pflag.StringP("workspace", "w", DefaultWorkspace, "Workspace location")
	pflag.Parse()

	// Setup viper configuration provider
	viper.SetConfigName("agent")
	viper.SetConfigType("yaml")
	viper.SetDefault("timestamps", DefaultTimestamps)
	viper.SetDefault("source", DefaultSourceFile)
	viper.SetDefault("project", DefaultProject)
	viper.SetDefault("subscription", DefaultSubscription)
	viper.SetDefault("workspace", DefaultWorkspace)
	viper.SetEnvPrefix("GZP")
	viper.AutomaticEnv()
	viper.BindEnv()
	viper.BindPFlags(pflag.CommandLine)

	// Unmarshal
	err := viper.Unmarshal(&Config)
	if err != nil {
		log.Fatalf("failed parsing configuration: %s\n", err)
	}

	// Update log flags
	var logFlags = 0
	if Config.Timestamps {
		logFlags = logFlags | log.Ldate | log.Ltime
	}
	if Config.SourceFile {
		logFlags = logFlags | log.Llongfile
	}
	log.SetFlags(logFlags)

	// Print configuration
	log.Println("Configuration:")
	log.Printf("  >> Project:       %s\n", Config.Project)
	log.Printf("  >> Subscription:  %s\n", Config.Subscription)
}
