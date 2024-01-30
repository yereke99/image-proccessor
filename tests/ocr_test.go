package tests

import (
	"ImageProcessor/pipeline"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"
)

func TestDlock(t *testing.T) {

	responseChan := make(chan int)
	errChan := make(chan error)

	go func() {
		time.Sleep(2 * time.Second)
		errChan <- errors.New("Hi")
	}()

	select {
	case rsp := <-responseChan:
		fmt.Println(rsp)
		return
	case err := <-errChan:
		fmt.Println(err)
		return
	}

}

func TestIOCR(t *testing.T) {

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:    5,
			MaxConnsPerHost: 20,
		},
	}

	serv := "192.168.0.75"

	b, err := os.ReadFile("./ss/1.png")
	if err != nil {
		log.Fatal(err)
	}

	pool := make(chan int, 300)  //300 requests in one time
	errs := make(map[string]int) //Errors counter
	var total float64 = 0
	var success float64 = 0
	var fail float64 = 0
	mt := sync.Mutex{}

	//V1

	for {
		pool <- 1
		go func() {
			total++
			defer func() {
				{
					<-pool
					printMap(total, success, fail, errs)
				}
			}()
			_, code, err := pipeline.IOCR(client, b, serv, 0)
			if err != nil {
				mt.Lock()
				if val, ok := errs[err.Error()]; !ok {
					errs[err.Error()] = 1
				} else {
					errs[err.Error()] = val + 1
				}
				mt.Unlock()
				fail++

				return
			}
			if code != 200 {
				println(code)
				mt.Lock()
				if val, ok := errs["rejected"]; !ok {
					errs["rejected"] = 1
				} else {
					errs["rejected"] = val + 1
				}
				mt.Unlock()
				fail++
			} else {
				success++
			}
		}()
		time.Sleep(10 * time.Millisecond)
	}
	//fmt.Println("==>", resp)

}

func printMap(total, success, fail float64, errs map[string]int) {
	successP := (success / (success + fail)) * 100
	log.Println("Success %: ", successP, "Total: ", total, "Fail: ", fail, "Success", success)
}
