package main

type Request struct {
	Body string `json:"body"`
}
type ResponseErr struct {
	Error string `json:"error"`
}

type ResponseValid struct {
	Valid bool `json:"valid"`
}

type ResponseClean struct {
	CleanedBody string `json:"cleaned_body"`
}
