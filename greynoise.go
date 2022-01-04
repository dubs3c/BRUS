package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"sync"
)

type GNoiseResponse struct {
	IP             string
	Noise          bool
	Riot           bool
	Classification string
	Name           string
	Link           string
	LastSeen       string
}

type GNoise struct {
	ApiKey string
	Http   *http.Transport
}

type Result struct {
	AmountOfNoise     int
	AmountOfNonNoise  int
	TopNoisyIP        []string
	TopClassification []string
	TopName           []string
}

func getTopValues(data map[string]int) []string {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool { return data[keys[i]] > data[keys[j]] })

	top := []string{}

	for k, _ := range data {
		top = append(top, k)
	}

	return top[:2]
}

func (gn *GNoise) gNoiseIpLookUp(ctx context.Context, ip string) (*GNoiseResponse, error) {
	apiPath := fmt.Sprintf("https://api.greynoise.io/v3/community/%s", ip)
	data := &GNoiseResponse{IP: ip}
	client := &http.Client{Transport: gn.Http}

	req, err := http.NewRequestWithContext(ctx, "GET", apiPath, nil)

	if err != nil {
		return &GNoiseResponse{}, err
	}

	req.Header.Add("key", gn.ApiKey)

	resp, err := client.Do(req)

	if err != nil {
		return &GNoiseResponse{}, err
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case 400:
		return &GNoiseResponse{}, errors.New("HTTP code 400 - Invalid Request")
	case 401:
		return &GNoiseResponse{}, errors.New("HTTP code 401 - Authentication Error")
	case 429:
		return &GNoiseResponse{}, errors.New("HTTP code 429 - Daily Rate-Limit Exceeded")
	case 500:
		return &GNoiseResponse{}, errors.New("HTTP code 500 - Internal Error")
	}

	body, err := io.ReadAll(resp.Body)

	if err := json.Unmarshal(body, data); err != nil {
		return &GNoiseResponse{}, err
	}

	return data, err
}

func (gn *GNoise) CheckNoise(ctx context.Context, directory string, days int) (*Result, error) {
	ips, err := ParseLogFiles(directory, days)

	if err != nil {
		log.Fatal(err)
	}

	wg := &sync.WaitGroup{}
	workers := 5
	ipChan := make(chan string, 1)
	done := make(chan bool, 1)
	output := make(chan *GNoiseResponse, 1)

	go func() {
		for _, ip := range ips {
			ipChan <- ip
		}
		close(ipChan)
	}()

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ip := range ipChan {
				data, _ := gn.gNoiseIpLookUp(ctx, ip)
				output <- data
			}
			done <- true
		}()
	}

	workerDone := 0
	responses := []*GNoiseResponse{}
	for {
		select {
		case r := <-output:
			responses = append(responses, r)
		case <-done:
			workerDone++
		default:
		}

		if workerDone == workers {
			break
		}
	}

	wg.Wait()

	// Perform calculation
	noise := 0
	topName := map[string]int{}
	topClassification := map[string]int{}

	// listan är ju unik ya fool........
	for _, data := range responses {
		if data.Noise {
			noise++
			topClassification[data.Classification] += 1
			topName[data.Name] += 1
		}
	}

	result := &Result{
		AmountOfNoise:     noise,
		AmountOfNonNoise:  len(responses) - noise,
		TopClassification: getTopValues(topClassification),
		TopName:           getTopValues(topName),
	}

	return result, nil
}
