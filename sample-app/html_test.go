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
	"html/template"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestHTML_Constant_NotEmpty tests HTML constant is not empty
func TestHTML_Constant_NotEmpty(t *testing.T) {
	if html == "" {
		t.Error("HTML constant should not be empty")
	}
}

// TestHTML_Constant_ValidHTML tests HTML is valid
func TestHTML_Constant_ValidHTML(t *testing.T) {
	// Check for required HTML structure
	if !strings.Contains(html, "<!doctype html>") {
		t.Error("HTML should contain doctype declaration")
	}

	if !strings.Contains(html, "<html>") {
		t.Error("HTML should contain opening html tag")
	}

	if !strings.Contains(html, "</html>") {
		t.Error("HTML should contain closing html tag")
	}

	if !strings.Contains(html, "<head>") {
		t.Error("HTML should contain head section")
	}

	if !strings.Contains(html, "<body>") {
		t.Error("HTML should contain body section")
	}
}

// TestHTML_Template_Parses tests template parses without error
func TestHTML_Template_Parses(t *testing.T) {
	_, err := template.New("test").Parse(html)
	if err != nil {
		t.Errorf("HTML template should parse without error: %v", err)
	}
}

// TestHTML_Template_HasPlaceholders tests template has expected placeholders
func TestHTML_Template_HasPlaceholders(t *testing.T) {
	expectedPlaceholders := []string{
		"{{.Name}}",
		"{{.Version}}",
		"{{.Id}}",
		"{{.Hostname}}",
		"{{.Zone}}",
		"{{.Project}}",
		"{{.InternalIP}}",
		"{{.ExternalIP}}",
		"{{.ClientIP}}",
		"{{.LBRequest}}",
		"{{.Error}}",
	}

	for _, placeholder := range expectedPlaceholders {
		if !strings.Contains(html, placeholder) {
			t.Errorf("HTML should contain placeholder %s", placeholder)
		}
	}
}

// TestHTML_Template_ExecutesWithData tests template execution
func TestHTML_Template_ExecutesWithData(t *testing.T) {
	tpl, err := template.New("test").Parse(html)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	testData := &Instance{
		Id:         "instance-123",
		Name:       "test-instance",
		Version:    "1.0.0",
		Hostname:   "hostname.example.com",
		Zone:       "us-east1-b",
		Project:    "my-gcp-project",
		InternalIP: "10.0.0.1",
		ExternalIP: "203.0.113.1",
		LBRequest:  "GET / HTTP/1.1",
		ClientIP:   "198.51.100.1",
		Error:      "",
	}

	w := httptest.NewRecorder()
	err = tpl.Execute(w, testData)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	output := w.Body.String()

	// Verify all data is rendered
	expectedValues := []string{
		"instance-123",
		"test-instance",
		"1.0.0",
		"hostname.example.com",
		"us-east1-b",
		"my-gcp-project",
		"10.0.0.1",
		"203.0.113.1",
		"GET / HTTP/1.1",
		"198.51.100.1",
	}

	for _, value := range expectedValues {
		if !strings.Contains(output, value) {
			t.Errorf("Template output should contain '%s'", value)
		}
	}
}

// TestHTML_Template_ExecutesWithEmptyData tests template with empty data
func TestHTML_Template_ExecutesWithEmptyData(t *testing.T) {
	tpl, err := template.New("test").Parse(html)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	emptyData := &Instance{}

	w := httptest.NewRecorder()
	err = tpl.Execute(w, emptyData)
	if err != nil {
		t.Fatalf("Template should execute with empty data: %v", err)
	}

	// Template should still render, just with empty values
	output := w.Body.String()
	if !strings.Contains(output, "<html>") {
		t.Error("Output should still contain HTML structure")
	}
}

// TestHTML_Template_ExecutesWithNilData tests template with nil data
func TestHTML_Template_ExecutesWithNilData(t *testing.T) {
	tpl, err := template.New("test").Parse(html)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	w := httptest.NewRecorder()
	err = tpl.Execute(w, nil)
	if err != nil {
		t.Fatalf("Template should handle nil data: %v", err)
	}

	output := w.Body.String()
	if !strings.Contains(output, "<html>") {
		t.Error("Output should still contain HTML structure")
	}
}

// TestHTML_Template_EscapesHTML tests HTML escaping
func TestHTML_Template_EscapesHTML(t *testing.T) {
	tpl, err := template.New("test").Parse(html)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	dangerousData := &Instance{
		Name:    "<script>alert('XSS')</script>",
		Error:   "<img src=x onerror=alert('XSS')>",
		Version: "1.0.0",
	}

	w := httptest.NewRecorder()
	err = tpl.Execute(w, dangerousData)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	output := w.Body.String()

	// Verify dangerous content is escaped
	if strings.Contains(output, "<script>alert") {
		t.Error("Script tags should be escaped")
	}

	// Check for properly escaped content - Go's html/template uses &#34; for quotes in attributes
	// and &lt; &gt; for angle brackets in text content
	if !strings.Contains(output, "&lt;") || !strings.Contains(output, "&gt;") {
		t.Error("HTML should be escaped with entities")
	}
}

// TestHTML_HasCSS tests HTML includes CSS
func TestHTML_HasCSS(t *testing.T) {
	if !strings.Contains(html, "css") && !strings.Contains(html, "CSS") {
		t.Error("HTML should reference CSS")
	}

	// Check for Materialize CSS
	if !strings.Contains(html, "materialize") {
		t.Error("HTML should include Materialize CSS")
	}
}

// TestHTML_HasTitle tests HTML has title
func TestHTML_HasTitle(t *testing.T) {
	if !strings.Contains(html, "<title>") {
		t.Error("HTML should have a title tag")
	}

	if !strings.Contains(html, "Frontend Web Server") {
		t.Error("HTML title should be 'Frontend Web Server'")
	}
}

// TestHTML_HasCards tests HTML has card structure
func TestHTML_HasCards(t *testing.T) {
	if !strings.Contains(html, "card") {
		t.Error("HTML should contain card elements")
	}

	if !strings.Contains(html, "Backend that serviced this request") {
		t.Error("HTML should contain backend card title")
	}

	if !strings.Contains(html, "Proxy that handled this request") {
		t.Error("HTML should contain proxy card title")
	}
}

// TestHTML_HasTables tests HTML has table structure
func TestHTML_HasTables(t *testing.T) {
	if !strings.Contains(html, "<table") {
		t.Error("HTML should contain table elements")
	}

	if !strings.Contains(html, "<tbody>") {
		t.Error("HTML should contain tbody")
	}

	if !strings.Contains(html, "<tr>") {
		t.Error("HTML should contain table rows")
	}

	if !strings.Contains(html, "<td>") {
		t.Error("HTML should contain table cells")
	}
}

// TestHTML_ResponsiveLayout tests responsive layout classes
func TestHTML_ResponsiveLayout(t *testing.T) {
	if !strings.Contains(html, "container") {
		t.Error("HTML should have container class")
	}

	if !strings.Contains(html, "row") {
		t.Error("HTML should have row class")
	}

	if !strings.Contains(html, "col") {
		t.Error("HTML should have column classes")
	}
}

// BenchmarkHTML_TemplateParse benchmarks template parsing
func BenchmarkHTML_TemplateParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = template.New("bench").Parse(html)
	}
}

// BenchmarkHTML_TemplateExecute benchmarks template execution
func BenchmarkHTML_TemplateExecute(b *testing.B) {
	tpl, _ := template.New("bench").Parse(html)
	data := &Instance{
		Name:    "test",
		Version: "1.0.0",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		tpl.Execute(w, data)
	}
}
