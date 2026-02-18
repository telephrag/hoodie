package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
)

var DontContinueOnMismatch bool = true

var errors = []error{}

type location struct {
	line, pos int
}

// I wish we could just XOR byte arrays...
// TODO: maybe rewrite later using `unsafe.StringData()`
func strCmp(resStr, expStr string) int {

	resLen, expLen := len(resStr), len(expStr)

	switch strings.Compare(resStr, expStr) {
	case -1:
		return resLen - 1
	case +1:
		return expLen - 1
	}

	if resLen == 0 { // strings should be equal I think
		return -1
	}

	for i := 0; i < len(expStr); i++ {
		if resStr[i] != expStr[i] {
			return i
		}
	}

	return -1
}

// Compares output of the program to the expected
// manually created one
func compareOutputs(result, expected string) ([]location, error) {
	mismatches := make([]location, 0)

	fresult, err := os.Open(result)
	if err != nil {
		return nil, err
	}
	fexpected, err := os.Open(expected)
	if err != nil {
		return nil, err
	}

	resScanner := bufio.NewScanner(fresult)
	expScanner := bufio.NewScanner(fexpected)

	duoScan := func() int {
		f := resScanner.Scan()
		e := expScanner.Scan()

		if f && e {
			return 0
		}
		if !f && !e {
			return 1
		}
		return -1
	}

	atLine := 0
	for ok := duoScan(); ok == 0; atLine++ {
		resLine := resScanner.Text()
		expLine := expScanner.Text()

		if atPos := strCmp(resLine, expLine); atPos != -1 {
			mismatches = append(mismatches, location{atLine, atPos})
			if DontContinueOnMismatch {
				return mismatches, nil
			}
		}
	}
	return mismatches, nil
}

func storeError(err error) {
	errors = append(errors, err)
}

func TestMain(t *testing.T) {
	Run("test_input/project/", "build.json", func(err error) { log.Fatal(err) })

	mismatches, err := compareOutputs(
		"test_input/project/output/laptop.vdf",
		"test_input/project/output/expected.vdf",
	)

	fmt.Println(mismatches, err)

	// Run("test_input/project/", storeError)
	// t.Log("\n", errors)
}
