package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/sparrc/go-ping"
)

func main() {
	flag.Parse()

	var input io.Reader

	if flag.NArg() > 0 {
		file, err := os.Open(flag.Arg(0))
		if err != nil {
			fmt.Printf("Failed to open input file: %s\n", err)
			os.Exit(1)
		}
		input = file
	}

	sc := bufio.NewScanner(input)

	urls := make(chan string, 128)
	concurrency := 30
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
				tryPing(url)
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

func tryPing(url string) {
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
