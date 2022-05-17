package query

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"net"
	"regexp"
	"strconv"
	"time"
)

var (
	colorRegex = regexp.MustCompile(`\^[\d:;]`)
)

func GetMasterServerResponse(masterServer string, gameId string, gameVersion string) []string {
	masterResponse, err := sendMessage(masterServer, fmt.Sprintf("getservers %s %s full empty", gameId, gameVersion), true)
	if err != nil {
		log.Println("couldn't get response from master server", err.Error())
		return []string{}
	}
	chunks := bytes.Split(masterResponse, []byte("\\"))
	servers := []string{}
	for i := 1; i < len(chunks)-1; i++ {
		server := chunks[i]
		if len(server) != 6 {
			continue
		}
		ip := net.IP(server[0:4]).String()
		port := strconv.Itoa(int(binary.BigEndian.Uint16(server[4:6])))
		servers = append(servers, ip+":"+port)
	}
	log.Printf("master server %q (%s) responded with %d servers", masterServer, gameId, len(servers))
	return servers
}

func GetSingleServerResponse(server string) (map[string]string, error) {
	challenge := make([]byte, 4)
	rand.Seed(time.Now().UnixNano())
	rand.Read(challenge)
	hex := fmt.Sprintf("%x", challenge)

	message := fmt.Sprintf("getinfo %s", hex)
	serverResponse, err := sendMessage(server, message, false)
	if err != nil {
		return nil, fmt.Errorf("couldn't get response from game server")
	}
	chunks := bytes.Split(serverResponse, []byte("\\"))[1:]
	if len(chunks)%2 != 0 {
		return nil, fmt.Errorf("malformed server response, key/value length not even")
	}

	info := map[string]string{}
	for i := 0; i < len(chunks)-1; i += 2 {
		info[string(chunks[i])] = string(chunks[i+1])
	}

	if returnedChallenge, present := info["challenge"]; present {
		if returnedChallenge != hex {
			return nil, fmt.Errorf("serverinfo challenge mismatch")
		}
	} else {
		return nil, fmt.Errorf("serverinfo challenge absent")
	}

	if hostname, present := info["hostname"]; present {
		info["hostname"] = colorRegex.ReplaceAllString(hostname, "")
	}
	info["ip"] = server

	log.Printf("got server response from server %s", server)

	return info, nil
}

func sendMessage(address string, message string, expectEot bool) ([]byte, error) {
	conn, err := net.DialTimeout("udp", address, 5*time.Second)
	if err != nil {
		return nil, err
	}
	rawMessage := []byte{0xFF, 0xFF, 0xFF, 0xFF}
	rawMessage = append(rawMessage, message...)
	conn.Write(rawMessage)

	response := make([]byte, 8192)
	reader := bufio.NewReader(conn)
	read, _ := reader.Read(response)
	response = response[:read]
	if expectEot {
		for {
			if bytes.Contains(response, []byte("EOF")) || bytes.Contains(response, []byte("EOT")) {
				break
			}
			tmp := make([]byte, 8192)
			read, _ := reader.Read(tmp)
			tmp = tmp[:read]
			response = append(response, tmp...)
		}
	}
	if err != nil {
		return nil, err
	}
	return response, nil
}
