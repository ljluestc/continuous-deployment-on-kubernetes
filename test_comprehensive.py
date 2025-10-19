#!/usr/bin/env python3
"""
Comprehensive Test Suite for Kubernetes Continuous Deployment Project
This script orchestrates unit tests, integration tests, benchmarks, and coverage analysis.
"""

import os
import sys
import subprocess
import json
import time
import argparse
from pathlib import Path
from typing import Dict, List, Tuple, Optional

class TestRunner:
    def __init__(self, project_root: str = "."):
        self.project_root = Path(project_root).resolve()
        self.sample_app_dir = self.project_root / "sample-app"
        self.results = {
            "unit_tests": {},
            "integration_tests": {},
            "benchmarks": {},
            "coverage": {},
            "static_analysis": {},
            "overall_status": "PENDING"
        }
        
    def run_command(self, cmd: List[str], cwd: Optional[Path] = None) -> Tuple[int, str, str]:
        """Run a command and return exit code, stdout, stderr"""
        try:
            result = subprocess.run(
                cmd, 
                cwd=cwd or self.sample_app_dir,
                capture_output=True, 
                text=True, 
                timeout=300
            )
            return result.returncode, result.stdout, result.stderr
        except subprocess.TimeoutExpired:
            return -1, "", "Command timed out"
        except Exception as e:
            return -1, "", str(e)
    
    def check_go_installation(self) -> bool:
        """Check if Go is installed and accessible"""
        print("🔍 Checking Go installation...")
        exit_code, stdout, stderr = self.run_command(["go", "version"])
        if exit_code == 0:
            print(f"✅ Go installed: {stdout.strip()}")
            return True
        else:
            print(f"❌ Go not found: {stderr}")
            return False
    
    def run_unit_tests(self) -> Dict:
        """Run unit tests with coverage"""
        print("\n🧪 Running unit tests...")
        
        # Run tests with coverage
        cmd = [
            "go", "test", "-v", "-race", 
            "-coverprofile=coverage.out", 
            "-covermode=atomic",
            "-mod=readonly",
            "./..."
        ]
        
        exit_code, stdout, stderr = self.run_command(cmd)
        
        result = {
            "exit_code": exit_code,
            "stdout": stdout,
            "stderr": stderr,
            "success": exit_code == 0
        }
        
        if exit_code == 0:
            print("✅ Unit tests passed")
            # Parse test output for summary
            lines = stdout.split('\n')
            for line in lines:
                if "coverage:" in line:
                    print(f"📊 {line.strip()}")
                    # Extract coverage percentage
                    if "coverage:" in line and "%" in line:
                        try:
                            coverage_str = line.split("coverage:")[1].split("%")[0].strip()
                            coverage = float(coverage_str)
                            if coverage >= 90:
                                print(f"🎉 Excellent coverage: {coverage}%")
                            elif coverage >= 80:
                                print(f"✅ Good coverage: {coverage}%")
                            elif coverage >= 70:
                                print(f"⚠️  Acceptable coverage: {coverage}% (target: 80%)")
                            else:
                                print(f"❌ Low coverage: {coverage}% (minimum: 70%)")
                        except (ValueError, IndexError):
                            pass
        else:
            print(f"❌ Unit tests failed: {stderr}")
        
        return result
    
    def run_integration_tests(self) -> Dict:
        """Run integration tests"""
        print("\n🔗 Running integration tests...")
        
        # Run integration tests specifically
        cmd = [
            "go", "test", "-v", "-race",
            "-tags", "integration",
            "-mod=readonly",
            "./..."
        ]
        
        exit_code, stdout, stderr = self.run_command(cmd)
        
        result = {
            "exit_code": exit_code,
            "stdout": stdout,
            "stderr": stderr,
            "success": exit_code == 0
        }
        
        if exit_code == 0:
            print("✅ Integration tests passed")
        else:
            print(f"❌ Integration tests failed: {stderr}")
        
        return result
    
    def run_benchmarks(self) -> Dict:
        """Run benchmark tests"""
        print("\n⚡ Running benchmarks...")
        
        # Run benchmarks separately to avoid test setup issues
        cmd = [
            "go", "test", "-run=^$", "-bench=.", 
            "-benchmem", "-mod=readonly",
            "./..."
        ]
        
        exit_code, stdout, stderr = self.run_command(cmd)
        
        result = {
            "exit_code": exit_code,
            "stdout": stdout,
            "stderr": stderr,
            "success": exit_code == 0
        }
        
        if exit_code == 0:
            print("✅ Benchmarks completed")
            # Extract benchmark results
            lines = stdout.split('\n')
            for line in lines:
                if "Benchmark" in line and "ns/op" in line:
                    print(f"📈 {line.strip()}")
        else:
            print(f"❌ Benchmarks failed: {stderr}")
        
        return result
    
    def run_security_tests(self) -> Dict:
        """Run security tests"""
        print("\n🔒 Running security tests...")
        
        # Run security tests specifically
        cmd = [
            "go", "test", "-v", "-race",
            "-tags", "security",
            "-mod=readonly",
            "./..."
        ]
        
        exit_code, stdout, stderr = self.run_command(cmd)
        
        result = {
            "exit_code": exit_code,
            "stdout": stdout,
            "stderr": stderr,
            "success": exit_code == 0
        }
        
        if exit_code == 0:
            print("✅ Security tests passed")
        else:
            print(f"❌ Security tests failed: {stderr}")
        
        return result
    
    def run_performance_tests(self) -> Dict:
        """Run performance tests"""
        print("\n⚡ Running performance tests...")
        
        # Run performance tests specifically
        cmd = [
            "go", "test", "-v", "-race",
            "-tags", "performance",
            "-mod=readonly",
            "./..."
        ]
        
        exit_code, stdout, stderr = self.run_command(cmd)
        
        result = {
            "exit_code": exit_code,
            "stdout": stdout,
            "stderr": stderr,
            "success": exit_code == 0
        }
        
        if exit_code == 0:
            print("✅ Performance tests passed")
        else:
            print(f"❌ Performance tests failed: {stderr}")
        
        return result
    
    def generate_coverage_report(self) -> Dict:
        """Generate detailed coverage report"""
        print("\n📊 Generating coverage report...")
        
        # Generate HTML coverage report
        cmd = ["go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html"]
        exit_code, stdout, stderr = self.run_command(cmd)
        
        if exit_code != 0:
            print(f"❌ Failed to generate HTML coverage report: {stderr}")
            return {"success": False, "error": stderr}
        
        # Generate coverage summary
        cmd = ["go", "tool", "cover", "-func=coverage.out"]
        exit_code, stdout, stderr = self.run_command(cmd)
        
        result = {
            "success": exit_code == 0,
            "html_report": "coverage.html",
            "summary": stdout if exit_code == 0 else stderr
        }
        
        if exit_code == 0:
            print("✅ Coverage report generated")
            # Extract total coverage percentage
            lines = stdout.split('\n')
            for line in lines:
                if "total:" in line:
                    print(f"📊 {line.strip()}")
        else:
            print(f"❌ Failed to generate coverage summary: {stderr}")
        
        return result
    
    def run_static_analysis(self) -> Dict:
        """Run static analysis tools"""
        print("\n🔍 Running static analysis...")
        
        results = {}
        
        # Run go vet
        print("  Running go vet...")
        cmd = ["go", "vet", "./..."]
        exit_code, stdout, stderr = self.run_command(cmd)
        results["go_vet"] = {
            "exit_code": exit_code,
            "stdout": stdout,
            "stderr": stderr,
            "success": exit_code == 0
        }
        
        if exit_code == 0:
            print("  ✅ go vet passed")
        else:
            print(f"  ❌ go vet failed: {stderr}")
        
        # Run go fmt check
        print("  Running go fmt check...")
        cmd = ["go", "fmt", "./..."]
        exit_code, stdout, stderr = self.run_command(cmd)
        results["go_fmt"] = {
            "exit_code": exit_code,
            "stdout": stdout,
            "stderr": stderr,
            "success": exit_code == 0
        }
        
        if exit_code == 0:
            print("  ✅ go fmt check passed")
        else:
            print(f"  ❌ go fmt check failed: {stderr}")
        
        # Run go mod tidy check
        print("  Running go mod tidy check...")
        cmd = ["go", "mod", "tidy"]
        exit_code, stdout, stderr = self.run_command(cmd)
        results["go_mod_tidy"] = {
            "exit_code": exit_code,
            "stdout": stdout,
            "stderr": stderr,
            "success": exit_code == 0
        }
        
        if exit_code == 0:
            print("  ✅ go mod tidy check passed")
        else:
            print(f"  ❌ go mod tidy check failed: {stderr}")
        
        return results
    
    def run_all_tests(self) -> Dict:
        """Run all tests and analysis"""
        print("🚀 Starting comprehensive test suite...")
        print(f"📁 Project root: {self.project_root}")
        print(f"📁 Sample app: {self.sample_app_dir}")
        
        # Check Go installation
        if not self.check_go_installation():
            return {"error": "Go not installed"}
        
        # Change to sample-app directory
        os.chdir(self.sample_app_dir)
        
        # Run all test types
        self.results["unit_tests"] = self.run_unit_tests()
        self.results["integration_tests"] = self.run_integration_tests()
        self.results["benchmarks"] = self.run_benchmarks()
        self.results["security_tests"] = self.run_security_tests()
        self.results["performance_tests"] = self.run_performance_tests()
        self.results["coverage"] = self.generate_coverage_report()
        self.results["static_analysis"] = self.run_static_analysis()
        
        # Determine overall status
        all_success = (
            self.results["unit_tests"]["success"] and
            self.results["integration_tests"]["success"] and
            self.results["benchmarks"]["success"] and
            self.results["security_tests"]["success"] and
            self.results["performance_tests"]["success"] and
            self.results["coverage"]["success"]
        )
        
        self.results["overall_status"] = "PASSED" if all_success else "FAILED"
        
        return self.results
    
    def generate_report(self) -> str:
        """Generate a comprehensive test report"""
        report = []
        report.append("# Comprehensive Test Report")
        report.append(f"Generated at: {time.strftime('%Y-%m-%d %H:%M:%S')}")
        report.append(f"Overall Status: {self.results['overall_status']}")
        report.append("")
        
        # Unit Tests
        report.append("## Unit Tests")
        ut = self.results["unit_tests"]
        report.append(f"Status: {'✅ PASSED' if ut['success'] else '❌ FAILED'}")
        if ut['stderr']:
            report.append(f"Errors: {ut['stderr']}")
        report.append("")
        
        # Integration Tests
        report.append("## Integration Tests")
        it = self.results["integration_tests"]
        report.append(f"Status: {'✅ PASSED' if it['success'] else '❌ FAILED'}")
        if it['stderr']:
            report.append(f"Errors: {it['stderr']}")
        report.append("")
        
        # Benchmarks
        report.append("## Benchmarks")
        bm = self.results["benchmarks"]
        report.append(f"Status: {'✅ PASSED' if bm['success'] else '❌ FAILED'}")
        if bm['stdout']:
            report.append("Results:")
            for line in bm['stdout'].split('\n'):
                if 'Benchmark' in line and 'ns/op' in line:
                    report.append(f"  {line.strip()}")
        report.append("")
        
        # Security Tests
        report.append("## Security Tests")
        st = self.results["security_tests"]
        report.append(f"Status: {'✅ PASSED' if st['success'] else '❌ FAILED'}")
        if st['stderr']:
            report.append(f"Errors: {st['stderr']}")
        report.append("")
        
        # Performance Tests
        report.append("## Performance Tests")
        pt = self.results["performance_tests"]
        report.append(f"Status: {'✅ PASSED' if pt['success'] else '❌ FAILED'}")
        if pt['stderr']:
            report.append(f"Errors: {pt['stderr']}")
        report.append("")
        
        # Coverage
        report.append("## Coverage")
        cov = self.results["coverage"]
        report.append(f"Status: {'✅ PASSED' if cov['success'] else '❌ FAILED'}")
        if cov['summary']:
            report.append("Summary:")
            for line in cov['summary'].split('\n'):
                if 'total:' in line or 'coverage:' in line:
                    report.append(f"  {line.strip()}")
        report.append("")
        
        # Static Analysis
        report.append("## Static Analysis")
        sa = self.results["static_analysis"]
        for tool, result in sa.items():
            status = "✅ PASSED" if result['success'] else "❌ FAILED"
            report.append(f"  {tool}: {status}")
            if result['stderr']:
                report.append(f"    Errors: {result['stderr']}")
        report.append("")
        
        return "\n".join(report)
    
    def save_results(self, filename: str = "test_results.json"):
        """Save results to JSON file"""
        with open(filename, 'w') as f:
            json.dump(self.results, f, indent=2)
        print(f"📄 Results saved to {filename}")

def main():
    parser = argparse.ArgumentParser(description="Comprehensive Test Suite")
    parser.add_argument("--project-root", default=".", help="Project root directory")
    parser.add_argument("--output", default="test_results.json", help="Output file for results")
    parser.add_argument("--report", default="test_report.md", help="Output file for markdown report")
    
    args = parser.parse_args()
    
    runner = TestRunner(args.project_root)
    results = runner.run_all_tests()
    
    # Generate and save report
    report = runner.generate_report()
    with open(args.report, 'w') as f:
        f.write(report)
    
    # Save results
    runner.save_results(args.output)
    
    print(f"\n📊 Test Results Summary:")
    print(f"  Overall Status: {results['overall_status']}")
    print(f"  Unit Tests: {'✅' if results['unit_tests']['success'] else '❌'}")
    print(f"  Integration Tests: {'✅' if results['integration_tests']['success'] else '❌'}")
    print(f"  Benchmarks: {'✅' if results['benchmarks']['success'] else '❌'}")
    print(f"  Security Tests: {'✅' if results['security_tests']['success'] else '❌'}")
    print(f"  Performance Tests: {'✅' if results['performance_tests']['success'] else '❌'}")
    print(f"  Coverage: {'✅' if results['coverage']['success'] else '❌'}")
    
    print(f"\n📄 Reports generated:")
    print(f"  JSON Results: {args.output}")
    print(f"  Markdown Report: {args.report}")
    print(f"  HTML Coverage: coverage.html")
    
    # Exit with appropriate code
    sys.exit(0 if results['overall_status'] == 'PASSED' else 1)

if __name__ == "__main__":
    main()