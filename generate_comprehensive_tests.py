#!/usr/bin/env python3
"""
Automatically generate comprehensive test coverage for all Go services
"""

import os
import re
import subprocess
from pathlib import Path

def analyze_go_file(file_path):
    """Analyze a Go file to find handlers and functions that need testing"""
    with open(file_path, 'r') as f:
        content = f.read()
    
    # Find all HTTP handlers
    handlers = re.findall(r'func (\w+Handler)\(w http\.ResponseWriter, r \*http\.Request\)', content)
    
    # Find all service methods
    methods = re.findall(r'func \(s \*\w+Service\) (\w+)\([^)]*\)', content)
    
    # Find helper functions
    helpers = re.findall(r'^func (\w+)\([^)]*\) [^{]*{', content, re.MULTILINE)
    helpers = [h for h in helpers if not h.endswith('Handler') and h != 'main']
    
    return {
        'handlers': handlers,
        'methods': methods,
        'helpers': list(set(helpers))
    }

def get_coverage_gaps(service_dir):
    """Run coverage and identify gaps"""
    try:
        result = subprocess.run(
            ['go', 'tool', 'cover', '-func=coverage.out'],
            cwd=service_dir,
            capture_output=True,
            text=True
        )
        
        if result.returncode != 0:
            return []
        
        gaps = []
        for line in result.stdout.split('\n'):
            if '%' in line and not line.startswith('total'):
                parts = line.split()
                if len(parts) >= 3:
                    func_name = parts[1].split(':')[1] if ':' in parts[1] else parts[1]
                    coverage = float(parts[2].rstrip('%'))
                    if coverage < 100:
                        gaps.append((func_name, coverage))
        
        return gaps
    except Exception as e:
        print(f"Error analyzing coverage: {e}")
        return []

def main():
    project_root = Path('/home/calelin/dev/continuous-deployment-on-kubernetes')
    
    services = [
        'services/dns',
        'services/webcrawler',
        'services/newsfeed',
        'services/loadbalancer',
        'services/tinyurl',
        'services/typeahead',
    ]
    
    print("=" * 80)
    print("ANALYZING SERVICES FOR COVERAGE GAPS")
    print("=" * 80)
    print()
    
    for service_path in services:
        service_dir = project_root / service_path
        service_name = service_path.split('/')[-1]
        
        print(f"\n{'='*60}")
        print(f"Service: {service_name}")
        print(f"{'='*60}")
        
        main_go = service_dir / 'main.go'
        if not main_go.exists():
            print(f"  ⚠️  main.go not found")
            continue
        
        # Analyze code
        analysis = analyze_go_file(main_go)
        print(f"\n  Found:")
        print(f"    - {len(analysis['handlers'])} HTTP handlers")
        print(f"    - {len(analysis['methods'])} service methods")
        print(f"    - {len(analysis['helpers'])} helper functions")
        
        # Check coverage gaps
        gaps = get_coverage_gaps(service_dir)
        if gaps:
            print(f"\n  Coverage gaps ({len(gaps)} functions):")
            for func_name, coverage in sorted(gaps, key=lambda x: x[1]):
                print(f"    - {func_name}: {coverage}%")
        
        # Print test recommendations
        print(f"\n  Recommended tests to add:")
        for handler in analysis['handlers']:
            print(f"    - Test{handler} (success case)")
            print(f"    - Test{handler}_InvalidMethod")
            print(f"    - Test{handler}_InvalidJSON (if POST)")
            print(f"    - Test{handler}_MissingParams")
            print(f"    - Test{handler}_NotFound")
        
        print()

if __name__ == '__main__':
    main()

