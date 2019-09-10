package main

import (
	"bufio"
	//"fmt"
	"io"
	//	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func LinesInFile(fileName string) []string {
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

func main() {
	fileName := `list`
	for _, line := range LinesInFile(fileName) {
		response, err := http.Get(line)
		if err != nil {
			log.Fatal(err)
		}

		defer response.Body.Close()

		n, err := io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("bytes:", n)
	}
}
