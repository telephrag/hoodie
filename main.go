package main

import (
	"encoding/json"
	"flag"
	"io"
	"io/fs"
	"log"
	"main/block"
	"main/hoodie"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	log.SetFlags(log.Flags() &^ log.LstdFlags)
}

func main() {
	workDir := flag.String("d", ".", "Specify sourcefile directory.")
	flag.Parse()
	if err := os.Chdir(*workDir); err != nil {
		log.Fatal(err)
	}

	// Just file realted shit. Scroll until the next comment.
	buildFile, err := os.Open("build.json")
	if err != nil {
		log.Fatal("build.json not found")
	}

	contents, err := io.ReadAll(buildFile)
	if err != nil {
		log.Fatalf("couldn't read build.json: %s\n", err)
	}

	buildSchema := map[string]string{}
	if err := json.Unmarshal(contents, &buildSchema); err != nil {
		log.Fatalf("failed to unmarshal build.json: %s\n", err)
	}

	srcPaths, srcNames := []string{}, []string{}
	err = filepath.WalkDir(".",
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if strings.HasSuffix(d.Name(), ".hoo") {
				srcPaths = append(srcPaths, path)
				srcNames = append(srcNames, d.Name())
			}
			return nil
		})
	if err != nil {
		log.Fatalf("errored while looking for files to compile: %s\n", err)
	}

	//	fmt.Println(srcNames)
	//	fmt.Println(srcPaths)
	//	fmt.Println(buildSchema)

	lf := len(srcPaths)
	lb := len(buildSchema)
	if lf != lb {
		log.Fatalf("build schema: %d entries; files found: %d\n", lb, lf)
	}

	// Actual work starts here
	hoodies := make([]*hoodie.Hoodie, len(srcPaths))
	for i := range srcPaths {
		f, err := os.Open(srcPaths[i])
		if err != nil {
			log.Fatalf("failed to open .hoo file %s: %s\n", srcPaths[i], err)
		}
		// fmt.Println(buildSchema[srcNames[i]], srcNames[i], srcPaths[i])
		hoodies[i] = hoodie.New(f, buildSchema[srcNames[i]], srcPaths[i])
	}

	for _, h := range hoodies {
		if err := h.Parse(); err != nil {
			log.Fatal(err)
		}
	}

	if err := block.ValidateTrates(); err != nil {
		log.Fatal(err)
	}

	for _, h := range hoodies {
		if err := h.ParseHead(); err != nil {
			log.Fatal(h.SrcPath(), err)
		}
	}

	for _, h := range hoodies {
		if err := h.WriteOutput(); err != nil {
			log.Fatal(err)
		}
	}
}

// TODO: Improve error handling. Provide line #
// 	where error has occured in all use-cases.
// TODO: Make tabulation in output files pretty
// TODO: Compile each file in a seperate thread
// TODO: Write documentation
