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
	"strconv"
	"strings"
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

func getHashFromData(text string) string {
	// take text get hash
	h := md5.New()
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil))
}

func main() {
	fileName := `list`

	for _, site := range linesInFile(fileName) {
		respCode, respData := getRespCodeAndSiteData(site)
		strCode := strconv.Itoa(respCode)
		hashedData := getHashFromData(respData)
		outputFile := strCode + "_" + hashedData

		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		tmpsiteName := strings.Replace(site, ":", "", -1)
		siteName := strings.Replace(tmpsiteName, "/", "_", -1)

		targetDir := filepath.Join(cwd, siteName)
		targetPath := filepath.Join(targetDir, outputFile)

		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			os.Mkdir(targetDir, os.ModeDir)
		}

		fmt.Printf("%d\n%s\n%s\n%s\n%s\n%s\n%s\n", respCode, strCode, hashedData, outputFile, cwd, targetDir, targetPath)

	}
}
