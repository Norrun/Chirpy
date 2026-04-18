package main

type Request struct {
	Body string `json:"body"`
}

type ResponseValid struct {
	Valid bool `json:"valid"`
}

type ResponseClean struct {
	CleanedBody string `json:"cleaned_body"`
}
