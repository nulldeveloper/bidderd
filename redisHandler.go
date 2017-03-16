package main

import (
	"log"

	redis "gopkg.in/redis.v5"
)

const channel = "Test"

var redisClient *redis.Client
var subscriber *redis.PubSub

func setupClient() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	log.Println("Setting up client")
	subscribe()
}

func subscribe() {
	log.Println("Starting subscriber")

	var err error
	subscriber, err = redisClient.Subscribe(channel)

	if err != nil {
		panic(err)
	}
}

func startRedisSubscriber() {
	log.Println("Waiting for Configuration Changes")
	for {
		msg, err := subscriber.ReceiveMessage()

		if err != nil {
			log.Println("error recieving message")
		} else {
			updateConfiguration(msg)
		}
	}
}

func stopRedisSubscriber() {
	subscriber.Close()
	subscriber.Unsubscribe(channel)
	redisClient.Close()
}

func updateConfiguration(msg *redis.Message) {
	log.Println("the message is", msg.Payload)

}
