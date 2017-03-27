package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/valyala/fasthttp"

	openrtb "gopkg.in/bsm/openrtb.v2"
)

var openRTBVersion = "2.2"
var l = logger{}

func fastHandleAuctions(ctx *fasthttp.RequestCtx, agent Agent) {
	var (
		ok    = true
		tmpOk = true
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

	// log.Println("INFO Received bid request", req.ID)
	var res *openrtb.BidResponse

	res, tmpOk = agent.DoBid(req)
	ok = tmpOk || ok

	if tmpOk {
		s.bidIncoming()
	}

	if ok {
		// ld := logData{AuctionData: req, BidData: res}
		// go l.log(ld)

		ctx.Response.Header.Set("Content-type", "application/json")
		ctx.Response.Header.Set("x-openrtb-version", openRTBVersion)
		ctx.SetStatusCode(http.StatusOK)

		bytes, _ := json.Marshal(res)
		ctx.SetBody(bytes)

		print(string(bytes))

		return
	}
	log.Println("No bid.")
	ctx.SetStatusCode(204)
}

func errorMux(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(http.StatusOK)
	s.bidEvent()
}

func winMux(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	s.bidWin()
}

func eventMux(ctx *fasthttp.RequestCtx) {
	log.Println("Event!!!!!")
	ctx.SetStatusCode(http.StatusOK)
	s.bidEvent()
}
