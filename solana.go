package solana_rand

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/rpc"
	"go.uber.org/ratelimit"
)

const DEBUG = false

const SOLANA_RPC_ENDPOINT = "https://api.mainnet-beta.solana.com/"

var ENDPOINTS = []string{SOLANA_RPC_ENDPOINT}

func GetRateLimit() int {
	return 4
}

var RateLimiter = ratelimit.New(GetRateLimit(), ratelimit.Per(1*time.Second))

func checkIfRateLimitErrorAndWait(err error) {
	x := fmt.Sprintf("%v", err)
	if strings.Contains(x, "429") {
		RateLimiter.Take()
	}
}

func GetLatestBlockhash() (blockhash rpc.GetLatestBlockhashValue, err error) {
	RateLimiter.Take()

	for _, endpoint := range ENDPOINTS {
		c := client.NewClient(endpoint)

		blockhash, err := c.GetLatestBlockhash(context.Background())

		if DEBUG {
			log.Printf("GetLatestBlockhash(%v): %v, err %+v\n", endpoint, blockhash, err)
		}

		if err != nil {
			if DEBUG {
				log.Printf("GetLatestBlockhash(%v): failed to get blockhash, err: %v\n", endpoint, err)
			}
			checkIfRateLimitErrorAndWait(err)
		} else {
			return blockhash, nil
		}
	}

	return blockhash, err
}
