package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"sync"
	"time"
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

	for k := range data {
		top = append(top, k)
	}

	if len(top) == 0 {
		return []string{}
	}

	return top[:2]
}

// ParseLogFiles Parses log files within a given directory that are not older than days
func ParseLogFiles(directory string, days int) (map[string]string, error) {

	files, err := ioutil.ReadDir(directory)

	if err != nil {
		log.Fatal(err)
	}

	var (
		ips    map[string]string = map[string]string{}
		ip     []byte            = []byte{}
		before time.Time         = time.Now().AddDate(0, 0, -days)
	)

	// Goroutines can be used to read multiple files, if needed
	for _, file := range files {
		t := GetCreationDate(file)
		if t.After(before) {
			filePath := path.Join(directory, file.Name())
			logFile, err := os.Open(filePath)

			if err != nil {
				log.Println(err)
				continue
			}

			scanner := bufio.NewScanner(logFile)

			if err := scanner.Err(); err != nil {
				return map[string]string{}, err
			}

			for scanner.Scan() {
				ip = []byte{}
				for _, b := range scanner.Bytes() {
					if b == 32 {
						break
					}
					ip = append(ip, b)
				}
				ips[string(ip)] = string(ip)
			}
		}
	}

	return ips, nil
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

	if len(ips) == 0 {
		return &Result{}, errors.New("no IPs parsed")
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
