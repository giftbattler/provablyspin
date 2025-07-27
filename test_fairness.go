package provablyspin

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

type TestResult struct {
	HashedSeed   string             `json:"hashed_seed"`
	RevealedSeed string             `json:"revealed_seed"`
	Nonce        int                `json:"nonce"`
	Players      []GamePlayer       `json:"players"`
	Result       Result[GamePlayer] `json:"result"`
}

func main() {
	fmt.Println("🎰 Testing Provably Fair Spin System")
	fmt.Println("=====================================")

	// Step 1: Create test players
	players := []GamePlayer{
		{ID: "player1", Name: "Alice", Stake: 100},
		{ID: "player2", Name: "Bob", Stake: 200},
		{ID: "player3", Name: "Charlie", Stake: 300},
	}

	// Step 2: Generate random seed
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		panic(err)
	}
	seed := hex.EncodeToString(randomBytes)

	// Step 3: Calculate hashed seed (this would be published before game)
	h := sha256.Sum256([]byte(seed))
	hashedSeed := hex.EncodeToString(h[:])

	nonce := 1

	fmt.Printf("1. 📋 Game Setup:\n")
	fmt.Printf("   Players: Alice(100), Bob(200), Charlie(300)\n")
	fmt.Printf("   Hashed Seed: %s\n", hashedSeed[:16]+"...")
	fmt.Printf("   Nonce: %d\n\n", nonce)

	// Step 4: Generate spin result
	result, err := Spin(seed, nonce, players)
	if err != nil {
		panic(err)
	}

	fmt.Printf("2. 🎲 Spin Result:\n")
	fmt.Printf("   Winner: %s (ID: %s)\n", result.Winner.GetName(), result.Winner.GetID())
	fmt.Printf("   Spin Value: %.10f\n", result.SpinFloat)
	fmt.Printf("   Ranges: %+v\n\n", result.Ranges)

	// Step 5: Create verification JSON
	testData := TestResult{
		HashedSeed:   hashedSeed,
		RevealedSeed: seed,
		Nonce:        nonce,
		Players:      players,
		Result:       result,
	}

	jsonData, err := json.MarshalIndent(testData, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Printf("3. 📄 Generated JSON for verification:\n")
	fmt.Printf("%s\n\n", string(jsonData))

	// Step 6: Verify the result using our verification function
	fmt.Printf("4. ✅ Verification Process:\n")

	// Check hash matches
	expectedHash := hex.EncodeToString(sha256.New().Sum([]byte(seed)))
	hashValid := expectedHash == hashedSeed
	fmt.Printf("   Hash validation: %v\n", hashValid)

	// Full verification
	isValid := VerifySpin(hashedSeed, seed, nonce, players, result)
	fmt.Printf("   Full verification: %v\n", isValid)

	// Step 7: Also test ControlJSON validation
	fmt.Printf("\n5. 🔍 ControlJSON Test:\n")
	fmt.Printf("   ControlJSON: %s\n", string(result.ControlJSON))

	var controlData ControlData[GamePlayer]
	if err := json.Unmarshal(result.ControlJSON, &controlData); err != nil {
		panic(err)
	}

	// Verify control data can reproduce the result
	controlResult, err := Spin(controlData.Seed, controlData.Nonce, controlData.Players)
	if err != nil {
		panic(err)
	}

	controlValid := controlResult.Winner.GetID() == result.Winner.GetID() &&
		controlResult.SpinFloat == result.SpinFloat &&
		controlResult.Nonce == result.Nonce

	fmt.Printf("   ControlJSON validation: %v\n", controlValid)

	// Final result
	fmt.Printf("\n🏆 FINAL RESULT:\n")
	if isValid && controlValid && hashValid {
		fmt.Println("✅ ALL TESTS PASSED - The spin is provably fair!")
		fmt.Printf("   Winner: %s with %.2f%% chance\n",
			result.Winner.GetName(),
			float64(result.Winner.GetStake())/6.0*100)
	} else {
		fmt.Println("❌ TESTS FAILED - The spin is NOT provably fair!")
	}
}
