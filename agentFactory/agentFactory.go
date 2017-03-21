package agentFactory

import (
	"encoding/json"
	"io/ioutil"
)

type Factory struct {
	Agents []Agent
}

func loadAgents(data []byte) ([]Agent, error) {
	type Agents []Agent
	var agents Agents

	err := json.Unmarshal(data, &agents)
	if err != nil {
		return nil, err
	}
	return agents, nil
}

// LoadAgentsFromFile Parse a JSON file and return a list of Agents.
func (factory *Factory) LoadAgentsFromFile(filepath string) ([]Agent, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	return loadAgents(data)
}

// LoadAgentsFromString ...
func LoadAgentsFromString(agentsString string) ([]Agent, error) {
	return loadAgents([]byte(agentsString))
}
