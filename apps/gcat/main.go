package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "gcat",
		Usage: "gcat - concatenate files and print on the standard output",
		Flags: []cli.Flag{
			&cli.IntFlag{Name: "goroutines", Aliases: []string{"g"}, Value: 16, Usage: "specify the threads"},
			&cli.StringFlag{Name: "word", Aliases: []string{"w"}, Value: "", Usage: "key word"},
			&cli.BoolFlag{Name: "status", Aliases: []string{"s"}, Value: false, Usage: "show status"},
			&cli.StringFlag{Name: "cache-dir", Aliases: []string{"d"}, Value: "./cache", Usage: "cache-dir path"},
		},
		Action: func(c *cli.Context) error {
			begin := time.Now()
			path := c.Args().Get(c.Args().Len() - 1)
			fi, err := os.Stat(path)
			if err != nil {
				return err
			}
			if fi.IsDir() || !strings.HasSuffix(fi.Name(), ".tar.gz") {
				return fmt.Errorf("invalid file, just support *.tar.gz")
			}

			goroutines := c.Int("goroutines")
			word := c.String("word")
			showCost := c.Bool("status")
			cacheDir := c.String("cache-dir")

			file, err := os.OpenFile(path, os.O_RDONLY, 0o644)
			if err != nil {
				return fmt.Errorf("open tar.gz failure, nest error: %v", err)
			}
			defer file.Close()

			data := make(chan string, 64)
			pipe := make(chan string, 32)

			go func() {
				if _, err := OpenTarGzFile(file, cacheDir, pipe); err != nil {
					log.Fatalf("open tar.gz failure, nest error: %v", err)
				}
				close(pipe)
			}()

			var wg sync.WaitGroup
			for i := 0; i < goroutines; i++ {
				wg.Add(1)
				go func() {
					for p := range pipe {
						step := time.Now()
						count, err := ReadTarGzFile(p, word, data)
						if err != nil {
							log.Fatalf("read tar.gz failure, nest error: %v", err)
						}
						if showCost {
							log.Printf("[I] Step: %s, cost: %v, count: %v", p, time.Since(step), count)
						}
						os.Remove(p)
					}
					wg.Done()
				}()
			}

			go func() {
				for d := range data {
					fmt.Println(d)
				}
				wg.Done()
			}()

			wg.Wait()
			close(data)
			wg.Add(1)
			wg.Wait()

			if showCost {
				log.Printf("[I] Search complete, cost: %v", time.Since(begin))
			}
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func OpenTarGzFile(file *os.File, dstDir string, pipe chan string) ([]string, error) {
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return nil, err
	}

	greader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("open tar.gz file failure, nest error: %v", err)
	}
	defer greader.Close()

	path := make([]string, 0, 32)
	treader := tar.NewReader(greader)
	for {
		header, err := treader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read next failure, nest error: %v", err)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			target := filepath.Join(dstDir, header.Name)
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return nil, fmt.Errorf("mkdir failure, nest error: %v", err)
			}
		case tar.TypeReg:
			if strings.HasSuffix(header.Name, ".tar.gz") {
				target := filepath.Join(dstDir, header.Name)
				dstFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
				if err != nil {
					return nil, fmt.Errorf("open file failure, nest error: %v", err)
				}
				defer dstFile.Close()

				if _, err := io.Copy(dstFile, treader); err != nil {
					return nil, fmt.Errorf("copy file failure, nest error: %v", err)
				}
				pipe <- target
			}

			if strings.HasSuffix(header.Name, ".txt") {
				reader := bufio.NewReader(treader)
				for {
					line, _, err := reader.ReadLine()
					if err == io.EOF {
						break
					}
					if err != nil {
						return nil, fmt.Errorf("read line failure, nest error: %v", err)
					}
					fmt.Println(string(line))
				}
			}

		default:
		}
	}

	return path, nil
}

func ReadTarGzFile(path string, word string, data chan string) (int64, error) {
	file, err := os.OpenFile(path, os.O_RDONLY, 0o755)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	greader, err := gzip.NewReader(file)
	if err != nil {
		return 0, fmt.Errorf("open tar.gz file failure, nest error: %v", err)
	}
	defer greader.Close()

	treader := tar.NewReader(greader)
	var count int64
	for {
		header, err := treader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("read tar header failure, nest error: %v", err)
		}
		if strings.HasSuffix(header.Name, ".txt") {
			reader := bufio.NewReader(treader)
			for {
				line, _, err := reader.ReadLine()
				if err == io.EOF {
					break
				}
				if err != nil {
					return 0, fmt.Errorf("read line failure, nest error: %v", err)
				}
				if strings.Contains(string(line), word) {
					count++
					data <- string(line)
				}
			}
		}
	}
	return count, nil
}
