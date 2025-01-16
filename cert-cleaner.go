package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	livePath                 = "/etc/letsencrypt/live"
	dulicateCertListFilePath = "/etc/letsencrypt/certswithduplicateentries.txt"
	perfectcertsFilePath     = "/etc/letsencrypt/perfectcerts.txt"
	fullchain                = "fullchain.pem"
	temp_fullchain           = "temp_fullchain.pem"
	old_fullchain            = "old_fullchain.pem"
)

func main() {

	if os.Args[1] == "list" {
		log.Printf("listing certificates with duplicte entries into %s and \n perfect certificates into %s", dulicateCertListFilePath, perfectcertsFilePath)
		listing()
	} else if os.Args[1] == "clean" {
		log.Printf("cleaning certificates from " + dulicateCertListFilePath)
		cleanup()
	} else if os.Args[1] == "renew" {
		renewCertificate()
	}
}

func listing() {
	log.Printf("reading %s", livePath)
	dirEntries, err := os.ReadDir(livePath)
	if err != nil {
		log.Fatalf("error while reading %s \\n %v", livePath, err.Error())
	}
	var certswithduplicateentries bytes.Buffer
	var perfectCerts bytes.Buffer
	for _, entry := range dirEntries {
		if entry.IsDir() {
			certDir := filepath.Join(livePath, entry.Name())
			certFilePath := filepath.Join(certDir, fullchain)
			_, err := os.Stat(certFilePath)
			if !os.IsNotExist(err) {
				flag := hasDuplicateEntry(certFilePath)
				if flag {
					log.Printf(certFilePath + " have duplicate entries")
					certswithduplicateentries.WriteString(entry.Name())
					certswithduplicateentries.WriteString("\n")
				} else {
					log.Printf(certFilePath + " doesn't have duplicate entries")
					perfectCerts.WriteString(entry.Name())
					perfectCerts.WriteString("\n")

				}

			}

		}
	}

	err = os.WriteFile(dulicateCertListFilePath, certswithduplicateentries.Bytes(), 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}

	fmt.Printf("certificates with duplicate entries have been written to %s\n", dulicateCertListFilePath)

	err = os.WriteFile(perfectcertsFilePath, perfectCerts.Bytes(), 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}

	fmt.Printf("perfect certificates have been written to %s\n", perfectcertsFilePath)
}

func cleanup() {

	// Open the file
	file, err := os.Open(dulicateCertListFilePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	// Create a new scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	// Loop through all the lines
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		log.Printf("cleaning " + trimmedLine)
		inputfile := filepath.Join(trimmedLine, fullchain)

		data, err := os.ReadFile(inputfile)
		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}

		certificates := bytes.Split(data, []byte("-----END CERTIFICATE-----"))
		uniqueCerts := make(map[string]bool)
		var cleanedCerts bytes.Buffer

		for _, cert := range certificates {
			if len(cert) > 0 && !isWhitespace(cert) {
				cert = append(cert, []byte("-----END CERTIFICATE-----")...)
				certStr := string(cert)
				if !uniqueCerts[certStr] {
					uniqueCerts[certStr] = true
					cleanedCerts.Write(cert)
					//cleanedCerts.WriteString("\n")
				}
			}
		}

		cleanedCerts.WriteString("\n")
		//outputFile := filepath.Join(trimmedLine, temp_fullchain)
		err = os.WriteFile(inputfile, cleanedCerts.Bytes(), 0o600)
		if err != nil {
			fmt.Println("Error writing file:", err)
			return
		}

		fmt.Printf("Cleaned certificates have been written to %s\n", inputfile)
	}

	// Check for any errors during scanning
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
	}

}
func renewCertificate() {
	// Open the file
	file, err := os.Open(dulicateCertListFilePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
	}
	defer file.Close()

	// Create a new scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	// Loop through all the lines
	for scanner.Scan() {
		line := scanner.Text()
		domain := strings.TrimSpace(line)
		log.Printf("sending request for " + domain)
		url := fmt.Sprintf("http://148.51.139.105:8888/api/certificate/renew?name=%s", domain)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(nil))
		if err != nil {
			log.Printf("failed to create request: %v", err)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("failed to send request: %v", err)
		}
		defer resp.Body.Close()
		log.Printf("status code: %d", resp.StatusCode)

	}
}

// isWhitespace checks if the given byte slice is only whitespace.
func isWhitespace(data []byte) bool {
	for _, b := range data {
		if b != ' ' && b != '\n' && b != '\t' {
			return false
		}
	}
	return true
}

// hasDuplicateEntry checks if there are duplicate certificate entries in the given PEM file.
func hasDuplicateEntry(pemFilePath string) bool {
	hasDuplicateEntry := false
	data, err := os.ReadFile(pemFilePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return false
	}

	certificates := bytes.Split(data, []byte("-----END CERTIFICATE-----"))
	uniqueCerts := make(map[string]bool)

	for _, cert := range certificates {
		if len(cert) > 0 && !isWhitespace(cert) {
			cert = append(cert, []byte("-----END CERTIFICATE-----")...)
			certStr := string(cert)
			if !uniqueCerts[certStr] {
				uniqueCerts[certStr] = true
			} else {
				hasDuplicateEntry = true
				break
			}
		}
	}

	return hasDuplicateEntry
}
