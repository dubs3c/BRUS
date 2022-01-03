package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/ini.v1"
)

func ParseFile(filename string) (map[string]string, error) {
	file, err := os.Open(filename)

	if err != nil {
		return map[string]string{}, err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	if err := scanner.Err(); err != nil {
		return map[string]string{}, err
	}

	var (
		ips map[string]string = map[string]string{}
		ip  []byte            = []byte{}
	)

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

	return ips, nil

}

func PreparePayload(message string, msgField string, additionalData string) ([]byte, error) {

	jayson := map[string]interface{}{
		msgField: message,
	}
	// Required for valid json
	additionalData = strings.ReplaceAll(additionalData, "'", "\"")
	if additionalData != "" {
		data := []byte(`` + additionalData + ``)
		var f interface{}
		if err := json.Unmarshal(data, &f); err != nil {
			return nil, err
		}
		m := f.(map[string]interface{})
		for k, v := range m {
			jayson[k] = v
		}
	}
	js, err := json.Marshal(jayson)
	if err != nil {
		return []byte{}, err
	}
	return js, nil
}

func main() {
	email := flag.Bool("smtp", false, "Send result using smtp")
	webhook := flag.Bool("webhook", false, "Send result using a webhook")

	flag.Parse()

	if len(os.Args) <= 1 {
		flag.Usage()
		os.Exit(0)
	}

	var cfg *ini.File
	User, err := user.Current()

	if err != nil {
		log.Fatal("Something went wrong trying to figure out your home directory", err)
	}

	configPath := filepath.FromSlash(User.HomeDir + "/.config/brus.ini")
	cfg, err = ini.Load(configPath)

	if err != nil {
		log.Fatal("Fail to read configuration file: ", err)
	}

	key := cfg.Section("GreyNoise").Key("key").String()

	if key == "" {
		log.Fatal("No API key for GreyNoise present")
	}

	gn := &GNoise{
		ApiKey: key,
		Http: &http.Transport{
			MaxIdleConns:    10,
			IdleConnTimeout: 30 * time.Second, // hmm do I need TimeoutContext?
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	result, err := gn.CheckNoise(ctx, "input.txt") // ska vara en katalog, inte en fil

	if err != nil {
		log.Fatal("GreyNoise failed: ", err)
	}

	message := fmt.Sprintf("*Results from BRUS the last 30 days*\nAmount of Noisy IPs: %d\nNon Noisy IPs: %d\nTop 3 Classification: %s\nTop 3 Names: %s\n", result.AmountOfNoise, result.AmountOfNonNoise, result.TopClassification, result.TopName)

	if *webhook {
		webhook := cfg.Section("Webhook").Key("webhook").String()
		textField := cfg.Section("Webhook").Key("textField").MustString("text")
		additionalData := cfg.Section("Webhook").Key("data").String()

		// MS Teams hack for properly showing rows
		if strings.HasPrefix(webhook, "https://outlook.office.com") {
			split := strings.Split(message, "\n")
			newMessage := ""
			for _, v := range split {
				newMessage += v + "\n\n"
			}
			message = newMessage
		}

		json, err := PreparePayload(message, textField, additionalData)

		if err != nil {
			log.Fatal("Could not prepare payload for webhook")
		}

		err = SendRequest(webhook, json)

		if err != nil {
			log.Fatal("Could not send data to webhook", err)
		}

		fmt.Println("Data sent to webhook")
	}

	if *email {
		emailUsername := cfg.Section("Email").Key("username").String()
		emailPassword := cfg.Section("Email").Key("password").String()
		emailRecipient := cfg.Section("Email").Key("recipient").String()
		emailPort := cfg.Section("Email").Key("port").String()
		emailServer := cfg.Section("Email").Key("server").String()
		emailSubject := cfg.Section("Email").Key("subject").String()

		emailConfig := EmailConfig{username: emailUsername, password: emailPassword,
			recipient: emailRecipient, port: emailPort, server: emailServer, subject: emailSubject,
			message: message}

		if SendEmail(emailConfig) != nil {
			log.Fatal("Could not email results", err)
		}

		fmt.Println("Data sent via email")
	}

}
