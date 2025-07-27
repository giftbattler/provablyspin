package provablyspin

import (
	"crypto/sha256"
	"encoding/hex"
)

func VerifySpin[T Player](hashedSeed, revealedSeed string, nonce int, players []T, result Result[T]) bool {
	// Verify that revealed seed matches the hashed seed
	h := sha256.Sum256([]byte(revealedSeed))
	expectedHash := hex.EncodeToString(h[:])

	if expectedHash != hashedSeed {
		return false
	}

	// Recalculate the spin with revealed seed
	expectedResult, err := Spin(revealedSeed, nonce, players)
	if err != nil {
		return false
	}

	// Verify the results match
	return expectedResult.Winner.GetID() == result.Winner.GetID() &&
		expectedResult.SpinFloat == result.SpinFloat &&
		expectedResult.Nonce == result.Nonce
}
