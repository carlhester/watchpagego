package main

import (
	"bufio"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	//"io/ioutil"
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
	"watchpagego/utils"
)

const userAgent = "watchpagego"

func main() {
	fileName := `list`

	var wg sync.WaitGroup

	// buffered channel with 1000 slots.
	// not efficient for what we're doing, but unsure best way to find the size of
	// our input list to best measure this with this approach, if the channel is
	// full, we're going to block and hang

	resultsChannel := make(chan [2]string, 1000)
	for _, line := range linesInFile(fileName) {
		line := strings.Split(line, ",")
		siteToCheck := line[0]
		identifierToCheck := line[1]
		//		fmt.Printf("Checking : %s  Identifier : %s\n", siteToCheck, identifierToCheck)

		wg.Add(1)
		go doTheWork(siteToCheck, identifierToCheck, resultsChannel, &wg)
	}

	// anonymous function / closure that won't close the channel until Wait is
	// resolved  wait will not get resolved until all of the wg.Done are finished
	// instead of making a buffered channel
	//go func() {
	//	wg.Wait()
	//	close(channel)
	//}()

	wg.Wait()
	close(resultsChannel)

	err := sendIfNewResults(resultsChannel)
	if err != nil {
		fmt.Println(err)
	}

	// this is here so the program doesn't exit right away
	time.Sleep(1 * time.Second)
}

func sendIfNewResults(resultsChannel chan [2]string) error {
	var resultsToSend string
	var numNewResults int
	for item := range resultsChannel {
		numNewResults += 1
		fmt.Println(item)
		resultsToSend += item[0] + "\t"
		resultsToSend += item[1] + "\r\n"
	}

	if numNewResults != 0 {
		err := utils.EmailResults(resultsToSend)
		if err != nil {
			return err
		}

	}
	return nil
}

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

func getRespCodeAndSiteData(site string, identifier string) (string, string, error) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	httpClient := &http.Client{}

	// Fetches a site and returns response code and data as strings
	request, err := http.NewRequest("GET", site, nil)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	request.Header.Set("User-Agent", userAgent)
	response, err := httpClient.Do(request)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	defer response.Body.Close()

	//httpResponseData, err := ioutil.ReadAll(response.Body)
	httpResponseData, err := goquery.NewDocumentFromReader(response.Body)
	checkError(err)

	extractedResponseData, err := httpResponseData.Find(identifier).Html()
	checkError(err)

	respCode := int(response.StatusCode)
	strCode := strconv.Itoa(respCode)
	respData := string(extractedResponseData)
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

func getOutputTargetDirAndFile(siteName string, outputFile string) (string, string, error) {
	cwd, err := os.Getwd()
	checkError(err)
	targetDir := filepath.Join(cwd, "output", siteName)
	targetFile := filepath.Join(targetDir, outputFile)
	return targetDir, targetFile, nil
}

func doTheWork(siteToCheck string, identifierToCheck string, resultsChannel chan [2]string, wg *sync.WaitGroup) {
	// make sure the entry looks like a url
	err := validateSiteFormat(siteToCheck)
	checkError(err)

	// get the data from the site and set a filesystem-safe name
	siteName, err := sanitizeSiteName(siteToCheck)
	strCode, respData, err := getRespCodeAndSiteData(siteToCheck, identifierToCheck)
	hashedData := getHashFromData(respData)
	outputFile := strCode + "_" + hashedData

	// determine output target filename and path
	targetDir, targetFile, err := getOutputTargetDirAndFile(siteName, outputFile)
	checkError(err)

	// if the output path doesn't exist, make the path
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		os.MkdirAll(targetDir, os.ModePerm)
	}

	// if the output file doesnt exist, create it and make noise
	if _, err := os.Stat(targetFile); os.IsNotExist(err) {
		//fmt.Printf("Changed! : %s : %s\n", siteToCheck, targetFile)
		dataFile, err := os.Create(targetFile)
		checkError(err)

		// can i skip this from returning anything if no error?
		bytesWritten, err := dataFile.WriteString(respData)
		checkError(err)
		dataFile.Close()

		if bytesWritten == 0 {
			fmt.Println(bytesWritten)
		}
		results := [2]string{siteToCheck, targetFile}

		resultsChannel <- results
	}

	wg.Done()
	return
}
