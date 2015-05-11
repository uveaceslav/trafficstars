package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
)

type Stats struct {
	FilePath   string
	Statistics map[string]int
}

const (
	ADD_STATS byte = 0
	GET_STATS byte = 1
)

var (
	server string
	dir    string
)

func main() {
	flag.StringVar(&server, "server", "127.0.0.1:3000", "server address")
	flag.StringVar(&dir, "dir", "./", "directory for walking")

	flag.Parse()

	fmt.Println("server:", server)
	fmt.Println("dir:", dir)

	run(dir)
}

func run(dir string) {
	fullPath, err := filepath.Abs(dir)
	if err != nil {
		log.Println("Error abs:", err)
		return
	}

	files := listFiles(fullPath)
	log.Println("Found files:", len(files))

	var wg sync.WaitGroup
	for _, fileName := range files {
		wg.Add(1)
		go worker(fileName, &wg)
	}
	wg.Wait()

	totalStats := getTotalStats()
	
	// Sort and print
	sm := NewSortedMap(totalStats.Statistics)
	sm.Sort()

	for i, _ := range sm.Keys {
		log.Println(sm.Keys[i], sm.Vals[i])
	}
}

func worker(fileName string, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Println(fileName, "in progress...")

	stats, err := count(fileName)
	if err != nil {
		log.Println("Error counting stats:", err)
		return
	}

	log.Println(fileName, "completed.")

	addStats(stats)
}

func listFiles(dir string) []string {
	var files []string
	filepath.Walk(dir, func(path string, f os.FileInfo, _ error) error {
		if !f.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files
}

func count(fileName string) (Stats, error) {
	result := make(map[string]int)

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Println("Error reading file:", err)
		return Stats{}, err
	}

	for _, b := range data {
		counter, _ := result[fmt.Sprintf("%#x", b)]
		result[fmt.Sprintf("%#x", b)] = counter + 1
	}

	return Stats{FilePath: fileName, Statistics: result}, nil
}

func getTotalStats() Stats {
	// Make request
	request := []byte{GET_STATS}

	// Send request
	response, err := sendRequest(request)
	if err != nil {
		log.Println("Error sending:", err.Error())
		return Stats{}
	}

	// Unmarshal stats
	var stats Stats
	err = json.Unmarshal(response, &stats)
	if err != nil {
		log.Println("Error unmarshaling:", err.Error())
		return Stats{}
	}

	return stats
}

func addStats(stats Stats) {
	// Marshal stats
	statsb, err := json.Marshal(stats)
	if err != nil {
		log.Println("Error marshaling:", err.Error())
		return
	}

	// Add command code
	statsb = append([]byte{ADD_STATS}, statsb...)

	// Send request
	_, err = sendRequest(statsb)
	if err != nil {
		log.Println("Error sending:", err.Error())
		return
	}

	log.Println("Completed")
}

func sendRequest(request []byte) ([]byte, error) {
	// Create connection
	conn, err := net.Dial("tcp", server)
	if err != nil {
		log.Println("Error connecting:", err.Error())
		return nil, err
	}
	defer conn.Close()

	// Write request
	log.Println("Write request")
	_, err = conn.Write(request)
	if err != nil {
		log.Println("Error writing:", err.Error())
		return nil, err
	}

	// Read response
	response, err := readFully(conn)
	if err != nil {
		log.Println("Error reading:", err.Error())
		return nil, err
	}

	return response, nil
}

func readFully(conn net.Conn) ([]byte, error) {
	result := bytes.NewBuffer(nil)
	var buf [1024]byte

	for {
		n, err := conn.Read(buf[0:])
		result.Write(buf[0:n])
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if n < 1024 {
			break
		}
	}

	return result.Bytes(), nil
}
