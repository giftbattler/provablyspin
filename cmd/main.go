package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/giftbattler/provablyspin"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type CLIInput struct {
	HashedSeed   string                                       `json:"hashed_seed"`
	RevealedSeed string                                       `json:"revealed_seed"`
	Nonce        int                                          `json:"nonce"`
	Players      []provablyspin.GamePlayer                    `json:"players"`
	Result       provablyspin.Result[provablyspin.GamePlayer] `json:"result"`
}

type SpinInput struct {
	Players []provablyspin.GamePlayer `json:"players"`
	Seed    string                    `json:"seed,omitempty"`
	Nonce   int                       `json:"nonce,omitempty"`
}

var rootCmd = &cobra.Command{
	Use:   "provablyspin",
	Short: "Provably fair spin verification tool",
	Long:  `A CLI tool to verify provably fair spins using HMAC-SHA256`,
}

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify a provably fair spin",
	Long:  `Verify that a spin result is valid by checking the revealed seed against the hashed seed and recalculating the spin`,
	Run:   verifySpin,
}

var spinCmd = &cobra.Command{
	Use:   "spin",
	Short: "Generate a provably fair spin result",
	Long:  `Generate a spin result from player JSON input and save raw data and results`,
	Run:   generateSpin,
}

func init() {
	verifyCmd.Flags().StringP("input", "i", "", "JSON input file path")
	verifyCmd.Flags().StringP("json", "j", "", "JSON input string")
	verifyCmd.MarkFlagRequired("input")

	spinCmd.Flags().StringP("input", "i", "", "JSON players file path")
	spinCmd.Flags().StringP("json", "j", "", "JSON players string")
	spinCmd.Flags().StringP("seed", "s", "", "Custom seed (optional)")
	spinCmd.Flags().IntP("nonce", "n", 1, "Nonce value")
	spinCmd.MarkFlagRequired("input")

	rootCmd.AddCommand(verifyCmd)
	rootCmd.AddCommand(spinCmd)
}

func verifySpin(cmd *cobra.Command, _ []string) {
	var input CLIInput

	jsonStr, _ := cmd.Flags().GetString("json")
	inputFile, _ := cmd.Flags().GetString("input")

	if jsonStr != "" {
		if err := json.Unmarshal([]byte(jsonStr), &input); err != nil {
			log.Fatalf("Error parsing JSON string: %v", err)
		}
	} else if inputFile != "" {
		data, err := os.ReadFile(inputFile)
		if err != nil {
			log.Fatalf("Error reading input file: %v", err)
		}

		if err := json.Unmarshal(data, &input); err != nil {
			log.Fatalf("Error parsing JSON file: %v", err)
		}
	} else {
		log.Fatal("Either --input or --json flag must be provided")
	}

	isValid := provablyspin.VerifySpin(
		input.HashedSeed,
		input.RevealedSeed,
		input.Nonce,
		input.Players,
		input.Result,
	)

	if isValid {
		fmt.Println("✅ Spin verification PASSED - The result is provably fair!")
	} else {
		fmt.Println("❌ Spin verification FAILED - The result is NOT valid!")
	}

	// Show details
	fmt.Printf("\nDetails:\n")
	fmt.Printf("Hashed Seed: %s\n", input.HashedSeed)
	fmt.Printf("Revealed Seed: %s\n", input.RevealedSeed)
	fmt.Printf("Nonce: %d\n", input.Nonce)
	fmt.Printf("Winner: %s (ID: %s)\n", input.Result.Winner.GetName(), input.Result.Winner.GetID())
	fmt.Printf("Spin Value: %.10f\n", input.Result.SpinFloat)
}

func generateSpin(cmd *cobra.Command, args []string) {
	var input SpinInput

	jsonStr, _ := cmd.Flags().GetString("json")
	inputFile, _ := cmd.Flags().GetString("input")
	customSeed, _ := cmd.Flags().GetString("seed")
	nonce, _ := cmd.Flags().GetInt("nonce")

	if jsonStr != "" {
		if err := json.Unmarshal([]byte(jsonStr), &input); err != nil {
			log.Fatalf("Error parsing JSON string: %v", err)
		}
	} else if inputFile != "" {
		data, err := os.ReadFile(inputFile)
		if err != nil {
			log.Fatalf("Error reading input file: %v", err)
		}

		if err := json.Unmarshal(data, &input); err != nil {
			log.Fatalf("Error parsing JSON file: %v", err)
		}
	} else {
		log.Fatal("Either --input or --json flag must be provided")
	}

	// Generate random seed if not provided
	var seed string
	if customSeed != "" {
		seed = customSeed
	} else {
		randomBytes := make([]byte, 32)
		if _, err := rand.Read(randomBytes); err != nil {
			log.Fatalf("Error generating random seed: %v", err)
		}
		seed = hex.EncodeToString(randomBytes)
	}

	// Calculate hashed seed
	h := sha256.Sum256([]byte(seed))
	hashedSeed := hex.EncodeToString(h[:])

	// Generate spin result
	result, err := provablyspin.Spin(seed, nonce, input.Players)
	if err != nil {
		log.Fatalf("Error generating spin: %v", err)
	}

	// Create directories
	if err := os.MkdirAll("spins", 0755); err != nil {
		log.Fatalf("Error creating spins directory: %v", err)
	}
	if err := os.MkdirAll("spins/results", 0755); err != nil {
		log.Fatalf("Error creating results directory: %v", err)
	}

	// Generate filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("spin_%s", timestamp)

	// Save raw data (includes revealed seed)
	rawData := CLIInput{
		HashedSeed:   hashedSeed,
		RevealedSeed: seed,
		Nonce:        nonce,
		Players:      input.Players,
		Result:       result,
	}

	rawJSON, err := json.MarshalIndent(rawData, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling raw data: %v", err)
	}

	rawPath := filepath.Join("spins", filename+"_raw.json")
	if err := os.WriteFile(rawPath, rawJSON, 0644); err != nil {
		log.Fatalf("Error writing raw file: %v", err)
	}

	// Save results (all data needed for verification)
	resultData := CLIInput{
		HashedSeed:   hashedSeed,
		RevealedSeed: seed,
		Nonce:        nonce,
		Players:      input.Players,
		Result:       result,
	}

	resultJSON, err := json.MarshalIndent(resultData, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling result data: %v", err)
	}

	resultPath := filepath.Join("spins", "results", filename+"_result.json")
	if err := os.WriteFile(resultPath, resultJSON, 0644); err != nil {
		log.Fatalf("Error writing result file: %v", err)
	}

	// Print summary
	fmt.Printf("🎰 Spin generated successfully!\n\n")
	fmt.Printf("Hashed Seed: %s\n", hashedSeed)
	fmt.Printf("Nonce: %d\n", nonce)
	fmt.Printf("Winner: %s (ID: %s)\n", result.Winner.GetName(), result.Winner.GetID())
	fmt.Printf("Spin Value: %.10f\n", result.SpinFloat)
	fmt.Printf("\nFiles saved:\n")
	fmt.Printf("  Raw data: %s\n", rawPath)
	fmt.Printf("  Results:  %s\n", resultPath)
}

func main() {
	viper.AutomaticEnv()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
