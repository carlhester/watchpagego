package main

import (
	"bufio"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

func linesInFile(fileName string) []string {
	// Reads and returns lines from fileName
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	result := []string{}

	for scanner.Scan() {
		line := scanner.Text()
		result = append(result, line)
	}
	return result
}

func getRespCodeAndSiteData(site string) (string, string, error) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	// Fetches a site and returns response code and data as strings
	response, err := http.Get(site)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	defer response.Body.Close()

	httpResponseData, err := ioutil.ReadAll(response.Body)
	checkError(err)
	respCode := int(response.StatusCode)
	strCode := strconv.Itoa(respCode)
	respData := string(httpResponseData)
	return strCode, respData, nil
}

func getHashFromData(text string) string {
	// take text return hash
	h := md5.New()
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil))
}

func checkError(err error) {
	// no idea if this works the way i think
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}

func validateSiteFormat(site string) error {
	// confirms we're dealing with a real URL and dies if not
	_, err := url.ParseRequestURI(site)
	if err != nil {
		fmt.Printf("\n\nBailing out: %s is not formatted correctly\n\n", site)
		panic(err)
	}
	return nil
}

func sanitizeSiteName(site string) (string, error) {
	// get rid of extraneous characters
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	checkError(err)
	siteName := reg.ReplaceAllString(site, "")
	return siteName, nil
}

func doTheWork(site string) {
	err := validateSiteFormat(site)
	checkError(err)

	siteName, err := sanitizeSiteName(site)
	strCode, respData, err := getRespCodeAndSiteData(site)
	hashedData := getHashFromData(respData)
	outputFile := strCode + "_" + hashedData

	cwd, err := os.Getwd()
	checkError(err)

	targetDir := filepath.Join(cwd, "output", siteName)
	targetFile := filepath.Join(targetDir, outputFile)

	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		os.MkdirAll(targetDir, os.ModePerm)
	}

	if _, err := os.Stat(targetFile); os.IsNotExist(err) {
		fmt.Printf("File does not exist; creating: %s\n", targetFile)
		dataFile, err := os.Create(targetFile)
		checkError(err)
		defer dataFile.Close()

		bytesWritten, err := dataFile.WriteString(respData)
		fmt.Println(bytesWritten)
		checkError(err)
	} else {
		fmt.Printf("File DOES exist: %s\n", targetFile)
	}

	//		fmt.Printf("%d\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n", respCode, strCode, hashedData, outputFile, cwd, targetDir, targetFile)
}

func main() {
	fileName := `list`
	for _, site := range linesInFile(fileName) {
		go doTheWork(site)
	}
	// this is here so the program doesn't exit right away
	time.Sleep(1 * time.Second)
}
