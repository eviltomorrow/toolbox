package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var path = flag.String("f", "", "hex file")

func main() {
	flag.Parse()

	f, err := os.OpenFile(*path, os.O_RDONLY, 0o644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	result := make([]string, 0, 128)

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, "   ") {
			attrs := strings.Split(line, "   ")
			if len(attrs) == 2 {
				for _, attr := range strings.Split(attrs[1], " ") {
					result = append(result, fmt.Sprintf("0x%s", attr))
				}
			}
		}
	}
	fmt.Printf("[]byte{%s}", strings.Join(result, ", "))
}
