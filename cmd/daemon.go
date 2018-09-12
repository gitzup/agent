package cmd

import (
	"context"
	"log"

	"cloud.google.com/go/pubsub"
	"github.com/gitzup/agent/pkg"
	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start the Gitzup agent daemon.",
	Long:  `This command will start the Gitzup agent daemon, processing build request coming in through the GCP Pub/Sub subscription.`,
	Args:  cobra.ExactArgs(2), // TODO: custom usage
	Run:   func(cmd *cobra.Command, args []string) { startDaemon(args[0], args[1]) },
}

func init() {
	rootCmd.AddCommand(daemonCmd)
}

func startDaemon(gcpProject string, gcpSubscriptionName string) {

	// Create context for the daemon
	ctx := context.Background()

	// Create the Pub/Sub client
	client, err := pubsub.NewClient(ctx, gcpProject)
	if err != nil {
		log.Fatalf("Failed to create GCP Pub/Sub client: %v", err)
	}
	defer client.Close()

	// Locate the subscription, fail if missing
	subscription := client.Subscription(gcpSubscriptionName)
	exists, err := subscription.Exists(ctx)
	if err != nil {
		log.Fatalf("Failed checking if subscription exists: %v", err)
	} else if exists == false {
		log.Fatalln("Subscription could not be found!")
	}

	// Start receiving messages (in separate goroutines)
	log.Printf("Subscribing to: %s", subscription)
	err = subscription.Receive(ctx, func(_ context.Context, msg *pubsub.Message) { handleMessage(msg) })
	if err != nil {
		log.Fatalf("Failed to subscribe to '%s': %v", subscription, err)
	}
}

func handleMessage(msg *pubsub.Message) {
	defer func() {
		err := recover()
		if err != nil {
			// TODO: re-publish this message to the errors topic
			switch t := err.(type) {
			case error:
				log.Printf("Fatal error processing message '%s': %s\n", msg.ID, t.Error())
			default:
				log.Printf("Fatal error processing message: %#v\n", t)
			}
		}
	}()

	msg.Ack()

	request, err := pkg.ParseBuildRequest(msg.ID, msg.Data, workspacePath)
	if err != nil {
		panic(err)
	}

	err = request.Apply()
	if err != nil {
		panic(err)
	}

	// TODO: receive apply result, and send Pub/Sub message with JSON of result
}
