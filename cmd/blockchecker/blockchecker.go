package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"time"
)

type Request struct {
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

type Response struct {
	Error  ErrorResponse          `json:"error,omitempty"`
	Result map[string]interface{} `json:"result,omitempty"`
}

type ResponseTrancing struct {
	Error  ErrorResponse          `json:"error,omitempty"`
	Result map[string]interface{} `json:"result,omitempty"`
}

type ErrorResponse struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

const url = "http://localhost:8545"

var requestId int = 0

func main() {
	ctx := context.Background()

	f, err := os.Create("output.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	logEvery := time.NewTicker(5 * time.Second)
	var lastTracedBlockNum uint64 = 0
Loop:
	for {
		select {
		case <-ctx.Done():
			break Loop
		case <-logEvery.C:
			blockNum, err := getBlockByNumber()
			if err != nil {
				log.Fatal(err)
			}
			if blockNum != 0 && lastTracedBlockNum != blockNum {
				lastTracedBlockNum = blockNum
				failingBlockMessage, err := traceBlock(blockNum)
				if err != nil {
					log.Fatal(err)
				}
				if failingBlockMessage != "" {
					message := fmt.Sprintf("fails at blockNumber=%d with error=%s\n", blockNum, failingBlockMessage)
					f.WriteString(message)
				}
			}
		default:
		}
	}
}

func traceBlock(blockNum uint64) (string, error) {
	var params []interface{}
	params = append(params, blockNum)
	v := Request{
		Jsonrpc: "2.0",
		Method:  "trace_block",
		Params:  params,
		ID:      requestId,
	}
	fmt.Println("tracing block", blockNum)

	json_data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		return "", err
	}
	requestId++

	var traceResponse Response
	json.NewDecoder(resp.Body).Decode(&traceResponse)
	if traceResponse.Error != (ErrorResponse{}) {
		fmt.Printf("block %d is naughty error=%v\n", blockNum, traceResponse.Error.Message)
		return fmt.Sprintf("block %d is naughty error=%v", blockNum, traceResponse.Error.Message), nil
	}

	return "", nil
}

func getBlockByNumber() (uint64, error) {
	var params []interface{}
	params = append(params, "latestExecuted")
	params = append(params, false)
	v := Request{
		Jsonrpc: "2.0",
		Method:  "eth_getBlockByNumber",
		Params:  params,
		ID:      requestId,
	}
	requestId++

	json_data, err := json.Marshal(v)
	if err != nil {
		return 0, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		return 0, err
	}

	var res Response
	json.NewDecoder(resp.Body).Decode(&res)

	if res.Error != (ErrorResponse{}) {
		return 0, fmt.Errorf("failed at getting BlockByNumber: err=%+v", res.Error)
	}
	if res.Result["number"] == nil {
		fmt.Println("No Response from eth_getBlockByNumber")
		return 0, nil
	}
	blockNumHex := res.Result["number"].(string)
	blockNum, _ := new(big.Int).SetString(blockNumHex[2:], 16)
	if blockNum == nil {
		return 0, fmt.Errorf("couldn't parse blockNum from %s", blockNumHex)
	}

	return blockNum.Uint64(), nil
}
