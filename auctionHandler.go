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
		ctx.Response.Header.Set("x-openrtb-version", "2.2")
		ctx.SetStatusCode(http.StatusOK)

		bytes, _ := json.Marshal(res)
		ctx.SetBody(bytes)

		return
	}
	log.Println("No bid.")
	ctx.SetStatusCode(204)
}
