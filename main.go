package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"

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

var af = agentFactory{}
var s = newStats()
var rh = newRedis()

// http client to pace agents (note that it's pointer)
var client = &http.Client{}

func printPortConfigs() {
	log.Printf("Bidder port: %d", bidderPort)
	log.Printf("Win port: %d", BidderWin)
	log.Printf("Event port: %d", BidderEvent)
}

func setupHandlers(agent Agent) {
	m := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/auctions":
			fastHandleAuctions(ctx, agent)
		default:
			ctx.Error("not found", fasthttp.StatusNotFound)
		}
	}

	go fasthttp.ListenAndServe(fmt.Sprintf(":%d", bidderPort), m)
	log.Println("Started Bid Mux")
}

func cleanup() {
	fmt.Println("Leaving...")
	// Implement remove agent from ACS
	// rh.stopRedisSubscriber()
	af.shutDownAgents()
	for i := 0; i < 3; i++ {
		wg.Done()
	}
}

func main() {
	var agentsConfigFile = flag.String("config", "", "Configuration file in JSON.")
	flag.IntVar(&bidderPort, "port", BiddingPort, "Port to listen on for router")
	flag.Parse()

	if *agentsConfigFile == "" {
		log.Fatal("You should provide a configuration file. Usage: bidderd --config filename.json ")
	}

	// rh.subscribe()
	// go rh.startRedisSubscriber()
	// wg.Add(1)

	printPortConfigs()

	// load configuration
	_agent, err := af.loadAgentsFromFile(*agentsConfigFile)

	if err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		<-c
		cleanup()
		os.Exit(1)
	}()

	af.setAgent(*_agent)

	StartStatOutput()
	setupHandlers(*_agent)

	go fasthttp.ListenAndServe(fmt.Sprintf(":%d", BidderEvent), eventMux)
	log.Println("Started event Mux")

	go fasthttp.ListenAndServe(fmt.Sprintf(":%d", BidderError), errorMux)
	log.Println("Started error Mux")

	go fasthttp.ListenAndServe(fmt.Sprintf(":%d", BidderWin), winMux)
	log.Println("Started Win Mux")

	wg.Add(3)

	wg.Wait()
}
