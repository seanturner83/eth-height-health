package main

import (
	"encoding/json"
	"fmt"
	"github.com/ybbus/jsonrpc"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
)

type Ethereum_v1 struct {
	Name             string    `json:"name"`
	Height           int       `json:"height"`
	Hash             string    `json:"hash"`
	Time             time.Time `json:"time"`
	LatestURL        string    `json:"latest_url"`
	PreviousHash     string    `json:"previous_hash"`
	PreviousURL      string    `json:"previous_url"`
	PeerCount        int       `json:"peer_count"`
	UnconfirmedCount int       `json:"unconfirmed_count"`
	HighGasPrice     int64     `json:"high_gas_price"`
	MediumGasPrice   int64     `json:"medium_gas_price"`
	LowGasPrice      int64     `json:"low_gas_price"`
	LastForkHeight   int       `json:"last_fork_height"`
	LastForkHash     string    `json:"last_fork_hash"`
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		var e Ethereum_v1

		resp, err := http.Get("https://api.blockcypher.com/v1/eth/main")
		if err != nil {
			http.Error(w, "HTTP request to blockcypher failed", 400)
		}

		defer resp.Body.Close()

		if resp.Body == nil {
			http.Error(w, "Empty response from blockcypher", 400)
			return
		}

		err = json.NewDecoder(resp.Body).Decode(&e)

		if err != nil {
			http.Error(w, "Failed to decode blockcypher json response", 500)
		}

		var h int64
		h = int64(e.Height)

		rpcClient := jsonrpc.NewClient("http://localhost:8545")
		rep, err := rpcClient.Call("eth_blockNumber")
		if err != nil {
			http.Error(w, "RPC request to parity failed", 400)
		}

		hex, err := rep.GetString()
		if err != nil {
			http.Error(w, "Failed to parse RPC response", 500)
		}

		dec, err := strconv.ParseInt(hex, 0, 64)
		if err != nil {
			http.Error(w, "Failed to parse hex into integer", 500)
		}
		var diff float64
		diff = float64(h - dec)
		comp := math.Abs(diff)

		if comp > 10 {
			http.Error(w, "Discrepancy between heights is over limit", 400)
		}

		fmt.Fprint(w, "Current chain height according to blockcypher; ", h, ", current height on this node; ", dec)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))

}
