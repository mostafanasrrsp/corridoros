#!/bin/bash

# CorridorOS Startup Script
# This script starts all CorridorOS services and performs health checks

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVICES=("corrd" "memqosd" "fabmand" "heliopassd" "attestd" "compatd" "metricsd" "securityd")
PORTS=(8080 8081 8083 8082 8084 8087 8088 8089)
HEALTH_ENDPOINTS=("/health" "/health" "/health" "/health" "/health" "/health" "/health" "/health")
TIMEOUT=30

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_dependencies() {
    log_info "Checking dependencies..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi
    
    # Check if Rust is installed
    if ! command -v cargo &> /dev/null; then
        log_error "Rust is not installed. Please install Rust 1.75 or later."
        exit 1
    fi
    
    # Check if curl is installed
    if ! command -v curl &> /dev/null; then
        log_error "curl is not installed. Please install curl."
        exit 1
    fi
    
    log_success "All dependencies are installed"
}

build_services() {
    log_info "Building CorridorOS services..."
    
    # Build corrd (Rust)
    log_info "Building corrd daemon..."
    cd daemon/corrd
    cargo build --release
    cd ../..
    
    # Build Go services
    for service in memqosd fabmand heliopassd attestd; do
        log_info "Building $service daemon..."
        cd daemon/$service
        go build -o $service .
        cd ../..
    done
    
    # Build CLI tools
    log_info "Building CLI tools..."
    cd cli
    go build -o corridor corridor/main.go
    go build -o ffm ffm/main.go
    cd ..
    
    # Build CorridorLabs
    log_info "Building CorridorLabs..."
    for lab in physics-decoder helio-sim; do
        log_info "Building $lab..."
        cd labs/$lab
        go build -o $lab .
        cd ../..
    done
    
    log_success "All services built successfully"
}

start_service() {
    local service=$1
    local port=$2
    local health_endpoint=$3
    
    log_info "Starting $service on port $port..."
    
    # Start service in background
    if [ "$service" = "corrd" ]; then
        ./daemon/corrd/target/release/corrd &
    else
        ./daemon/$service/$service &
    fi
    
    local pid=$!
    echo $pid > /tmp/corridoros-$service.pid
    
    # Wait for service to start
    local count=0
    while [ $count -lt $TIMEOUT ]; do
        if curl -s -f "http://localhost:$port$health_endpoint" > /dev/null 2>&1; then
            log_success "$service is running (PID: $pid)"
            return 0
        fi
        sleep 1
        count=$((count + 1))
    done
    
    log_error "$service failed to start within $TIMEOUT seconds"
    return 1
}

start_services() {
    log_info "Starting CorridorOS services..."
    
    for i in "${!SERVICES[@]}"; do
        local service=${SERVICES[$i]}
        local port=${PORTS[$i]}
        local health_endpoint=${HEALTH_ENDPOINTS[$i]}
        
        if ! start_service "$service" "$port" "$health_endpoint"; then
            log_error "Failed to start $service"
            stop_services
            exit 1
        fi
    done
    
    log_success "All CorridorOS services are running"
}

stop_services() {
    log_info "Stopping CorridorOS services..."
    
    for service in "${SERVICES[@]}"; do
        local pid_file="/tmp/corridoros-$service.pid"
        if [ -f "$pid_file" ]; then
            local pid=$(cat "$pid_file")
            if kill -0 "$pid" 2>/dev/null; then
                log_info "Stopping $service (PID: $pid)..."
                kill "$pid"
                rm -f "$pid_file"
            fi
        fi
    done
    
    log_success "All services stopped"
}

show_status() {
    log_info "CorridorOS Service Status:"
    echo "=========================="
    
    for i in "${!SERVICES[@]}"; do
        local service=${SERVICES[$i]}
        local port=${PORTS[$i]}
        local health_endpoint=${HEALTH_ENDPOINTS[$i]}
        
        if curl -s -f "http://localhost:$port$health_endpoint" > /dev/null 2>&1; then
            echo -e "  $service: ${GREEN}RUNNING${NC} (port $port)"
        else
            echo -e "  $service: ${RED}STOPPED${NC} (port $port)"
        fi
    done
}

show_usage() {
    echo "Usage: $0 {start|start-only|start-with-labs|start-labs|stop-labs|status-labs|stop|restart|status|build|demo}"
    echo ""
    echo "Commands:"
    echo "  start   - Start all CorridorOS services"
    echo "  start-only - Start services without building"
    echo "  start-with-labs - Start services and Labs bundle"
    echo "  start-labs - Start Labs bundle only (physics-decoder, helio-sim, synchrony-analytics)"
    echo "  stop-labs - Stop Labs bundle"
    echo "  status-labs - Show Labs status"
    echo "  stop    - Stop all CorridorOS services"
    echo "  restart - Restart all CorridorOS services"
    echo "  status  - Show service status"
    echo "  build   - Build all services"
    echo "  demo    - Run demonstration"
    echo ""
}

run_demo() {
    log_info "Running CorridorOS demonstration..."
    
    # Check if services are running
    local all_running=true
    for i in "${!SERVICES[@]}"; do
        local service=${SERVICES[$i]}
        local port=${PORTS[$i]}
        local health_endpoint=${HEALTH_ENDPOINTS[$i]}
        
        if ! curl -s -f "http://localhost:$port$health_endpoint" > /dev/null 2>&1; then
            all_running=false
            break
        fi
    done
    
    if [ "$all_running" = false ]; then
        log_error "Not all services are running. Please start services first."
        exit 1
    fi
    
    # Run FFM demo
    log_info "Running FFM demo..."
    cd examples/ffm-demo
    go run main.go
    cd ../..
    
    # Run corridor demo
    log_info "Running corridor demo..."
    cd examples/corridor-demo
    go run main.go
    cd ../..
    
    log_success "Demonstration completed"
}

# Main script logic
case "${1:-start}" in
    start)
        check_dependencies
        build_services
        start_services
        show_status
        ;;
    start-only)
        start_services
        show_status
        ;;
    start-with-labs)
        start_services
        ./scripts/run-labs.sh start || true
        show_status
        ./scripts/run-labs.sh status || true
        ;;
    start-labs)
        ./scripts/run-labs.sh start
        ;;
    stop-labs)
        ./scripts/run-labs.sh stop
        ;;
    status-labs)
        ./scripts/run-labs.sh status
        ;;
    stop)
        stop_services
        ;;
    restart)
        stop_services
        sleep 2
        start_services
        show_status
        ;;
    status)
        show_status
        ;;
    build)
        check_dependencies
        build_services
        ;;
    demo)
        run_demo
        ;;
    *)
        show_usage
        exit 1
        ;;
esac
