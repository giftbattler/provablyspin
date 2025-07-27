package provablyspin

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
)

func Spin[T Player](seed string, nonce int, players []T) (Result[T], error) {
	if seed == "" || len(players) == 0 {
		return Result[T]{}, errors.New("invalid input")
	}

	r := spinFloat(seed, nonce)

	var (
		total int64
		accum float64
	)
	for _, p := range players {
		total += p.GetStake()
	}

	ranges := make(map[string]float64)
	var winner T
	var winnerFound bool

	for _, p := range players {
		accum += float64(p.GetStake()) / float64(total)
		ranges[p.GetID()] = accum
		if r < accum && !winnerFound {
			winner = p
			winnerFound = true
		}
	}

	if !winnerFound {
		winner = players[len(players)-1]
	}

	// Create control JSON for verification
	controlData := ControlData[T]{
		Seed:    seed,
		Nonce:   nonce,
		Players: players,
	}

	controlJSON, err := json.Marshal(controlData)
	if err != nil {
		return Result[T]{}, fmt.Errorf("failed to create control JSON: %w", err)
	}

	return Result[T]{
		Winner:      winner,
		SpinFloat:   r,
		Nonce:       nonce,
		Ranges:      ranges,
		ControlJSON: controlJSON,
	}, nil
}

func spinFloat(seed string, nonce int) float64 {
	h := hmac.New(sha256.New, []byte(seed))
	h.Write([]byte(fmt.Sprintf("%d", nonce)))
	hash := h.Sum(nil)
	return float64(binary.BigEndian.Uint64(hash[:8])) / float64(math.MaxUint64)
}
