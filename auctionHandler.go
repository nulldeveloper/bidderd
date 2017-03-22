package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/valyala/fasthttp"
	openrtb "gopkg.in/bsm/openrtb.v2"
)

func fastHandleAuctions(ctx *fasthttp.RequestCtx, agents []Agent) {
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
		ctx.Response.Header.Set("x-openrtb-version", "2.2")
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
	BidEvent()
}

func winMux(ctx *fasthttp.RequestCtx) {
	// log.Println(ctx.PostBody())
	ctx.SetStatusCode(fasthttp.StatusOK)
	BidWin()
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
