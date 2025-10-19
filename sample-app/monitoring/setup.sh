#!/bin/bash

# Monitoring Setup Script for Kubernetes Continuous Deployment Project
# This script sets up Prometheus, Grafana, and monitoring infrastructure

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="monitoring"
PROMETHEUS_VERSION="v2.45.0"
GRAFANA_VERSION="10.0.0"
ALERTMANAGER_VERSION="v0.25.0"

echo -e "${GREEN}🚀 Setting up monitoring infrastructure...${NC}"

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}❌ kubectl is not installed or not in PATH${NC}"
    exit 1
fi

# Check if helm is available
if ! command -v helm &> /dev/null; then
    echo -e "${YELLOW}⚠️  Helm is not installed. Installing helm...${NC}"
    curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
fi

# Create monitoring namespace
echo -e "${GREEN}📁 Creating monitoring namespace...${NC}"
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# Add Prometheus Helm repository
echo -e "${GREEN}📦 Adding Prometheus Helm repository...${NC}"
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Install Prometheus
echo -e "${GREEN}🔍 Installing Prometheus...${NC}"
helm upgrade --install prometheus prometheus-community/kube-prometheus-stack \
    --namespace $NAMESPACE \
    --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false \
    --set prometheus.prometheusSpec.podMonitorSelectorNilUsesHelmValues=false \
    --set prometheus.prometheusSpec.ruleSelectorNilUsesHelmValues=false \
    --set prometheus.prometheusSpec.retention=30d \
    --set prometheus.prometheusSpec.retentionSize=10GB \
    --set prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage=10Gi \
    --set grafana.adminPassword=admin \
    --set grafana.service.type=LoadBalancer \
    --set grafana.service.port=80 \
    --set grafana.service.targetPort=3000 \
    --set grafana.persistence.enabled=true \
    --set grafana.persistence.size=5Gi \
    --set alertmanager.alertmanagerSpec.retention=120h \
    --set alertmanager.alertmanagerSpec.storage.volumeClaimTemplate.spec.resources.requests.storage=2Gi \
    --wait

# Wait for Prometheus to be ready
echo -e "${GREEN}⏳ Waiting for Prometheus to be ready...${NC}"
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=prometheus -n $NAMESPACE --timeout=300s

# Wait for Grafana to be ready
echo -e "${GREEN}⏳ Waiting for Grafana to be ready...${NC}"
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=grafana -n $NAMESPACE --timeout=300s

# Create ServiceMonitor for sample-app
echo -e "${GREEN}📊 Creating ServiceMonitor for sample-app...${NC}"
cat <<EOF | kubectl apply -f -
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: sample-app
  namespace: $NAMESPACE
  labels:
    app: sample-app
spec:
  selector:
    matchLabels:
      app: sample-app
  endpoints:
  - port: http
    path: /prometheus
    interval: 15s
    scrapeTimeout: 10s
EOF

# Create PrometheusRule for sample-app alerts
echo -e "${GREEN}🚨 Creating PrometheusRule for sample-app alerts...${NC}"
cat <<EOF | kubectl apply -f -
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: sample-app-alerts
  namespace: $NAMESPACE
  labels:
    app: sample-app
spec:
  groups:
  - name: sample-app
    rules:
    - alert: HighErrorRate
      expr: rate(error_count_total[5m]) / rate(request_count_total[5m]) > 0.05
      for: 2m
      labels:
        severity: warning
      annotations:
        summary: "High error rate detected"
        description: "Error rate is {{ \$value | humanizePercentage }} for the last 5 minutes"
    - alert: VeryHighErrorRate
      expr: rate(error_count_total[5m]) / rate(request_count_total[5m]) > 0.1
      for: 1m
      labels:
        severity: critical
      annotations:
        summary: "Very high error rate detected"
        description: "Error rate is {{ \$value | humanizePercentage }} for the last 5 minutes"
    - alert: HighResponseTime
      expr: avg_response_time_seconds > 2
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "High response time detected"
        description: "Average response time is {{ \$value }}s"
    - alert: ServiceDown
      expr: up{job="sample-app"} == 0
      for: 1m
      labels:
        severity: critical
      annotations:
        summary: "Sample app is down"
        description: "Sample app has been down for more than 1 minute"
EOF

# Get Grafana service information
echo -e "${GREEN}📊 Getting Grafana service information...${NC}"
GRAFANA_SERVICE=$(kubectl get svc -n $NAMESPACE -l app.kubernetes.io/name=grafana -o jsonpath='{.items[0].metadata.name}')
GRAFANA_PORT=$(kubectl get svc -n $NAMESPACE $GRAFANA_SERVICE -o jsonpath='{.spec.ports[0].port}')

# Get Prometheus service information
echo -e "${GREEN}🔍 Getting Prometheus service information...${NC}"
PROMETHEUS_SERVICE=$(kubectl get svc -n $NAMESPACE -l app.kubernetes.io/name=prometheus -o jsonpath='{.items[0].metadata.name}')
PROMETHEUS_PORT=$(kubectl get svc -n $NAMESPACE $PROMETHEUS_SERVICE -o jsonpath='{.spec.ports[0].port}')

# Get AlertManager service information
echo -e "${GREEN}🚨 Getting AlertManager service information...${NC}"
ALERTMANAGER_SERVICE=$(kubectl get svc -n $NAMESPACE -l app.kubernetes.io/name=alertmanager -o jsonpath='{.items[0].metadata.name}')
ALERTMANAGER_PORT=$(kubectl get svc -n $NAMESPACE $ALERTMANAGER_SERVICE -o jsonpath='{.spec.ports[0].port}')

# Display access information
echo -e "${GREEN}✅ Monitoring setup completed successfully!${NC}"
echo ""
echo -e "${YELLOW}📊 Access Information:${NC}"
echo -e "Grafana: http://localhost:$GRAFANA_PORT (admin/admin)"
echo -e "Prometheus: http://localhost:$PROMETHEUS_PORT"
echo -e "AlertManager: http://localhost:$ALERTMANAGER_PORT"
echo ""
echo -e "${YELLOW}🔧 Port Forward Commands:${NC}"
echo -e "Grafana: kubectl port-forward -n $NAMESPACE svc/$GRAFANA_SERVICE $GRAFANA_PORT:80"
echo -e "Prometheus: kubectl port-forward -n $NAMESPACE svc/$PROMETHEUS_SERVICE $PROMETHEUS_PORT:9090"
echo -e "AlertManager: kubectl port-forward -n $NAMESPACE svc/$ALERTMANAGER_SERVICE $ALERTMANAGER_PORT:9093"
echo ""
echo -e "${YELLOW}📈 Dashboard Import:${NC}"
echo -e "1. Access Grafana at http://localhost:$GRAFANA_PORT"
echo -e "2. Login with admin/admin"
echo -e "3. Import dashboard from monitoring/grafana-dashboard.json"
echo ""
echo -e "${YELLOW}🔍 Verify Setup:${NC}"
echo -e "kubectl get pods -n $NAMESPACE"
echo -e "kubectl get svc -n $NAMESPACE"
echo -e "kubectl get servicemonitor -n $NAMESPACE"
echo -e "kubectl get prometheusrule -n $NAMESPACE"
echo ""
echo -e "${GREEN}🎉 Monitoring setup completed!${NC}"
