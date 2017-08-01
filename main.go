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

func worker(id int, jobs <-chan string, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := range jobs {
		resp, err := http.Get(j)
		if err != nil {
			fmt.Printf("Worker failed to get %s with error %s\n", j, err)
		} else if resp.StatusCode == http.StatusOK {
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Worker failed to decode response from %s"+
					" with error %s\n", j, err)
			} else {
				body := string(bodyBytes)
				strCount := strings.Count(body, "Go")
				fmt.Printf("Count for %s: %d worker %d\n", j, strCount, id)
				resp.Body.Close()
				results <- strCount
			}
		} else {
			fmt.Printf("Page %s returned code %d\n", j, resp.StatusCode)
		}
	}
}

func receiver(results <-chan int, stopChannel chan bool) {
	totalCount := 0
	for {
		select {
		case i := <-results:
			{
				totalCount += i
			}
		case <-stopChannel:
			fmt.Printf("Total: %d ", totalCount)
			stopChannel <- true
			return
		}
	}

}

func main() {
	k := 5
	scanner := bufio.NewScanner(os.Stdin)

	jobs := make(chan string, k)
	results := make(chan int)
	stopChannel := make(chan bool)

	var workerWaiter sync.WaitGroup
	for j := 0; j < k; j++ {
		workerWaiter.Add(1)
		go worker(j, jobs, results, &workerWaiter)
	}
	go receiver(results, stopChannel)

	for scanner.Scan() {
		jobs <- scanner.Text()
	}

	if scanner.Err() != nil {
		panic(scanner.Err())
	}

	close(jobs)

	workerWaiter.Wait()

	stopChannel <- true

	<-stopChannel
}
