package main

import (
	"fmt"
	"sort"
	"strconv"
	"sync"
)

func DataSignerCrc32(data string) string {
	return data
}

func DataSignerMd5(data string) string {
	return data
}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for data := range in {
		wg.Add(1)
		go func(data interface{}) {
			defer wg.Done()
			dataInt := data.(int)
			dataStr := strconv.Itoa(dataInt)
			md5 := DataSignerMd5(dataStr)
			crc32 := DataSignerCrc32(dataStr)
			crc32Md5 := DataSignerCrc32(md5)
			out <- crc32 + "~" + crc32Md5
		}(data)
	}
	wg.Wait()
	close(out)
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for data := range in {
		wg.Add(1)
		go func(data interface{}) {
			defer wg.Done()
			dataStr := data.(string)
			multiHash := make([]string, 6)
			wgInner := &sync.WaitGroup{}
			for th := 0; th <= 5; th++ {
				wgInner.Add(1)
				go func(th int) {
					defer wgInner.Done()
					multiHash[th] = DataSignerCrc32(strconv.Itoa(th) + dataStr)
				}(th)
			}
			wgInner.Wait()
			out <- multiHash
		}(data)
	}
	wg.Wait()
	close(out)
}

func CombineResults(in, out chan interface{}) {
	var results []string
	for data := range in {
		multiHash := data.([]string)
		for _, hash := range multiHash {
			results = append(results, hash)
		}
	}
	sort.Strings(results)
	out <- results
	close(out)
}

func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{})
	for _, job := range jobs {
		wg.Add(1)
		out := make(chan interface{})
		go func(job job, in, out chan interface{}) {
			defer wg.Done()
			defer close(out)
			job(in, out)
		}(job, in, out)
		in = out
	}
	wg.Wait()
}

func main() {
	inputData := []int{0, 1, 2, 3, 4, 5, 6}
	output := make(chan interface{})
	go ExecutePipeline(SingleHash, MultiHash, CombineResults)
	for _, data := range inputData {
		inputData <- data
	}
	close(inputData)
	result := <-output
	fmt.Println(result)
}
