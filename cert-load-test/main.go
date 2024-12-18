package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
)

type Request struct {
	SCN string   `json:"scn"`
	SAN []string `json:"san"`
}

// generateRandomString creates a random string of lowercase letters with the given length.
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

var logger *log.Logger

type Config struct {
	Loads []string `json:"loads"`
	URI   string   `json:"uri"`
}

func main() {

	// Base domain
	//baseDomain := "load.edgedrive.com"

	// homeDir, _ := os.UserHomeDir()
	// logfilePath := filepath.Join(homeDir, "cert-load-test.log")
	logfilePath := "/var/log/letsencrypt/cert-load-test.log"
	logger = initLogger(logfilePath)
	domain := "edgedrive.com"
	baseDomains := make([]string, 0, 1000) //"ld01","ld02","ld03", "ld04","ld05", "ld06","ld07","ld08","ld09","ld10","ld11", "ld12", "ld12"
	// Read the JSON string from the file and deserialize it back to a struct

	filename := "cert-load-config.json"

	config := *getConfig(filename)
	//loads := []string{"ld13"}
	for _, load := range config.Loads {
		logger.Println("Loading " + load)
		for i := 0; i < 10000; i++ {
			hexa := strconv.FormatInt(int64(i), 16)
			//load := fmt.Sprintf("%s%s", "load", strconv.Itoa(i))
			load := fmt.Sprintf("%s%s", load, hexa)
			baseDomain := fmt.Sprintf("%s.%s", load, domain)

			baseDomains = append(baseDomains, baseDomain)
		}

		for _, baseDomain := range baseDomains {
			// Create the slice for subdomains
			subDomains := make([]string, 0, 49)

			// Generate 49 random subdomains
			for i := 0; i < 49; i++ {
				subDomain := generateRandomString(4)
				subDomains = append(subDomains, subDomain)
			}

			// Initialize the slice of full domain strings
			domains := make([]string, 0, 50)

			// Add the base domain first
			//domains = append(domains, baseDomain)

			// Add the subdomain variations
			for _, subDomain := range subDomains {
				fullDomain := fmt.Sprintf("%s.%s", subDomain, baseDomain)
				domains = append(domains, fullDomain)
			}

			// Print the list of domains
			for _, domain := range domains {
				logger.Println(domain)
			}

			// request := &Request{
			// 	SCN: baseDomain,
			// 	SAN: domains,
			// }
			logger.Printf("firing request for %s", baseDomain)
			//fireRequest(*request, config.URI)
			//time.Sleep(time.Second * 2)
			logger.Println("waiting for 5 sec to fire next request")

		}

	}

}
func fireRequest(request Request, uri string) {
	payloadJson, err := json.Marshal(request)
	if err != nil {
		logger.Print(err)
		return
	}
	url := uri + "/api/certificate/create"
	method := "POST"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewReader(payloadJson))

	if err != nil {
		logger.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		logger.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Println(err)
		return
	}
	logger.Println(string(body))
}
func initLogger(logfilePath string) *log.Logger {
	file, err := os.OpenFile(logfilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	multiWriter := io.MultiWriter(file, os.Stdout)
	logger := log.New(multiWriter, "info", log.LstdFlags)
	return logger
}

func getConfig(filename string) *Config {
	var data *Config
	var jsonData []byte
	var err error
	defaultData := Config{
		Loads: []string{"ld01", "ld02", "ld03", "ld04", "ld05", "ld06", "ld07", "ld08", "ld09", "ld10", "ld11", "ld12"},
		URI:   "http://148.51.139.41:8888",
	}

	// Check if the file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// File does not exist, create it with default values
		jsonData, err = json.MarshalIndent(defaultData, "", "    ")
		if err != nil {
			fmt.Println("Error serializing data:", err)
			return nil
		}

		err = os.WriteFile(filename, jsonData, 0644)
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return nil
		}

		fmt.Println("File not found. Created cert-load-config.json with default values.")
	} else {
		// File exists, read and deserialize it
		jsonData, err = os.ReadFile(filename)
		if err != nil {
			fmt.Println("Error reading file:", err)
			return nil
		}
	}
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		fmt.Println("Error deserializing data:", err)
		return nil
	}
	return data
}
