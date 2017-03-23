package main

import (
	"encoding/json"
	"io/ioutil"
)

type agentFactory struct {
	Agents []Agent
}

func (af *agentFactory) loadAgents(data []byte) ([]Agent, error) {
	type Agents []Agent
	var agents Agents

	err := json.Unmarshal(data, &agents)
	if err != nil {
		return nil, err
	}
	return agents, nil
}

// LoadAgentsFromFile Parse a JSON file and return a list of Agents.
func (af *agentFactory) LoadAgentsFromFile(filepath string) ([]Agent, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	return af.loadAgents(data)
}

// LoadAgentsFromString ...
func (af *agentFactory) LoadAgentsFromString(agentsString string) ([]Agent, error) {
	return af.loadAgents([]byte(agentsString))
}
