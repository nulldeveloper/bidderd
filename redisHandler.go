package main

import (
	"log"

	"encoding/json"

	redis "gopkg.in/redis.v5"
)

type redisData struct {
	Name string `json:"name"`
}

type redisHandler struct {
	client     *redis.Client
	subscriber *redis.PubSub
	channel    string
}

const hashName = "AGENTCONFIGURATIONS"

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
			return
		}

		var rd redisData
		err = json.Unmarshal([]byte(msg.Payload), &rd)

		if err != nil {
			log.Println("Unable read message from redis subscriber")
			return
		}

		rh.getAgentConfigFromRedis(rd)
	}
}

func (rh *redisHandler) stopRedisSubscriber() {
	rh.subscriber.Unsubscribe(rh.channel)
	rh.subscriber.Close()
	rh.client.Close()
}

func (rh *redisHandler) getAgentConfigFromRedis(rd redisData) {
	if rd.Name != af.Agent.Name {
		return
	}

	data, err := rh.client.HMGet(hashName, rd.Name).Result()

	if err != nil {
		log.Println("Unable to read agent configuration from redis")
		return
	}

	if len(data) > 0 {
		af.updateAgentConfiguration(data[0].(string))
	}
}
