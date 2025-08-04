package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"math/bits"
	"os"
	"time"
)

var inputFilePathFlag = flag.String("i", "input.txt", "input file path with IP addresses")

func main() {
	flag.Parse()
	if inputFilePathFlag == nil {
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
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ipAddr, errParse := parseIPAddr(scanner.Text())
		if errParse != nil {
			return 0, err
		}
		mapIndex := ipAddr >> 6
		// take last 6 bits of addr
		mapElemShift := ipAddr & 0x3f
		// and shift 1 to the left by this value to find bit index
		addressesMap[mapIndex] |= 1 << mapElemShift
	}
	if err = scanner.Err(); err != nil {
		return 0, fmt.Errorf("scan file %s: %w", filePath, err)
	}

	result := 0
	for _, addressesElem := range addressesMap {
		result += bits.OnesCount64(addressesElem)
	}

	return result, nil
}

func parseIPAddr(line string) (uint32, error) {
	var ipAddrParts [4]byte
	_, err := fmt.Sscanf(line, "%d.%d.%d.%d", &ipAddrParts[0], &ipAddrParts[1], &ipAddrParts[2], &ipAddrParts[3])
	if err != nil {
		return 0, fmt.Errorf("parse ip addr %s: %w", line, err)
	}
	// we can use any endianness
	return binary.BigEndian.Uint32(ipAddrParts[:]), nil
}
