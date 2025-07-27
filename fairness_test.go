package provablyspin

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
)

func TestProvablyFairSpin(t *testing.T) {
	// Create test players
	players := []GamePlayer{
		{ID: "player1", Name: "Alice", Stake: 100},
		{ID: "player2", Name: "Bob", Stake: 200},
		{ID: "player3", Name: "Charlie", Stake: 300},
	}

	// Generate random seed
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		t.Fatalf("Failed to generate random seed: %v", err)
	}
	seed := hex.EncodeToString(randomBytes)

	// Calculate hashed seed
	h := sha256.Sum256([]byte(seed))
	hashedSeed := hex.EncodeToString(h[:])

	nonce := 1

	// Generate spin result
	result, err := Spin(seed, nonce, players)
	if err != nil {
		t.Fatalf("Failed to generate spin: %v", err)
	}

	// Test 1: Verify hash matches
	hashBytes := sha256.Sum256([]byte(seed))
	expectedHash := hex.EncodeToString(hashBytes[:])
	if expectedHash != hashedSeed {
		t.Errorf("Hash validation failed: expected %s, got %s", hashedSeed, expectedHash)
	}

	// Test 2: Full verification using VerifySpin
	isValid := VerifySpin(hashedSeed, seed, nonce, players, result)
	if !isValid {
		t.Error("VerifySpin validation failed")
	}

	// Test 3: ControlJSON validation
	var controlData ControlData[GamePlayer]
	if err := json.Unmarshal(result.ControlJSON, &controlData); err != nil {
		t.Fatalf("Failed to unmarshal ControlJSON: %v", err)
	}

	// Verify control data can reproduce the result
	controlResult, err := Spin(controlData.Seed, controlData.Nonce, controlData.Players)
	if err != nil {
		t.Fatalf("Failed to spin with control data: %v", err)
	}

	if controlResult.Winner.GetID() != result.Winner.GetID() {
		t.Errorf("ControlJSON winner mismatch: expected %s, got %s",
			result.Winner.GetID(), controlResult.Winner.GetID())
	}

	if controlResult.SpinFloat != result.SpinFloat {
		t.Errorf("ControlJSON spin value mismatch: expected %f, got %f",
			result.SpinFloat, controlResult.SpinFloat)
	}

	if controlResult.Nonce != result.Nonce {
		t.Errorf("ControlJSON nonce mismatch: expected %d, got %d",
			result.Nonce, controlResult.Nonce)
	}

	// Test 4: Verify ranges are correct
	var total int64
	for _, p := range players {
		total += p.GetStake()
	}

	var accum float64
	for _, p := range players {
		accum += float64(p.GetStake()) / float64(total)
		expectedRange := accum
		actualRange, ok := result.Ranges[p.GetID()]
		if !ok {
			t.Errorf("Missing range for player %s", p.GetID())
		}
		if actualRange != expectedRange {
			t.Errorf("Range mismatch for player %s: expected %f, got %f",
				p.GetID(), expectedRange, actualRange)
		}
	}

	t.Logf("✅ All tests passed!")
	t.Logf("Winner: %s (ID: %s)", result.Winner.GetName(), result.Winner.GetID())
	t.Logf("Spin Value: %.10f", result.SpinFloat)
	t.Logf("Hashed Seed: %s", hashedSeed[:20]+"...")
}

func TestMultipleSpinsWithSameSeed(t *testing.T) {
	players := []GamePlayer{
		{ID: "player1", Name: "Alice", Stake: 100},
		{ID: "player2", Name: "Bob", Stake: 200},
	}

	seed := "test-seed-123"
	nonce := 1

	// Generate same spin multiple times
	result1, err := Spin(seed, nonce, players)
	if err != nil {
		t.Fatalf("Failed to generate first spin: %v", err)
	}

	result2, err := Spin(seed, nonce, players)
	if err != nil {
		t.Fatalf("Failed to generate second spin: %v", err)
	}

	// Results should be identical
	if result1.Winner.GetID() != result2.Winner.GetID() {
		t.Error("Deterministic test failed: winners differ")
	}

	if result1.SpinFloat != result2.SpinFloat {
		t.Error("Deterministic test failed: spin values differ")
	}

	t.Logf("✅ Deterministic test passed - same seed produces same result")
}

func TestInvalidVerification(t *testing.T) {
	players := []GamePlayer{
		{ID: "player1", Name: "Alice", Stake: 100},
	}

	seed := "test-seed"
	wrongSeed := "wrong-seed"
	nonce := 1

	// Generate result with correct seed
	result, err := Spin(seed, nonce, players)
	if err != nil {
		t.Fatalf("Failed to generate spin: %v", err)
	}

	// Try to verify with wrong seed
	h := sha256.Sum256([]byte(wrongSeed))
	wrongHashedSeed := hex.EncodeToString(h[:])

	isValid := VerifySpin(wrongHashedSeed, wrongSeed, nonce, players, result)
	if isValid {
		t.Error("Verification should fail with wrong seed but it passed")
	}

	t.Logf("✅ Invalid verification test passed - wrong seed correctly rejected")
}
