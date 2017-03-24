package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	openrtb "gopkg.in/bsm/openrtb.v2"
)

var outputChannel = make(chan bool)

var be = bidEngine{}

// Creative ...
type Creative struct {
	Format         string           `json:"format"`
	ID             int              `json:"id"`
	UUID           string           `json:"uuid,omitempty"`
	Name           string           `json:"name"`
	ProviderConfig *json.RawMessage `json:"providerConfig"`
}

// AgentConfig This is the agent configuration that will be sent to RTBKIT's ACS
type AgentConfig struct {
	// We use `RawMessage` for Augmentations and BidcControl, because we
	// don't need it, we just cache it.
	Account            []string         `json:"account"`
	Augmentations      *json.RawMessage `json:"augmentations"`
	BidControl         *json.RawMessage `json:"bidControl"`
	BidProbability     float64          `json:"bidProbability"`
	Creatives          []Creative       `json:"creatives"`
	ErrorFormat        string           `json:"errorFormat"`
	External           bool             `json:"external"`
	ExternalID         int              `json:"externalId"`
	LossFormat         string           `json:"lossFormat"`
	MinTimeAvailableMs float64          `json:"minTimeAvailableMs"`
	ProviderConfig     *json.RawMessage `json:"providerConfig"`
	WinFormat          string           `json:"winFormat"`
	BidderInterface    string           `json:"bidderInterface"`
}

// Agent This represents a RTBKIT Agent
type Agent struct {
	Name   string      `json:"name"`
	Config AgentConfig `json:"config"`

	// This is the price the agent will pay per impression. "Fixed price bidder".
	Price      float64 `json:"price"`
	Percentage float64 `json:"percentage"`

	// For pacing the budgeting
	Period  int `json:"period"`
	Balance int `json:"balance"`

	// private state of each agent
	registered bool      // did we register the configuration in the ACS?
	pacer      chan bool // go routine updating balance in the banker
	bidID      int       // unique id for response
}

// RegisterAgent in the ACS sending a HTTP request to the service on `acsIp`:`acsPort`
func (agent *Agent) RegisterAgent(httpClient *http.Client, acsIP string, acsPort int) {
	url := fmt.Sprintf("http://%s:%d/v1/agents/%s/config", acsIP, acsPort, agent.Name)
	body, _ := json.Marshal(agent.Config)
	reader := bytes.NewReader(body)
	req, _ := http.NewRequest("POST", url, reader)
	req.Header.Add("Accept", "application/json")
	res, err := httpClient.Do(req)
	if err != nil {
		log.Printf("ACS registration failed with %s\n", err)
		return
	}
	agent.registered = true
	res.Body.Close()
}

// UnregisterAgent Removes the agent configuration from the ACS
func (agent *Agent) UnregisterAgent(httpClient *http.Client, acsIP string, acsPort int) {
	url := fmt.Sprintf("http://%s:%d/v1/agents/%s/config", acsIP, acsPort, agent.Name)
	req, _ := http.NewRequest("DELETE", url, bytes.NewBufferString(""))
	res, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Unregister failed with %s\n", err)
		return
	}
	agent.registered = false
	res.Body.Close()
}

func pace(httpClient *http.Client, url string, body string) {
	log.Println("Pacing...")
	req, _ := http.NewRequest("POST", url, strings.NewReader(body))
	req.Header.Add("Accept", "application/json")
	res, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Balance failed with %s\n", err)
		return
	}
	res.Body.Close()
}

// StartPacer Starts a go routine which periodically updates the balance on the agents account.
func (agent *Agent) StartPacer(
	httpClient *http.Client, bankerIP string, bankerPort int) {

	accounts := agent.Config.Account

	url := fmt.Sprintf("http://%s:%d/v1/accounts/%s/balance",
		bankerIP, bankerPort, strings.Join(accounts, ":"))
	body := fmt.Sprintf("{\"USD/1M\": %d}", agent.Balance)
	ticker := time.NewTicker(time.Duration(agent.Period) * time.Millisecond)
	agent.pacer = make(chan bool)

	//Run pacer on startup
	go pace(httpClient, url, body)

	go func() {
		for {
			select {
			case <-ticker.C:
				// make this a new go routine?
				go pace(httpClient, url, body)
			case <-agent.pacer:
				ticker.Stop()
				return
			}
		}
	}()
}

// StartStatOutput is responsible for displaying the number of wins and events per timer tick
func StartStatOutput() {
	tickerChannel := time.NewTicker(time.Second * time.Duration(s.outputPerSeconds)).C

	go func() {
		for {
			select {
			case <-tickerChannel:
				s.printStats()
			case <-outputChannel:
				return
			}
		}
	}()
}

// StopPacer Stops the go routine updating the bank balance.
func (agent *Agent) StopPacer() {
	close(outputChannel)
	close(agent.pacer)
}

// DoBid Adds to the bid response the bid by the agent. The Bid is added to
// the only seat of the response. It picks a random creative from
// the list of creatives from the `Agent.Config.Creative` and places it
// in the bid.
func (agent *Agent) DoBid(req *openrtb.BidRequest) (*openrtb.BidResponse, bool) {
	res := be.bid(*req, *agent)
	return &res, len(res.SeatBid[0].Bid) > 0
}
