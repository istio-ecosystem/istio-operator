// Copyright 2019 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
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

	if err := ioutil.WriteFile(filePath, []byte(strings.Join(out, "\n")), 0644); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
