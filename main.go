package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Request struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []int  `json:"params"`
	ID      int    `json:"id"`
}

type Response struct {
	Error  interface{} `json:"error,omitempty"`
	Result interface{} `json:"result,omitempty"`
}

const url = "http://max-desktop:8545"

func main() {
	start := 7507000
	end := 7508000

	var naughtyBlocks = make(map[int]interface{})

	for i := start; i <= end; i++ {
		resp, err := traceBlock(i)
		if err != nil {
			fmt.Println(err)
			naughtyBlocks[i] = resp
		}
	}

	fmt.Println("done!")
}

func traceBlock(blockno int) (interface{}, error) {

	v := Request{
		Jsonrpc: "2.0",
		Method:  "trace_block",
		Params:  []int{blockno},
		ID:      0,
	}

	json_data, err := json.Marshal(v)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		log.Fatal(err)
	}

	var res Response
	json.NewDecoder(resp.Body).Decode(&res)

	if res.Error != nil {
		return res, fmt.Errorf("block %d is naughty", blockno)
	}

	return res, nil
}
