package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

var (
	host        = flag.String("host", "localhost", "CarrotDB Router host")
	port        = flag.String("port", "8000", "CarrotDB Router port")
	duration    = flag.Duration("duration", 10*time.Second, "Duration of each experiment")
	concurrency = flag.Int("concurrency", 10, "Number of concurrent workers")
)

type Result struct {
	Ops      int
	Latencies []time.Duration
}

func main() {
	flag.Parse()

	addr := fmt.Sprintf("%s:%s", *host, *port)
	color.Cyan("🧪 CarrotDB Lab Bench: Starting Experiments...")
	color.Cyan("Connecting to cluster at %s (Workers: %d, Duration: %v)", addr, *concurrency, *duration)

	// 1. Get Cluster Topology
	topology, err := getTopology(addr)
	if err != nil {
		color.Red("Failed to get topology: %v", err)
		os.Exit(1)
	}

	numShards := len(strings.Split(topology, ";")) - 1
	color.Green("Cluster Topology detected: %d active shards.", numShards)

	// 2. Experiment 1: Write Throughput (SET)
	writeResult := runExperiment(addr, "SET", *concurrency, *duration)
	printReport("Experiment 1: Write Throughput (SET)", writeResult, numShards)

	// 3. Experiment 2: Key Discovery (KEYS)
	keysResult := runExperiment(addr, "KEYS", *concurrency, *duration)
	printReport("Experiment 2: Cluster Scanning (KEYS)", keysResult, numShards)

	color.HiWhite("\n🎓 Lab Bench Complete. Knowledge is power!")
}

func getTopology(addr string) (string, error) {
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	fmt.Fprintln(conn, "CLUSTER")
	resp, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(resp, "+") {
		return resp[1:], nil
	}
	return "", fmt.Errorf("unexpected response: %s", resp)
}

func runExperiment(addr, cmd string, workers int, d time.Duration) Result {
	var wg sync.WaitGroup
	results := make(chan Result, workers)
	start := time.Now()

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				return
			}
			defer conn.Close()

			res := Result{}
			reader := bufio.NewReader(conn)
			
			// Warm up
			time.Sleep(10 * time.Millisecond)

			for time.Since(start) < d {
				command := ""
				if cmd == "SET" {
					command = fmt.Sprintf("SET bench:w%d:%d val\n", workerID, res.Ops)
				} else if cmd == "KEYS" {
					command = "KEYS bench:\n"
				}

				reqStart := time.Now()
				fmt.Fprint(conn, command)
				_, err := reader.ReadString('\n')
				if err != nil {
					break
				}
				res.Latencies = append(res.Latencies, time.Since(reqStart))
				res.Ops++
			}
			results <- res
		}(i)
	}

	wg.Wait()
	close(results)

	final := Result{}
	for r := range results {
		final.Ops += r.Ops
		final.Latencies = append(final.Latencies, r.Latencies...)
	}
	return final
}

func printReport(name string, r Result, numShards int) {
	color.HiWhite("\n--- %s ---", name)
	
	totalSeconds := flag.Lookup("duration").Value.(flag.Getter).Get().(time.Duration).Seconds()
	opsPerSec := float64(r.Ops) / totalSeconds

	sort.Slice(r.Latencies, func(i, j int) bool {
		return r.Latencies[i] < r.Latencies[j]
	})

	p50 := time.Duration(0)
	p99 := time.Duration(0)
	if len(r.Latencies) > 0 {
		p50 = r.Latencies[len(r.Latencies)/2]
		p99 = r.Latencies[int(float64(len(r.Latencies))*0.99)]
	}

	fmt.Printf("Total Ops:  %d\n", r.Ops)
	fmt.Printf("Throughput: %.2f ops/sec\n", opsPerSec)
	fmt.Printf("Latency:    P50: %v | P99: %v\n", p50, p99)

	// The "Professor" section
	color.Yellow("\n👨‍🏫 The Professor's Insight:")
	if strings.Contains(name, "Write Throughput") {
		fmt.Printf("Since CarrotDB uses Sharding, each of your %d shards handles a piece of the load.\n", numShards)
		if numShards > 1 {
			color.HiGreen("Your write throughput is scaling horizontally! More shards = More writes.")
		} else {
			color.HiYellow("You only have 1 shard. Adding more shards would allow keys to be processed in parallel.")
		}
	} else if strings.Contains(name, "Cluster Scanning") {
		fmt.Printf("To find these keys, the Router had to 'Fan-out' to %d different shards.\n", numShards)
		if numShards > 1 {
			color.HiRed("Notice the latency? The more shards you have, the slower global commands like 'KEYS' become because we must wait for the slowest shard to respond.")
		} else {
			color.HiGreen("With only 1 shard, cluster scanning is fast because there is no network fan-out overhead.")
		}
	}
	fmt.Println()
}
