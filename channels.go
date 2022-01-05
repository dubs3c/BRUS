package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"strings"
	"time"
)

// EmailConfig contains email data
type EmailConfig struct {
	username  string
	password  string
	server    string
	port      string
	recipient string
	subject   string
	message   string
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
			return []byte{}, err
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

// Email Send messages via email
func SendEmail(email EmailConfig) error {
	// Set up authentication information.
	auth := smtp.PlainAuth("", email.username, email.password, email.server)

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	to := []string{email.recipient}
	msg := []byte("To: " + email.recipient + "\r\n" +
		"Subject: " + email.subject + "\r\n" +
		"\r\n" +
		email.message + "\r\n")
	err := smtp.SendMail(email.server+":"+email.port, auth, email.username, to, msg)
	if err != nil {
		return err
	}

	return nil
}

// SendRequest Send the request to the webhook
func SendRequest(endpoint string, data []byte) error {

	tr := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}

	var resp *http.Response
	var err error

	client := &http.Client{Transport: tr}

	resp, err = client.Post(endpoint, "application/json", bytes.NewBuffer(data))

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, respErr := ioutil.ReadAll(resp.Body)
		if respErr != nil {
			return respErr
		}
		errorMessage := fmt.Sprintf("Response HTTP Status code: %d\nResponse HTTP Body: %s", resp.StatusCode, string(body))
		return errors.New(errorMessage)
	}

	return nil
}
