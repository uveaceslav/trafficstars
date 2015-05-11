package main

import (
	"bytes"
	"encoding/json"
	"io"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

type Stats struct {
	ID         bson.ObjectId  `bson:"_id,omitempty"`
	FilePath   string         `bson:"filepath"`
	Statistics map[string]int `bson:"stats"`
}

const (
	ADD_STATS byte = 0
	GET_STATS byte = 1
)

var (
	statistics   map[string]int = make(map[string]int)
	mongoSession *mgo.Session
)

func main() {
	// Init DB
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:   []string{"127.0.0.1:27017"},
		Timeout: 60 * time.Second,
	}

	var err error
	mongoSession, err = mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		log.Println("Error creating session:", err.Error())
		os.Exit(1)
	}

	// Start server
	l, err := net.Listen("tcp", "127.0.0.1:3000")
	if err != nil {
		log.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	log.Println("Listening on 127.0.0.1:3000")

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("Error accepting:", err.Error())
			continue
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	defer conn.Close()

	log.Println("Reading request...")

	// Read request
	request, err := readFully(conn)
	if err != nil {
		log.Println("Error reading:", err.Error())
		return
	}

	log.Println("Reading request completed.")

	// Parse request
	code := request[0]
	body := request[1:]

	log.Println("Parsing request completed.")

	if code == ADD_STATS {
		addStats(body)
	} else if code == GET_STATS {
		response := getStats()
		// Write response
		_, err := conn.Write(response)
		if err != nil {
			log.Println("Error writing:", err.Error())
			return
		}
	} else {
		log.Println("Error unknown request code:", code)
		return
	}
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

func addStats(statsb []byte) {
	log.Println("Adding stats")

	// Unmarshal
	var stats Stats
	err := json.Unmarshal(statsb, &stats)
	if err != nil {
		log.Println("Error unmarshaling:", err.Error())
		return
	}

	// Saving
	save(stats)

	log.Println("Adding to total stats...")

	// Add to total stats
	var l sync.Mutex
	l.Lock()
	for k, v := range stats.Statistics {
		count, _ := statistics[k]
		statistics[k] = count + v
	}
	l.Unlock()

	log.Println("Adding to total stats completed")
}

func getStats() []byte {
	log.Println("Getting total stats...")

	// Marshal  common stats
	statsb, err := json.Marshal(Stats{Statistics: statistics})
	if err != nil {
		log.Println("Error marshaling:", err.Error())
		return nil
	}

	log.Println("Getting total stats completed.")
	return statsb
}

// Database funcs
func save(stats Stats) {
	log.Println("Saving stats...")

	collection := mongoSession.DB("trafficstars").C("trafficstars")

	err := collection.Insert(stats)
	if err != nil {
		log.Println("Error saving:", err.Error())
	}

	log.Println("Saving stats completed.")
}
