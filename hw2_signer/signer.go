package main

import (
	"sync"
	"log"
	"sort"
	"bytes"
	"strconv"
)

func ExecutePipeline(jobs ...job) {
	wgResult := new(sync.WaitGroup)

	channels := make([]chan interface{}, 0)

	for i := 0; i <= len(jobs); i++ {
		ch := make(chan interface{}, 100)
		channels = append(channels, ch)
	}

	for i, _ := range jobs {
		wgResult.Add(1)

		go func(currentJobIndex int) {
			outChannel := channels[(currentJobIndex + 1)]
			inChannel := channels[currentJobIndex]

			jobs[currentJobIndex](inChannel, outChannel)

			close(outChannel)
			wgResult.Done()
		}(i)
	}

	wgResult.Wait()
}

func CombineResults(in, out chan interface{}) {
	buffer := make([]string, 0)
	done := make(chan bool)

	log.Printf("CombineResults starts")

	go func() {
		for {
			dataRawP, more := <-in

			if !more {
				log.Printf("finish CombineResults")
				done <- true
				return
			}

			data, _ := dataRawP.(string)
			buffer = append(buffer, data)
		}
	}()

	<-done

	sort.Strings(buffer)

	var stringBuffer bytes.Buffer

	for i, el := range(buffer) {
		stringBuffer.WriteString(el)

		if i != len(buffer) - 1 {
			stringBuffer.WriteString("_")
		}
	}

	result := stringBuffer.String()
	out <- result

	log.Printf("CombineResults %s", result)
}

func MultiHash(in, out chan interface{}) {
	wgResult := new(sync.WaitGroup)

	for {
		dataRawP, more := <-in

		if !more {
			wgResult.Wait()
			log.Printf("finish MultiHash")
			return
		}
		wgResult.Add(1)

		go func(dataRaw interface{}) {
			data, _ := dataRaw.(string)

			wg := new(sync.WaitGroup)

			calcHash := func(wg *sync.WaitGroup, resultCh chan<- string, index string) {
				crc32Data := DataSignerCrc32(index + data)
				log.Printf("%s MultiHash crc32(th+step1)) %s %s", data, index, crc32Data)

				resultCh <- crc32Data
				wg.Done()
			}

			resultChannels := make([]chan string, 0)

			for i := 0; i < 6; i++ {
				r := make(chan string, 1)
				resultChannels = append(resultChannels, r)

				wg.Add(1)
				go calcHash(wg, r, strconv.Itoa(i))
			}

			wg.Wait()

			var buffer bytes.Buffer
			for _, resultChannel := range resultChannels {
				buffer.WriteString(<-resultChannel)
			}

			result := buffer.String()
			log.Printf("%s MultiHash result: %s", data, result)

			out <- result
			wgResult.Done()
		}(dataRawP)
	}
}

func SingleHash(in, out chan interface{}) {
	wgResult := new(sync.WaitGroup)

	for {
		dataRawP, more := <-in

		if !more {
			wgResult.Wait()
			log.Printf("finish SingleHash")
			return
		}

		wgResult.Add(1)

		go func(dataRaw interface{}) {
			data := strconv.Itoa(dataRaw.(int))
			log.Printf("%s SingleHash data %s", data, data)

			wg := new(sync.WaitGroup)
			wg.Add(2)

			result1 := make(chan string, 1)
			result2 := make(chan string, 1)

			go func(wg *sync.WaitGroup, md5out chan<- string) {
				md5Data := DataSignerMd5Proxy(data)
				log.Printf("%s SingleHash md5(data) %s", data, md5Data)

				crc32md5 := DataSignerCrc32(md5Data)
				log.Printf("%s SingleHash crc32(md5(data)) %s", data, crc32md5)

				md5out <- crc32md5
				wg.Done()
			}(wg, result1)

			go func(wg *sync.WaitGroup, crc32out chan<- string) {
				crc32data := DataSignerCrc32(data)
				log.Printf("%s SingleHash crc32(data) %s", data, crc32data)

				crc32out <- crc32data
				wg.Done()
			}(wg, result2)

			wg.Wait()

			result := <-result2 + "~" + <-result1
			log.Printf("%s SingleHash result %s", data, result)

			out <- result
			wgResult.Done()
		}(dataRawP)
	}
}

var mutex = &sync.Mutex{}

func DataSignerMd5Proxy(data string) string {
	mutex.Lock()
	defer mutex.Unlock()

	return DataSignerMd5(data)
}