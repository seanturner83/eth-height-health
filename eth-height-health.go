package main

import (
	"encoding/json"
	"fmt"
	jsonrpc2 "github.com/ybbus/jsonrpc"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
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

type NP struct {
	Status bool `json:"status"`
	Data   int  `json:"data"`
}

type RPC2 struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  string `json:"result"`
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {


		var ethereum_node = "http://localhost:8545"
		var threshold = 10

		i := 0
		// blockcypher REST API
		var e Ethereum_v1
		var h int64

		resp, err := http.Get("https://api.blockcypher.com/v1/eth/main")
		if err != nil || resp.StatusCode != 200 {
			i++
		} else {

			defer resp.Body.Close()

			err = json.NewDecoder(resp.Body).Decode(&e)

			if err != nil {
				http.Error(w, "Failed to decode blockcypher json response", 500)
			}

			h = int64(e.Height)
		}

		// nanopool RPC over http
		var nh int64
		response, err := http.Get("https://api.nanopool.org/v1/eth/network/lastblocknumber/")
		if err != nil || response.StatusCode != 200 {
			i++
		} else {
			defer response.Body.Close()
			contents, err := ioutil.ReadAll(response.Body)
			if err != nil {
				http.Error(w, "Reading nanopool response body failed", 500)
			}
			str := string(contents)
			res := strings.NewReader(str)
			var n NP
			err = json.NewDecoder(res).Decode(&n)
			if err != nil {
				http.Error(w, "Failed to decode nanopool json response", 500)
			}
			nh = int64(n.Data)
		}

		// etherscan RPC2.0 over http (hex)
		var eh int64
		respo, err := http.Get("https://api.etherscan.io/api?module=proxy&action=eth_blockNumber")
		if err != nil || respo.StatusCode != 200 {
			i++
		} else {
			defer respo.Body.Close()
			cont, err := ioutil.ReadAll(respo.Body)
			if err != nil {
				http.Error(w, "Reading etherscan response body failed", 500)
			}
			stri := string(cont)
			ress := strings.NewReader(stri)
			var ee RPC2
			err = json.NewDecoder(ress).Decode(&ee)
			if err != nil {
				http.Error(w, "Failed to decode etherscan json response", 500)
			}
			eh, err = strconv.ParseInt(ee.Result, 0, 64)
			if err != nil {
				http.Error(w, "Failed to parse etherscan hex into integer", 500)
			}
		}

		// ethereum (parity) node jsonrpc2.0 (hex)
		var dec int64
		rpcClient2 := jsonrpc2.NewClient(ethereum_node)
		rep, err := rpcClient2.Call("eth_blockNumber")
		if err != nil {
			http.Error(w, "RPC request to parity failed", 400)
		} else {

			hex, err := rep.GetString()
			if err != nil {
				http.Error(w, "Failed to parse parity RPC response", 500)
			}

			dec, err = strconv.ParseInt(hex, 0, 64)
			if err != nil {
				http.Error(w, "Failed to parse parity hex into integer", 500)
			}
		}

		if i == 3 {
			http.Error(w, "Couldn't contact any external API", 408)
		}

		var th int64
		th = int64(threshold)
		us := dec + th

		if h > us {
			http.Error(w, "Blockcypher is ahead of this node!", 418)
		}
		if nh > us {
			http.Error(w, "Nanopool is ahead of this node!", 400)
		}
		if eh > us {
			http.Error(w, "Etherscan is ahead of this node!", 400)
		}

		fmt.Fprint(w, "Current chain height according to blockcypher; ", h, ", nanopool; ", nh, ", etherscan; ", eh, ", current height on this node; ", dec)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))

}
