package solana_rand

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/mr-tron/base58"
	"log"
	"math"
	mathrand "math/rand"
	"time"
)

const (
	SeedLength      = 8
	MinYieldTime    = 1
	MaxYieldTime    = 5
	BlockHashLength = 32
)

func getRandomUint64() (uint64, error) {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		log.Printf("getRandomUint64: %v\n", err)
		return 0, err
	}

	return binary.LittleEndian.Uint64(b), nil
}

func getRandomIntn(n uint64) (uint64, error) {
	if n <= 0 {
		log.Printf("getRandomIntn: n must be positive\n")
		return 0, errors.New("n must be positive")
	}

	maxMultipleOfN := n * (math.MaxUint64 / n)

	for {
		r, err := getRandomUint64()
		if err != nil {
			log.Printf("getRandomIntn: %v\n", err)
			return 0, err
		}

		if r < maxMultipleOfN {
			return r % n, nil
		}

		// retry
	}
}

func GetSeedFromBlockchain() (int64, []string, error) {
	blockHashes := []string{}
	for i := 0; i < SeedLength; i++ {
		latestBlockhash, err := GetLatestBlockhash()
		if err != nil {
			log.Printf("Can't continue without blockhashes\n")
			return 0, []string{}, errors.New("Can't continue without blockhashes")
		}

		log.Printf("GetSeedFromBlockchain: Fetching blockhash %v from chain: %v [Block Height %v]\n",
			i+1, latestBlockhash.Blockhash, latestBlockhash.LatestValidBlockHeight)

		blockHashes = append(blockHashes, latestBlockhash.Blockhash)
		rndInt, err := getRandomIntn(MaxYieldTime - MinYieldTime + 1)
		if err != nil {
			log.Printf("GetSeedFromBlockchain: %v\n", err)
			return 0, []string{}, err
		}

		n := time.Second * time.Duration(MinYieldTime+rndInt)
		log.Printf("GetSeedFromBlockchain: Waiting for %v\n", n)
		time.Sleep(n)
	}

	seed, err := GetSeedFromBlockHashes(blockHashes)
	return seed, blockHashes, err
}

func GetSeedFromBlockHashes(blockHashes []string) (int64, error) {
	if len(blockHashes) != SeedLength {
		log.Printf("GetSeedFromBlockHashes: must provide %v blockhashes\n", SeedLength)
		return 0, fmt.Errorf("must provide %v blockhashes", SeedLength)
	}

	seedHash := sha256.New()
	for i := 0; i < SeedLength; i++ {
		b, err := base58.Decode(blockHashes[i])
		if err != nil {
			log.Printf("GetSeedFromBlockHashes: i %v, blockhash %v, err %v\n", i, blockHashes[i], err)
			return 0, err
		}

		if len(b) != BlockHashLength {
			log.Printf("GetSeedFromBlockHashes: i %v, blockhash %v, len(b) == %v\n", i, blockHashes[i], len(b))
			return 0, fmt.Errorf("len(b) == %v", len(b))
		}

		seedHash.Write(b)
	}

	seedBytes := seedHash.Sum(nil)

	return int64(binary.LittleEndian.Uint64(seedBytes)), nil
}

func RandFromSeed(seed int64) *mathrand.Rand {
	return mathrand.New(mathrand.NewSource(seed))
}
