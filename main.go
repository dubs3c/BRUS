package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/ini.v1"
)

func main() {

	var (
		email     bool
		webhook   bool
		quite     bool
		days      int
		directory string
		configLoc string
		conf      string
	)

	flag.BoolVar(&email, "smtp", false, "Send result using smtp")
	flag.BoolVar(&webhook, "webhook", false, "Send result using a webhook")
	flag.BoolVar(&quite, "quite", false, "Suppress output")
	flag.IntVar(&days, "days", 30, "Parse log files within the last X days")
	flag.StringVar(&directory, "directory", "", "Required. Directory that contains log files to be parsed. Must be absolute path")
	flag.StringVar(&configLoc, "config", "", "Specify an alternative config location")

	flag.Parse()

	if len(os.Args) <= 1 {
		flag.Usage()
		os.Exit(0)
	}

	if directory == "" {
		log.Fatal("Directory must be present")
		os.Exit(1)
	}

	if !path.IsAbs(directory) {
		log.Fatal("Directory must be absolute path")
		os.Exit(1)
	}

	if !webhook && !email {
		log.Fatal("You need to specify either -webhook or -email")
		os.Exit(1)
	}

	var cfg *ini.File
	User, err := user.Current()

	if err != nil {
		log.Fatal("Something went wrong trying to figure out your home directory", err)
	}

	if configLoc != "" {
		conf = configLoc
	} else {
		conf = User.HomeDir + "/.config/brus.ini"
	}

	configPath := filepath.FromSlash(conf)
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

	result, err := gn.CheckNoise(ctx, directory, days)

	if err != nil {
		log.Fatal("GreyNoise failed: ", err)
	}

	message := fmt.Sprintf(`
# Results from BRUS the last 30 days
- Amount of Noisy IPs: %d
- Non Noisy IPs: %d
- Top 3 Classification: %s
- Top 3 Names: %s`, result.AmountOfNoise, result.AmountOfNonNoise,
		strings.Join(result.TopClassification, ", "), strings.Join(result.TopName, ", "))

	if webhook {
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

		if !quite {
			fmt.Println("🚀 Data sent to webhook")
		}
	}

	if email {
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
		if !quite {
			fmt.Println("📧 Data sent via email")
		}
	}

	if !quite {
		fmt.Println(message)
	}
}
