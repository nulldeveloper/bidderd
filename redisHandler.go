package main

import (
	"log"

	redis "gopkg.in/redis.v5"
)

const channel = "Test"

var redisClient *redis.Client
var subscriber *redis.PubSub

type redisHandler struct {
}

func (rh *redisHandler) setupClient() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	log.Println("Setting up client")
	rh.subscribe()
}

func (rh *redisHandler) subscribe() {
	log.Println("Starting subscriber")

	var err error
	subscriber, err = redisClient.Subscribe(channel)

	if err != nil {
		panic(err)
	}
}

func (rh *redisHandler) startRedisSubscriber() {
	log.Println("Waiting for Configuration Changes")
	for {
		msg, err := subscriber.ReceiveMessage()

		if err != nil {
			log.Println("error recieving message")
		} else {
			rh.updateConfiguration(msg)
		}
	}
}

func (rh *redisHandler) stopRedisSubscriber() {
	subscriber.Close()
	subscriber.Unsubscribe(channel)
	redisClient.Close()
}

func (rh *redisHandler) updateConfiguration(msg *redis.Message) {
	log.Println("the message is", msg.Payload)

	af := agentFactory{}
	agents, err := af.LoadAgentsFromString(msg.Payload)

	if err != nil {
		log.Fatal("Bad json configuration, sucka!!!!!!")
	}

	shutDownAgents(_agents)
	_agents = agents
	startAgents(_agents)
}
