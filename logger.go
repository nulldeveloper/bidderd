package main

import (
	"log"

	"encoding/json"

	"github.com/valyala/fasthttp"
	"gopkg.in/bsm/openrtb.v2"
)

type logData struct {
	AuctionData *openrtb.BidRequest  `json:"auction_data"`
	BidData     *openrtb.BidResponse `json:"bid_data"`
}

type logger struct {
}

func (l *logger) log(data logData) {
	b, err := json.Marshal(&data)

	if err != nil {
		log.Println("Error marshalling log data")
	}

	req := fasthttp.AcquireRequest()
	req.Header.SetMethod("POST")
	req.SetRequestURI("http://127.0.0.1:1111/bids")
	// req.SetHost("127.0.0.1:1111/bids")
	req.SetBody(b)
	req.Header.SetContentType("application/json")

	resp := fasthttp.AcquireResponse()
	client := &fasthttp.Client{}
	err = client.Do(req, resp)

	if err != nil {
		log.Println("error is", err)
	}

	bodyBytes := resp.Body()
	println(string(bodyBytes))
}
