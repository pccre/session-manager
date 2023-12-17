package main

type Response struct {
  Method  string      `json:"method"`
  Content interface{} `json:"response"`
}
