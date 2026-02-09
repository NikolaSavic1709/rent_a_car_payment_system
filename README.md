# Rent a Car Payment System - Deployment Guide

## Architecture Overview

This is a microservices-based payment system with multiple Go backends and Next.js frontends.

### Services Architecture

**Backend Services (Go)**
- **Webshop Service** (Port 8080): Main rental service
- **Bank Gateway Service** (Port 8081): Routes payments to banks
- **Bank Service** (Port 8082): Bank integration
- **PSP Service** (Port 8084): Payment Service Provider - routes to payment methods
- **Crypto Service** (Port 8086): Cryptocurrency payment processing

**Frontend Applications (Next.js)**
- **rentacar-front** (Port 3000): Customer-facing webshop
- **psp-front** (Port 3001): Card/QR/PayPal payment pages
- **crypto-front** (Port 3002): Cryptocurrency payment interface

**Databases (PostgreSQL)**
- psql_webshop (Port 5432)
- psql_bank_gateway (Port 5433)
- psql_erstebank (Port 5434)
- psql_psp (Port 5436)
- psql_crypto (Port 5438)

## Deployment Instructions

### Prerequisites
- Docker and Docker Compose installed
- Node.js 18+ and npm installed
- Ports 3000-3002, 8080-8086, 5432-5438 available

### Step 1: Start Backend Services

Navigate to the core directory:
```bash
cd core
```

Start all microservices with Docker Compose:
```bash
docker compose up --build
```

This will start:
- All Go microservices
- All PostgreSQL databases
- Automatic database schema initialization

**Wait for all services to be healthy** (check logs for "database system is ready to accept connections")

To start services in detached mode:
```bash
docker compose up -d
```

To stop all services:
```bash
docker compose down
```

To stop and remove volumes (clean slate):
```bash
docker compose down -v
```

### Step 2: Start Frontend Applications

Open **3 separate terminal windows** for the frontends.

#### Terminal 1: Start rentacar-front (Port 3000)
```bash
cd front/rentacar-front
npm install
npm run dev
```

Access at: http://localhost:3000

#### Terminal 2: Start psp-front (Port 3001)
```bash
cd front/psp-front
npm install
npm run dev
```

Access at: http://localhost:3001

#### Terminal 3: Start crypto-front (Port 3002)
```bash
cd front/crypto-front
npm install
npm run dev
```

Access at: http://localhost:3002

### Service Health Checks

Check if services are running:

**Backend Services:**
```bash
curl http://localhost:8080/health  # Webshop
curl http://localhost:8084/health  # PSP
curl http://localhost:PORT/health  # ...
```

**Frontend Services:**
- http://localhost:3000 (rentacar-front)
- http://localhost:3001 (psp-front - card page)
- http://localhost:3002 (crypto-front)


## Port Reference

| Service | Port | URL |
|---------|------|-----|
| Webshop Service | 8080 | http://localhost:8080 |
| Bank Gateway | 8081 | http://localhost:8081 |
| Erste Bank | 8082 | http://localhost:8082 |
| PSP Service | 8084 | http://localhost:8084 |
| Crypto Service | 8086 | http://localhost:8086 |
| rentacar-front | 3000 | http://localhost:3000 |
| psp-front | 3001 | http://localhost:3001 |
| crypto-front | 3002 | http://localhost:3002 |


### Viewing Logs

View all service logs:
```bash
docker compose logs -f
```

View specific service logs:
```bash
docker compose logs -f crypto_service
docker compose logs -f psp_service
```
