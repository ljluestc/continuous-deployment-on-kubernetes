#!/usr/bin/env python3
"""
Comprehensive Test Orchestrator for Algorithm Visualization Project
This script runs all types of tests and generates comprehensive reports.
"""

import os
import sys
import subprocess
import json
import time
from pathlib import Path
from datetime import datetime
from typing import Dict, List, Any, Optional

class TestOrchestrator:
    def __init__(self, project_root: str):
        self.project_root = Path(project_root)
        self.results = {
            "timestamp": datetime.now().isoformat(),
            "project": "algorithm-visualization",
            "tests": {},
            "coverage": {},
            "benchmarks": {},
            "summary": {}
        }
        
    def run_command(self, cmd: List[str], cwd: Optional[str] = None) -> tuple:
        """Run a command and return exit code, stdout, stderr"""
        try:
            result = subprocess.run(
                cmd,
                cwd=cwd or self.project_root,
                capture_output=True,
                text=True,
                timeout=300  # 5 minute timeout
            )
            return result.returncode, result.stdout, result.stderr
        except subprocess.TimeoutExpired:
            return -1, "", "Command timed out"
        except Exception as e:
            return -1, "", str(e)
    
    def run_go_tests(self, test_path: str, test_type: str) -> Dict[str, Any]:
        """Run Go tests for a specific path and type"""
        print(f"🧪 Running {test_type} tests: {test_path}")
        
        # Run tests with coverage
        cmd = ["go", "test", "-v", "-cover", "-coverprofile=coverage.out", test_path]
        exit_code, stdout, stderr = self.run_command(cmd)
        
        coverage_percent = 0.0
        if exit_code == 0 and (self.project_root / "coverage.out").exists():
            # Get coverage percentage
            coverage_cmd = ["go", "tool", "cover", "-func=coverage.out"]
            _, cover_stdout, _ = self.run_command(coverage_cmd)
            
            for line in cover_stdout.split('\n'):
                if "total:" in line:
                    parts = line.split('\t')
                    if len(parts) >= 3:
                        try:
                            coverage_percent = float(parts[2].replace('%', ''))
                        except ValueError:
                            pass
                    break
        
        result = {
            "type": test_type,
            "path": test_path,
            "exit_code": exit_code,
            "success": exit_code == 0,
            "coverage": coverage_percent,
            "stdout": stdout,
            "stderr": stderr
        }
        
        if exit_code == 0:
            print(f"✅ {test_type} tests passed: {coverage_percent:.1f}% coverage")
        else:
            print(f"❌ {test_type} tests failed: {stderr}")
            
        return result
    
    def run_benchmarks(self, benchmark_path: str) -> Dict[str, Any]:
        """Run Go benchmarks"""
        print(f"📊 Running benchmarks: {benchmark_path}")
        
        cmd = ["go", "test", "-bench=.", "-benchmem", "-run=^$", benchmark_path]
        exit_code, stdout, stderr = self.run_command(cmd)
        
        result = {
            "path": benchmark_path,
            "exit_code": exit_code,
            "success": exit_code == 0,
            "output": stdout,
            "error": stderr
        }
        
        if exit_code == 0:
            print(f"✅ Benchmarks completed")
        else:
            print(f"❌ Benchmarks failed: {stderr}")
            
        return result
    
    def run_race_tests(self, test_path: str) -> Dict[str, Any]:
        """Run Go tests with race detection"""
        print(f"🔍 Running race detection tests: {test_path}")
        
        cmd = ["go", "test", "-race", "-v", test_path]
        exit_code, stdout, stderr = self.run_command(cmd)
        
        result = {
            "path": test_path,
            "exit_code": exit_code,
            "success": exit_code == 0,
            "output": stdout,
            "error": stderr
        }
        
        if exit_code == 0:
            print(f"✅ Race detection tests passed")
        else:
            print(f"❌ Race detection tests failed: {stderr}")
            
        return result
    
    def run_static_analysis(self) -> Dict[str, Any]:
        """Run static analysis tools"""
        print("🔧 Running static analysis...")
        
        results = {}
        
        # Go vet
        exit_code, stdout, stderr = self.run_command(["go", "vet", "./..."])
        results["go_vet"] = {
            "exit_code": exit_code,
            "success": exit_code == 0,
            "output": stdout,
            "error": stderr
        }
        
        # Go fmt check
        exit_code, stdout, stderr = self.run_command(["gofmt", "-l", "."])
        results["go_fmt"] = {
            "exit_code": exit_code,
            "success": len(stdout.strip()) == 0,
            "output": stdout,
            "error": stderr
        }
        
        if results["go_vet"]["success"]:
            print("✅ Go vet passed")
        else:
            print("❌ Go vet found issues")
            
        if results["go_fmt"]["success"]:
            print("✅ Code formatting is correct")
        else:
            print("❌ Code formatting issues found")
            
        return results
    
    def generate_coverage_report(self) -> Dict[str, Any]:
        """Generate comprehensive coverage report"""
        print("📊 Generating coverage report...")
        
        # Run tests with coverage for all packages
        cmd = ["go", "test", "-coverprofile=coverage.out", "-covermode=atomic", "./..."]
        exit_code, stdout, stderr = self.run_command(cmd)
        
        if exit_code != 0:
            return {"error": "Failed to generate coverage", "stderr": stderr}
        
        # Generate HTML coverage report
        html_cmd = ["go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html"]
        self.run_command(html_cmd)
        
        # Get coverage summary
        func_cmd = ["go", "tool", "cover", "-func=coverage.out"]
        _, func_stdout, _ = self.run_command(func_cmd)
        
        coverage_data = {
            "total_coverage": 0.0,
            "package_coverage": {},
            "functions": []
        }
        
        for line in func_stdout.split('\n'):
            if line.strip():
                parts = line.split('\t')
                if len(parts) >= 3:
                    if "total:" in line:
                        try:
                            coverage_data["total_coverage"] = float(parts[2].replace('%', ''))
                        except ValueError:
                            pass
                    else:
                        func_info = {
                            "function": parts[0],
                            "statements": parts[1],
                            "coverage": parts[2]
                        }
                        coverage_data["functions"].append(func_info)
                        
                        # Extract package name
                        pkg_name = parts[0].split('/')[-1].split(':')[0]
                        if pkg_name not in coverage_data["package_coverage"]:
                            coverage_data["package_coverage"][pkg_name] = []
                        coverage_data["package_coverage"][pkg_name].append(func_info)
        
        print(f"📈 Total coverage: {coverage_data['total_coverage']:.1f}%")
        return coverage_data
    
    def run_all_tests(self) -> Dict[str, Any]:
        """Run all tests and return comprehensive results"""
        print("🚀 Starting comprehensive test suite...")
        print("=" * 60)
        
        # Ensure we're in the right directory
        if not (self.project_root / "go.mod").exists():
            print(f"❌ Error: go.mod not found in {self.project_root}")
            return {"error": "Invalid project directory"}
        
        # Download dependencies
        print("📦 Downloading dependencies...")
        exit_code, _, stderr = self.run_command(["go", "mod", "tidy"])
        if exit_code != 0:
            print(f"❌ Failed to download dependencies: {stderr}")
            return {"error": "Failed to download dependencies"}
        print("✅ Dependencies downloaded")
        print()
        
        # Run unit tests
        test_paths = [
            ("./tests/unit/...", "unit"),
            ("./tests/integration/...", "integration"),
            ("./tests/performance/...", "performance"),
        ]
        
        for test_path, test_type in test_paths:
            if (self.project_root / test_path.replace("./", "").replace("/...", "")).exists():
                result = self.run_go_tests(test_path, test_type)
                self.results["tests"][test_type] = result
            else:
                print(f"⚠️  Skipping {test_type} tests: path not found")
        print()
        
        # Run benchmarks
        benchmark_paths = ["./tests/unit/...", "./tests/performance/..."]
        for benchmark_path in benchmark_paths:
            if (self.project_root / benchmark_path.replace("./", "").replace("/...", "")).exists():
                result = self.run_benchmarks(benchmark_path)
                self.results["benchmarks"][benchmark_path] = result
        print()
        
        # Run race detection
        race_result = self.run_race_tests("./...")
        self.results["race_tests"] = race_result
        print()
        
        # Run static analysis
        static_results = self.run_static_analysis()
        self.results["static_analysis"] = static_results
        print()
        
        # Generate coverage report
        coverage_data = self.generate_coverage_report()
        self.results["coverage"] = coverage_data
        print()
        
        # Generate summary
        self.generate_summary()
        
        return self.results
    
    def generate_summary(self):
        """Generate test summary"""
        total_tests = len(self.results["tests"])
        passed_tests = sum(1 for test in self.results["tests"].values() if test["success"])
        
        total_coverage = self.results["coverage"].get("total_coverage", 0.0)
        
        self.results["summary"] = {
            "total_tests": total_tests,
            "passed_tests": passed_tests,
            "failed_tests": total_tests - passed_tests,
            "success_rate": (passed_tests / total_tests * 100) if total_tests > 0 else 0,
            "total_coverage": total_coverage,
            "race_tests_passed": self.results.get("race_tests", {}).get("success", False),
            "static_analysis_passed": all(
                result["success"] for result in self.results.get("static_analysis", {}).values()
            )
        }
    
    def save_results(self, filename: str = None):
        """Save results to JSON file"""
        if filename is None:
            timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
            filename = f"test_results_{timestamp}.json"
        
        filepath = self.project_root / filename
        with open(filepath, 'w') as f:
            json.dump(self.results, f, indent=2)
        
        print(f"💾 Results saved to: {filepath}")
        return filepath
    
    def print_summary(self):
        """Print test summary"""
        summary = self.results["summary"]
        
        print("=" * 60)
        print("📊 TEST SUMMARY")
        print("=" * 60)
        
        print(f"🎯 Overall Status: {'PASSED' if summary['success_rate'] == 100 else 'FAILED'}")
        print(f"✅ Passed: {summary['passed_tests']}/{summary['total_tests']}")
        print(f"❌ Failed: {summary['failed_tests']}/{summary['total_tests']}")
        print(f"📈 Success Rate: {summary['success_rate']:.1f}%")
        print(f"📊 Total Coverage: {summary['total_coverage']:.1f}%")
        print(f"🔍 Race Tests: {'PASSED' if summary['race_tests_passed'] else 'FAILED'}")
        print(f"🔧 Static Analysis: {'PASSED' if summary['static_analysis_passed'] else 'FAILED'}")
        
        print("\n📋 Individual Test Results:")
        print("-" * 40)
        for test_type, result in self.results["tests"].items():
            status = "✅ PASS" if result["success"] else "❌ FAIL"
            coverage = f"{result['coverage']:.1f}%" if result["success"] else "N/A"
            print(f"{test_type:15} {status:8} Coverage: {coverage}")
        
        print("=" * 60)

def main():
    if len(sys.argv) > 1:
        project_root = sys.argv[1]
    else:
        project_root = "."
    
    orchestrator = TestOrchestrator(project_root)
    
    try:
        results = orchestrator.run_all_tests()
        
        if "error" in results:
            print(f"❌ Error: {results['error']}")
            sys.exit(1)
        
        orchestrator.print_summary()
        orchestrator.save_results()
        
        # Exit with appropriate code
        if orchestrator.results["summary"]["success_rate"] == 100:
            print("\n🎉 All tests passed!")
            sys.exit(0)
        else:
            print("\n❌ Some tests failed!")
            sys.exit(1)
            
    except KeyboardInterrupt:
        print("\n⏹️  Test execution interrupted by user")
        sys.exit(1)
    except Exception as e:
        print(f"\n💥 Unexpected error: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()

