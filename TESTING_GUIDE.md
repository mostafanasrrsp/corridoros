# CorridorOS Testing Guide

## Quick Start Testing (No Hardware Required)

### 1. Set Up Mock Environment
```bash
cd /Users/mnasr/Desktop/COS/CorridorOS
chmod +x scripts/start-corridoros.sh
./scripts/start-corridoros.sh build
./scripts/start-corridoros.sh start-with-labs
```

### 2. Test CLI Commands
```bash
# Photonic lanes allocation
./cli/corridor_cli.py lanes-alloc --lanes 8 --lambda-start 1550 --min-gbps 400 --latency-ns 250

# Free-Form Memory allocation  
./cli/corridor_cli.py ffm-alloc --bytes $((256*1024*1024*1024)) --tier T2 --bw-floor 150

# List allocated resources
./cli/corridor_cli.py list
```

### 3. Run CorridorLabs Simulations
```bash
# HELIOPASS calibration simulation
cd labs/helio-sim
./helio-sim --lanes 8 --target-ber 1e-12

# Physics formula validation
cd ../physics-decoder  
./physics-decoder --validate "E=mc^2" --units "J,kg,m/s"
```

### 4. API Testing
```bash
# Test photonic corridors API
curl -X POST http://localhost:8080/v1/corridors \
  -H "Content-Type: application/json" \
  -d '{
    "type": "SiCorridor",
    "lanes": 8,
    "lambda_nm": [1550,1551,1552,1553,1554,1555,1556,1557],
    "min_gbps": 400,
    "latency_budget_ns": 250
  }'

# Test FFM allocation API  
curl -X POST http://localhost:8081/v1/ffm/alloc \
  -H "Content-Type: application/json" \
  -d '{
    "bytes": 274877906944,
    "latency_class": "T2", 
    "bandwidth_floor_GBs": 150
  }'
```

## Hardware Development Phases

### Phase 1: Development Machine (Current)
**Purpose**: Software development, simulation, algorithm testing
**Requirements**: 
- CPU: 8+ cores (i7/Ryzen 7+)
- RAM: 32GB recommended
- Storage: 1TB NVMe SSD
- Budget: $1,200-$1,800

**Recommended Models**:
- ThinkPad P16s Gen 1 (~$1,400)
- Framework Laptop 16 (~$1,600) 
- ASUS ProArt StudioBook (~$1,500)

### Phase 2: CXL Integration (3-6 months)
**Purpose**: Real CXL device testing, FFM development
**Requirements**:
- CXL 3.0 compatible motherboard
- Intel Sapphire Rapids or AMD Genoa CPU
- CXL Type-3 memory devices
- Budget: $8,000-$15,000

### Phase 3: Photonic Integration (6+ months)  
**Purpose**: Real optical I/O testing, full corridor validation
**Requirements**:
- Photonic development kits
- Co-packaged optics
- Optical test equipment
- Budget: $25,000-$75,000

## Performance Testing

### Latency Testing
```bash
# Test memory latency by tier
./cli/ffm benchmark --tier T0 --pattern random-read
./cli/ffm benchmark --tier T1 --pattern random-read  
./cli/ffm benchmark --tier T2 --pattern random-read

# Test photonic latency
./cli/corridor benchmark --lanes 8 --pattern ping-pong
```

### Bandwidth Testing
```bash
# Test FFM bandwidth floors
./cli/ffm bandwidth --allocation ffm-001 --target-gbps 200
./cli/ffm telemetry ffm-001  # Verify achieved bandwidth

# Test photonic bandwidth
./cli/corridor telemetry cor-001  # Check achieved vs. min Gbps
```

### BER/Eye Testing (Photonic)
```bash
# Monitor BER over time
./cli/corridor watch cor-001

# Trigger recalibration
./cli/corridor calibrate cor-001 --target-ber 1e-15
```

## Security Testing

### Attestation Testing
```bash
# Generate attestation tickets
./scripts/gen-attestation.sh --type corridor --policy gold
./scripts/gen-attestation.sh --type ffm --policy tenant-a

# Test with attestation
./cli/corridor lanes-alloc --lanes 8 --attestation-ticket att-abc123
```

### Confidential Compute Testing
```bash
# Test with TEE/enclave integration
./cli/ffm alloc 1GB --tier T1 --security-domain confidential
```

## Integration Testing

### Docker Testing
```bash
# Run in containerized environment
docker-compose up corridoros-stack
docker-compose exec corrd corridor lanes-alloc --lanes 4
```

### Kubernetes Testing (Future)
```bash
# Deploy as K8s operators
kubectl apply -f k8s/corridoros-operator.yaml
kubectl create corridorallocation my-lanes --lanes=8
```

## Monitoring & Observability

### Metrics Collection
```bash
# Start metrics collection
./scripts/start-corridoros.sh start
# Metrics available at:
# http://localhost:8088/metrics (Prometheus)
# http://localhost:3000 (Grafana dashboard)
```

### Log Analysis
```bash
# View service logs
tail -f /var/log/corridoros/corrd.log
tail -f /var/log/corridoros/memqosd.log

# Structured log analysis
jq '.level="ERROR"' /var/log/corridoros/corrd.log
```
