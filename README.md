# Foomlet - E-Wallet Application

Foomlet is a modern e-wallet application built with Go that provides secure digital wallet services including top-up, payments, and money transfers. The application features asynchronous transfer processing using RabbitMQ for improved performance and reliability.

## Features

- **User Management**
  - User registration with phone number and PIN
  - Secure login with JWT authentication
  - Profile management

- **Wallet Operations**
  - Digital wallet creation for each user
  - Balance tracking and history
  - Secure transaction processing

- **Financial Transactions**
  - **Top-Up**: Add money to wallet
  - **Payments**: Make payments with remarks
  - **Transfers**: Send money to other users
  - Transaction history with detailed records

- **Asynchronous Processing**
  - RabbitMQ integration for transfer processing
  - Quorum queues for high availability
  - Fallback to synchronous processing when queue is unavailable

- **Security**
  - Argon2 password hashing
  - JWT access and refresh tokens
  - Database transaction integrity

## Technology Stack

### Backend
- **Go 1.23.5** - Primary programming language
- **Gin** - HTTP web framework
- **PostgreSQL 15** - Primary database
- **RabbitMQ 3.12** - Message broker for asynchronous processing

### Authentication & Security
- **JWT** - JSON Web Tokens for authentication
- **Argon2** - Password hashing algorithm

### Database
- **pgx/v5** - PostgreSQL driver and connection pooling
- **golang-migrate** - Database migration tool

### DevOps & Deployment
- **Docker & Docker Compose** - Containerization
- **Alpine Linux** - Lightweight container base

### Additional Libraries
- **godotenv** - Environment variable management
- **uuid** - UUID generation
- **amqp091-go** - RabbitMQ client

## Project Structure

```
e:\coding\rabbitmq\
├── cmd/
│   ├── main.go                 # Main application entry point
│   └── seeder/
│       └── seed.main.go        # Database seeder
├── internal/
│   ├── config/                 # Configuration management
│   ├── handlers/               # HTTP request handlers
│   ├── middlewares/            # Custom middlewares
│   ├── models/                 # Data models and DTOs
│   ├── repositories/           # Data access layer
│   └── routes/                 # Route definitions
├── migrations/                 # Database migrations
│   └── seed/                   # Seed data
├── pkg/                        # Shared utilities
├── docker-compose.yml          # Docker services configuration
├── dockerfile                  # Application container
├── .env                        # Environment variables
└── README.md
```

## API Endpoints

### Authentication
- `POST /api/auth` - User login
- `POST /api/auth/new` - User registration
- `POST /api/auth/refresh` - Refresh access token

### Profile Management
- `PATCH /api/profile` - Update user profile

### Transactions
- `POST /api/topup` - Add money to wallet
- `POST /api/payments` - Make payment
- `POST /api/transfers` - Transfer money to another user
- `GET /api/transactions` - Get transaction history

### Health Check
- `GET /ping` - Application health check

## How to Run the Project

### Prerequisites
- Docker and Docker Compose installed
- Git for cloning the repository

### Method 1: Using Docker Compose (Recommended)

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd foomlet
   ```

2. **Set up environment variables**
   ```bash
   # Create .env file with your database credentials
   # Example configuration is provided below
   ```

3. **Start all services**
   ```bash
   docker-compose up --build -d
   ```

4. **Check service status**
   ```bash
   docker-compose logs -f app
   ```

5. **Access the application**
   - API: http://localhost:8080
   - RabbitMQ Management: http://localhost:15672 (guest/guest)
   - Database: localhost:5432

### Method 2: Local Development

1. **Install dependencies**
   ```bash
   # Install Go 1.23.5
   # Install PostgreSQL 15
   # Install RabbitMQ 3.12
   ```

2. **Set up database**
   ```bash
   # Create PostgreSQL database
   createdb -U postgres foomlet
   
   # Run migrations
   migrate -path ./migrations -database "your-database-url" up
   ```

3. **Seed initial data**
   ```bash
   go run cmd/seeder/seed.main.go
   ```

4. **Start the application**
   ```bash
   go run cmd/main.go
   ```

### Available Commands

```bash
# Build and start services
docker-compose up --build -d

# Stop services
docker-compose down

# View logs
docker-compose logs -f app

# Clean up
docker-compose down -v

# Run migrations manually
docker-compose exec app migrate -path ./migrations -database "$DB_URL" up

# Seed data manually
docker-compose exec app ./seeder

# Access database
docker-compose exec db psql -U youruser -d yourdb
```

## Initial Test Data

The application comes with pre-seeded test data:

### Test Users
- **User 1**: 
  - Phone: `08123456789`
  - PIN: `123456`
  - Name: John Doe
  - Initial Balance: 350,000 (after transactions)

- **User 2**: 
  - Phone: `08987654321` 
  - PIN: `123456`
  - Name: Jane Smith
  - Initial Balance: 100,000 (received from transfer)

### Sample Transactions
- Top-up: 500,000
- Payment: 50,000 (electricity bill)
- Transfer: 100,000 (birthday gift to User 2)

## Environment Variables

Create a `.env` file with the following configuration:

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=youruser
DB_PASSWORD=yourpassword
DB_NAME=yourdb
DB_URL="postgresql://youruser:yourpassword@localhost:5432/yourdb?search_path=public&sslmode=disable"

# JWT Configuration
JWT_ACCESS_SECRET=your_access_secret_here
JWT_REFRESH_SECRET=your_refresh_secret_here
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

# RabbitMQ Configuration
RABBITMQ_URL=amqp://guest:guest@localhost:5672/

# Server Configuration
PORT=8080
```

## Architecture Highlights

### Asynchronous Transfer Processing
- Transfers are queued in RabbitMQ for processing
- Worker processes consume transfer messages
- Graceful fallback to synchronous processing
- Quorum queues ensure message durability

### Database Design
- PostgreSQL with proper foreign key relationships
- Money type for accurate financial calculations
- Transaction atomicity with proper rollback handling
- Separate tables for different transaction types

### Security Features
- Argon2 password hashing with salt
- JWT-based stateless authentication
- Input validation and sanitization
- Database transaction integrity

## Monitoring & Management

### RabbitMQ Management UI
Access http://localhost:15672 with credentials `guest/guest` to:
- Monitor queue status
- View message rates
- Manage exchanges and queues
- Debug transfer processing

### Database Management
Connect to PostgreSQL to:
- View transaction history
- Monitor wallet balances
- Analyze user data
- Debug database issues

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

This project is developed for educational and demonstration purposes.
This project is developed for educational and demonstration purposes.
