package main

import "log"

type stats struct {
	outputPerSeconds int
	wins             int
	events           int
	bids             int
}

func newStats() *stats {
	s := &stats{}
	s.outputPerSeconds = 10
	s.wins = 0
	s.events = 0
	s.bids = 0
	return s
}

func (s *stats) bidWin() {
	s.wins++
}

func (s *stats) bidEvent() {
	s.events++
}

func (s *stats) bidIncoming() {
	s.bids++
}

func (s *stats) printStats() {
	tempWins := s.wins
	s.wins = 0
	winsPerSecond := tempWins / s.outputPerSeconds
	tempEvents := s.events
	s.events = 0
	eventsPerSecond := tempEvents / s.outputPerSeconds
	tempBids := s.bids
	s.bids = 0
	bidsPerSecond := tempBids / s.outputPerSeconds
	log.Println("***********************")
	log.Printf("Bids: %d (%d/second)", tempBids, bidsPerSecond)
	log.Printf("Wins: %d (%d/second)", tempWins, winsPerSecond)
	log.Printf("Events: %d (%d/second)", tempEvents, eventsPerSecond)
}
