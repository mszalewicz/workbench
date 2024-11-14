package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"os"
	"strconv"
)

func main() {
	programType := os.Args[1]

	if programType != "writer" && programType != "reader" {
		log.Fatal("Usage: 'program reader' or 'program writter indentifier', where identifier has to be integer between 0 and 255.")
	}

	if programType == "reader" {
		listener, err := net.Listen("tcp", ":44444")
		if err != nil {
			fmt.Println("Error starting server:", err)
			return
		}
		defer listener.Close()

		log.Println("Server started on :44444")

		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("Error accepting connection:", err)
				continue
			}
			go receiveMsgs(conn)
		}
	}

	if programType == "writer" {
		findLocalAddresses()
		// conn, err := net.Dial("tcp", "localhost:8080")
		// if err != nil {
		// 	fmt.Println("Error connecting to server:", err)
		// 	os.Exit(1)
		// }
		// defer conn.Close()
		// fmt.Println("Connected to server at localhost:8080")
	}
}

func findLocalAddresses() {
	interfaces, err := net.Interfaces()

	if err != nil {
		log.Fatal(err)
	}

	for _, intrface := range interfaces {
		addresses, err := intrface.Addrs()

		if err != nil {
			log.Fatal(err)
		}

		for _, address := range addresses {
			ip, _, err := net.ParseCIDR(address.String())
			if err != nil {
				log.Fatal(err)
			}

			if ip.To4() == nil { // Check if addres is IP v4
				continue
			}

			fmt.Println(ip)
		}
	}
}

func receiveMsgs(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		lengthBytes := make([]byte, 4)
		if _, err := reader.Read(lengthBytes); err != nil {
			fmt.Println("Error reading message length:", err)
			return
		}

		messageLength := binary.BigEndian.Uint32(lengthBytes)

		messageBytes := make([]byte, messageLength)
		if _, err := reader.Read(messageBytes); err != nil {
			fmt.Println("Error reading message:", err)
			return
		}

		fmt.Println("Sender:", int(messageBytes[0]), " Received message:", string(messageBytes[1:]))
	}
}

func sendMessage(conn net.Conn) {
	writer := bufio.NewWriter(conn)
	msgLength, msg := randString()

	identifier, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}

	combinedMsg := []byte{byte(identifier)}

	for _, char := range msg {
		combinedMsg = append(combinedMsg, byte(char))
	}

	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, uint32(msgLength+1))
	writer.Write(header)
	writer.Write(combinedMsg)
}

func randString() (int, string) {
	const chars = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"0123456789"

	rndLength := rand.IntN(20) + 5

	result := make([]byte, rndLength)
	for i := range result {
		result[i] = chars[rand.Int64()%int64(len(chars))]
	}
	return rndLength, string(result)
}
