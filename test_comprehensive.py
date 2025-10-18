#!/usr/bin/env python3
"""
Comprehensive Test Orchestration Script for gceme Application

This script orchestrates all testing activities including:
- Unit tests
- Integration tests
- Benchmark tests
- Coverage reporting
- Performance analysis
- Test report generation

Usage:
    python3 test_comprehensive.py --all
    python3 test_comprehensive.py --unit
    python3 test_comprehensive.py --integration
    python3 test_comprehensive.py --benchmark
    python3 test_comprehensive.py --coverage
    python3 test_comprehensive.py --report
"""

import argparse
import os
import subprocess
import sys
import json
import time
from datetime import datetime
from pathlib import Path
import html


class Colors:
    """ANSI color codes for terminal output"""
    HEADER = '\033[95m'
    OKBLUE = '\033[94m'
    OKCYAN = '\033[96m'
    OKGREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'
    UNDERLINE = '\033[4m'


class TestOrchestrator:
    """Main orchestrator for running comprehensive tests"""

    def __init__(self, verbose=False):
        self.verbose = verbose
        self.test_dir = Path(__file__).parent / "sample-app"
        self.report_dir = Path(__file__).parent / "test-reports"
        self.coverage_file = self.test_dir / "coverage.out"
        self.results = {
            "timestamp": datetime.now().isoformat(),
            "tests": {},
            "coverage": {},
            "benchmarks": {},
            "summary": {}
        }

        # Create report directory
        self.report_dir.mkdir(exist_ok=True)

    def log(self, message, level="INFO"):
        """Log a message with color coding"""
        colors = {
            "INFO": Colors.OKBLUE,
            "SUCCESS": Colors.OKGREEN,
            "WARNING": Colors.WARNING,
            "ERROR": Colors.FAIL,
            "HEADER": Colors.HEADER
        }
        color = colors.get(level, Colors.ENDC)
        print(f"{color}[{level}] {message}{Colors.ENDC}")

    def run_command(self, cmd, cwd=None, capture=True):
        """Run a shell command and return output"""
        self.log(f"Running: {cmd}", "INFO")
        try:
            if capture:
                result = subprocess.run(
                    cmd,
                    shell=True,
                    cwd=cwd or self.test_dir,
                    capture_output=True,
                    text=True,
                    timeout=300
                )
                if self.verbose:
                    print(result.stdout)
                if result.stderr:
                    print(result.stderr)
                return result.returncode == 0, result.stdout, result.stderr
            else:
                result = subprocess.run(
                    cmd,
                    shell=True,
                    cwd=cwd or self.test_dir,
                    timeout=300
                )
                return result.returncode == 0, "", ""
        except subprocess.TimeoutExpired:
            self.log("Command timed out!", "ERROR")
            return False, "", "Timeout"
        except Exception as e:
            self.log(f"Command failed: {e}", "ERROR")
            return False, "", str(e)

    def run_unit_tests(self):
        """Run all unit tests with coverage"""
        self.log("Running unit tests...", "HEADER")

        # Run tests with coverage and race detection
        cmd = f"go test -v -race -coverprofile={self.coverage_file} -covermode=atomic ./..."
        success, stdout, stderr = self.run_command(cmd)

        # Parse test results
        test_count = stdout.count("--- PASS:")
        fail_count = stdout.count("--- FAIL:")

        self.results["tests"]["unit"] = {
            "success": success,
            "passed": test_count,
            "failed": fail_count,
            "output": stdout,
            "errors": stderr
        }

        if success:
            self.log(f"Unit tests passed: {test_count} tests", "SUCCESS")
        else:
            self.log(f"Unit tests failed: {fail_count} failures", "ERROR")

        return success

    def run_integration_tests(self):
        """Run integration tests"""
        self.log("Running integration tests...", "HEADER")

        # Run tests with integration tag
        cmd = "go test -v -tags=integration -timeout=30s ./..."
        success, stdout, stderr = self.run_command(cmd)

        test_count = stdout.count("--- PASS:")
        fail_count = stdout.count("--- FAIL:")

        self.results["tests"]["integration"] = {
            "success": success,
            "passed": test_count,
            "failed": fail_count,
            "output": stdout,
            "errors": stderr
        }

        if success:
            self.log(f"Integration tests passed: {test_count} tests", "SUCCESS")
        else:
            self.log(f"Integration tests failed: {fail_count} failures", "ERROR")

        return success

    def run_benchmark_tests(self):
        """Run benchmark tests"""
        self.log("Running benchmark tests...", "HEADER")

        # Run benchmarks with memory stats
        benchmark_file = self.report_dir / "benchmark.txt"
        cmd = f"go test -bench=. -benchmem -benchtime=3s ./... | tee {benchmark_file}"
        success, stdout, stderr = self.run_command(cmd)

        self.results["benchmarks"] = {
            "success": success,
            "output": stdout,
            "file": str(benchmark_file)
        }

        if success:
            self.log("Benchmarks completed successfully", "SUCCESS")
        else:
            self.log("Benchmarks failed", "ERROR")

        return success

    def generate_coverage_report(self):
        """Generate coverage reports in multiple formats"""
        self.log("Generating coverage reports...", "HEADER")

        if not self.coverage_file.exists():
            self.log("Coverage file not found. Run unit tests first.", "ERROR")
            return False

        # Generate HTML coverage report
        html_report = self.report_dir / "coverage.html"
        cmd = f"go tool cover -html={self.coverage_file} -o {html_report}"
        success, _, _ = self.run_command(cmd)

        # Generate function-level coverage
        cmd = f"go tool cover -func={self.coverage_file}"
        success, func_output, _ = self.run_command(cmd)

        # Parse overall coverage percentage
        coverage_pct = 0.0
        for line in func_output.split('\n'):
            if 'total:' in line.lower():
                parts = line.split()
                for part in parts:
                    if '%' in part:
                        coverage_pct = float(part.replace('%', ''))

        # Save function coverage to file
        func_report = self.report_dir / "coverage_functions.txt"
        func_report.write_text(func_output)

        self.results["coverage"] = {
            "percentage": coverage_pct,
            "html_report": str(html_report),
            "function_report": str(func_report),
            "coverage_file": str(self.coverage_file),
            "details": func_output
        }

        self.log(f"Coverage: {coverage_pct}%", "SUCCESS" if coverage_pct == 100 else "WARNING")
        self.log(f"HTML report: {html_report}", "INFO")

        return coverage_pct >= 100.0

    def run_static_analysis(self):
        """Run static analysis tools"""
        self.log("Running static analysis...", "HEADER")

        results = {}

        # go fmt check
        self.log("Checking code formatting...", "INFO")
        success, stdout, _ = self.run_command("gofmt -l .")
        results["fmt"] = {
            "success": len(stdout.strip()) == 0,
            "files_to_format": stdout.strip().split('\n') if stdout.strip() else []
        }

        # go vet
        self.log("Running go vet...", "INFO")
        success, stdout, stderr = self.run_command("go vet ./...")
        results["vet"] = {
            "success": success,
            "output": stdout,
            "errors": stderr
        }

        # go mod verify
        self.log("Verifying go modules...", "INFO")
        success, stdout, stderr = self.run_command("go mod verify")
        results["mod_verify"] = {
            "success": success,
            "output": stdout
        }

        self.results["static_analysis"] = results

        all_passed = all(r["success"] for r in results.values())
        if all_passed:
            self.log("All static analysis checks passed", "SUCCESS")
        else:
            self.log("Some static analysis checks failed", "WARNING")

        return all_passed

    def generate_html_report(self):
        """Generate comprehensive HTML test report"""
        self.log("Generating HTML test report...", "HEADER")

        report_file = self.report_dir / "test_report.html"

        html_content = f"""
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>gceme Test Report</title>
    <style>
        body {{
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            margin: 0;
            padding: 20px;
            background: #f5f5f5;
        }}
        .container {{
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }}
        h1 {{
            color: #2c3e50;
            border-bottom: 3px solid #3498db;
            padding-bottom: 10px;
        }}
        h2 {{
            color: #34495e;
            margin-top: 30px;
            border-left: 4px solid #3498db;
            padding-left: 10px;
        }}
        .metric {{
            display: inline-block;
            margin: 10px 20px 10px 0;
            padding: 15px 25px;
            border-radius: 5px;
            font-size: 14px;
        }}
        .success {{
            background: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
        }}
        .warning {{
            background: #fff3cd;
            color: #856404;
            border: 1px solid #ffeaa7;
        }}
        .error {{
            background: #f8d7da;
            color: #721c24;
            border: 1px solid #f5c6cb;
        }}
        .info {{
            background: #d1ecf1;
            color: #0c5460;
            border: 1px solid #bee5eb;
        }}
        table {{
            width: 100%;
            border-collapse: collapse;
            margin: 20px 0;
        }}
        th, td {{
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }}
        th {{
            background: #3498db;
            color: white;
            font-weight: bold;
        }}
        tr:hover {{
            background: #f5f5f5;
        }}
        pre {{
            background: #2c3e50;
            color: #ecf0f1;
            padding: 15px;
            border-radius: 5px;
            overflow-x: auto;
            font-size: 12px;
        }}
        .badge {{
            display: inline-block;
            padding: 5px 10px;
            border-radius: 3px;
            font-size: 12px;
            font-weight: bold;
        }}
        .badge.pass {{
            background: #27ae60;
            color: white;
        }}
        .badge.fail {{
            background: #e74c3c;
            color: white;
        }}
        .timestamp {{
            color: #7f8c8d;
            font-size: 14px;
        }}
    </style>
</head>
<body>
    <div class="container">
        <h1>gceme Comprehensive Test Report</h1>
        <p class="timestamp">Generated: {self.results['timestamp']}</p>

        <h2>Summary</h2>
        <div>
            <div class="metric {self._get_status_class(self.results.get('coverage', {}).get('percentage', 0) >= 100)}">
                <strong>Coverage:</strong> {self.results.get('coverage', {}).get('percentage', 0):.2f}%
            </div>
            <div class="metric {self._get_status_class(self.results.get('tests', {}).get('unit', {}).get('success', False))}">
                <strong>Unit Tests:</strong> {self._get_status_text(self.results.get('tests', {}).get('unit', {}).get('success', False))}
            </div>
            <div class="metric {self._get_status_class(self.results.get('tests', {}).get('integration', {}).get('success', False))}">
                <strong>Integration Tests:</strong> {self._get_status_text(self.results.get('tests', {}).get('integration', {}).get('success', False))}
            </div>
            <div class="metric {self._get_status_class(self.results.get('benchmarks', {}).get('success', False))}">
                <strong>Benchmarks:</strong> {self._get_status_text(self.results.get('benchmarks', {}).get('success', False))}
            </div>
        </div>

        <h2>Unit Tests</h2>
        <table>
            <tr>
                <th>Status</th>
                <th>Passed</th>
                <th>Failed</th>
            </tr>
            <tr>
                <td><span class="badge {self._get_badge_class(self.results.get('tests', {}).get('unit', {}).get('success', False))}">{self._get_status_text(self.results.get('tests', {}).get('unit', {}).get('success', False))}</span></td>
                <td>{self.results.get('tests', {}).get('unit', {}).get('passed', 0)}</td>
                <td>{self.results.get('tests', {}).get('unit', {}).get('failed', 0)}</td>
            </tr>
        </table>

        <h2>Integration Tests</h2>
        <table>
            <tr>
                <th>Status</th>
                <th>Passed</th>
                <th>Failed</th>
            </tr>
            <tr>
                <td><span class="badge {self._get_badge_class(self.results.get('tests', {}).get('integration', {}).get('success', False))}">{self._get_status_text(self.results.get('tests', {}).get('integration', {}).get('success', False))}</span></td>
                <td>{self.results.get('tests', {}).get('integration', {}).get('passed', 0)}</td>
                <td>{self.results.get('tests', {}).get('integration', {}).get('failed', 0)}</td>
            </tr>
        </table>

        <h2>Code Coverage</h2>
        <div class="metric info">
            <strong>Total Coverage:</strong> {self.results.get('coverage', {}).get('percentage', 0):.2f}%
        </div>
        <p>
            <a href="coverage.html" target="_blank">View detailed HTML coverage report</a><br>
            <a href="coverage_functions.txt" target="_blank">View function-level coverage</a>
        </p>

        <h2>Coverage Details</h2>
        <pre>{html.escape(self.results.get('coverage', {}).get('details', 'No coverage data'))}</pre>

        <h2>Benchmark Results</h2>
        <p><a href="benchmark.txt" target="_blank">View benchmark results</a></p>

        <h2>Static Analysis</h2>
        <table>
            <tr>
                <th>Check</th>
                <th>Status</th>
                <th>Details</th>
            </tr>
            {self._generate_static_analysis_rows()}
        </table>
    </div>
</body>
</html>
"""

        report_file.write_text(html_content)
        self.log(f"HTML report generated: {report_file}", "SUCCESS")
        return report_file

    def _get_status_class(self, success):
        return "success" if success else "error"

    def _get_badge_class(self, success):
        return "pass" if success else "fail"

    def _get_status_text(self, success):
        return "PASS" if success else "FAIL"

    def _generate_static_analysis_rows(self):
        rows = []
        for check, result in self.results.get('static_analysis', {}).items():
            status = self._get_status_text(result.get('success', False))
            badge = self._get_badge_class(result.get('success', False))
            details = result.get('output', '')[:100] if result.get('output') else 'OK'
            rows.append(f"""
            <tr>
                <td>{check}</td>
                <td><span class="badge {badge}">{status}</span></td>
                <td>{html.escape(details)}</td>
            </tr>
            """)
        return '\n'.join(rows)

    def save_results_json(self):
        """Save test results as JSON"""
        json_file = self.report_dir / "test_results.json"
        with open(json_file, 'w') as f:
            json.dump(self.results, f, indent=2)
        self.log(f"JSON results saved: {json_file}", "INFO")

    def print_summary(self):
        """Print a summary of all test results"""
        print("\n" + "="*80)
        self.log("TEST SUMMARY", "HEADER")
        print("="*80)

        # Unit tests
        unit = self.results.get('tests', {}).get('unit', {})
        self.log(f"Unit Tests: {self._get_status_text(unit.get('success', False))} "
                f"({unit.get('passed', 0)} passed, {unit.get('failed', 0)} failed)",
                "SUCCESS" if unit.get('success') else "ERROR")

        # Integration tests
        integration = self.results.get('tests', {}).get('integration', {})
        self.log(f"Integration Tests: {self._get_status_text(integration.get('success', False))} "
                f"({integration.get('passed', 0)} passed, {integration.get('failed', 0)} failed)",
                "SUCCESS" if integration.get('success') else "WARNING")

        # Coverage
        coverage = self.results.get('coverage', {})
        coverage_pct = coverage.get('percentage', 0)
        self.log(f"Code Coverage: {coverage_pct:.2f}%",
                "SUCCESS" if coverage_pct >= 100 else "WARNING")

        # Benchmarks
        benchmarks = self.results.get('benchmarks', {})
        self.log(f"Benchmarks: {self._get_status_text(benchmarks.get('success', False))}",
                "SUCCESS" if benchmarks.get('success') else "WARNING")

        print("="*80)

        # Overall status
        all_passed = (
            unit.get('success', False) and
            coverage_pct >= 100
        )

        if all_passed:
            self.log("ALL TESTS PASSED! 100% COVERAGE ACHIEVED!", "SUCCESS")
        else:
            self.log("Some tests failed or coverage target not met", "WARNING")

        print("="*80 + "\n")

        return all_passed


def main():
    parser = argparse.ArgumentParser(description="Comprehensive test orchestration for gceme")
    parser.add_argument("--all", action="store_true", help="Run all tests and generate reports")
    parser.add_argument("--unit", action="store_true", help="Run unit tests only")
    parser.add_argument("--integration", action="store_true", help="Run integration tests only")
    parser.add_argument("--benchmark", action="store_true", help="Run benchmarks only")
    parser.add_argument("--coverage", action="store_true", help="Generate coverage reports")
    parser.add_argument("--static", action="store_true", help="Run static analysis")
    parser.add_argument("--report", action="store_true", help="Generate HTML report")
    parser.add_argument("--verbose", "-v", action="store_true", help="Verbose output")

    args = parser.parse_args()

    # If no specific test selected, run all
    if not any([args.unit, args.integration, args.benchmark, args.coverage, args.static, args.report]):
        args.all = True

    orchestrator = TestOrchestrator(verbose=args.verbose)

    overall_success = True

    try:
        if args.all or args.static:
            if not orchestrator.run_static_analysis():
                overall_success = False

        if args.all or args.unit:
            if not orchestrator.run_unit_tests():
                overall_success = False

        if args.all or args.integration:
            if not orchestrator.run_integration_tests():
                overall_success = False

        if args.all or args.benchmark:
            if not orchestrator.run_benchmark_tests():
                overall_success = False

        if args.all or args.coverage:
            if not orchestrator.generate_coverage_report():
                overall_success = False

        if args.all or args.report:
            orchestrator.generate_html_report()

        # Always save JSON results and print summary
        orchestrator.save_results_json()
        final_success = orchestrator.print_summary()

        # Exit with appropriate code
        sys.exit(0 if final_success and overall_success else 1)

    except KeyboardInterrupt:
        orchestrator.log("\nTest execution interrupted by user", "WARNING")
        sys.exit(130)
    except Exception as e:
        orchestrator.log(f"Fatal error: {e}", "ERROR")
        import traceback
        traceback.print_exc()
        sys.exit(1)


if __name__ == "__main__":
    main()
