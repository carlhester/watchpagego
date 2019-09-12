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
	"strings"
	"sync"
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

func doTheWork(siteToCheck string, identifierToCheck string, resultsChannel chan string, wg *sync.WaitGroup) {
	// make sure the entry looks like a url
	err := validateSiteFormat(siteToCheck)
	checkError(err)

	// get a filename safe version of the site
	siteName, err := sanitizeSiteName(siteToCheck)
	strCode, respData, err := getRespCodeAndSiteData(siteToCheck)
	hashedData := getHashFromData(respData)
	outputFile := strCode + "_" + hashedData

	// determine output filename and path
	cwd, err := os.Getwd()
	checkError(err)
	targetDir := filepath.Join(cwd, "output", siteName)
	targetFile := filepath.Join(targetDir, outputFile)

	// if the output path doesn't exist, make the path
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		os.MkdirAll(targetDir, os.ModePerm)
	}

	// if the output file doesnt exist, create it and make noise
	if _, err := os.Stat(targetFile); os.IsNotExist(err) {
		fmt.Printf("Changed! : %s : %s\n", siteToCheck, targetFile)
		dataFile, err := os.Create(targetFile)
		checkError(err)

		// can i skip this from returning anything if no error?
		bytesWritten, err := dataFile.WriteString(respData)
		checkError(err)
		dataFile.Close()
		fmt.Println(bytesWritten)
		resultsChannel <- siteToCheck
	} else {
		fmt.Printf("No change: %s\n", siteToCheck)
	}

	//		fmt.Printf("%d\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n", respCode, strCode, hashedData, outputFile, cwd, targetDir, targetFile)
	wg.Done()
	return
}

func main() {
	fileName := `list`

	var wg sync.WaitGroup

	// buffered channel with 1000 slots.
	// not efficient for what we're doing, but unsure best way to find the size of our input list to best measure this
	// with this approach, if the channel is full, we're going to block and hang
	resultsChannel := make(chan string, 1000)
	for _, line := range linesInFile(fileName) {
		line := strings.Split(line, ",")
		siteToCheck := line[0]
		identifierToCheck := line[1]
		fmt.Printf("Checking : %s.  Identifier : %s.\n", siteToCheck, identifierToCheck)

		wg.Add(1)
		go doTheWork(siteToCheck, identifierToCheck, resultsChannel, &wg)
	}

	// anonymous function / closure that won't close the channel until Wait is resolved
	// wait will not get resolved until all of the wg.Done are finished
	// instead of making a buffered channel
	//go func() {
	//	wg.Wait()
	//	close(channel)
	//}()

	wg.Wait()
	close(resultsChannel)

	for item := range resultsChannel {
		fmt.Println(item)
	}

	// this is here so the program doesn't exit right away
	time.Sleep(1 * time.Second)
}
