package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/valyala/fasthttp"

	openrtb "gopkg.in/bsm/openrtb.v2"
)

const (
	ACSIp       string = "127.0.0.1"
	ACSPort            = 9986
	BankerIp           = "127.0.0.1"
	BankerPort         = 9985
	BidderWin          = 7653
	BidderEvent        = 7652
	BidderError        = 7651
)

var BidderPort int

func printPortConfigs() {
	log.Printf("Bidder port: %d", BidderPort)
	log.Printf("Win port: %d", BidderWin)
	log.Printf("Event port: %d", BidderEvent)
}

func fastHandleAuctions(ctx *fasthttp.RequestCtx, agents []Agent) {
	var (
		ok    bool = true
		tmpOk bool = true
	)

	// enc := json.NewEncoder(w)
	// body, _ := ioutil.ReadAll(r.Body)
	// fmt.Println(string(body))
	var req *openrtb.BidRequest
	err := json.Unmarshal(ctx.PostBody(), &req)
	// req, err := openrtb.ParseRequest(r.Body)

	if err != nil {
		log.Println("ERROR", err.Error())
		ctx.SetStatusCode(fasthttp.StatusNoContent)
		return
	}

	if req.Test == 1 {
		log.Println("the test is true")
	} else {
		log.Println("test is not true", req.Test)
	}

	// log.Println("INFO Received bid request", req.ID)

	ids := externalIdsFromRequest(req)
	res := emptyResponseWithOneSeat(req)

	for _, agent := range agents {
		res, tmpOk = agent.DoBid(req, res, ids)
		ok = tmpOk || ok

		if tmpOk {
			BidIncoming()
		}
	}

	if ok {
		ctx.Response.Header.Set("Content-type", "application/json")
		ctx.Response.Header.Set("x-openrtb-version", "2.1")
		ctx.SetStatusCode(http.StatusOK)

		bytes, _ := json.Marshal(res)
		ctx.SetBody(bytes)

		return
	}
	log.Println("No bid.")
	ctx.SetStatusCode(204)
}

func main() {
	var agentsConfigFile = flag.String("config", "agents.json", "Configuration file in JSON.")
	flag.IntVar(&BidderPort, "port", 7654, "Port to listen on for router")

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

	m := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/auctions":
			fastHandleAuctions(ctx, agents)
		default:
			ctx.Error("not found", fasthttp.StatusNotFound)
		}
	}

	go fasthttp.ListenAndServe(fmt.Sprintf(":%d", BidderPort), m)
	log.Println("Started Bid Mux")

	eventmux := func(ctx *fasthttp.RequestCtx) {
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

	go fasthttp.ListenAndServe(fmt.Sprintf(":%d", BidderEvent), eventmux)
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

	// evemux := http.NewServeMux()
	// evemux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	log.Println(r.Header)
	//
	// 	defer r.Body.Close()
	// 	w.WriteHeader(http.StatusOK)
	// 	io.WriteString(w, "")
	//
	// 	BidEvent()
	// })
	// go http.ListenAndServe(fmt.Sprintf(":%d", BidderEvent), evemux)

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
