package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/connectedinteractive/bidderd/agent"
	"github.com/valyala/fasthttp"

	openrtb "gopkg.in/bsm/openrtb.v2"
)

var openRTBVersion = "2.2"

func fastHandleAuctions(ctx *fasthttp.RequestCtx, agents []BiddingAgent.Agent) {
	var (
		ok    = true
		tmpOk = true
	)

	log.Println("Got a bid!!")

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

	ids := BiddingAgent.ExternalIdsFromRequest(req)
	res := BiddingAgent.EmptyResponseWithOneSeat(req)

	for _, agent := range agents {
		res, tmpOk = agent.DoBid(req, res, ids)
		ok = tmpOk || ok

		if tmpOk {
			BiddingAgent.BidIncoming()
		}
	}

	if ok {
		ctx.Response.Header.Set("Content-type", "application/json")
		ctx.Response.Header.Set("x-openrtb-version", openRTBVersion)
		ctx.SetStatusCode(http.StatusOK)

		bytes, _ := json.Marshal(res)
		ctx.SetBody(bytes)

		return
	}
	log.Println("No bid.")
	ctx.SetStatusCode(204)
}

func errorMux(ctx *fasthttp.RequestCtx) {
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
	BiddingAgent.BidEvent()
}

func winMux(ctx *fasthttp.RequestCtx) {
	// log.Println(ctx.PostBody())
	ctx.SetStatusCode(fasthttp.StatusOK)
	BiddingAgent.BidWin()
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
	BiddingAgent.BidEvent()
}
