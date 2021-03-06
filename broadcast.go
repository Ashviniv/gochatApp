package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"sync"
	"time"

	pb "github.com/gautamrege/gochat/api"
)

// Broadcast Listener
// Listens on 33333 and updates the Global Handles list
func registerHandle(wg *sync.WaitGroup, exit chan bool) {
	defer wg.Done()
	// Check if the handle is already in HANDLES. If not, add a new one!

	handle := Handle{}
	for {
		// listen to port 33333
		localAddress, _ := net.ResolveUDPAddr("udp4", "192.168.1.255:33333")
		connection, err := net.ListenUDP("udp", localAddress)
		if err != nil {
			fmt.Println(err)
		}

		// read the data and add to handlers. Igore the handle with same host
		inputBytes := make([]byte, 4096)
		length, _, _ := connection.ReadFromUDP(inputBytes)
		buffer := bytes.NewBuffer(inputBytes[:length])
		decoder := gob.NewDecoder(buffer)
		decoder.Decode(&handle)
		if handle.Host != *host {
			//fmt.Println("listened data %s\n > ", handle)
			HANDLES.Insert(handle.Handle)
		}

		// close the connection
		connection.Close()
	}
}

// isAlive go-routine that publishes it's Handle on 33333
func isAlive(wg *sync.WaitGroup, exit chan bool) {
	defer wg.Done()

	// Broadcast immediately at the start
	broadcastIsAlive()

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	for {
		select {
		case <-exit:
			break
		case <-ticker.C:
			broadcastIsAlive()
		}
	}
}

// broadcast on 33333 every 30 seconds with Handler
// - name
// - port
// - host
// - current timestamp
func broadcastIsAlive() {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: []byte{192, 168, 1, 255}, Port: 33333})
	if err != nil {
		fmt.Println(err)
	}
	handle := Handle{
		Handle: pb.Handle{
			Name: *name,
			Port: int32(*port),
			Host: *host,
		},
		Created_at: time.Now(),
	}

	encoder.Encode(handle)
	conn.Write(buffer.Bytes())
	buffer.Reset()
	conn.Close()
}
