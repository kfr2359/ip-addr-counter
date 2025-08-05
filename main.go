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
	fmt.Printf("IP addresses were loaded in bitmap for for %s\n", endTime.Sub(startTime))

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

	bufSize := *inputBufferSize
	leftoverBuffer := make([]byte, 0, 16) // max 3 symbols per ip byte, 3 dots, 1 newline
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

		ipsRaw := bytes.Split(processBuf, []byte{'\n'})
		for i := 0; i < len(ipsRaw)-1; i++ {
			if needContinue := loadIPRaw(ipsRaw[i], addressesMap); !needContinue {
				break
			}
		}
		leftoverBuffer = ipsRaw[len(ipsRaw)-1]
	}
	loadIPRaw(leftoverBuffer, addressesMap)

	return addressesMap, nil
}

func countIPAddressesBitMap(addressesMap []uint64) int {
	result := 0
	for _, addressesElem := range addressesMap {
		result += bits.OnesCount64(addressesElem)
	}

	return result
}

func parseIPAddr(line []byte) uint32 {
	ipAddrPartsBytes := bytes.Split(line, []byte{'.'})
	if len(ipAddrPartsBytes) != 4 {
		log.Fatalf("unexprected number of parts in ip \"%s\"", string(line))
	}
	var ipAddrParts [4]byte
	for i := range 4 {
		ipAddrPart := byteAtoi(ipAddrPartsBytes[i])
		ipAddrParts[i] = ipAddrPart
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

func loadIPRaw(ipRaw []byte, addressesMap []uint64) bool {
	if len(ipRaw) == 0 || ipRaw[0] == 0 {
		// reached end of chunk
		return false
	}
	ipAddr := parseIPAddr(ipRaw)
	mapIndex := ipAddr >> 6
	// take last 6 bits of addr
	mapElemShift := ipAddr & 0x3f
	// and shift 1 to the left by this value to find bit index
	addressesMap[mapIndex] |= 1 << mapElemShift

	return true
}
