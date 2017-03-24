package main

import (
	"log"

	redis "gopkg.in/redis.v5"
)

type redisHandler struct {
	client     *redis.Client
	subscriber *redis.PubSub
	channel    string
}

func newRedis() *redisHandler {
	rh := &redisHandler{}
	rh.client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	rh.channel = "Test"
	return rh
}

func (rh *redisHandler) subscribe() {
	log.Println("Starting subscriber")

	var err error
	rh.subscriber, err = rh.client.Subscribe(rh.channel)

	if err != nil {
		panic(err)
	}
}

func (rh *redisHandler) startRedisSubscriber() {
	log.Println("Waiting for Configuration Changes")
	for {
		msg, err := rh.subscriber.ReceiveMessage()

		if err != nil {
			log.Println("error recieving message")
		} else {
			rh.updateConfiguration(msg)
		}
	}
}

func (rh *redisHandler) stopRedisSubscriber() {
	rh.subscriber.Close()
	rh.subscriber.Unsubscribe(rh.channel)
	rh.client.Close()
}

func (rh *redisHandler) updateConfiguration(msg *redis.Message) {
	log.Println("the message is", msg.Payload)

	agents, err := af.loadAgentsFromString(msg.Payload)

	if err != nil {
		log.Fatal("Bad json configuration, sucka!!!!!!")
	}

	af.shutDownAgents()
	af.Agents = agents
	af.startAgents()
}
