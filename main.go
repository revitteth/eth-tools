package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
)

type Request struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []int  `json:"params"`
	ID      int    `json:"id"`
}

type Response struct {
	Error  *Error      `json:"error,omitempty"`
	Result interface{} `json:"result,omitempty"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

const url = "http://localhost:8545"
const reg = "0[xX][0-9a-fA-F]+"

func main() {
	start := 7506000
	end := 7508000

	re := regexp.MustCompile(reg)

	var naughtyBlocks = make(map[int]*Error)

	for i := start; i <= end; i++ {
		resp, err := traceBlock(i)
		if err != nil {
			naughtyBlocks[i] = resp.Error
			match := re.FindString(resp.Error.Message)
			fmt.Printf("block, %d, addr, %s, msg, %s\n", i, match, resp.Error.Message)
		}
	}

	fmt.Println("done!")
}

func traceBlock(blockno int) (Response, error) {

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
		return res, fmt.Errorf("block %d errored", blockno)
	}

	return res, nil
}
