package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"github.com/connectedinteractive/bidderd/agent"
	"github.com/valyala/fasthttp"
)

const (
	ACSIP       = "127.0.0.1"
	ACSPort     = 9986
	BankerIP    = "127.0.0.1"
	BankerPort  = 9985
	BidderWin   = 7653
	BidderEvent = 7652
	BidderError = 7651
	BiddingPort = 7654
)

var bidderPort int
var wg sync.WaitGroup
var _agents []BiddingAgent.Agent

// http client to pace agents (note that it's pointer)
var client = &http.Client{}

func printPortConfigs() {
	log.Printf("Bidder port: %d", bidderPort)
	log.Printf("Win port: %d", BidderWin)
	log.Printf("Event port: %d", BidderEvent)
}

func setupHandlers(agents []BiddingAgent.Agent) {
	m := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/auctions":
			fastHandleAuctions(ctx, agents)
		default:
			ctx.Error("not found", fasthttp.StatusNotFound)
		}
	}

	go fasthttp.ListenAndServe(fmt.Sprintf(":%d", bidderPort), m)
	log.Println("Started Bid Mux")
}

func cleanup(agents []BiddingAgent.Agent) {
	// stopRedisSubscriber()
	// Implement remove agent from ACS
	shutDownAgents(agents)
	fmt.Println("Leaving...")
	for i := 0; i < 3; i++ {
		wg.Done()
	}
}

func startAgents(agents []BiddingAgent.Agent) {
	log.Printf("Starting Up %d Agents", len(agents))
	for _, agent := range agents {
		agent.RegisterAgent(client, ACSIP, ACSPort)
		agent.StartPacer(client, BankerIP, BankerPort)
	}
}

func shutDownAgents(agents []BiddingAgent.Agent) {
	log.Println("Shutting Down Agents")
	for _, agent := range agents {
		agent.UnregisterAgent(client, ACSIP, ACSPort)
	}
}

func main() {
	var agentsConfigFile = flag.String("config", "agents.json", "Configuration file in JSON.")
	flag.IntVar(&bidderPort, "port", BiddingPort, "Port to listen on for router")
	flag.Parse()

	if *agentsConfigFile == "" {
		log.Fatal("You should provide a configuration file.")
	}

	// setupClient()
	// go startRedisSubscriber()
	// wg.Add(1)

	printPortConfigs()

	// load configuration
	_agents, err := BiddingAgent.LoadAgentsFromFile(*agentsConfigFile)

	if err != nil {
		log.Fatal(err)
	}

	startAgents(_agents)

	BiddingAgent.StartStatOutput()
	setupHandlers(_agents)

	go fasthttp.ListenAndServe(fmt.Sprintf(":%d", BidderEvent), eventMux)
	log.Println("Started event Mux")

	go fasthttp.ListenAndServe(fmt.Sprintf(":%d", BidderError), errorMux)
	log.Println("Started error Mux")

	go fasthttp.ListenAndServe(fmt.Sprintf(":%d", BidderWin), winMux)
	log.Println("Started Win Mux")

	wg.Add(3)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	go func() {
		<-c
		cleanup(_agents)
		os.Exit(1)
	}()

	wg.Wait()
}
