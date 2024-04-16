package config

import (
	"encoding/json"
	"os"
)

// Helper functions to retrieve access tokens from a central location
// This should be replaced with a more secure solution.
// Also, it deliberately ignores error-handling, since this is temporary code.

func GetAviationStackCredentials() (apiKey string) {

	data, _ := os.ReadFile("./config/demo-tokens.json")

	type aviationStackCredentials struct {
		AviationStackApiKey string
	}

	var cred aviationStackCredentials
	_ = json.Unmarshal(data, &cred)

	return cred.AviationStackApiKey
}

func GetAmadeusCredentials() (clientId string, clientSecret string) {

	data, _ := os.ReadFile("./config/demo-tokens.json")

	type amadeusCredentials struct {
		AmadeusClientId     string
		AmadeusClientSecret string
	}

	var cred amadeusCredentials
	_ = json.Unmarshal(data, &cred)

	return cred.AmadeusClientId, cred.AmadeusClientSecret
}
