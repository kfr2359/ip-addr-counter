package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"math/bits"
	"os"
	"time"
)

var inputFilePathFlag = flag.String("i", "input.txt", "input file path with IP addresses")
var inputNumReadWorkers = flag.Int("w", 6, "number of read worker")
var inputIPChunkSize = flag.Int("c", 10000, "size of raw ip chunk that is passed to workers")

func main() {
	flag.Parse()
	if inputFilePathFlag == nil || inputNumReadWorkers == nil || inputIPChunkSize == nil {
		flag.Usage()
		os.Exit(1)
	}
	startTime := time.Now()
	result, err := countIPAddressesBitMap(*inputFilePathFlag)
	if err != nil {
		log.Fatal(err)
	}
	endTime := time.Now()
	fmt.Printf("Unique IP addresses count: %d\n", result)
	fmt.Printf("Counted for %s\n", endTime.Sub(startTime))
}

func countIPAddressesBitMap(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("open file %s: %w", filePath, err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Printf("error closing file %s: %v", filePath, err)
		}
	}()

	// allocate bitmap for each possible IPv4 address
	// let addr be 0x12345678, it's already a kind of index for our bitmap
	// elem of bitmap - 64 bits, can hold 64 (2^6) unique addresses or last 6 bits
	addressesMap := make([]uint64, 2<<(32-6))

	numReadWorkers := uint(*inputNumReadWorkers)
	readChan := make(chan [][]byte, 100)
	for range numReadWorkers {
		go func() {
			for ipRawChunk := range readChan {
				for _, ipRaw := range ipRawChunk {
					if len(ipRaw) == 0 {
						// reached end of chunk
						break
					}
					ipAddr := parseIPAddr(ipRaw)
					mapIndex := ipAddr >> 6
					// take last 6 bits of addr
					mapElemShift := ipAddr & 0x3f
					// and shift 1 to the left by this value to find bit index
					addressesMap[mapIndex] |= 1 << mapElemShift
				}
			}
		}()
	}

	ipChunkSize := *inputIPChunkSize
	ipRawChunk := make([][]byte, ipChunkSize)
	i := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ipRaw := scanner.Bytes()
		ipRawChunk[i] = make([]byte, len(ipRaw))
		copy(ipRawChunk[i], ipRaw)
		i++
		if i == ipChunkSize {
			readChan <- ipRawChunk
			ipRawChunk = make([][]byte, ipChunkSize)
			i = 0
		}
	}
	if i > 0 {
		readChan <- ipRawChunk
	}
	if err = scanner.Err(); err != nil {
		return 0, fmt.Errorf("scan file %s: %w", filePath, err)
	}

	close(readChan)

	result := 0
	for _, addressesElem := range addressesMap {
		result += bits.OnesCount64(addressesElem)
	}

	return result, nil
}

func parseIPAddr(line []byte) uint32 {
	ipAddrPartsBytes := bytes.Split(line, []byte{'.'})
	var ipAddrParts [4]byte
	for i := range 4 {
		ipAddrPart := byteAtoi(ipAddrPartsBytes[i])
		ipAddrParts[i] = byte(ipAddrPart)
	}
	// we can use any endianness
	return binary.BigEndian.Uint32(ipAddrParts[:])
}

// simplified implementation of strings.Atoi for byte slice
func byteAtoi(raw []byte) byte {
	var result byte
	for i := 0; i < len(raw); i++ {
		result = result*10 + raw[i] - '0'
	}
	return result
}
