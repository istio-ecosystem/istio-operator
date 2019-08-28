package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

const (
	replaceStructToken = "// GOSTRUCT: "
)

func usage() {
	fmt.Println("Usage: fixup-structs -f <filename>")
	os.Exit(1)
}

func main() {
	var filePath string
	flag.StringVar(&filePath, "f", "", "path to input file")
	flag.Parse()

	if filePath == "" {
		usage()
	}

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	var out []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		l := scanner.Text()
		newLine := l
		if strings.Contains(l, replaceStructToken) {
			newLine = strings.ReplaceAll(l, replaceStructToken, "")
			scanner.Scan()
		}
		out = append(out, newLine)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(strings.Join(out, "\n"))
}
