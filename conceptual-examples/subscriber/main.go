package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// messageHandler is called whenever a new message arrives on a subscribed topic.
var messageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	// In a real app, we would push this to a channel for processing.
	// For now, we just print the raw payload.
	fmt.Printf("[MQTT] Received on topic: %s\nPayload: %s\n", msg.Topic(), msg.Payload())
}

func main() {
	// 1. Configure the Client
	opts := mqtt.NewClientOptions()
	opts.AddBroker("tcp://localhost:1883") // Point to your Broker (e.g., Mosquitto)
	opts.SetClientID("go-gateway-listener")
	opts.SetKeepAlive(60 * time.Second)
	// Set the default handler for messages
	opts.SetDefaultPublishHandler(messageHandler)

	// 2. Connect
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect to broker: %v", token.Error())
	}
	fmt.Println("Connected to MQTT Broker")

	// 3. Subscribe to the "Wildcard" Topic
	// The '#' character is a wildcard. We are listening to EVERYTHING under "factory".
	topic := "factory/#"
	if token := client.Subscribe(topic, 0, nil); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to subscribe: %v", token.Error())
	}
	fmt.Printf("Subscribed to %s\n", topic)

	// 4. Block until interupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	// Graceful Disconnect
	client.Disconnect(250)
	fmt.Println("Disconnected.")
}
