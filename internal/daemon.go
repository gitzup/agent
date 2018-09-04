package internal

import (
	"cloud.google.com/go/pubsub"
	"golang.org/x/net/context"
	"log"
)

func StartDaemon() {

	// Create context for the daemon
	ctx := context.Background()

	// Validate configuration
	if Config.Project == "" {
		Usage("GCP project is required")
	} else if Config.Subscription == "" {
		Usage("GCP Pub/Sub subscription is required")
	}

	// Create the Pub/Sub client
	client, err := pubsub.NewClient(ctx, Config.Project)
	if err != nil {
		log.Fatalf("Failed to create GCP Pub/Sub client: %v", err)
	}
	defer client.Close()

	// Locate the subscription, fail if missing
	subscription := client.Subscription(Config.Subscription)
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

	request, err := ParseBuildRequest(msg.ID, msg.Data)
	if err != nil {
		panic(err)
	}

	err = request.Apply()
	if err != nil {
		panic(err)
	}
}
