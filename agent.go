package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	openrtb "gopkg.in/bsm/openrtb.v2"
)

//OutputPerSeconds = number of seconds between stat output
var OutputPerSeconds = 10

//Wins = counter for # of wins in the last OutputPerSeconds
var Wins = 0

//Events = counter for # of events in the last OutputPerSeconds
var Events = 0

//Bids = counter for # of bids in the last OutputPerSeconds
var Bids = 0

var outputChannel = make(chan bool)

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
	Price float64 `json:"price"`

	// For pacing the budgeting
	Period  int `json:"period"`
	Balance int `json:"balance"`

	// private state of each agent
	registered bool      // did we register the configuration in the ACS?
	pacer      chan bool // go routine updating balance in the banker
	bidID      int       // unique id for response
}

// CreativesKey This is used to make a mapping between an impression and the
// external-id of an agent to the creatives that can be sent to the
// exchange for that impression.
type CreativesKey struct {
	ImpID string
	ExtID int
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
	tickerChannel := time.NewTicker(time.Second * time.Duration(OutputPerSeconds)).C

	go func() {
		for {
			select {
			case <-tickerChannel:
				printStats()
			case <-outputChannel:
				return
			}
		}
	}()
}

// BidWin ...
func BidWin() {
	Wins++
}

// BidEvent ...
func BidEvent() {
	Events++
}

// BidIncoming ...
func (a *Agent) BidIncoming() {
	Bids++
}

func printStats() {
	tempWins := Wins
	Wins = 0
	winsPerSecond := tempWins / OutputPerSeconds
	tempEvents := Events
	Events = 0
	eventsPerSecond := tempEvents / OutputPerSeconds
	tempBids := Bids
	Bids = 0
	bidsPerSecond := tempBids / OutputPerSeconds
	log.Println("***********************")
	log.Printf("Bids: %d (%d/second)", tempBids, bidsPerSecond)
	log.Printf("Wins: %d (%d/second)", tempWins, winsPerSecond)
	log.Printf("Events: %d (%d/second)", tempEvents, eventsPerSecond)
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
func (agent *Agent) DoBid(req *openrtb.BidRequest, res *openrtb.BidResponse, ids map[CreativesKey]interface{}) (*openrtb.BidResponse, bool) {

	for _, imp := range req.Imp {
		key := CreativesKey{ImpID: imp.ID, ExtID: agent.Config.ExternalID}
		if ids[key] == nil {
			continue
		}

		creativeList := ids[key].([]interface{})
		// pick a random creative
		n := rand.Intn(len(creativeList))

		// Extract a usable creative ID from the JSON parse
		crid := strconv.Itoa(int(creativeList[n].(float64)))

		bidID := strconv.Itoa(agent.bidID)

		rp := randomPrice{percentage: 0.25, price: 1.25}
		price := rp.randomPrice()

		ext := map[string]interface{}{"priority": 1.0, "external-id": agent.Config.ExternalID}
		jsonExt, _ := json.Marshal(ext)
		bid := openrtb.Bid{ID: bidID, ImpID: imp.ID, CreativeID: crid, Price: price, Ext: jsonExt}
		agent.bidID++
		res.SeatBid[0].Bid = append(res.SeatBid[0].Bid, bid)
	}

	res.Currency = "USD"
	res.BidID = strconv.Itoa(agent.bidID)

	return res, len(res.SeatBid[0].Bid) > 0
}

// ExternalIdsFromRequest makes a mappping with a range of type (Impression Id, External Id)
// to a slice of "creative indexes" (See the agent configuration "creative").
// We use this auxiliary function in `DoBid` to match the `BidRequest` to the
// creatives of the agent and create a response.
func ExternalIdsFromRequest(req *openrtb.BidRequest) map[CreativesKey]interface{} {
	ids := make(map[CreativesKey]interface{})

	for _, imp := range req.Imp {
		log.Print("")
		var extJSON map[string]interface{}
		_ = json.Unmarshal(imp.Ext, &extJSON)

		for _, extID := range extJSON["external-ids"].([]interface{}) {
			extID = int(extID.(float64))
			key := CreativesKey{ImpID: imp.ID, ExtID: extID.(int)}
			creatives := (extJSON["creative-ids"].(map[string]interface{}))[strconv.Itoa(extID.(int))]
			ids[key] = creatives.(interface{})
		}
	}
	return ids
}

// EmptyResponseWithOneSeat adds a Seat to the Response.
// Seat: A buyer entity that uses a Bidder to obtain impressions on its behalf.
func EmptyResponseWithOneSeat(req *openrtb.BidRequest) *openrtb.BidResponse {
	seat := openrtb.SeatBid{Bid: make([]openrtb.Bid, 0)}
	seatbid := []openrtb.SeatBid{seat}
	res := &openrtb.BidResponse{ID: req.ID, SeatBid: seatbid}
	return res
}

// LoadAgent Parse a JSON file and return an Agent.
func LoadAgent(filepath string) (Agent, error) {
	var agent Agent
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return Agent{}, err
	}
	err = json.Unmarshal(data, &agent)
	if err != nil {
		return Agent{}, err
	}
	return agent, nil
}

// FindCreativeIndexFromID takes a creative ID and an AgentConfig,
// returning a creative index usable by the RTBKit router
func FindCreativeIndexFromID(crid int, agent AgentConfig) (string, error) {
	for creativeIndex, creative := range agent.Creatives {
		if creative.ID == crid {
			return strconv.Itoa(creativeIndex), nil
		}
	}
	return "", errors.New("Unable to find matching creative")
}
