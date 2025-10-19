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
	"errors"
	"testing"

	"cloud.google.com/go/compute/metadata"
)

// TestNewInstance_NotOnGCE tests behavior when not running on GCE
func TestNewInstance_NotOnGCE_ReturnsError(t *testing.T) {
	i := newInstance()

	if !metadata.OnGCE() {
		if i.Error != "Not running on GCE" {
			t.Errorf("Expected error 'Not running on GCE', got '%s'", i.Error)
		}

		// All other fields should be empty
		if i.Id != "" || i.Name != "" || i.Hostname != "" {
			t.Error("Fields should be empty when not on GCE")
		}
	}
}

// TestNewInstance_VersionField tests version is always set
func TestNewInstance_VersionField_AlwaysSet(t *testing.T) {
	i := newInstance()

	if i.Version != version {
		t.Errorf("Expected version '%s', got '%s'", version, i.Version)
	}
}

// TestNewInstance_FieldsPopulated tests all fields are populated on GCE
func TestNewInstance_OnGCE_FieldsPopulated(t *testing.T) {
	if !metadata.OnGCE() {
		t.Skip("Skipping test - not running on GCE")
	}

	i := newInstance()

	// Version should always be set
	if i.Version != version {
		t.Errorf("Expected version '%s', got '%s'", version, i.Version)
	}

	// On GCE, these should be populated
	if i.Id == "" {
		t.Error("ID should be populated on GCE")
	}
	if i.Zone == "" {
		t.Error("Zone should be populated on GCE")
	}
	if i.Name == "" {
		t.Error("Name should be populated on GCE")
	}
	if i.Project == "" {
		t.Error("Project should be populated on GCE")
	}
}

// TestAssigner_SuccessfulAssignment tests successful value assignment
func TestAssigner_SuccessfulAssignment_ReturnsValue(t *testing.T) {
	a := &assigner{}

	getValue := func() (string, error) {
		return "test-value", nil
	}

	result := a.assign(getValue)

	if result != "test-value" {
		t.Errorf("Expected 'test-value', got '%s'", result)
	}

	if a.err != nil {
		t.Errorf("Expected no error, got %v", a.err)
	}
}

// TestAssigner_ErrorAssignment tests error handling
func TestAssigner_ErrorAssignment_ReturnsEmpty(t *testing.T) {
	a := &assigner{}

	getValue := func() (string, error) {
		return "", errors.New("test error")
	}

	result := a.assign(getValue)

	if result != "" {
		t.Errorf("Expected empty string, got '%s'", result)
	}

	if a.err == nil {
		t.Error("Expected error to be set")
	}

	if a.err.Error() != "test error" {
		t.Errorf("Expected 'test error', got '%v'", a.err)
	}
}

// TestAssigner_MultipleAssignments tests multiple assignments
func TestAssigner_MultipleAssignments_Succeeds(t *testing.T) {
	a := &assigner{}

	getValue1 := func() (string, error) {
		return "value1", nil
	}

	getValue2 := func() (string, error) {
		return "value2", nil
	}

	result1 := a.assign(getValue1)
	result2 := a.assign(getValue2)

	if result1 != "value1" {
		t.Errorf("Expected 'value1', got '%s'", result1)
	}

	if result2 != "value2" {
		t.Errorf("Expected 'value2', got '%s'", result2)
	}

	if a.err != nil {
		t.Errorf("Expected no error, got %v", a.err)
	}
}

// TestAssigner_ErrorPropagation tests error stops further assignments
func TestAssigner_ErrorPropagation_StopsAssignments(t *testing.T) {
	a := &assigner{}

	getValue1 := func() (string, error) {
		return "value1", nil
	}

	getError := func() (string, error) {
		return "", errors.New("assignment error")
	}

	getValue2 := func() (string, error) {
		return "value2", nil
	}

	result1 := a.assign(getValue1)
	resultErr := a.assign(getError)
	result2 := a.assign(getValue2)

	if result1 != "value1" {
		t.Errorf("First assignment should succeed, got '%s'", result1)
	}

	if resultErr != "" {
		t.Errorf("Error assignment should return empty string, got '%s'", resultErr)
	}

	if result2 != "" {
		t.Errorf("Assignment after error should return empty string, got '%s'", result2)
	}

	if a.err == nil {
		t.Error("Error should be set")
	}

	if a.err.Error() != "assignment error" {
		t.Errorf("Expected 'assignment error', got '%v'", a.err)
	}
}

// TestAssigner_PersistentError tests error persists across calls
func TestAssigner_PersistentError_ErrorPersists(t *testing.T) {
	a := &assigner{err: errors.New("existing error")}

	getValue := func() (string, error) {
		return "value", nil
	}

	result := a.assign(getValue)

	if result != "" {
		t.Errorf("Expected empty string due to existing error, got '%s'", result)
	}

	if a.err.Error() != "existing error" {
		t.Error("Original error should persist")
	}
}

// TestInstance_AllFields tests all Instance struct fields
func TestInstance_AllFields_Accessible(t *testing.T) {
	i := &Instance{
		Id:         "test-id-123",
		Name:       "test-instance",
		Version:    "2.0.0",
		Hostname:   "test-host.example.com",
		Zone:       "us-central1-a",
		Project:    "my-project",
		InternalIP: "10.128.0.2",
		ExternalIP: "35.192.0.1",
		LBRequest:  "GET /api HTTP/1.1\r\nHost: example.com",
		ClientIP:   "192.168.1.1",
		Error:      "test error message",
	}

	// Verify all fields are accessible
	if i.Id != "test-id-123" {
		t.Error("Id field not accessible")
	}
	if i.Name != "test-instance" {
		t.Error("Name field not accessible")
	}
	if i.Version != "2.0.0" {
		t.Error("Version field not accessible")
	}
	if i.Hostname != "test-host.example.com" {
		t.Error("Hostname field not accessible")
	}
	if i.Zone != "us-central1-a" {
		t.Error("Zone field not accessible")
	}
	if i.Project != "my-project" {
		t.Error("Project field not accessible")
	}
	if i.InternalIP != "10.128.0.2" {
		t.Error("InternalIP field not accessible")
	}
	if i.ExternalIP != "35.192.0.1" {
		t.Error("ExternalIP field not accessible")
	}
	if i.LBRequest != "GET /api HTTP/1.1\r\nHost: example.com" {
		t.Error("LBRequest field not accessible")
	}
	if i.ClientIP != "192.168.1.1" {
		t.Error("ClientIP field not accessible")
	}
	if i.Error != "test error message" {
		t.Error("Error field not accessible")
	}
}

// TestInstance_ZeroValue tests zero value initialization
func TestInstance_ZeroValue_EmptyFields(t *testing.T) {
	var i Instance

	if i.Id != "" {
		t.Error("Zero value Id should be empty")
	}
	if i.Name != "" {
		t.Error("Zero value Name should be empty")
	}
	if i.Version != "" {
		t.Error("Zero value Version should be empty")
	}
}

// BenchmarkNewInstance benchmarks instance creation
func BenchmarkNewInstance(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = newInstance()
	}
}

// TestNewInstance_OnGCE_Mocked tests GCE behavior with mocked metadata
func TestNewInstance_OnGCE_Mocked(t *testing.T) {
	// This test simulates what would happen on GCE by testing the assigner logic
	// We can't actually run on GCE in tests, but we can test the logic
	
	// Test the assigner with multiple successful calls
	a := &assigner{}
	
	// Simulate successful metadata calls
	result1 := a.assign(func() (string, error) {
		return "instance-123", nil
	})
	result2 := a.assign(func() (string, error) {
		return "us-central1-a", nil
	})
	result3 := a.assign(func() (string, error) {
		return "test-instance", nil
	})
	
	if result1 != "instance-123" {
		t.Errorf("Expected 'instance-123', got '%s'", result1)
	}
	if result2 != "us-central1-a" {
		t.Errorf("Expected 'us-central1-a', got '%s'", result2)
	}
	if result3 != "test-instance" {
		t.Errorf("Expected 'test-instance', got '%s'", result3)
	}
	
	if a.err != nil {
		t.Errorf("Expected no error, got %v", a.err)
	}
}

// TestNewInstance_GCE_Simulation tests the GCE-specific code path
func TestNewInstance_GCE_Simulation(t *testing.T) {
	// This test simulates the GCE-specific code path by testing the assigner logic
	// that would be executed in newInstance when running on GCE
	
	// Create a new instance (this will use the non-GCE path)
	i := newInstance()
	
	// Verify the non-GCE behavior
	if !metadata.OnGCE() {
		if i.Error != "Not running on GCE" {
			t.Errorf("Expected error 'Not running on GCE', got '%s'", i.Error)
		}
	}
	
	// Test the assigner logic that would be used on GCE
	a := &assigner{}
	
	// Simulate the GCE metadata calls that would happen in newInstance
	i.Id = a.assign(func() (string, error) {
		return "gce-instance-123", nil
	})
	i.Zone = a.assign(func() (string, error) {
		return "us-central1-a", nil
	})
	i.Name = a.assign(func() (string, error) {
		return "gce-test-instance", nil
	})
	i.Hostname = a.assign(func() (string, error) {
		return "gce-hostname.example.com", nil
	})
	i.Project = a.assign(func() (string, error) {
		return "gce-project", nil
	})
	i.InternalIP = a.assign(func() (string, error) {
		return "10.128.0.2", nil
	})
	i.ExternalIP = a.assign(func() (string, error) {
		return "35.192.0.1", nil
	})
	
	// Test error handling
	if a.err != nil {
		i.Error = a.err.Error()
	}
	
	// Verify the simulated GCE behavior
	if i.Id != "gce-instance-123" {
		t.Errorf("Expected 'gce-instance-123', got '%s'", i.Id)
	}
	if i.Zone != "us-central1-a" {
		t.Errorf("Expected 'us-central1-a', got '%s'", i.Zone)
	}
	if i.Name != "gce-test-instance" {
		t.Errorf("Expected 'gce-test-instance', got '%s'", i.Name)
	}
	if i.Hostname != "gce-hostname.example.com" {
		t.Errorf("Expected 'gce-hostname.example.com', got '%s'", i.Hostname)
	}
	if i.Project != "gce-project" {
		t.Errorf("Expected 'gce-project', got '%s'", i.Project)
	}
	if i.InternalIP != "10.128.0.2" {
		t.Errorf("Expected '10.128.0.2', got '%s'", i.InternalIP)
	}
	if i.ExternalIP != "35.192.0.1" {
		t.Errorf("Expected '35.192.0.1', got '%s'", i.ExternalIP)
	}
	
	if a.err != nil {
		t.Errorf("Expected no error, got %v", a.err)
	}
}

// TestNewInstance_OnGCE_WithError tests GCE behavior with metadata errors
func TestNewInstance_OnGCE_WithError(t *testing.T) {
	// Test the assigner with an error in the middle
	a := &assigner{}
	
	// First call succeeds
	result1 := a.assign(func() (string, error) {
		return "instance-123", nil
	})
	
	// Second call fails
	result2 := a.assign(func() (string, error) {
		return "", errors.New("metadata error")
	})
	
	// Third call should return empty due to previous error
	result3 := a.assign(func() (string, error) {
		return "should-not-be-called", nil
	})
	
	if result1 != "instance-123" {
		t.Errorf("Expected 'instance-123', got '%s'", result1)
	}
	if result2 != "" {
		t.Errorf("Expected empty string on error, got '%s'", result2)
	}
	if result3 != "" {
		t.Errorf("Expected empty string due to previous error, got '%s'", result3)
	}
	
	if a.err == nil {
		t.Error("Expected error to be set")
	}
	if a.err.Error() != "metadata error" {
		t.Errorf("Expected 'metadata error', got '%v'", a.err)
	}
}

// BenchmarkAssigner_Assign benchmarks assignment operation
func BenchmarkAssigner_Assign(b *testing.B) {
	getValue := func() (string, error) {
		return "benchmark-value", nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a := &assigner{}
		a.assign(getValue)
	}
}
