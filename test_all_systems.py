#!/usr/bin/env python3
"""
Comprehensive test suite for all microservices
Tests: TinyURL, Newsfeed, LoadBalancer, Typeahead, Messaging, DNS, WebCrawler, GoogleDocs, Quora
"""

import subprocess
import os
import json
from pathlib import Path
from typing import Dict, List

class SystemTester:
    def __init__(self, project_root: str = "."):
        self.project_root = Path(project_root).resolve()
        self.services_dir = self.project_root / "services"
        self.sample_app_dir = self.project_root / "sample-app"
        self.results = {}
        
    def run_command(self, cmd: List[str], cwd: str = None) -> tuple:
        """Run a command and return exit code, stdout, stderr"""
        try:
            result = subprocess.run(
                cmd,
                cwd=cwd or self.project_root,
                capture_output=True,
                text=True,
                timeout=60
            )
            return result.returncode, result.stdout, result.stderr
        except subprocess.TimeoutExpired:
            return -1, "", "Command timed out"
        except Exception as e:
            return -1, "", str(e)
    
    def test_service(self, service_name: str) -> Dict:
        """Test a single service"""
        print(f"\n🧪 Testing {service_name}...")
        service_dir = self.services_dir / service_name
        
        if not service_dir.exists():
            print(f"❌ Service directory not found: {service_dir}")
            return {"success": False, "coverage": 0, "error": "Directory not found"}
        
        # Run unit tests with coverage
        cmd = ["go", "test", "-tags=unit", "-v", "-coverprofile=coverage.out", "./..."]
        exit_code, stdout, stderr = self.run_command(cmd, cwd=str(service_dir))
        
        result = {
            "exit_code": exit_code,
            "stdout": stdout,
            "stderr": stderr,
            "success": exit_code == 0
        }
        
        # Extract coverage
        if exit_code == 0:
            for line in stdout.split('\n'):
                if "coverage:" in line and "%" in line:
                    try:
                        coverage_str = line.split("coverage:")[1].split("%")[0].strip()
                        result["coverage"] = float(coverage_str)
                        print(f"✅ {service_name}: {result['coverage']}% coverage")
                    except (ValueError, IndexError):
                        result["coverage"] = 0
        else:
            result["coverage"] = 0
            print(f"❌ {service_name}: Tests failed")
            print(f"Error: {stderr}")
        
        return result
    
    def test_sample_app(self) -> Dict:
        """Test the sample-app"""
        print(f"\n🧪 Testing sample-app...")
        
        cmd = ["go", "test", "-tags=unit", "-v", "-coverprofile=coverage.out", "./..."]
        exit_code, stdout, stderr = self.run_command(cmd, cwd=str(self.sample_app_dir))
        
        result = {
            "exit_code": exit_code,
            "stdout": stdout,
            "stderr": stderr,
            "success": exit_code == 0
        }
        
        # Extract coverage
        if exit_code == 0:
            for line in stdout.split('\n'):
                if "coverage:" in line and "%" in line:
                    try:
                        coverage_str = line.split("coverage:")[1].split("%")[0].strip()
                        result["coverage"] = float(coverage_str)
                        print(f"✅ sample-app: {result['coverage']}% coverage")
                    except (ValueError, IndexError):
                        result["coverage"] = 0
        else:
            result["coverage"] = 0
            print(f"❌ sample-app: Tests failed")
        
        return result
    
    def test_all_services(self) -> Dict:
        """Test all services"""
        services = [
            "tinyurl",
            "newsfeed",
            "loadbalancer",
            "typeahead",
            "messaging",
            "dns",
            "webcrawler",
            "googledocs",
            "quora"
        ]
        
        print("=" * 80)
        print("🚀 COMPREHENSIVE SYSTEM TEST SUITE")
        print("=" * 80)
        
        # Test sample-app
        self.results["sample-app"] = self.test_sample_app()
        
        # Test all services
        for service in services:
            self.results[service] = self.test_service(service)
        
        # Calculate overall statistics
        total_services = len(services) + 1  # +1 for sample-app
        passed_services = sum(1 for r in self.results.values() if r.get("success", False))
        total_coverage = sum(r.get("coverage", 0) for r in self.results.values())
        avg_coverage = total_coverage / total_services if total_services > 0 else 0
        
        self.results["summary"] = {
            "total_services": total_services,
            "passed_services": passed_services,
            "failed_services": total_services - passed_services,
            "average_coverage": avg_coverage,
            "overall_status": "PASSED" if passed_services == total_services else "FAILED"
        }
        
        return self.results
    
    def print_summary(self):
        """Print test summary"""
        print("\n" + "=" * 80)
        print("📊 TEST SUMMARY")
        print("=" * 80)
        
        summary = self.results.get("summary", {})
        
        print(f"\n🎯 Overall Status: {summary.get('overall_status', 'UNKNOWN')}")
        print(f"✅ Passed: {summary.get('passed_services', 0)}/{summary.get('total_services', 0)}")
        print(f"❌ Failed: {summary.get('failed_services', 0)}")
        print(f"📈 Average Coverage: {summary.get('average_coverage', 0):.1f}%")
        
        print("\n📋 Individual Service Results:")
        print("-" * 80)
        
        for service, result in sorted(self.results.items()):
            if service == "summary":
                continue
            
            status = "✅ PASS" if result.get("success", False) else "❌ FAIL"
            coverage = result.get("coverage", 0)
            print(f"{service:20s} {status:10s} Coverage: {coverage:5.1f}%")
        
        print("=" * 80)
        
        # Coverage breakdown
        print("\n📊 Coverage Breakdown:")
        print("-" * 80)
        
        excellent = []
        good = []
        acceptable = []
        needs_improvement = []
        
        for service, result in self.results.items():
            if service == "summary":
                continue
            coverage = result.get("coverage", 0)
            if coverage >= 80:
                excellent.append((service, coverage))
            elif coverage >= 70:
                good.append((service, coverage))
            elif coverage >= 60:
                acceptable.append((service, coverage))
            else:
                needs_improvement.append((service, coverage))
        
        if excellent:
            print(f"\n🎉 Excellent (≥80%): {len(excellent)} services")
            for service, cov in excellent:
                print(f"   {service}: {cov:.1f}%")
        
        if good:
            print(f"\n✅ Good (70-79%): {len(good)} services")
            for service, cov in good:
                print(f"   {service}: {cov:.1f}%")
        
        if acceptable:
            print(f"\n⚠️  Acceptable (60-69%): {len(acceptable)} services")
            for service, cov in acceptable:
                print(f"   {service}: {cov:.1f}%")
        
        if needs_improvement:
            print(f"\n❗ Needs Improvement (<60%): {len(needs_improvement)} services")
            for service, cov in needs_improvement:
                print(f"   {service}: {cov:.1f}%")
        
        print("\n" + "=" * 80)
    
    def save_results(self, output_file: str = "test_all_systems_results.json"):
        """Save results to JSON file"""
        output_path = self.project_root / output_file
        with open(output_path, 'w') as f:
            json.dump(self.results, f, indent=2)
        print(f"\n💾 Results saved to: {output_path}")

def main():
    tester = SystemTester()
    tester.test_all_services()
    tester.print_summary()
    tester.save_results()
    
    # Exit with appropriate code
    summary = tester.results.get("summary", {})
    exit_code = 0 if summary.get("overall_status") == "PASSED" else 1
    exit(exit_code)

if __name__ == "__main__":
    main()

