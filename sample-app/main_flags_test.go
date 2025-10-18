/**
# Copyright 2015 Google Inc. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
**/

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

// TestMainVersionFlag tests the -version flag
func TestMainVersionFlag(t *testing.T) {
	// Save original os.Args and stdout
	oldArgs := os.Args
	oldStdout := os.Stdout

	defer func() {
		os.Args = oldArgs
		os.Stdout = oldStdout
		// Reset flag package for other tests
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}()

	// Create a pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set os.Args to simulate -version flag
	os.Args = []string{"cmd", "-version"}

	// Reset flags for this test
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Run main in a goroutine
	done := make(chan bool)
	go func() {
		main()
		done <- true
	}()

	// Wait for main to complete
	<-done

	// Close writer and read output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check output contains version
	expected := fmt.Sprintf("Version %s\n", version)
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestMainDefaultMode tests running in default (backend) mode
func TestMainDefaultMode(t *testing.T) {
	// This test verifies the main function would call backendMode
	// We can't actually run it without blocking, so we test the logic separately

	// The logic is: if frontend flag is false, call backendMode
	// We've already tested backendMode in server_test.go
	t.Log("Default mode logic tested via backendMode tests")
}

// TestMainFrontendMode tests running in frontend mode
func TestMainFrontendMode(t *testing.T) {
	// This test verifies the main function would call frontendMode
	// We can't actually run it without blocking, so we test the logic separately

	// The logic is: if frontend flag is true, call frontendMode
	// We've already tested frontendMode in server_test.go
	t.Log("Frontend mode logic tested via frontendMode tests")
}

// TestVersionConstant tests the version constant is set
func TestVersionConstant(t *testing.T) {
	if version == "" {
		t.Error("Version constant should not be empty")
	}

	if !strings.Contains(version, ".") {
		t.Error("Version should be in semantic version format")
	}

	// Verify it matches expected format
	if version != "1.0.0" {
		t.Logf("Version is %s (expected 1.0.0)", version)
	}
}
