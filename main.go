package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/valyala/fasthttp"
)

const (
	ACSIp       = "127.0.0.1"
	ACSPort     = 9986
	BankerIp    = "127.0.0.1"
	BankerPort  = 9985
	BidderWin   = 7653
	BidderEvent = 7652
	BidderError = 7651
)

var bidderPort int

func printPortConfigs() {
	log.Printf("Bidder port: %d", bidderPort)
	log.Printf("Win port: %d", BidderWin)
	log.Printf("Event port: %d", BidderEvent)
}

func setupHandlers(agents []Agent) {
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

func eventMux(ctx *fasthttp.RequestCtx) {
	// var f interface{}

	// s := string(ctx.Request.Header.Header()[:])
	// s := string(ctx.Request.Body()[:])
	// log.Println("string is", s)

	// for name, headers := range ctx.Request.Header.Header() {
	// 	name = strings.ToLower(name)
	// 	for _, h := range headers {
	// 		log.Println(name, h)
	// 		// 	request = append(request, fmt.Sprintf(“%v: %v”, name, h))
	// 	}
	// }

	log.Println("Event!!!!!")
	ctx.SetStatusCode(http.StatusOK)
	BidEvent()
}

func main() {
	var agentsConfigFile = flag.String("config", "agents.json", "Configuration file in JSON.")
	flag.IntVar(&bidderPort, "port", 7654, "Port to listen on for router")

	flag.Parse()
	if *agentsConfigFile == "" {
		log.Fatal("You should provide a configuration file.")
	}

	printPortConfigs()

	// http client to pace agents (note that it's pointer)
	client := &http.Client{}

	// load configuration
	agents, err := LoadAgentsFromFile(*agentsConfigFile)

	if err != nil {
		log.Fatal(err)
	}
	for _, agent := range agents {
		agent.RegisterAgent(client, ACSIp, ACSPort)
		agent.StartPacer(client, BankerIp, BankerPort)
	}

	StartStatOutput()

	setupHandlers(agents)

	go fasthttp.ListenAndServe(fmt.Sprintf(":%d", BidderEvent), eventMux)
	log.Println("Started event Mux")

	errormux := func(ctx *fasthttp.RequestCtx) {
		// var f interface{}

		s := string(ctx.Request.Header.Header()[:])
		log.Println("string is", s)

		// for name, headers := range ctx.Request.Header.Header() {
		// 	name = strings.ToLower(name)
		// 	for _, h := range headers {
		// 		log.Println(name, h)
		// 		// 	request = append(request, fmt.Sprintf(“%v: %v”, name, h))
		// 	}
		// }

		ctx.SetStatusCode(http.StatusOK)
		BidEvent()
	}

	go fasthttp.ListenAndServe(fmt.Sprintf(":%d", BidderError), errormux)
	log.Println("Started error Mux")

	winmux := func(ctx *fasthttp.RequestCtx) {
		// log.Println(ctx.PostBody())
		ctx.SetStatusCode(fasthttp.StatusOK)
		BidWin()
	}

	go fasthttp.ListenAndServe(fmt.Sprintf(":%d", BidderWin), winmux)
	log.Println("Started Win Mux")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	select {
	case <-c:
		// Implement remove agent from ACS
		for _, agent := range agents {
			agent.UnregisterAgent(client, ACSIp, ACSPort)
		}
		fmt.Println("Leaving...")
	}
}
