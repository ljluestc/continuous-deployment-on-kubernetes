# CI/CD Pipeline Documentation

## Table of Contents
1. [Overview](#overview)
2. [Pipeline Architecture](#pipeline-architecture)
3. [GitHub Actions Workflows](#github-actions-workflows)
4. [Pre-commit Hooks](#pre-commit-hooks)
5. [Deployment Process](#deployment-process)
6. [Monitoring and Alerts](#monitoring-and-alerts)

## Overview

The gceme application uses a fully automated CI/CD pipeline that ensures code quality, runs comprehensive tests, and safely deploys to production.

### Pipeline Objectives
- ✅ Automated testing on every commit
- ✅ Code quality enforcement
- ✅ Security scanning
- ✅ Performance benchmarking
- ✅ Automated deployment
- ✅ Rollback capabilities

### Technology Stack
- **CI/CD Platform**: GitHub Actions
- **Pre-commit**: Python pre-commit framework
- **Container**: Docker
- **Orchestration**: Kubernetes (GKE)
- **Jenkins**: Build automation (see main README)

## Pipeline Architecture

### Pipeline Stages

```
┌─────────────┐
│   Commit    │
└──────┬──────┘
       │
       ├─────────────────┐
       │                 │
┌──────▼──────┐   ┌─────▼──────┐
│ Pre-commit  │   │  GitHub    │
│   Hooks     │   │  Actions   │
└──────┬──────┘   └─────┬──────┘
       │                │
       │          ┌─────▼──────┐
       │          │    Lint    │
       │          └─────┬──────┘
       │                │
       │          ┌─────▼──────┐
       │          │   Tests    │
       │          └─────┬──────┘
       │                │
       │          ┌─────▼──────┐
       │          │  Security  │
       │          └─────┬──────┘
       │                │
       │          ┌─────▼──────┐
       │          │   Build    │
       │          └─────┬──────┘
       │                │
       │          ┌─────▼──────┐
       │          │   Deploy   │
       │          └────────────┘
       │
       └──────────────────────────> Success/Failure
```

### Stage Details

| Stage | Duration | Purpose |
|-------|----------|---------|
| Pre-commit | <30s | Local validation before commit |
| Lint | 1-2 min | Code quality checks |
| Tests | 2-3 min | Unit and integration tests |
| Security | 1-2 min | Vulnerability scanning |
| Build | 2-3 min | Docker image creation |
| Deploy | 3-5 min | Kubernetes deployment |

## GitHub Actions Workflows

### Main CI/CD Workflow

Location: `.github/workflows/ci.yml`

#### Trigger Events
```yaml
on:
  push:
    branches: [ master, main, canary, develop ]
  pull_request:
    branches: [ master, main, canary ]
  workflow_dispatch:  # Manual trigger
```

#### Jobs

##### 1. Lint Job
**Purpose**: Static analysis and code quality checks

```yaml
steps:
  - go fmt check
  - go vet
  - golangci-lint
  - go mod tidy check
```

**Success Criteria**:
- No formatting issues
- No vet warnings
- No lint errors
- go.mod is tidy

##### 2. Test Job
**Purpose**: Run unit tests with coverage

```yaml
steps:
  - Run unit tests with race detection
  - Generate coverage report
  - Verify coverage threshold (90%+)
  - Upload coverage to Codecov
```

**Success Criteria**:
- All tests pass
- No race conditions
- Coverage ≥ 90%

##### 3. Integration Job
**Purpose**: End-to-end testing

```yaml
steps:
  - Run integration tests
  - Verify component interaction
```

**Success Criteria**:
- All integration tests pass

##### 4. Benchmark Job
**Purpose**: Performance testing

```yaml
steps:
  - Run benchmarks
  - Compare with baseline
  - Upload results
```

**Success Criteria**:
- No significant performance regression

##### 5. Security Job
**Purpose**: Vulnerability scanning

```yaml
steps:
  - Trivy filesystem scan
  - gosec security scanner
  - Upload results to GitHub Security
```

**Success Criteria**:
- No critical vulnerabilities
- No high-severity issues

##### 6. Build Job
**Purpose**: Create Docker image

```yaml
steps:
  - Build Docker image
  - Tag with commit SHA
  - Save as artifact
```

**Success Criteria**:
- Image builds successfully
- Image size is reasonable

##### 7. Test Orchestration Job
**Purpose**: Run comprehensive Python test script

```yaml
steps:
  - Execute test_comprehensive.py
  - Generate HTML reports
  - Upload reports
```

**Success Criteria**:
- All test suites pass
- Reports generated successfully

### Workflow Configuration

#### Environment Variables
```yaml
env:
  GO_VERSION: '1.20'
  COVERAGE_THRESHOLD: 90
```

#### Caching
```yaml
- uses: actions/cache@v3
  with:
    path: ~/go/pkg/mod
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
```

### Artifacts

Generated artifacts available for 90 days:
- Coverage reports (HTML)
- Benchmark results
- Test reports
- Docker images (1 day retention)

## Pre-commit Hooks

### Installation

```bash
# Install pre-commit
pip install pre-commit

# Install hooks in repository
cd /path/to/repo
pre-commit install
```

### Configured Hooks

Location: `.pre-commit-config.yaml`

#### Go Hooks
1. **go-fmt**: Format Go code
2. **go-vet**: Static analysis
3. **go-imports**: Import management
4. **go-mod-tidy**: Dependency management
5. **go-unit-tests**: Run unit tests
6. **go-coverage-check**: Verify coverage
7. **golangci-lint**: Comprehensive linting

#### Python Hooks
1. **black**: Code formatting
2. **flake8**: Linting

#### General Hooks
1. **check-yaml**: YAML validation
2. **check-json**: JSON validation
3. **check-merge-conflict**: Merge conflict detection
4. **check-added-large-files**: Large file prevention
5. **trailing-whitespace**: Whitespace cleanup
6. **end-of-file-fixer**: EOF normalization
7. **mixed-line-ending**: Line ending fixes

#### Security Hooks
1. **detect-secrets**: Secret detection
2. **hadolint**: Dockerfile linting
3. **markdownlint**: Markdown linting

### Usage

#### Automatic (on commit)
```bash
git commit -m "Your message"
# Hooks run automatically
```

#### Manual (all files)
```bash
pre-commit run --all-files
```

#### Manual (specific hook)
```bash
pre-commit run go-fmt
```

#### Skip hooks (emergency only)
```bash
git commit --no-verify -m "Skip hooks"
```

### Hook Configuration

#### Coverage Threshold
Modify in `.pre-commit-config.yaml`:
```yaml
- id: go-coverage-check
  entry: bash -c 'cd sample-app && go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out | grep total | awk "{if (\$3 < 90.0) {print \"Coverage below 90%: \" \$3; exit 1}}"'
```

## Deployment Process

### Branch Strategy

```
master/main     → Production deployment
canary          → Canary deployment (partial rollout)
develop         → Development environment
feature/*       → Feature branches (no deployment)
```

### Deployment Flow

#### 1. Development
```bash
git checkout -b feature/new-feature
# Make changes
git commit -m "Add feature"
git push origin feature/new-feature
# Create PR → Triggers CI checks
```

#### 2. Canary Deployment
```bash
git checkout canary
git merge feature/new-feature
git push origin canary
# Triggers canary deployment (1/5 instances)
```

#### 3. Production Deployment
```bash
git checkout master
git merge canary
git push origin master
# Triggers full production deployment
```

### Kubernetes Deployment

#### Namespaces
- `production`: Production environment
- `canary`: Canary environment (within production namespace)
- `new-feature`: Development branches

#### Deployment Configuration
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gceme-frontend-production
  namespace: production
spec:
  replicas: 4
  selector:
    matchLabels:
      app: gceme
      role: frontend
      env: production
```

### Rollback Process

#### Automatic Rollback
- Triggered on health check failures
- Monitored by Kubernetes liveness/readiness probes

#### Manual Rollback
```bash
# Rollback to previous version
kubectl rollout undo deployment/gceme-frontend-production -n production

# Rollback to specific revision
kubectl rollout undo deployment/gceme-frontend-production -n production --to-revision=2

# Check rollout status
kubectl rollout status deployment/gceme-frontend-production -n production
```

## Monitoring and Alerts

### Metrics Monitored

1. **Pipeline Metrics**
   - Build success rate
   - Average build time
   - Test pass rate
   - Coverage trends

2. **Application Metrics**
   - Request rate
   - Error rate
   - Response time (p50, p95, p99)
   - CPU/Memory usage

3. **Security Metrics**
   - Vulnerabilities detected
   - Security scan results
   - Failed authentication attempts

### GitHub Actions Insights

View pipeline analytics:
```
Repository → Actions → View workflow runs
```

Metrics available:
- Success/failure rates
- Duration trends
- Resource usage

### Notifications

#### Success
- ✅ Green check on commit
- GitHub notification

#### Failure
- ❌ Red X on commit
- Email notification
- GitHub notification
- Slack webhook (if configured)

## Troubleshooting

### Common Pipeline Issues

#### Tests Fail in CI but Pass Locally
**Cause**: Environment differences
**Solution**:
```bash
# Run tests exactly as CI does
go test -v -race -coverprofile=coverage.out ./...
```

#### Coverage Below Threshold
**Cause**: New code not tested
**Solution**:
```bash
# Find uncovered code
go tool cover -html=coverage.out
# Add tests for uncovered lines
```

#### Docker Build Fails
**Cause**: Missing dependencies, invalid Dockerfile
**Solution**:
```bash
# Test Docker build locally
docker build -t gceme:test .
```

#### Deployment Hangs
**Cause**: Image pull errors, resource limits
**Solution**:
```bash
# Check deployment status
kubectl describe deployment gceme-frontend-production -n production
# Check pod status
kubectl get pods -n production
# Check logs
kubectl logs -f <pod-name> -n production
```

### Debug Pipeline

#### Enable Debug Logging
Add to workflow:
```yaml
env:
  ACTIONS_STEP_DEBUG: true
  ACTIONS_RUNNER_DEBUG: true
```

#### View Detailed Logs
GitHub Actions → Select workflow run → View logs

#### Download Artifacts
GitHub Actions → Select workflow run → Artifacts → Download

## Best Practices

### DO ✅
- Run pre-commit hooks before pushing
- Write tests for all new code
- Keep pipeline fast (<10 minutes)
- Monitor pipeline health
- Review security scan results
- Update dependencies regularly

### DON'T ❌
- Skip CI checks with --no-verify
- Merge failing PRs
- Ignore security vulnerabilities
- Deploy without testing
- Hardcode secrets in code
- Leave flaky tests

## Security

### Secrets Management
- Use GitHub Secrets for sensitive data
- Never commit credentials
- Rotate secrets regularly
- Use least-privilege access

### Secret Configuration
```yaml
# In workflow
- name: Deploy
  env:
    GCP_SA_KEY: ${{ secrets.GCP_SA_KEY }}
```

## Performance Optimization

### Pipeline Speed Optimization
1. **Caching**: Cache dependencies
2. **Parallel Jobs**: Run independent jobs concurrently
3. **Matrix Builds**: Test multiple versions in parallel
4. **Skip Steps**: Use conditions to skip unnecessary steps

### Example Optimization
```yaml
strategy:
  matrix:
    go-version: [1.19, 1.20, 1.21]
  fail-fast: false
```

## Maintenance

### Regular Tasks
- [ ] Review and update dependencies (weekly)
- [ ] Check for GitHub Actions updates (monthly)
- [ ] Review security scan results (weekly)
- [ ] Update Go version (as needed)
- [ ] Clean up old artifacts (automated)

### Quarterly Reviews
- Pipeline performance analysis
- Cost optimization review
- Security posture assessment
- Dependency audit

## Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Pre-commit Framework](https://pre-commit.com/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Go CI/CD Best Practices](https://golang.org/doc/)

---
**Last Updated**: 2025-10-17
**Pipeline Version**: 1.0
**Maintainer**: Engineering Team
