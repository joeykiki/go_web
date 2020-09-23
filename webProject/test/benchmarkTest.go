package main

import (
	"fmt"
	//"io/ioutil"
	"log"
	"math"
	"net"
	"net/url"
	"strings"

	//"net/url"
	"sync"
	//"testing"
	"strconv"
	"net/http"
	"sync/atomic"
	"time"
)


//func routine(remaining *int32, req []byte){
//	for atomic.AddInt32(remaining, -1) >= 0{
//		http.Post("localhost:3000/login", "application/json", bytes.NewBuffer(req))
//	}
//}

//func BenchmarkLogin(b *testing.B){
//	var username string
//	var password string
//	var remaining int32
//	for i := 0 ; i < 10000 ; i++{
//		b.StopTimer()
//		username = strconv.Itoa(i%200)
//		password = strconv.Itoa(i%200)
//		req, _ := json.Marshal(map[string]string{
//			"username": username,
//			"password": password,
//		})
//		b.StartTimer()
//		go routine(&remaining, req)
//	}
//}

func benchmarkBasicN(totalN, concurrency int32) (elapsed time.Duration) {
	readyGo := make(chan bool)
	var wg sync.WaitGroup
	wg.Add(int(concurrency))

	remaining := totalN

	var transport http.RoundTripper = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          int(concurrency),
		MaxIdleConnsPerHost:   int(concurrency),
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
	}

	var username string
	var password string

	cliRoutine := func(no int32) {

		for atomic.AddInt32(&remaining, -1) >= 0 {

			username = strconv.Itoa(int(remaining)%200)
			password = strconv.Itoa(int(remaining)%200)
			req := url.Values{"username": {username}, "password": {password}}
			request, _ := http.NewRequest("POST", "http://127.0.0.1:3000/login", strings.NewReader(req.Encode()))
			request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			cookie := http.Cookie{Name:"Token", Value:"MTQ3MTU5NTA2NzQ3MA=="}
			request.AddCookie(&cookie)
			_, err := client.Do(request)
			if err != nil {
				log.Println(err)
			}
			//defer resp.Body.Close()
			//_, err = ioutil.ReadAll(resp.Body)
			//if err != nil {
			//	log.Println(err)
			//}
		}

		wg.Done()
	}

	for i := int32(0); i < concurrency; i++ {
		go cliRoutine(i)
	}

	close(readyGo)
	start := time.Now()

	wg.Wait()

	return time.Since(start)
}

func main(){
	totalN  := int32(1000)
	concurrency := int32(200)
	elapsed := benchmarkBasicN(totalN, concurrency)
	fmt.Println("HTTP server benchmark done:")
	fmt.Printf("\tTotal Requests(%v) - Concurrency(%v) - Cost(%s) - QPS(%v/sec)\n",
		totalN, concurrency, elapsed, math.Ceil(float64(totalN)/(float64(elapsed)/1000000000)))
}
