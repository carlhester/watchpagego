package main

import (
	"bufio"
	"fmt"
	//"io"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
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

func getRespCodeAndSiteData(site string) (int, string) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	// Fetches a site and returns response code and data
	response, err := http.Get(site)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	httpResponseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	respCode := int(response.StatusCode)
	respData := string(httpResponseData)
	return respCode, respData
}

func getHashFromData(respData string) string {
	h := md5.New()
	h.Write([]byte(respData))
	return hex.EncodeToString(h.Sum(nil))
}

func main() {
	fileName := `list`

	for _, site := range linesInFile(fileName) {
		respCode, respData := getRespCodeAndSiteData(site)
		strCode := strconv.Itoa(respCode)
		hashedData := getHashFromData(respData)
		outputFile := strCode + "_" + hashedData

		fmt.Printf("%d_%s_%s_%s\n", respCode, strCode, hashedData, outputFile)

	}
}
