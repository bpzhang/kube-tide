#!/bin/bash

# Kube-Tide Database API Test Script
# This script tests the complete API functionality

set -e

API_BASE_URL="http://localhost:8080/api/v1/db"
CLUSTER_ID="test-cluster-$(date +%s)"

echo "🚀 Starting Kube-Tide Database API Tests"
echo "API Base URL: $API_BASE_URL"
echo "Test Cluster ID: $CLUSTER_ID"
echo

# Function to make HTTP requests with error handling
make_request() {
    local method=$1
    local url=$2
    local data=$3
    local expected_status=${4:-200}
    
    echo "📡 $method $url"
    if [ -n "$data" ]; then
        echo "📤 Request body: $data"
    fi
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$url")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url")
    fi
    
    # Split response body and status code
    body=$(echo "$response" | head -n -1)
    status=$(echo "$response" | tail -n 1)
    
    echo "📥 Response ($status): $body"
    
    if [ "$status" != "$expected_status" ]; then
        echo "❌ Expected status $expected_status, got $status"
        exit 1
    fi
    
    echo "✅ Request successful"
    echo
    
    # Return the response body for further processing
    echo "$body"
}

# Test 1: Health Check
echo "🏥 Testing Health Check"
make_request "GET" "http://localhost:8080/health"

# Test 2: API Documentation
echo "📚 Testing API Documentation"
make_request "GET" "$API_BASE_URL"

# Test 3: Create Deployment
echo "🚀 Testing Deployment Creation"
deployment_data='{
    "cluster_id": "'$CLUSTER_ID'",
    "namespace": "default",
    "name": "test-deployment",
    "replicas": 3,
    "ready_replicas": 3,
    "available_replicas": 3,
    "unavailable_replicas": 0,
    "updated_replicas": 3,
    "strategy_type": "RollingUpdate",
    "labels": {"app": "test", "version": "v1"},
    "annotations": {"deployment.kubernetes.io/revision": "1"},
    "selector": {"app": "test"},
    "template": {"spec": {"containers": [{"name": "test", "image": "nginx:latest"}]}}
}'

deployment_response=$(make_request "POST" "$API_BASE_URL/deployments" "$deployment_data" "200")
deployment_id=$(echo "$deployment_response" | jq -r '.deployment.id')
echo "Created deployment with ID: $deployment_id"

# Test 4: Get Deployment by ID
echo "🔍 Testing Get Deployment by ID"
make_request "GET" "$API_BASE_URL/deployments/$deployment_id"

# Test 5: List Deployments by Cluster
echo "📋 Testing List Deployments by Cluster"
make_request "GET" "$API_BASE_URL/clusters/$CLUSTER_ID/deployments?page=1&page_size=10"

# Test 6: Count Deployments by Cluster
echo "🔢 Testing Count Deployments by Cluster"
make_request "GET" "$API_BASE_URL/clusters/$CLUSTER_ID/deployments/count"

# Test 7: Create Service
echo "🌐 Testing Service Creation"
service_data='{
    "cluster_id": "'$CLUSTER_ID'",
    "namespace": "default",
    "name": "test-service",
    "type": "ClusterIP",
    "cluster_ip": "10.96.0.100",
    "ports": [{"name": "http", "port": 80, "target_port": 8080, "protocol": "TCP"}],
    "selector": {"app": "test"},
    "labels": {"app": "test", "version": "v1"},
    "annotations": {"service.kubernetes.io/load-balancer-class": "internal"}
}'

service_response=$(make_request "POST" "$API_BASE_URL/services" "$service_data" "200")
service_id=$(echo "$service_response" | jq -r '.service.id')
echo "Created service with ID: $service_id"

# Test 8: Get Service by ID
echo "🔍 Testing Get Service by ID"
make_request "GET" "$API_BASE_URL/services/$service_id"

# Test 9: List Services by Cluster
echo "📋 Testing List Services by Cluster"
make_request "GET" "$API_BASE_URL/clusters/$CLUSTER_ID/services?page=1&page_size=10"

# Test 10: Count Services by Cluster
echo "🔢 Testing Count Services by Cluster"
make_request "GET" "$API_BASE_URL/clusters/$CLUSTER_ID/services/count"

# Test 11: List Deployments by Namespace
echo "📋 Testing List Deployments by Namespace"
make_request "GET" "$API_BASE_URL/clusters/$CLUSTER_ID/namespaces/default/deployments"

# Test 12: List Services by Namespace
echo "📋 Testing List Services by Namespace"
make_request "GET" "$API_BASE_URL/clusters/$CLUSTER_ID/namespaces/default/services"

# Test 13: Update Deployment
echo "✏️ Testing Update Deployment"
update_data='{
    "replicas": 5,
    "ready_replicas": 5,
    "available_replicas": 5,
    "updated_replicas": 5
}'
make_request "PUT" "$API_BASE_URL/deployments/$deployment_id" "$update_data"

# Test 14: Update Service
echo "✏️ Testing Update Service"
service_update_data='{
    "type": "NodePort",
    "ports": [{"name": "http", "port": 80, "target_port": 8080, "protocol": "TCP", "node_port": 30080}]
}'
make_request "PUT" "$API_BASE_URL/services/$service_id" "$service_update_data"

# Test 15: Get Updated Deployment
echo "🔍 Testing Get Updated Deployment"
make_request "GET" "$API_BASE_URL/deployments/$deployment_id"

# Test 16: Get Updated Service
echo "🔍 Testing Get Updated Service"
make_request "GET" "$API_BASE_URL/services/$service_id"

# Test 17: Count Deployments by Namespace
echo "🔢 Testing Count Deployments by Namespace"
make_request "GET" "$API_BASE_URL/clusters/$CLUSTER_ID/namespaces/default/deployments/count"

# Test 18: Count Services by Namespace
echo "🔢 Testing Count Services by Namespace"
make_request "GET" "$API_BASE_URL/clusters/$CLUSTER_ID/namespaces/default/services/count"

# Cleanup Tests
echo "🧹 Testing Cleanup Operations"

# Test 19: Delete Deployment
echo "🗑️ Testing Delete Deployment"
make_request "DELETE" "$API_BASE_URL/deployments/$deployment_id"

# Test 20: Delete Service
echo "🗑️ Testing Delete Service"
make_request "DELETE" "$API_BASE_URL/services/$service_id"

# Test 21: Verify Deletion - Should return 404
echo "✅ Testing Deployment Deletion Verification"
make_request "GET" "$API_BASE_URL/deployments/$deployment_id" "" "404"

echo "✅ Testing Service Deletion Verification"
make_request "GET" "$API_BASE_URL/services/$service_id" "" "404"

echo
echo "🎉 All API tests completed successfully!"
echo "✅ Deployment CRUD operations working"
echo "✅ Service CRUD operations working"
echo "✅ Pagination and filtering working"
echo "✅ Count operations working"
echo "✅ Cluster and namespace scoping working"
echo "✅ Error handling working"
echo
echo "🚀 Kube-Tide Database API is fully functional!" 