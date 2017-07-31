package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
)

func worker(jobs <-chan string, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := range jobs {
		resp, err := http.Get(j)
		if err != nil {
			fmt.Printf("Worker failed to get %s with error %s\n", j, err)
			continue
		}
		if resp.StatusCode == http.StatusOK {
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Worker failed to decode response from %s"+
					" with error %s\n", j, err)
				continue
			}
			body := string(bodyBytes)
			strCount := strings.Count(body, "Go")
			fmt.Printf("Count for %s: %d \n", j, strCount)
			resp.Body.Close()
			results <- strCount
		} else {
			fmt.Printf("Page %s returned code %d\n", j, resp.StatusCode)
		}
	}
}

func main() {
	var lines []string
	k := 5
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if scanner.Err() != nil {
		panic(scanner.Err())
	}

	jobs := make(chan string, len(lines))
	results := make(chan int, len(lines))
	var workerWaiter sync.WaitGroup
	for j := 0; j < k; j++ {
		workerWaiter.Add(1)
		go worker(jobs, results, &workerWaiter)
	}

	for _, url := range lines {
		jobs <- url
	}
	close(jobs)

	workerWaiter.Wait()
	totalCount := 0
	resultsCount := len(results)

	for j := 0; j < resultsCount; j++ {
		totalCount += <-results
	}
	fmt.Printf("Total: %d", totalCount)
}
