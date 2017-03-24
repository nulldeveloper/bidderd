package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type agentFactory struct {
	Agents agents
}

func (af *agentFactory) loadAgents(data []byte) ([]Agent, error) {
	var a agents

	err := json.Unmarshal(data, &a)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (af *agentFactory) loadAgentsFromFile(filepath string) (agents, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	return af.loadAgents(data)
}

func (af *agentFactory) loadAgentsFromString(agentsString string) (agents, error) {
	return af.loadAgents([]byte(agentsString))
}

func (af *agentFactory) startAgents() {
	log.Printf("Starting Up %d Agents", len(af.Agents))
	for _, agent := range af.Agents {
		agent.RegisterAgent(client, ACSIP, ACSPort)
		agent.StartPacer(client, BankerIP, BankerPort)
	}
}

func (af *agentFactory) shutDownAgents() {
	log.Println("Shutting Down Agents")
	for _, agent := range af.Agents {
		agent.UnregisterAgent(client, ACSIP, ACSPort)
	}
}
