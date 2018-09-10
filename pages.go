package main

type Pages struct {
	Current     int    `json:"current"`
	Last        int    `json:"last"`
	NextURL     string `json:"next_url"`
	PreviousURL string `json:"previous_url"`
}
