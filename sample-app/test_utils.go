//go:build test
// +build test

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
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestConfig represents test configuration
type TestConfig struct {
	TestSettings struct {
		TimeoutSeconds      int    `json:"timeout_seconds"`
		ConcurrentRequests  int    `json:"concurrent_requests"`
		StressTestDuration  string `json:"stress_test_duration"`
		BenchmarkIterations int    `json:"benchmark_iterations"`
	} `json:"test_settings"`
	TestData struct {
		ValidInstance   Instance `json:"valid_instance"`
		InvalidInstance Instance `json:"invalid_instance"`
	} `json:"test_data"`
	TestScenarios struct {
		UnitTests struct {
			Enabled           bool `json:"enabled"`
			CoverageThreshold int  `json:"coverage_threshold"`
		} `json:"unit_tests"`
		IntegrationTests struct {
			Enabled bool   `json:"enabled"`
			Timeout string `json:"timeout"`
		} `json:"integration_tests"`
		BenchmarkTests struct {
			Enabled       bool `json:"enabled"`
			MinIterations int  `json:"min_iterations"`
		} `json:"benchmark_tests"`
		StressTests struct {
			Enabled         bool   `json:"enabled"`
			Duration        string `json:"duration"`
			ConcurrentUsers int    `json:"concurrent_users"`
		} `json:"stress_tests"`
	} `json:"test_scenarios"`
}

// LoadTestConfig loads test configuration from testdata/test_config.json
func LoadTestConfig(t *testing.T) *TestConfig {
	configPath := filepath.Join("testdata", "test_config.json")
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	var config TestConfig
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("Failed to parse test config: %v", err)
	}

	return &config
}

// LoadTestInstances loads test instances from testdata/test_instances.json
func LoadTestInstances(t *testing.T) []Instance {
	instancesPath := filepath.Join("testdata", "test_instances.json")
	data, err := ioutil.ReadFile(instancesPath)
	if err != nil {
		t.Fatalf("Failed to load test instances: %v", err)
	}

	var instances []Instance
	if err := json.Unmarshal(data, &instances); err != nil {
		t.Fatalf("Failed to parse test instances: %v", err)
	}

	return instances
}

// CreateTestInstance creates a test instance with the given parameters
func CreateTestInstance(name, version, id, zone, project, hostname, internalIP, externalIP string) *Instance {
	return &Instance{
		Name:       name,
		Version:    version,
		Id:         id,
		Zone:       zone,
		Project:    project,
		Hostname:   hostname,
		InternalIP: internalIP,
		ExternalIP: externalIP,
	}
}

// CreateValidTestInstance creates a valid test instance
func CreateValidTestInstance() *Instance {
	return CreateTestInstance(
		"test-instance",
		"1.0.0",
		"test-123",
		"us-central1-a",
		"test-project",
		"test.example.com",
		"10.0.0.1",
		"35.192.0.1",
	)
}

// CreateInvalidTestInstance creates an invalid test instance
func CreateInvalidTestInstance() *Instance {
	return CreateTestInstance(
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
	)
}

// WaitForCondition waits for a condition to be true or timeout
func WaitForCondition(condition func() bool, timeout time.Duration, interval time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(interval)
	}
	return false
}

// IsTestEnvironment checks if we're running in a test environment
func IsTestEnvironment() bool {
	return os.Getenv("GO_TESTING") == "1" || os.Getenv("TEST_ENV") == "1"
}

// SetTestEnvironment sets the test environment variable
func SetTestEnvironment() {
	os.Setenv("GO_TESTING", "1")
	os.Setenv("TEST_ENV", "1")
}

// CleanupTestEnvironment cleans up test environment variables
func CleanupTestEnvironment() {
	os.Unsetenv("GO_TESTING")
	os.Unsetenv("TEST_ENV")
}

// TestHelper provides common test helper functions
type TestHelper struct {
	t *testing.T
}

// NewTestHelper creates a new test helper
func NewTestHelper(t *testing.T) *TestHelper {
	return &TestHelper{t: t}
}

// AssertEqual asserts that two values are equal
func (th *TestHelper) AssertEqual(expected, actual interface{}, message string) {
	if expected != actual {
		th.t.Errorf("%s: expected %v, got %v", message, expected, actual)
	}
}

// AssertNotEqual asserts that two values are not equal
func (th *TestHelper) AssertNotEqual(expected, actual interface{}, message string) {
	if expected == actual {
		th.t.Errorf("%s: expected %v to not equal %v", message, expected, actual)
	}
}

// AssertTrue asserts that a condition is true
func (th *TestHelper) AssertTrue(condition bool, message string) {
	if !condition {
		th.t.Errorf("%s: expected true, got false", message)
	}
}

// AssertFalse asserts that a condition is false
func (th *TestHelper) AssertFalse(condition bool, message string) {
	if condition {
		th.t.Errorf("%s: expected false, got true", message)
	}
}

// AssertNil asserts that a value is nil
func (th *TestHelper) AssertNil(value interface{}, message string) {
	if value != nil {
		th.t.Errorf("%s: expected nil, got %v", message, value)
	}
}

// AssertNotNil asserts that a value is not nil
func (th *TestHelper) AssertNotNil(value interface{}, message string) {
	if value == nil {
		th.t.Errorf("%s: expected not nil, got nil", message)
	}
}

// AssertContains asserts that a string contains a substring
func (th *TestHelper) AssertContains(str, substr, message string) {
	if !contains(str, substr) {
		th.t.Errorf("%s: expected '%s' to contain '%s'", message, str, substr)
	}
}

// AssertNotContains asserts that a string does not contain a substring
func (th *TestHelper) AssertNotContains(str, substr, message string) {
	if contains(str, substr) {
		th.t.Errorf("%s: expected '%s' to not contain '%s'", message, str, substr)
	}
}

// Helper function to check if a string contains a substring
func contains(str, substr string) bool {
	return len(str) >= len(substr) && str[:len(substr)] == substr ||
		len(str) > len(substr) && contains(str[1:], substr)
}

// TestDataManager manages test data
type TestDataManager struct {
	config    *TestConfig
	instances []Instance
}

// NewTestDataManager creates a new test data manager
func NewTestDataManager(t *testing.T) *TestDataManager {
	return &TestDataManager{
		config:    LoadTestConfig(t),
		instances: LoadTestInstances(t),
	}
}

// GetConfig returns the test configuration
func (tdm *TestDataManager) GetConfig() *TestConfig {
	return tdm.config
}

// GetInstances returns the test instances
func (tdm *TestDataManager) GetInstances() []Instance {
	return tdm.instances
}

// GetValidInstance returns a valid test instance
func (tdm *TestDataManager) GetValidInstance() *Instance {
	return &tdm.config.TestData.ValidInstance
}

// GetInvalidInstance returns an invalid test instance
func (tdm *TestDataManager) GetInvalidInstance() *Instance {
	return &tdm.config.TestData.InvalidInstance
}

// GetRandomInstance returns a random test instance
func (tdm *TestDataManager) GetRandomInstance() *Instance {
	if len(tdm.instances) == 0 {
		return CreateValidTestInstance()
	}
	return &tdm.instances[0] // For simplicity, return the first instance
}
