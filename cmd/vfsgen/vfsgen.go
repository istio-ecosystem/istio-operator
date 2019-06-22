// +build ignore

package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/shurcooL/vfsgen"
)

func main() {
	var cwd, _ = os.Getwd()
	templates := http.Dir(filepath.Join(cwd, "../data"))
	if err := vfsgen.Generate(templates, vfsgen.Options{
		Filename:    "../pkg/vfsgen/vfsgen_data.go",
		PackageName: "vfsgen",
		//		BuildTags:    "deploy_build",
		VariableName: "Assets",
	}); err != nil {
		log.Fatalln(err)
	}
}
