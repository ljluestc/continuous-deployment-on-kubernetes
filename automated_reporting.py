#!/usr/bin/env python3
"""
Automated Reporting System for Continuous Deployment on Kubernetes
Generates comprehensive test reports, coverage analysis, and trend tracking
"""

import os
import sys
import json
import time
import subprocess
from pathlib import Path
from datetime import datetime
from typing import Dict, List, Tuple
import argparse


class AutomatedReporter:
    def __init__(self, project_root: str = "."):
        self.project_root = Path(project_root).resolve()
        self.report_dir = self.project_root / "test-reports"
        self.report_dir.mkdir(exist_ok=True)
        
        self.services = [
            "sample-app",
            "services/googledocs",
            "services/quora",
            "services/messaging",
            "services/dns",
            "services/webcrawler",
            "services/newsfeed",
            "services/loadbalancer",
            "services/tinyurl",
            "services/typeahead",
        ]
        
        self.results = {
            "timestamp": datetime.now().isoformat(),
            "services": {},
            "summary": {},
            "trends": {},
        }
    
    def run_command(self, cmd: List[str], cwd: Path) -> Tuple[int, str, str]:
        """Run a command and return exit code, stdout, stderr"""
        try:
            result = subprocess.run(
                cmd,
                cwd=cwd,
                capture_output=True,
                text=True,
                timeout=300
            )
            return result.returncode, result.stdout, result.stderr
        except subprocess.TimeoutExpired:
            return -1, "", "Command timed out"
        except Exception as e:
            return -1, "", str(e)
    
    def get_service_coverage(self, service_path: Path) -> Dict:
        """Get coverage information for a service"""
        print(f"  📊 Analyzing coverage for {service_path.name}...")
        
        # Run tests
        cmd = ["go", "test", "-tags=unit", "-v", "-coverprofile=coverage.out", "./..."]
        exit_code, stdout, stderr = self.run_command(cmd, service_path)
        
        if exit_code != 0:
            return {
                "status": "FAILED",
                "coverage": 0.0,
                "error": stderr or stdout
            }
        
        # Get coverage percentage
        cmd = ["go", "tool", "cover", "-func=coverage.out"]
        exit_code, stdout, stderr = self.run_command(cmd, service_path)
        
        if exit_code != 0:
            return {
                "status": "PASSED",
                "coverage": 0.0,
                "details": "No coverage data"
            }
        
        # Parse coverage
        total_coverage = 0.0
        function_coverage = []
        
        for line in stdout.split('\n'):
            if line.strip():
                parts = line.split()
                if len(parts) >= 3:
                    func_name = parts[1] if len(parts) > 2 else parts[0]
                    coverage_str = parts[-1].rstrip('%')
                    try:
                        coverage = float(coverage_str)
                        if 'total:' in line:
                            total_coverage = coverage
                        else:
                            function_coverage.append({
                                "function": func_name,
                                "coverage": coverage
                            })
                    except ValueError:
                        pass
        
        # Generate HTML report
        html_path = service_path / "coverage.html"
        cmd = ["go", "tool", "cover", "-html=coverage.out", "-o", str(html_path)]
        self.run_command(cmd, service_path)
        
        return {
            "status": "PASSED",
            "coverage": total_coverage,
            "function_coverage": function_coverage,
            "html_report": str(html_path.relative_to(self.project_root))
        }
    
    def analyze_all_services(self):
        """Analyze all services and generate reports"""
        print("\n" + "="*80)
        print("AUTOMATED REPORTING SYSTEM")
        print("="*80)
        print(f"Timestamp: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        print(f"Project Root: {self.project_root}")
        print("="*80 + "\n")
        
        total_coverage = 0.0
        passed_count = 0
        failed_count = 0
        
        for service_path_str in self.services:
            service_path = self.project_root / service_path_str
            service_name = service_path.name
            
            print(f"\n{'─'*60}")
            print(f"Service: {service_name}")
            print(f"{'─'*60}")
            
            if not service_path.exists():
                print(f"  ⚠️  Service directory not found: {service_path}")
                continue
            
            result = self.get_service_coverage(service_path)
            self.results["services"][service_name] = result
            
            if result["status"] == "PASSED":
                passed_count += 1
                total_coverage += result["coverage"]
                print(f"  ✅ Status: PASSED")
                print(f"  📈 Coverage: {result['coverage']:.1f}%")
            else:
                failed_count += 1
                print(f"  ❌ Status: FAILED")
                print(f"  ❗ Error: {result.get('error', 'Unknown error')}")
        
        # Calculate summary
        total_services = passed_count + failed_count
        avg_coverage = total_coverage / passed_count if passed_count > 0 else 0.0
        
        self.results["summary"] = {
            "total_services": total_services,
            "passed": passed_count,
            "failed": failed_count,
            "average_coverage": round(avg_coverage, 2),
            "pass_rate": round((passed_count / total_services * 100) if total_services > 0 else 0, 2)
        }
        
        print("\n" + "="*80)
        print("SUMMARY")
        print("="*80)
        print(f"Total Services: {total_services}")
        print(f"✅ Passed: {passed_count}")
        print(f"❌ Failed: {failed_count}")
        print(f"📊 Average Coverage: {avg_coverage:.2f}%")
        print(f"🎯 Pass Rate: {self.results['summary']['pass_rate']}%")
        print("="*80 + "\n")
    
    def generate_html_report(self) -> str:
        """Generate HTML report"""
        html = f"""
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Test Coverage Report - {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}</title>
    <style>
        body {{
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background: #f5f5f5;
        }}
        .header {{
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            border-radius: 10px;
            margin-bottom: 30px;
        }}
        .summary {{
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }}
        .metric {{
            background: white;
            padding: 20px;
            border-radius: 10px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }}
        .metric-value {{
            font-size: 2.5em;
            font-weight: bold;
            color: #667eea;
        }}
        .metric-label {{
            color: #666;
            margin-top: 10px;
        }}
        .service {{
            background: white;
            padding: 20px;
            margin-bottom: 15px;
            border-radius: 10px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }}
        .service-header {{
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 15px;
        }}
        .service-name {{
            font-size: 1.2em;
            font-weight: bold;
        }}
        .status-badge {{
            padding: 5px 15px;
            border-radius: 20px;
            font-weight: bold;
            font-size: 0.9em;
        }}
        .status-passed {{
            background: #d4edda;
            color: #155724;
        }}
        .status-failed {{
            background: #f8d7da;
            color: #721c24;
        }}
        .coverage-bar {{
            background: #e0e0e0;
            height: 30px;
            border-radius: 15px;
            overflow: hidden;
            margin: 10px 0;
        }}
        .coverage-fill {{
            height: 100%;
            background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
            display: flex;
            align-items: center;
            justify-content: center;
            color: white;
            font-weight: bold;
        }}
        .coverage-excellent {{ background: linear-gradient(90deg, #00c9ff 0%, #92fe9d 100%); }}
        .coverage-good {{ background: linear-gradient(90deg, #667eea 0%, #764ba2 100%); }}
        .coverage-fair {{ background: linear-gradient(90deg, #f093fb 0%, #f5576c 100%); }}
        .coverage-poor {{ background: linear-gradient(90deg, #fa709a 0%, #fee140 100%); }}
        .timestamp {{
            text-align: center;
            color: #666;
            margin-top: 30px;
        }}
    </style>
</head>
<body>
    <div class="header">
        <h1>🚀 Test Coverage Report</h1>
        <p>Kubernetes Continuous Deployment Project</p>
        <p>{datetime.now().strftime('%Y-%m-%d %H:%M:%S')}</p>
    </div>
    
    <div class="summary">
        <div class="metric">
            <div class="metric-value">{self.results['summary']['total_services']}</div>
            <div class="metric-label">Total Services</div>
        </div>
        <div class="metric">
            <div class="metric-value">{self.results['summary']['passed']}</div>
            <div class="metric-label">✅ Passed</div>
        </div>
        <div class="metric">
            <div class="metric-value">{self.results['summary']['failed']}</div>
            <div class="metric-label">❌ Failed</div>
        </div>
        <div class="metric">
            <div class="metric-value">{self.results['summary']['average_coverage']:.1f}%</div>
            <div class="metric-label">📊 Average Coverage</div>
        </div>
        <div class="metric">
            <div class="metric-value">{self.results['summary']['pass_rate']}%</div>
            <div class="metric-label">🎯 Pass Rate</div>
        </div>
    </div>
    
    <h2>Service Details</h2>
"""
        
        for service_name, result in sorted(self.results['services'].items()):
            status_class = "status-passed" if result['status'] == "PASSED" else "status-failed"
            coverage = result.get('coverage', 0)
            
            if coverage >= 90:
                coverage_class = "coverage-excellent"
            elif coverage >= 80:
                coverage_class = "coverage-good"
            elif coverage >= 70:
                coverage_class = "coverage-fair"
            else:
                coverage_class = "coverage-poor"
            
            html += f"""
    <div class="service">
        <div class="service-header">
            <span class="service-name">{service_name}</span>
            <span class="status-badge {status_class}">{result['status']}</span>
        </div>
        <div class="coverage-bar">
            <div class="coverage-fill {coverage_class}" style="width: {coverage}%">
                {coverage:.1f}%
            </div>
        </div>
"""
            if result['status'] == "PASSED" and 'html_report' in result:
                html += f'        <a href="{result["html_report"]}" target="_blank">View Detailed Coverage Report →</a>\n'
            
            html += "    </div>\n"
        
        html += f"""
    <div class="timestamp">
        Generated at {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}
    </div>
</body>
</html>
"""
        return html
    
    def save_reports(self):
        """Save all reports"""
        # Save JSON report
        json_path = self.report_dir / f"report_{datetime.now().strftime('%Y%m%d_%H%M%S')}.json"
        with open(json_path, 'w') as f:
            json.dump(self.results, f, indent=2)
        print(f"✅ JSON report saved: {json_path}")
        
        # Save latest JSON (for trend tracking)
        latest_json = self.report_dir / "latest.json"
        with open(latest_json, 'w') as f:
            json.dump(self.results, f, indent=2)
        print(f"✅ Latest report saved: {latest_json}")
        
        # Save HTML report
        html_path = self.report_dir / "test_report.html"
        with open(html_path, 'w') as f:
            f.write(self.generate_html_report())
        print(f"✅ HTML report saved: {html_path}")
        
        print(f"\n📊 Open {html_path} in your browser to view the report")


def main():
    parser = argparse.ArgumentParser(description="Automated Reporting System")
    parser.add_argument("--project-root", default=".", help="Project root directory")
    
    args = parser.parse_args()
    
    reporter = AutomatedReporter(args.project_root)
    reporter.analyze_all_services()
    reporter.save_reports()
    
    # Exit with error if any tests failed
    if reporter.results['summary']['failed'] > 0:
        sys.exit(1)
    
    sys.exit(0)


if __name__ == "__main__":
    main()

