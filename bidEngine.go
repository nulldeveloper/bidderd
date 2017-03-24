package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"strconv"

	openrtb "gopkg.in/bsm/openrtb.v2"
)

// CreativesKey This is used to make a mapping between an impression and the
// external-id of an agent to the creatives that can be sent to the
// exchange for that impression.
type CreativesKey struct {
	ImpID string
	ExtID int
}

type bidEngine struct {
}

func (be *bidEngine) bid(req openrtb.BidRequest, agent Agent) openrtb.BidResponse {
	ids := be.externalIdsFromRequest(req)
	res := be.emptyResponseWithOneSeat(req)

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

		rp := randomPrice{percentage: agent.Percentage, price: agent.Price}
		price := rp.randomPrice()

		ext := map[string]interface{}{"priority": 1.0, "external-id": agent.Config.ExternalID}
		jsonExt, _ := json.Marshal(ext)
		bid := openrtb.Bid{ID: bidID, ImpID: imp.ID, CreativeID: crid, Price: price, Ext: jsonExt}
		agent.bidID++
		res.SeatBid[0].Bid = append(res.SeatBid[0].Bid, bid)
	}

	res.Currency = "USD"
	res.BidID = strconv.Itoa(agent.bidID)

	return res
}

func (be *bidEngine) externalIdsFromRequest(req openrtb.BidRequest) map[CreativesKey]interface{} {
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
func (be *bidEngine) emptyResponseWithOneSeat(req openrtb.BidRequest) openrtb.BidResponse {
	seat := openrtb.SeatBid{Bid: make([]openrtb.Bid, 0)}
	seatbid := []openrtb.SeatBid{seat}
	res := openrtb.BidResponse{ID: req.ID, SeatBid: seatbid}
	return res
}
