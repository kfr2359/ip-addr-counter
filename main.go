package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"math/bits"
	"os"
	"sync"
	"time"
)

var inputFilePathFlag = flag.String("i", "input.txt", "input file path with IP addresses")
var inputNumReadWorkers = flag.Int("w", 6, "number of read worker")
var inputBufferSize = flag.Int("c", 1024*1024, "size of input read buffer")

func main() {
	flag.Parse()
	if inputFilePathFlag == nil || inputNumReadWorkers == nil || inputBufferSize == nil {
		flag.Usage()
		os.Exit(1)
	}

	startTime := time.Now()
	addressesMap, err := loadIPAddresses(*inputFilePathFlag)
	if err != nil {
		log.Fatal(err)
	}
	endTime := time.Now()
	fmt.Printf("IP addresses were loaded in bitmap for %s\n", endTime.Sub(startTime))

	startTime = time.Now()
	result := countIPAddressesBitMap(addressesMap)
	endTime = time.Now()
	fmt.Printf("IP addresses were counted for %s\n", endTime.Sub(startTime))

	fmt.Printf("Unique IP addresses count: %d\n", result)
}

func loadIPAddresses(filePath string) ([]uint64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file %s: %w", filePath, err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Printf("error closing file %s: %v", filePath, err)
		}
	}()

	// allocate bitmap for each possible IPv4 address
	// let addr be 0x12345678, it's already a kind of index for our bitmap
	// we need 2^32 bits to store ore or 2^(32-1) bytes
	// elem of bitmap - 64 bits, can hold 64 (2^6) unique addresses or last 6 bits of address
	addressesMap := make([]uint64, 2<<(32-1-6))

	ipsRawChan := make(chan [][]byte, 200)
	wg := sync.WaitGroup{}
	for range *inputNumReadWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ipsRaw := range ipsRawChan {
				for i := 0; i < len(ipsRaw); i++ {
					loadIPRaw(ipsRaw[i], addressesMap)
				}
				ipsRaw = nil
			}
		}()
	}

	leftoverBuffer := make([]byte, 0, 16) // max 3 symbols per ip byte, 3 dots, 1 newline
	bufSize := *inputBufferSize
	processBuf := make([]byte, bufSize+16)
	for {
		processBuf = processBuf[:len(leftoverBuffer)+bufSize]
		copy(processBuf, leftoverBuffer)
		bytesRead, err := file.Read(processBuf[len(leftoverBuffer):])
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read input chunk: %w", err)
		}
		processBuf = processBuf[:len(leftoverBuffer)+bytesRead]

		processBufCpy := make([]byte, len(processBuf))
		copy(processBufCpy, processBuf)
		ipsRaw := bytes.Split(processBufCpy, []byte{'\n'})
		ipsRawChan <- ipsRaw[:len(ipsRaw)-1]
		leftoverBuffer = ipsRaw[len(ipsRaw)-1]
	}
	if len(leftoverBuffer) > 0 && leftoverBuffer[0] != 0 {
		loadIPRaw(leftoverBuffer, addressesMap)
	}

	close(ipsRawChan)
	wg.Wait()

	return addressesMap, nil
}

func countIPAddressesBitMap(addressesMap []uint64) int {
	result := 0
	for _, addressesElem := range addressesMap {
		result += bits.OnesCount64(addressesElem)
	}

	return result
}

func loadIPRaw(ipRaw []byte, addressesMap []uint64) {
	ipAddr := parseIPAddr(ipRaw)
	mapIndex := ipAddr >> 6
	// take last 6 bits of addr
	mapElemShift := ipAddr & 0x3f
	// and shift 1 to the left by this value to find bit index
	addressesMap[mapIndex] |= 1 << mapElemShift
}

func parseIPAddr(line []byte) uint32 {
	// no safety and easyness for reading, all for speed
	var dotIndexes [3]int
	dotIndexes[0] = bytes.IndexByte(line, '.')
	dotIndexes[1] = bytes.IndexByte(line[dotIndexes[0]+1:], '.') + dotIndexes[0] + 1
	dotIndexes[2] = bytes.IndexByte(line[dotIndexes[1]+1:], '.') + dotIndexes[1] + 1

	var ipAddrParts [4]byte
	ipAddrParts[0] = byteAtoi(line[:dotIndexes[0]])
	ipAddrParts[1] = byteAtoi(line[dotIndexes[0]+1 : dotIndexes[1]])
	ipAddrParts[2] = byteAtoi(line[dotIndexes[1]+1 : dotIndexes[2]])
	ipAddrParts[3] = byteAtoi(line[dotIndexes[2]+1:])

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
