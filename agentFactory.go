package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type agentFactory struct {
	Agent Agent
}

func (af *agentFactory) setAgent(a Agent) {
	af.shutDownAgents()
	af.Agent = a
	af.startAgents()
}

func (af *agentFactory) loadAgent(data []byte) (*Agent, error) {
	var a Agent

	err := json.Unmarshal(data, &a)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (af *agentFactory) loadAgentsFromFile(filepath string) (*Agent, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	return af.loadAgent(data)
}

func (af *agentFactory) loadAgentsFromString(agentsString string) (*Agent, error) {
	return af.loadAgent([]byte(agentsString))
}

func (af *agentFactory) updateAgentConfiguration(config string) {
	log.Println("the message is", config)

	agent, err := af.loadAgentsFromString(config)

	if err != nil {
		log.Println("Bad json configuration, sucka!!!!!!")
	} else {
		af.setAgent(*agent)
	}
}

func (af *agentFactory) startAgents() {
	log.Println("Starting up Agent")
	af.Agent.RegisterAgent(client, ACSIP, ACSPort)
	af.Agent.StartPacer(client, BankerIP, BankerPort)
}

func (af *agentFactory) shutDownAgents() {
	log.Println("Shutting Down Agents")
	af.Agent.UnregisterAgent(client, ACSIP, ACSPort)
}
