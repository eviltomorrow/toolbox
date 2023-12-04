package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
)

const (
	KB = 1024
	MB = 1024 * 1024
)

func main() {
	if len(os.Args) == 1 {
		fi, err := os.Stdin.Stat()
		if err != nil {
			log.Fatal(err)
		}
		if fi.Size() > 0 {
			var buf [32 * KB]byte
			for {
				n, err := os.Stdin.Read(buf[0:])
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Fatal(err)
				}
				data := parseEscapeCharacter(buf[:n])
				fmt.Print(data)
			}
		}

	} else {
		files := os.Args[1:]
		for _, path := range files {
			if err := readFile(path); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func readFile(path string) error {
	file, err := os.OpenFile(path, os.O_RDONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	var buf [32 * KB]byte
	for {
		n, err := file.Read(buf[0:])
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		data := parseEscapeCharacter(buf[:n])
		fmt.Print(data)

	}
	return nil
}

func parseEscapeCharacter(buf []byte) string {
	var buffer bytes.Buffer
	for i := 0; i < len(buf); i++ {
		b := buf[i]
		if b == '\\' && i < len(buf)-1 {
			b1 := buf[i+1]
			switch b1 {
			case '\\':
				if i < len(buf)-2 {
					b2 := buf[i+2]
					if b2 != '\\' {
						continue
					} else {
						buffer.WriteByte('\\')
						i += 2
					}
				}
				continue
			case 'r':
				buffer.WriteByte('\r')
				i++
			case 't':
				buffer.WriteByte('\t')
				i++
			case 'n':
				buffer.WriteByte('\n')
				i++
			default:
			}
		} else {
			buffer.WriteByte(b)
		}

	}
	return buffer.String()
}
