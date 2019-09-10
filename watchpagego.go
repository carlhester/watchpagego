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
	"path/filepath"
	"regexp"
	"strconv"
	//"strings"
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

func getRespCodeAndSiteData(site string) (string, string) {
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
	strCode := strconv.Itoa(respCode)
	respData := string(httpResponseData)
	return strCode, respData
}

func getHashFromData(text string) string {
	// take text get hash
	h := md5.New()
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil))
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}

}

func main() {
	fileName := `list`

	for _, site := range linesInFile(fileName) {
		strCode, respData := getRespCodeAndSiteData(site)
		hashedData := getHashFromData(respData)
		outputFile := strCode + "_" + hashedData

		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		reg, err := regexp.Compile("[^a-zA-Z0-9]+")
		if err != nil {
			log.Fatal(err)
		}
		siteName := reg.ReplaceAllString(site, "")

		targetDir := filepath.Join(cwd, "output", siteName)
		targetFilePath := filepath.Join(targetDir, outputFile)

		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			os.MkdirAll(targetDir, os.ModePerm)
		}

		if _, err := os.Stat(targetFilePath); os.IsNotExist(err) {
			dataFile, err := os.Create(targetFilePath)
			if err != nil {
				log.Fatal(err)
				return
			}
			defer dataFile.Close()

			l, err := dataFile.WriteString(respData)
			fmt.Println(l)
			if err != nil {
				log.Fatal(err)
				return
			}

		}

		//		fmt.Printf("%d\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n", respCode, strCode, hashedData, outputFile, cwd, targetDir, targetFilePath)

	}
}
