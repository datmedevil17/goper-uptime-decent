# Uptime Monitoring Platform

Decentralized uptime monitoring system with Solana-based validator payments.

## 🏗️ Architecture

- **API Server** (Port 8080): REST API for website management
- **Hub Server** (Port 8081): WebSocket coordinator for validators
- **Validators**: Distributed nodes performing health checks
- **PostgreSQL**: GORM-based data persistence
- **RabbitMQ**: Asynchronous payout processing
- **Solana Devnet**: Validator payment settlement

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- Solana CLI (optional, for keypair generation)

### Setup

1. **Clone repository**
```bash
git clone <your-repo>
cd uptime-monitor
```

2. **Install dependencies**
```bash
make install
```

3. **Configure environment**
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. **Generate Solana keypair (optional)**
```bash
make generate-keypair
# Copy the base58 private key to .env
```

5. **Start services**
```bash
make docker-up
```

6. **View logs**
```bash
make docker-logs
```

## 📡 API Endpoints

### Website Management (Authenticated)
- `POST /api/v1/website` - Create website
- `GET /api/v1/websites` - List all websites
- `GET /api/v1/website/status?websiteId=xxx` - Get website status
- `DELETE /api/v1/website` - Delete website

### Validator Payouts
- `POST /api/v1/payout/:validatorId` - Request payout
- `GET /api/v1/validator/:validatorId/balance` - Check balance

### Health
- `GET /health` - Service health check

## 🛠️ Development

### Run locally without Docker

Terminal 1 - Database:
```bash
docker-compose up postgres rabbitmq -d
```

Terminal 2 - Hub:
```bash
make run-hub
```

Terminal 3 - API:
```bash
make run-api
```

Terminal 4 - Validator:
```bash
VALIDATOR_PRIVATE_KEY=<your-key> make run-validator
```

### Build binaries
```bash
make build
./bin/api
./bin/hub
./bin/validator
```

## 🔧 Configuration

Edit `.env`:
- `DATABASE_URL`: PostgreSQL connection string
- `RABBITMQ_URL`: RabbitMQ connection string
- `PLATFORM_PRIVATE_KEY`: Solana wallet for payouts

- `VALIDATOR_PRIVATE_KEY`: Individual validator keypair

## 📊 Database Schema (GORM)

- **User**: User accounts
- **Website**: Monitored websites
- **Validator**: Registered validators
- **WebsiteTick**: Health check results
- **PayoutTransaction**: Payment history

Migrations run automatically on startup.

## 🧪 Testing

```bash
make test
```

## 📦 Docker

Scale validators:
```bash
docker-compose up -d --scale validator=5
```

Rebuild containers:
```bash
make docker-rebuild
```

## 🔒 Security

- JWT authentication for API endpoints
- Cryptographic signatures for validator messages
- Database row locking for payout safety
- Transaction-based payout processing

## 📝 License

MIT# goper-uptime-decent
