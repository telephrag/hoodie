package main

import (
	"testing"
)

var errors = []error{}

func storeError(err error) {
	errors = append(errors, err)
}

func TestMain(t *testing.T) {

	Run("test_input/", storeError)
	t.Log("\n", errors)
}
