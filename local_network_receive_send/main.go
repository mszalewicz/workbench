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
	"strings"
	"sync"
	"time"
)

func main() {
	programType := os.Args[1]

	if (programType != "writer" && programType != "reader") || (programType == "writer" && len(os.Args) != 3) || programType == "usage" {
		fmt.Printf("\nUsage:\n\n   'program reader' or 'program writter _indentifier_', where _identifier_ has to be integer between 0 and 255.\n\nbinary e.x.:     program reader\nbinary e.x.:     program writer 1\nrunning e.x.:    go run . writer 2\n\n")

		os.Exit(0)
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
		candidateAddresses := findLocalAddresses()

		if len(candidateAddresses) == 0 {
			fmt.Println("No valid response from canditate(s).\nShutting down writer.")
			os.Exit(0)
		}

		timeout := 500 * time.Millisecond
		var wg sync.WaitGroup
		for _, address := range candidateAddresses {
			wg.Add(1)
			go func() {
				defer wg.Done()

				conn, err := net.DialTimeout("tcp", address, timeout)
				if err != nil {
					log.Fatal(err)
				}

				for {
					sendMessage(conn)
					time.Sleep(time.Millisecond * 300)
				}
			}()
		}
		wg.Wait()
	}
}

func findLocalAddresses() []string {
	var candidateAddresses []string

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

			ip, ipNet, err := net.ParseCIDR(address.String())
			if err != nil {
				log.Fatal(err)
			}

			if ip.To4() == nil {
				continue
			}

			var wg sync.WaitGroup
			targetPort := "44444"
			timeout := 500 * time.Millisecond

			if strings.Contains(ip.String(), "192.168") {
				fmt.Println("Local network:", address.String())

				for ip := ip.Mask(ipNet.Mask); ipNet.Contains(ip); incrementIP(ip) {
					wg.Add(1)

					go func(ip string) {
						defer wg.Done()

						address := fmt.Sprintf("%s:%s", ip, targetPort)
						conn, err := net.DialTimeout("tcp", address, timeout)

						if err == nil {
							fmt.Printf("Server candidate found at %s\n", address)
							conn.Close()
							candidateAddresses = append(candidateAddresses, address)
						}
					}(ip.String())
				}
				wg.Wait()
				return candidateAddresses
			}
		}
	}

	return candidateAddresses
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

	fmt.Println("Sent message:", msg)

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

	_, err = writer.Write(header)
	if err != nil {
		log.Fatal(err)
	}

	err = writer.Flush()
	if err != nil {
		log.Fatal(err)
	}

	_, err = writer.Write(combinedMsg)
	if err != nil {
		log.Fatal(err)
	}

	err = writer.Flush()
	if err != nil {
		log.Fatal(err)
	}
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

func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
