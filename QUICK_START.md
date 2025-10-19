# Quick Start Guide

## 🚀 Getting Started

### Prerequisites
- Go 1.21 or later
- Python 3.9 or later
- Docker (optional)
- Kubernetes (optional)

### Quick Test All Systems
```bash
python3 test_all_systems.py
```

### Test Individual Services

#### TinyURL
```bash
cd services/tinyurl
go test -tags=unit -v ./...
```

#### Newsfeed
```bash
cd services/newsfeed
go test -tags=unit -v ./...
```

#### Load Balancer
```bash
cd services/loadbalancer
go test -tags=unit -v ./...
```

#### Typeahead
```bash
cd services/typeahead
go test -tags=unit -v ./...
```

#### Messaging
```bash
cd services/messaging
go test -tags=unit -v ./...
```

#### DNS
```bash
cd services/dns
go test -tags=unit -v ./...
```

#### Web Crawler
```bash
cd services/webcrawler
go test -tags=unit -v ./...
```

#### Google Docs
```bash
cd services/googledocs
go test -tags=unit -v ./...
```

#### Quora
```bash
cd services/quora
go test -tags=unit -v ./...
```

#### Sample App
```bash
cd sample-app
go test -tags=unit -v ./...
```

### Run Services

#### Start TinyURL Service
```bash
cd services/tinyurl
go run main.go
# Service runs on http://localhost:8080
```

#### Start Newsfeed Service
```bash
cd services/newsfeed
go run main.go
# Service runs on http://localhost:8081
```

#### Start Load Balancer
```bash
cd services/loadbalancer
go run main.go
# Service runs on http://localhost:8082
```

### API Examples

#### TinyURL API
```bash
# Create short URL
curl -X POST http://localhost:8080/create \
  -H "Content-Type: application/json" \
  -d '{"long_url": "https://example.com/very/long/url"}'

# Get stats
curl http://localhost:8080/stats?short_url=abc123

# Redirect (use browser or curl -L)
curl -L http://localhost:8080/abc123
```

#### Newsfeed API
```bash
# Create user
curl -X POST http://localhost:8081/user/create \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user1", "username": "john"}'

# Create post
curl -X POST http://localhost:8081/post/create \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user1", "content": "Hello World!"}'

# Get newsfeed
curl http://localhost:8081/newsfeed?user_id=user1
```

#### DNS API
```bash
# Add DNS record
curl -X POST http://localhost:8085/add \
  -H "Content-Type: application/json" \
  -d '{"domain": "example.com", "ip_address": "192.168.1.1", "type": "A", "ttl": 300}'

# Resolve domain
curl http://localhost:8085/resolve?domain=example.com
```

### View Coverage Reports

#### Generate HTML Coverage Report
```bash
cd services/tinyurl
go test -tags=unit -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### Run Coverage Analysis
```bash
cd sample-app
python3 coverage_analysis.py
```

### Docker Deployment

#### Build Docker Image
```bash
cd services/tinyurl
docker build -t tinyurl:latest .
```

#### Run Docker Container
```bash
docker run -p 8080:8080 tinyurl:latest
```

### Monitoring

#### Start Prometheus and Grafana
```bash
cd sample-app/monitoring
./setup.sh
```

#### Access Dashboards
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)

### Pre-commit Hooks

#### Install Pre-commit
```bash
pip install pre-commit
pre-commit install
```

#### Run Pre-commit Manually
```bash
pre-commit run --all-files
```

## 📊 Project Statistics

- **Total Services**: 10
- **Test Pass Rate**: 100%
- **Average Coverage**: 63.7%
- **Total Test Files**: 20+
- **Documentation Files**: 8+

## 📚 Documentation

- `README.md` - Project overview
- `FINAL_REPORT.md` - Complete implementation report
- `SYSTEMS_IMPLEMENTATION_SUMMARY.md` - Systems summary
- `docs/PRD.md` - Product requirements
- `docs/TASK_MASTER.md` - Task tracking
- `docs/CI_CD.md` - CI/CD documentation
- `docs/TESTING.md` - Testing guidelines
- `docs/MONITORING.md` - Monitoring setup

## 🎯 Next Steps

1. Review the FINAL_REPORT.md for complete details
2. Run test_all_systems.py to verify everything works
3. Explore individual services
4. Set up monitoring with Prometheus/Grafana
5. Deploy to Kubernetes (see k8s/ directory)

## 🆘 Troubleshooting

### Tests Failing
```bash
# Clean and rebuild
go clean -cache
go mod tidy
go test -v ./...
```

### Port Already in Use
```bash
# Find and kill process
lsof -ti:8080 | xargs kill -9
```

### Coverage Not Generating
```bash
# Ensure test tags are used
go test -tags=unit -coverprofile=coverage.out ./...
```

## ✅ Verification Checklist

- [ ] All tests pass: `python3 test_all_systems.py`
- [ ] Coverage reports generated
- [ ] Services start without errors
- [ ] API endpoints respond correctly
- [ ] Monitoring dashboards accessible
- [ ] Documentation reviewed

---

**Ready to deploy!** 🚀
