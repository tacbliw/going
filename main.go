package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/sparrc/go-ping"
)

func main() {
	var (
		concurrentFlag int
		fileFlag       string
		timeoutFlag    int
	)
	flag.IntVar(&concurrentFlag, "c", 30, "Number of concurrent goroutines.")
	flag.StringVar(&fileFlag, "f", "", "Use input from file instead of stdin.")
	flag.IntVar(&timeoutFlag, "t", 500, "Timeout duration for each ping in miliseconds.")
	flag.Parse()

	input := os.Stdin
	if fileFlag != "" {
		file, err := os.Open(fileFlag)
		if err != nil {
			fmt.Printf("Failed to open input file: %s\n", err)
			os.Exit(1)
		}
		input = file
	}

	sc := bufio.NewScanner(input)

	urls := make(chan string, 128)
	concurrency := concurrentFlag
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for url := range urls {
				if !resolves(url) {
					// fmt.Printf("Cannot resolve: %s\n", url)
					continue
				}
				tryPing(url, timeoutFlag)
			}
		}()
	}

	for sc.Scan() {
		urls <- sc.Text()
	}
	close(urls)

	if sc.Err() != nil {
		fmt.Printf("error: %s\n", sc.Err())
	}

	wg.Wait()
}

func resolves(u string) bool {
	_, err := net.LookupIP(u)
	if err != nil {
		return false
	}
	return true
}

func tryPing(url string, timeout int) {
	pinger, err := ping.NewPinger(url)
	if err != nil {
		// fmt.Printf("Cannot initialize Pinger: %s", err)
		return
	}
	pinger.Count = 1
	pinger.OnRecv = func(pkt *ping.Packet) {
		fmt.Printf("%s\n", url)
	}
	go func(p *ping.Pinger) {
		time.Sleep(500 * time.Millisecond)
		if pinger.PacketsRecv == 1 {
			return
		}
		pinger.Stop()
	}(pinger)
	pinger.Run()
}
