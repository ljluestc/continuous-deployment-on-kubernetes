#!/usr/bin/env python3
"""
Coverage Analysis Script for Kubernetes Continuous Deployment Project
This script analyzes test coverage and provides recommendations for improvement.
"""

import json
import subprocess
import sys
from pathlib import Path
from typing import Dict, List, Tuple

class CoverageAnalyzer:
    def __init__(self, project_root: str = "."):
        self.project_root = Path(project_root).resolve()
        # If we're already in sample-app directory, use current directory
        if self.project_root.name == "sample-app":
            self.sample_app_dir = self.project_root
        else:
            self.sample_app_dir = self.project_root / "sample-app"
        self.coverage_config = self.load_coverage_config()
        
    def load_coverage_config(self) -> Dict:
        """Load coverage configuration from coverage_config.json"""
        config_path = self.sample_app_dir / "coverage_config.json"
        try:
            with open(config_path, 'r') as f:
                return json.load(f)
        except FileNotFoundError:
            print(f"Warning: Coverage config not found at {config_path}")
            return self.get_default_config()
    
    def get_default_config(self) -> Dict:
        """Get default coverage configuration"""
        return {
            "coverage_requirements": {
                "overall": {
                    "minimum": 70,
                    "target": 80,
                    "excellent": 90
                }
            }
        }
    
    def run_coverage_analysis(self) -> Dict:
        """Run coverage analysis and return results"""
        print("🔍 Running coverage analysis...")
        
        # Change to sample-app directory
        import os
        os.chdir(self.sample_app_dir)
        
        # Run tests with coverage
        result = subprocess.run([
            "go", "test", "-v", "-coverprofile=coverage.out", 
            "-covermode=atomic", "-tags=unit", "./..."
        ], capture_output=True, text=True)
        
        if result.returncode != 0:
            print(f"❌ Tests failed: {result.stderr}")
            return {"error": "Tests failed"}
        
        # Generate coverage report
        coverage_result = subprocess.run([
            "go", "tool", "cover", "-func=coverage.out"
        ], capture_output=True, text=True)
        
        if coverage_result.returncode != 0:
            print(f"❌ Coverage analysis failed: {coverage_result.stderr}")
            return {"error": "Coverage analysis failed"}
        
        # Parse coverage output
        coverage_data = self.parse_coverage_output(coverage_result.stdout)
        
        # Analyze coverage
        analysis = self.analyze_coverage(coverage_data)
        
        # Generate recommendations
        recommendations = self.generate_recommendations(analysis)
        
        return {
            "coverage_data": coverage_data,
            "analysis": analysis,
            "recommendations": recommendations
        }
    
    def parse_coverage_output(self, output: str) -> Dict:
        """Parse coverage output from go tool cover"""
        lines = output.strip().split('\n')
        coverage_data = {
            "files": {},
            "total": {}
        }
        
        for line in lines:
            if line.strip():
                # Check for total line first
                if "total:" in line:
                    # Handle format: "total:\t\t(statements)\t84.7%"
                    # Split by tab and find the percentage
                    all_parts = line.split('\t')
                    for part in all_parts:
                        if '%' in part:
                            coverage_str = part.replace('%', '')
                            try:
                                coverage = float(coverage_str)
                                coverage_data["total"] = {
                                    "coverage": coverage,
                                    "statements": "0"
                                }
                                break
                            except ValueError:
                                continue
                else:
                    parts = line.split('\t')
                    if len(parts) >= 3:
                        file_path = parts[0]
                        coverage_str = parts[2].replace('%', '')
                        
                        try:
                            coverage = float(coverage_str)
                            coverage_data["files"][file_path] = {
                                "coverage": coverage,
                                "statements": parts[1] if len(parts) > 1 else "0"
                            }
                        except ValueError:
                            continue
        
        # If no total found, try to extract from the last line
        if not coverage_data["total"] and lines:
            last_line = lines[-1]
            if "total:" in last_line:
                parts = last_line.split('\t')
                if len(parts) >= 3:
                    coverage_str = parts[2].replace('%', '')
                    try:
                        coverage = float(coverage_str)
                        coverage_data["total"] = {
                            "coverage": coverage,
                            "statements": parts[1] if len(parts) > 1 else "0"
                        }
                    except ValueError:
                        pass
        
        return coverage_data
    
    def analyze_coverage(self, coverage_data: Dict) -> Dict:
        """Analyze coverage data and determine status"""
        # Get total coverage, with fallback to 0 if not found
        if "total" in coverage_data and "coverage" in coverage_data["total"]:
            total_coverage = coverage_data["total"]["coverage"]
        else:
            total_coverage = 0.0
        
        requirements = self.coverage_config["coverage_requirements"]["overall"]
        
        # Determine status
        if total_coverage >= requirements["excellent"]:
            status = "excellent"
            status_emoji = "🎉"
        elif total_coverage >= requirements["target"]:
            status = "good"
            status_emoji = "✅"
        elif total_coverage >= requirements["minimum"]:
            status = "acceptable"
            status_emoji = "⚠️"
        else:
            status = "poor"
            status_emoji = "❌"
        
        # Analyze individual files
        file_analysis = {}
        for file_path, data in coverage_data["files"].items():
            file_coverage = data["coverage"]
            
            if file_coverage >= 90:
                file_status = "excellent"
            elif file_coverage >= 80:
                file_status = "good"
            elif file_coverage >= 70:
                file_status = "acceptable"
            else:
                file_status = "needs_improvement"
            
            file_analysis[file_path] = {
                "coverage": file_coverage,
                "status": file_status,
                "needs_attention": file_coverage < 70
            }
        
        return {
            "total_coverage": total_coverage,
            "status": status,
            "status_emoji": status_emoji,
            "meets_minimum": total_coverage >= requirements["minimum"],
            "meets_target": total_coverage >= requirements["target"],
            "meets_excellent": total_coverage >= requirements["excellent"],
            "file_analysis": file_analysis,
            "requirements": requirements
        }
    
    def generate_recommendations(self, analysis: Dict) -> List[str]:
        """Generate recommendations for improving coverage"""
        recommendations = []
        
        total_coverage = analysis["total_coverage"]
        requirements = analysis["requirements"]
        
        # Overall recommendations
        if not analysis["meets_minimum"]:
            recommendations.append(f"❌ Coverage {total_coverage:.1f}% is below minimum requirement of {requirements['minimum']}%")
            recommendations.append("Focus on increasing overall test coverage to meet minimum requirements")
        elif not analysis["meets_target"]:
            recommendations.append(f"⚠️  Coverage {total_coverage:.1f}% is below target of {requirements['target']}%")
            recommendations.append("Consider adding more tests to reach the target coverage")
        elif not analysis["meets_excellent"]:
            recommendations.append(f"✅ Coverage {total_coverage:.1f}% meets target but could be improved to reach {requirements['excellent']}%")
            recommendations.append("Focus on edge cases and error handling for excellent coverage")
        else:
            recommendations.append(f"🎉 Excellent coverage of {total_coverage:.1f}%!")
            recommendations.append("Maintain current coverage levels and focus on test quality")
        
        # File-specific recommendations
        files_needing_attention = [
            file_path for file_path, data in analysis["file_analysis"].items()
            if data["needs_attention"]
        ]
        
        if files_needing_attention:
            recommendations.append(f"Files needing attention: {', '.join(files_needing_attention)}")
            
            for file_path in files_needing_attention:
                file_data = analysis["file_analysis"][file_path]
                recommendations.append(f"  - {file_path}: {file_data['coverage']:.1f}% (target: 70%+)")
        
        # Specific recommendations based on file types
        for file_path, data in analysis["file_analysis"].items():
            if "main.go" in file_path and data["coverage"] < 80:
                recommendations.append("Consider adding more tests for main.go error handling and edge cases")
            elif "test" in file_path and data["coverage"] < 100:
                recommendations.append(f"Test file {file_path} should have 100% coverage")
            elif data["coverage"] < 70:
                recommendations.append(f"Add comprehensive tests for {file_path}")
        
        return recommendations
    
    def print_analysis(self, results: Dict):
        """Print coverage analysis results"""
        if "error" in results:
            print(f"❌ Error: {results['error']}")
            return
        
        analysis = results["analysis"]
        recommendations = results["recommendations"]
        
        print(f"\n📊 Coverage Analysis Results")
        print(f"{'='*50}")
        print(f"Total Coverage: {analysis['total_coverage']:.1f}% {analysis['status_emoji']}")
        print(f"Status: {analysis['status'].upper()}")
        print(f"Meets Minimum ({analysis['requirements']['minimum']}%): {'✅' if analysis['meets_minimum'] else '❌'}")
        print(f"Meets Target ({analysis['requirements']['target']}%): {'✅' if analysis['meets_target'] else '❌'}")
        print(f"Meets Excellent ({analysis['requirements']['excellent']}%): {'✅' if analysis['meets_excellent'] else '❌'}")
        
        print(f"\n📁 File-by-File Analysis")
        print(f"{'='*50}")
        for file_path, data in analysis["file_analysis"].items():
            status_emoji = "🎉" if data["status"] == "excellent" else "✅" if data["status"] == "good" else "⚠️" if data["status"] == "acceptable" else "❌"
            print(f"{file_path}: {data['coverage']:.1f}% {status_emoji}")
        
        print(f"\n💡 Recommendations")
        print(f"{'='*50}")
        for i, rec in enumerate(recommendations, 1):
            print(f"{i}. {rec}")
        
        print(f"\n🎯 Next Steps")
        print(f"{'='*50}")
        if analysis["meets_excellent"]:
            print("1. Maintain current coverage levels")
            print("2. Focus on test quality and maintainability")
            print("3. Consider property-based testing")
        elif analysis["meets_target"]:
            print("1. Focus on edge cases and error handling")
            print("2. Add integration tests for complex scenarios")
            print("3. Improve test coverage for low-coverage files")
        elif analysis["meets_minimum"]:
            print("1. Increase coverage to reach target (80%)")
            print("2. Add tests for critical functionality")
            print("3. Focus on files with lowest coverage")
        else:
            print("1. Prioritize increasing coverage to meet minimum (70%)")
            print("2. Add basic tests for all major functions")
            print("3. Focus on critical paths and error handling")

def main():
    if len(sys.argv) > 1:
        project_root = sys.argv[1]
    else:
        project_root = "."
    
    analyzer = CoverageAnalyzer(project_root)
    results = analyzer.run_coverage_analysis()
    analyzer.print_analysis(results)
    
    # Exit with appropriate code
    if "error" in results:
        sys.exit(1)
    
    analysis = results["analysis"]
    if not analysis["meets_minimum"]:
        sys.exit(1)
    else:
        sys.exit(0)

if __name__ == "__main__":
    main()
