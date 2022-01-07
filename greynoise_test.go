package main

import (
	"context"
	"testing"
)

type GNoiseMock struct{}

func (f *GNoiseMock) ParseLogFiles(directory string, days int) (map[string]string, error) {
	return map[string]string{"1.1.1.1": "1.1.1.1"}, nil
}

func (f *GNoiseMock) IPLookup(ctx context.Context, ip string) (*GNoiseResponse, error) {
	return &GNoiseResponse{
		IP:             "1.1.1.1",
		Noise:          true,
		Riot:           false,
		Classification: "malicious",
		Name:           "Shit face",
		Link:           "dat link tho",
		LastSeen:       "never",
	}, nil
}

func TestParseLogFiles(t *testing.T) {

	gn := &GNoise{}

	// need to update creation date of test file everytime this code runs :(
	ips, err := gn.ParseLogFiles("./testdata", 30)
	if err != nil {
		t.Fatal(err)
	}

	l := len(ips)
	if l != 18 {
		t.Errorf("Amount of IPs is incorrect, expected %d, got %d", 18, l)
	}
}

func TestCheckNoise(t *testing.T) {
	ctx := context.Background()
	gn := &GNoiseMock{}

	result, err := CheckNoise(ctx, gn, "some dir", 1337)

	if err != nil {
		t.Fatal(err)
	}

	if result.AmountOfNoise != 1 {
		t.Errorf("expected amount of noise = %d, got %d", 1, result.AmountOfNoise)
	}
}
