package main

import (
	"encoding/json"
	"flag"
	"fmt"
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

func Run(workDir, buildSchemaFileName string, ef func(error)) {

	workDir, err := filepath.Abs(workDir)
	if err != nil {
		log.Fatalf("invalid path to project: %s\n", err)
	}

	if err := os.Chdir(workDir); err != nil {
		log.Fatal(err)
	}

	// Just file realted shit. Scroll until the next comment.
	var buildFilePath string
	var buildFileFound bool
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

			if d.Name() == buildSchemaFileName {
				if buildFileFound {
					log.Fatalf("another %s found at %s\n", buildSchemaFileName, path)
				}

				buildFilePath += path
				buildFileFound = true

			}

			return nil
		})
	if err != nil {
		log.Fatalf("error walking project directory: %s\n", err)
	}

	if !buildFileFound {
		log.Fatalf("%s not found\n", buildSchemaFileName)
	}

	buildFile, err := os.Open(buildFilePath)
	if err != nil {
		log.Fatalf("failed to open %s: %s\n", buildFilePath, err)
	}

	contents, err := io.ReadAll(buildFile)
	if err != nil {
		log.Fatalf("couldn't read %s: %s\n", buildFilePath, err)
	}

	buildSchema := map[string]string{}
	if err := json.Unmarshal(contents, &buildSchema); err != nil {
		log.Fatalf("failed to unmarshal %s: %s\n", buildFilePath, err)
	}

	lf, lb := len(srcPaths), len(buildSchema)
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
		hoodies[i] = hoodie.New(f, buildSchema[srcNames[i]], srcPaths[i])
	}

	for _, h := range hoodies {
		if err := h.Parse(); err != nil {
			ef(err)
		}
	}

	if err := block.ValidateTrates(); err != nil {
		ef(err)
	}

	for _, h := range hoodies {
		if err := h.ParseHead(); err != nil {
			ef(err)
		}
	}

	for _, h := range hoodies {
		if err := h.WriteOutput(); err != nil {
			ef(err)
		}
	}
}

func main() {
	var errFunc = func(err error) { log.Fatal(err) }

	var workDir = flag.String("d", ".", "Sourcecode directory.")
	var buildSchemaFileName = flag.String("s", "build.json", "Build schema file's name.")
	flag.BoolFunc("c", "Wether to continue parsing project on error.",
		func(string) error {
			errFunc = func(err error) { log.Println(err) }
			return nil
		},
	)
	flag.Parse()

	fmt.Println(*workDir)

	Run(*workDir, *buildSchemaFileName, errFunc)
}

// Do all errors contain src path?

// DONE: Problems finding build.json, try ./main test_input/project/
// DONE: Allow passing name of .json build schema

// TODO: Improve error handling.
//	> Provide line # where error has occured in all use-cases.
// 	> Write test or files where error is supposed to occure
//  > Return specific error types to be more concise
// TODO: Make tabulation in output files pretty
// TODO: Compile each file in a seperate thread
// TODO: Allow the same output destination for multiple .hoo files
// TODO: Allow #include and #base statements
// 	> why? hoodie does the same but differently
