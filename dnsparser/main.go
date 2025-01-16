package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type StorageDNS struct {
	DNS               string `json:"dns"`
	CloudflareID      string `json:"cloudflare_id"`
	CloudflareID2     string `json:"cloudflare_id2"`
	Used              int    `json:"used"`
	AddedOn           string `json:"added_on"`
	UpdatedOn         string `json:"updated_on"`
	Comments          string `json:"comments"`
	Domain            string `json:"domain"`
	Capacity          int    `json:"capacity"`
	IsNodeActive      int    `json:"is_node_active"`
	PublicIP          string `json:"public_ip"`
	PrivateIP         string `json:"private_ip"`
	DomainName        string `json:"domain_name"`
	Identifier        string `json:"identifier"`
	LastPingTime      string `json:"last_ping_time"`
	TotalDiskSpace    int64  `json:"total_disk_space"`
	DiskSpaceUsed     int64  `json:"disk_space_used"`
	StorageNodeStatus int    `json:"storage_node_status"`
	ZoneName          string `json:"zone_name"`
	Region            string `json:"region"`
}

type StorageDNSFile struct {
	StorageDNS []StorageDNS `json:"storage_dns"`
}

func main() {
	// Open and read the file
	dnsdir := "/home/dnslist"
	livePath := "/etc/letsencrypt/live"
	live_filtered_Path := "/etc/letsencrypt/live_filtered"
	// Create a map with CNAME as keys and list of DNS as values
	cnameMap := make(map[string][]string)

	dirEntries, err := os.ReadDir(dnsdir)
	if err != nil {
		log.Fatalf("Failed to read file: %s", err)
	}
	for _, entry := range dirEntries {
		dnsFilePath := filepath.Join(dnsdir, entry.Name())

		file, err := os.Open(dnsFilePath)
		if err != nil {
			log.Fatalf("Failed to open file: %s", err)
		}
		defer file.Close()

		byteValue, err := io.ReadAll(file)
		if err != nil {
			log.Fatalf("Failed to read file: %s", err)
		}

		// Parse the JSON content
		var storageDNSFile StorageDNSFile
		if err := json.Unmarshal(byteValue, &storageDNSFile); err != nil {
			log.Fatalf("Failed to unmarshal JSON: %s", err)
		}

		for _, dnsEntry := range storageDNSFile.StorageDNS {
			parts := strings.Split(dnsEntry.DNS, ".")
			if len(parts) > 2 {
				cname := strings.Join(parts[1:], ".")
				cnameMap[cname] = append(cnameMap[cname], dnsEntry.DNS)
			}
		}
	}
	// Write the map to a file
	outputFile, err := os.Create("/home/cnamelist.txt")
	if err != nil {
		log.Fatalf("Failed to create output file: %s", err)
	}
	defer outputFile.Close()
	// Print the map
	for cname, dnsList := range cnameMap {
		fmt.Printf("CNAME: %s, DNS List: %v\n", cname, dnsList)
		inputLiveFile := filepath.Join(livePath, cname)
		outputLiveFile := filepath.Join(live_filtered_Path, cname)
		copyDir(inputLiveFile, outputLiveFile)
		if _, err := outputFile.WriteString(cname + "\n"); err != nil {
			log.Fatalf("Failed to write to output file: %s", err)
		}
	}
}

func copyDir(src string, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, os.ModePerm); err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	return destinationFile.Sync()
}
