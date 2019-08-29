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

/*
// GOTYPE: map[string]interface{}
type TypeMapStringInterface struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}
*/

const (
	goTypeToken = "// GOTYPE: "
)

func usage() {
	fmt.Println("Usage: fixup-structs -f <filename>")
	os.Exit(1)
}

func getFileLines(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	var out []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		out = append(out, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func main() {
	var filePath string
	flag.StringVar(&filePath, "f", "", "path to input file")
	flag.Parse()

	if filePath == "" {
		usage()
	}

	lines, err := getFileLines(filePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	subs := make(map[string]string)

	var tmp []string

	for i, l := range lines {
		if strings.Contains(l, goTypeToken) {
			v := strings.ReplaceAll(l, goTypeToken, "")
			nl := lines[i+1]
			nlv := strings.Split(nl, " ")
			if len(nlv) != 4 || nlv[0] != "type" || nlv[2] != "struct" || nlv[3] != "{" {
				fmt.Printf("Bad GOTYPE target: %s\n", nl)
				os.Exit(1)
			}
			subs[nlv[1]] = v
			for ; lines[i] != ""; i++ {
			}
			l = lines[i]
		}

		tmp = append(tmp, l)
	}

	lines = tmp
	var out []string

	for _, l := range lines {
		for k, v := range subs {
			if strings.Contains(l, v) {
				l = strings.ReplaceAll(l, v, k)
			}
		}
		out = append(out, l)
	}

	if err := ioutil.WriteFile(filePath, []byte(strings.Join(out, "\n")), 0644); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
