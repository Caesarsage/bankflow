# BankFlow - Fintech Microservices Platform
## Complete Architecture & Design Document

---

## Table of Contents

1. [System Overview](#system-overview)
2. [Architecture Principles](#architecture-principles)
3. [Microservices Design](#microservices-design)
4. [Technology Stack](#technology-stack)
5. [Data Architecture](#data-architecture)
6. [Event-Driven Architecture](#event-driven-architecture)
7. [Security & Compliance](#security--compliance)
8. [Infrastructure Architecture](#infrastructure-architecture)
9. [Project Structure](#project-structure)
10. [API Design](#api-design)

---

## System Overview

### What We're Building

**BankFlow**: A complete digital banking platform with modern microservices architecture

**Core Features:**
- User registration and authentication (KYC)
- Account management (checking, savings)
- Fund transfers (internal & external)
- Payment processing
- Transaction history & statements
- Real-time notifications
- Fraud detection
- Admin dashboard

**Business Capabilities:**
```
┌─────────────────────────────────────────┐
│           BankFlow Platform             │
├─────────────────────────────────────────┤
│                                         │
│  Customer Management                    │
│  ├─ Registration & KYC                  │
│  ├─ Profile Management                  │
│  └─ Authentication & Authorization      │
│                                         │
│  Account Services                       │
│  ├─ Account Creation                    │
│  ├─ Balance Management                  │
│  └─ Account Types (Checking, Savings)   │
│                                         │
│  Transaction Processing                 │
│  ├─ Fund Transfers                      │
│  ├─ Payment Processing                  │
│  ├─ Transaction History                 │
│  └─ Statement Generation                │
│                                         │
│  Notification Services                  │
│  ├─ Email Notifications                 │
│  ├─ SMS Alerts                          │
│  └─ Push Notifications                  │
│                                         │
│  Fraud & Compliance                     │
│  ├─ Fraud Detection                     │
│  ├─ Transaction Monitoring              │
│  └─ Compliance Reporting                │
│                                         │
└─────────────────────────────────────────┘
```

---

## Architecture Principles

### 1. Domain-Driven Design (DDD)

**Bounded Contexts:**
- **Identity Context**: User authentication, authorization
- **Customer Context**: Customer profiles, KYC
- **Account Context**: Accounts, balances
- **Transaction Context**: Transfers, payments
- **Notification Context**: Alerts, notifications
- **Fraud Context**: Detection, monitoring

### 2. Microservices Principles

**Single Responsibility:**
- Each service owns one business capability
- Independent deployment and scaling
- Loose coupling, high cohesion

**Data Ownership:**
- Each service owns its data
- No shared databases
- Communication via APIs and events

**Technology Diversity:**
- Choose best tool for each job
- Golang for performance-critical services
- Java for enterprise features
- Node.js for I/O-intensive tasks

### 3. Event-Driven Architecture

**Async Communication via Kafka:**
- Services publish domain events
- Other services subscribe to events
- Eventual consistency
- Fault tolerance

---

## Microservices Design

### Service Catalog

```
┌─────────────────────────────────────────────────────────┐
│                    API Gateway (Kong)                    │
│              Single Entry Point for All Clients          │
└────────────────────────┬────────────────────────────────┘
                         │
         ┌───────────────┼───────────────┐
         ▼               ▼               ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│   Identity   │  │   Customer   │  │   Account    │
│   Service    │  │   Service    │  │   Service    │
│   (Golang)   │  │   (Java)     │  │   (Golang)   │
└──────┬───────┘  └──────┬───────┘  └──────┬───────┘
       │                 │                 │
       └─────────────────┼─────────────────┘
                         │
         ┌───────────────┼───────────────┐
         ▼               ▼               ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ Transaction  │  │Notification  │  │    Fraud     │
│   Service    │  │   Service    │  │   Service    │
│  (Node.js)   │  │  (Node.js)   │  │   (Golang)   │
└──────┬───────┘  └──────┬───────┘  └──────┬───────┘
       │                 │                 │
       └─────────────────┼─────────────────┘
                         │
                         ▼
              ┌──────────────────┐
              │  Kafka Event Bus  │
              │  (Event Backbone) │
              └──────────────────┘
```

---

### 1. Identity Service (Golang)

**Responsibility**: Authentication, Authorization, JWT tokens

**Why Golang**: High performance, excellent concurrency for auth operations

**Capabilities:**
- User registration and login
- JWT token generation and validation
- OAuth2 integration
- Password management
- Session management
- Multi-factor authentication (MFA)

**Technology Stack:**
- **Language**: Go 1.21+
- **Framework**: Gin or Fiber
- **Database**: PostgreSQL (user credentials)
- **Cache**: Redis (sessions, tokens)
- **Auth**: JWT, bcrypt

**API Endpoints:**
```
POST   /api/v1/auth/register      - Register new user
POST   /api/v1/auth/login         - Login user
POST   /api/v1/auth/logout        - Logout user
POST   /api/v1/auth/refresh       - Refresh JWT token
POST   /api/v1/auth/forgot        - Forgot password
POST   /api/v1/auth/reset         - Reset password
POST   /api/v1/auth/verify        - Verify email/phone
GET    /api/v1/auth/me            - Get current user
```

**Database Schema:**
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    is_verified BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE sessions (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    token VARCHAR(500) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

**Events Published:**
- `user.registered` - When user registers
- `user.logged_in` - When user logs in
- `user.verified` - When email/phone verified

---

### 2. Customer Service (Java Spring Boot)

**Responsibility**: Customer profiles, KYC, document management

**Why Java**: Enterprise-grade features, mature ecosystem, excellent for business logic

**Capabilities:**
- Customer profile management
- KYC (Know Your Customer) workflow
- Document upload and verification
- Customer search and lookup
- Profile updates

**Technology Stack:**
- **Language**: Java 17+
- **Framework**: Spring Boot 3.x, Spring Cloud
- **Database**: PostgreSQL (customer data)
- **Storage**: MinIO (documents)
- **Validation**: Hibernate Validator

**API Endpoints:**
```
POST   /api/v1/customers                - Create customer profile
GET    /api/v1/customers/:id            - Get customer by ID
PUT    /api/v1/customers/:id            - Update customer
GET    /api/v1/customers/search         - Search customers
POST   /api/v1/customers/:id/kyc        - Submit KYC documents
GET    /api/v1/customers/:id/kyc/status - Get KYC status
POST   /api/v1/customers/:id/documents  - Upload document
GET    /api/v1/customers/:id/documents  - List documents
```

**Database Schema:**
```sql
CREATE TABLE customers (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL UNIQUE,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    date_of_birth DATE NOT NULL,
    ssn_encrypted VARCHAR(255),
    address_line1 VARCHAR(255),
    address_line2 VARCHAR(255),
    city VARCHAR(100),
    state VARCHAR(50),
    zip_code VARCHAR(20),
    country VARCHAR(100),
    kyc_status VARCHAR(50) DEFAULT 'PENDING',
    kyc_verified_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE kyc_documents (
    id UUID PRIMARY KEY,
    customer_id UUID REFERENCES customers(id),
    document_type VARCHAR(50) NOT NULL,
    document_url VARCHAR(500) NOT NULL,
    status VARCHAR(50) DEFAULT 'PENDING',
    uploaded_at TIMESTAMP DEFAULT NOW(),
    verified_at TIMESTAMP
);
```

**Events Published:**
- `customer.created` - Customer profile created
- `customer.updated` - Profile updated
- `kyc.submitted` - KYC documents submitted
- `kyc.approved` - KYC approved
- `kyc.rejected` - KYC rejected

---

### 3. Account Service (Golang)

**Responsibility**: Account management, balance operations

**Why Golang**: High performance for balance calculations, excellent concurrency

**Capabilities:**
- Create accounts (checking, savings)
- Get account details
- Check balance
- Account statements
- Account closure
- Freeze/unfreeze accounts

**Technology Stack:**
- **Language**: Go 1.21+
- **Framework**: Gin
- **Database**: PostgreSQL (accounts)
- **Cache**: Redis (balance cache)

**API Endpoints:**
```
POST   /api/v1/accounts              - Create account
GET    /api/v1/accounts/:id          - Get account details
GET    /api/v1/accounts/user/:userId - Get user's accounts
GET    /api/v1/accounts/:id/balance  - Get account balance
GET    /api/v1/accounts/:id/statement - Get account statement
POST   /api/v1/accounts/:id/freeze   - Freeze account
POST   /api/v1/accounts/:id/unfreeze - Unfreeze account
DELETE /api/v1/accounts/:id          - Close account
```

**Database Schema:**
```sql
CREATE TABLE accounts (
    id UUID PRIMARY KEY,
    account_number VARCHAR(20) UNIQUE NOT NULL,
    customer_id UUID NOT NULL,
    account_type VARCHAR(50) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    balance DECIMAL(15,2) DEFAULT 0.00,
    available_balance DECIMAL(15,2) DEFAULT 0.00,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    opened_at TIMESTAMP DEFAULT NOW(),
    closed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_accounts_customer ON accounts(customer_id);
CREATE INDEX idx_accounts_number ON accounts(account_number);
```

**Events Published:**
- `account.created` - Account opened
- `account.balance.updated` - Balance changed
- `account.frozen` - Account frozen
- `account.closed` - Account closed

---

### 4. Transaction Service (Node.js)

**Responsibility**: Fund transfers, payment processing, transaction history

**Why Node.js**: Excellent for I/O operations, good for integrations

**Capabilities:**
- Internal transfers (between accounts)
- External transfers (ACH, wire)
- Payment processing
- Transaction history
- Transaction reversal
- Scheduled transfers

**Technology Stack:**
- **Language**: Node.js 20+
- **Framework**: Express.js
- **Database**: PostgreSQL (transactions)
- **Queue**: Bull (job queue for async processing)
- **Cache**: Redis

**API Endpoints:**
```
POST   /api/v1/transactions/transfer       - Initiate transfer
POST   /api/v1/transactions/payment        - Process payment
GET    /api/v1/transactions/:id            - Get transaction details
GET    /api/v1/transactions/account/:accId - Get account transactions
POST   /api/v1/transactions/:id/reverse    - Reverse transaction
GET    /api/v1/transactions/search         - Search transactions
POST   /api/v1/transactions/scheduled      - Create scheduled transfer
```

**Database Schema:**
```sql
CREATE TABLE transactions (
    id UUID PRIMARY KEY,
    transaction_ref VARCHAR(50) UNIQUE NOT NULL,
    from_account_id UUID,
    to_account_id UUID,
    amount DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    transaction_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) DEFAULT 'PENDING',
    description TEXT,
    metadata JSONB,
    processed_at TIMESTAMP,
    reversed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_transactions_from_account ON transactions(from_account_id);
CREATE INDEX idx_transactions_to_account ON transactions(to_account_id);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_created ON transactions(created_at);
```

**Events Published:**
- `transaction.initiated` - Transaction started
- `transaction.processing` - Being processed
- `transaction.completed` - Successfully completed
- `transaction.failed` - Transaction failed
- `transaction.reversed` - Transaction reversed

---

### 5. Notification Service (Node.js)

**Responsibility**: Send notifications (email, SMS, push)

**Why Node.js**: Great for I/O operations, easy integration with notification providers

**Capabilities:**
- Email notifications
- SMS notifications
- Push notifications
- Notification templates
- Notification preferences
- Notification history

**Technology Stack:**
- **Language**: Node.js 20+
- **Framework**: Express.js
- **Database**: MongoDB (notification logs)
- **Email**: SendGrid / AWS SES
- **SMS**: Twilio
- **Push**: Firebase Cloud Messaging

**API Endpoints:**
```
POST   /api/v1/notifications/email     - Send email
POST   /api/v1/notifications/sms       - Send SMS
POST   /api/v1/notifications/push      - Send push notification
GET    /api/v1/notifications/user/:id  - Get user notifications
PUT    /api/v1/notifications/preferences - Update preferences
GET    /api/v1/notifications/templates - Get templates
```

**Database Schema (MongoDB):**
```javascript
{
  _id: ObjectId,
  userId: UUID,
  type: "email" | "sms" | "push",
  recipient: String,
  subject: String,
  body: String,
  templateId: String,
  status: "pending" | "sent" | "failed",
  metadata: Object,
  sentAt: Date,
  createdAt: Date
}
```

**Events Consumed:**
- `user.registered` → Send welcome email
- `transaction.completed` → Send transaction confirmation
- `kyc.approved` → Send approval notification
- `account.created` → Send account details

---

### 6. Fraud Detection Service (Golang)

**Responsibility**: Real-time fraud detection and prevention

**Why Golang**: High performance for real-time analysis, excellent concurrency

**Capabilities:**
- Real-time transaction analysis
- Fraud scoring
- Pattern detection
- Velocity checks
- Blacklist/whitelist management
- Risk scoring

**Technology Stack:**
- **Language**: Go 1.21+
- **Framework**: Gin
- **Database**: PostgreSQL (fraud rules, cases)
- **Cache**: Redis (velocity checks)
- **ML**: Optional - Python service for ML models

**API Endpoints:**
```
POST   /api/v1/fraud/analyze          - Analyze transaction
GET    /api/v1/fraud/score/:txnId     - Get fraud score
POST   /api/v1/fraud/rules            - Create fraud rule
GET    /api/v1/fraud/rules            - List rules
GET    /api/v1/fraud/cases            - Get fraud cases
POST   /api/v1/fraud/whitelist        - Add to whitelist
POST   /api/v1/fraud/blacklist        - Add to blacklist
```

**Database Schema:**
```sql
CREATE TABLE fraud_rules (
    id UUID PRIMARY KEY,
    rule_name VARCHAR(100) NOT NULL,
    rule_type VARCHAR(50) NOT NULL,
    conditions JSONB NOT NULL,
    action VARCHAR(50) NOT NULL,
    priority INT DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE fraud_cases (
    id UUID PRIMARY KEY,
    transaction_id UUID,
    account_id UUID,
    fraud_score DECIMAL(5,2) NOT NULL,
    risk_level VARCHAR(50) NOT NULL,
    flags JSONB,
    status VARCHAR(50) DEFAULT 'OPEN',
    resolved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);
```

**Events Consumed:**
- `transaction.initiated` → Analyze for fraud
- `account.created` → Check account patterns

**Events Published:**
- `fraud.detected` - Fraud detected
- `fraud.high_risk` - High risk transaction
- `account.suspicious` - Suspicious activity

---

### 7. API Gateway (Kong)

**Responsibility**: Single entry point, routing, rate limiting, auth

**Why Kong**: Industry-standard, plugin ecosystem, high performance

**Capabilities:**
- Request routing
- Authentication (JWT validation)
- Rate limiting
- Request/response transformation
- API analytics
- CORS handling

**Plugins Used:**
- JWT authentication
- Rate limiting
- CORS
- Request transformer
- Response transformer
- Logging

---

### 8. Frontend Application (React)

**Responsibility**: User interface for customers and admins

**Technology Stack:**
- **Framework**: React 18+ with TypeScript
- **State Management**: Redux Toolkit
- **UI Library**: Material-UI or Ant Design
- **Forms**: React Hook Form
- **API Client**: Axios with interceptors
- **Charts**: Recharts or Chart.js

**Application Structure:**
```
frontend/
├── public/
├── src/
│   ├── components/
│   │   ├── common/          # Reusable components
│   │   ├── auth/            # Login, Register
│   │   ├── dashboard/       # Dashboard widgets
│   │   ├── accounts/        # Account components
│   │   ├── transactions/    # Transaction components
│   │   └── profile/         # User profile
│   ├── pages/
│   │   ├── LoginPage.tsx
│   │   ├── DashboardPage.tsx
│   │   ├── AccountsPage.tsx
│   │   ├── TransfersPage.tsx
│   │   └── TransactionsPage.tsx
│   ├── services/
│   │   ├── api.ts           # Axios instance
│   │   ├── authService.ts
│   │   ├── accountService.ts
│   │   └── transactionService.ts
│   ├── store/
│   │   ├── store.ts
│   │   ├── authSlice.ts
│   │   ├── accountSlice.ts
│   │   └── transactionSlice.ts
│   ├── utils/
│   ├── hooks/
│   ├── types/
│   └── App.tsx
└── package.json
```

**Key Features:**
- Customer dashboard with account overview
- Transaction history with search/filter
- Fund transfer interface
- Profile management
- KYC document upload
- Real-time notifications
- Admin panel (separate routes)

---

## Data Architecture

### Database Strategy

**Database per Service:**
```
Identity Service     → PostgreSQL (users, sessions)
Customer Service     → PostgreSQL (customers, KYC)
Account Service      → PostgreSQL (accounts)
Transaction Service  → PostgreSQL (transactions)
Notification Service → MongoDB (notifications)
Fraud Service        → PostgreSQL (rules, cases)
```

**Why Different Databases:**
- Service independence
- Technology fit (MongoDB for logs)
- Scalability (scale services independently)
- Failure isolation

### Data Consistency Patterns

**1. Saga Pattern for Distributed Transactions**

Example: Creating a new account
```
1. Customer Service creates customer → Success
2. Publish: customer.created event
3. Account Service listens → Creates account
4. If account creation fails → Compensating transaction
5. Publish: account.creation.failed
6. Customer Service listens → Rollback customer
```

**2. Event Sourcing for Transactions**

All transaction state changes stored as events:
```
transaction.initiated
transaction.fraud_check.passed
transaction.balance_reserved
transaction.processed
transaction.completed
```

**3. CQRS (Command Query Responsibility Segregation)**

Separate read and write models:
```
Command: POST /api/v1/transactions/transfer
  ↓
Write to transactions table
  ↓
Publish event
  ↓
Update read model (transaction history view)
```

---

## Event-Driven Architecture

### Kafka Topics

```
┌─────────────────────────────────────────┐
│         Kafka Event Topics              │
├─────────────────────────────────────────┤
│                                         │
│ identity-events                         │
│   ├─ user.registered                    │
│   ├─ user.logged_in                     │
│   └─ user.verified                      │
│                                         │
│ customer-events                         │
│   ├─ customer.created                   │
│   ├─ customer.updated                   │
│   ├─ kyc.submitted                      │
│   ├─ kyc.approved                       │
│   └─ kyc.rejected                       │
│                                         │
│ account-events                          │
│   ├─ account.created                    │
│   ├─ account.balance.updated            │
│   ├─ account.frozen                     │
│   └─ account.closed                     │
│                                         │
│ transaction-events                      │
│   ├─ transaction.initiated              │
│   ├─ transaction.processing             │
│   ├─ transaction.completed              │
│   ├─ transaction.failed                 │
│   └─ transaction.reversed               │
│                                         │
│ fraud-events                            │
│   ├─ fraud.detected                     │
│   ├─ fraud.high_risk                    │
│   └─ account.suspicious                 │
│                                         │
│ notification-events                     │
│   └─ notification.sent                  │
│                                         │
└─────────────────────────────────────────┘
```

### Event Schema (CloudEvents Standard)

```json
{
  "specversion": "1.0",
  "type": "com.bankflow.transaction.completed",
  "source": "transaction-service",
  "id": "A234-1234-1234",
  "time": "2024-12-18T12:00:00Z",
  "datacontenttype": "application/json",
  "data": {
    "transactionId": "txn_123456",
    "fromAccountId": "acc_789",
    "toAccountId": "acc_456",
    "amount": 100.00,
    "currency": "USD",
    "status": "completed"
  }
}
```

---

## Security & Compliance

### Authentication & Authorization

**JWT Token Flow:**
```
1. User logs in → Identity Service
2. Identity Service validates credentials
3. Returns JWT token (15 min expiry) + Refresh token (7 days)
4. Client includes JWT in Authorization header
5. API Gateway validates JWT
6. Routes to appropriate service
7. Service validates token claims
```

**JWT Claims:**
```json
{
  "sub": "user_id",
  "email": "user@example.com",
  "roles": ["customer"],
  "customer_id": "cust_123",
  "iat": 1703001600,
  "exp": 1703005200
}
```

### Security Layers

**1. API Gateway Level:**
- Rate limiting (100 req/min per user)
- JWT validation
- CORS policies
- DDoS protection

**2. Service Level:**
- Input validation
- SQL injection prevention
- XSS protection
- Authorization checks

**3. Data Level:**
- Encryption at rest (PII data)
- Encryption in transit (TLS)
- PCI DSS compliance
- Data masking in logs

### Compliance Requirements

**PCI DSS (Payment Card Industry):**
- Secure transmission of card data
- Strong access control
- Regular security testing
- Monitoring and logging

**KYC/AML (Anti-Money Laundering):**
- Customer identity verification
- Transaction monitoring
- Suspicious activity reporting

---

## Infrastructure Architecture

### Kubernetes Resources

```
┌─────────────────────────────────────────────────────┐
│                 EKS Cluster                         │
├─────────────────────────────────────────────────────┤
│                                                     │
│  Namespace: bankflow-prod                          │
│                                                     │
│  ┌─────────────────────────────────────────────┐  │
│  │  API Gateway (Kong)                         │  │
│  │  - Deployment: 3 replicas                   │  │
│  │  - Service: LoadBalancer                    │  │
│  │  - HPA: 3-10 replicas                       │  │
│  └─────────────────────────────────────────────┘  │
│                                                     │
│  ┌─────────────────────────────────────────────┐  │
│  │  Microservices                              │  │
│  │  ├─ Identity Service (3 replicas)           │  │
│  │  ├─ Customer Service (3 replicas)           │  │
│  │  ├─ Account Service (3 replicas)            │  │
│  │  ├─ Transaction Service (5 replicas)        │  │
│  │  ├─ Notification Service (3 replicas)       │  │
│  │  └─ Fraud Service (3 replicas)              │  │
│  └─────────────────────────────────────────────┘  │
│                                                     │
│  ┌─────────────────────────────────────────────┐  │
│  │  Data Layer                                 │  │
│  │  ├─ PostgreSQL (StatefulSet)                │  │
│  │  ├─ MongoDB (StatefulSet)                   │  │
│  │  └─ Redis (StatefulSet)                     │  │
│  └─────────────────────────────────────────────┘  │
│                                                     │
│  ┌─────────────────────────────────────────────┐  │
│  │  Message Broker                             │  │
│  │  └─ Kafka (StatefulSet - 3 brokers)         │  │
│  │     └─ Zookeeper (StatefulSet - 3 nodes)    │  │
│  └─────────────────────────────────────────────┘  │
│                                                     │
│  ┌─────────────────────────────────────────────┐  │
│  │  Observability                              │  │
│  │  ├─ Prometheus (metrics)                    │  │
│  │  ├─ Grafana (dashboards)                    │  │
│  │  ├─ Loki (logs)                             │  │
│  │  └─ Jaeger (tracing)                        │  │
│  └─────────────────────────────────────────────┘  │
│                                                     │
└─────────────────────────────────────────────────────┘
```

### AWS Services Integration

```
┌─────────────────────────────────────────────┐
│           AWS Services                      │
├─────────────────────────────────────────────┤
│                                             │
│  EKS (Elastic Kubernetes Service)           │
│  ├─ Managed Kubernetes control plane        │
│  ├─ Worker nodes (EC2)                      │
│  └─ Node groups (on-demand + spot)          │
│                                             │
│  RDS (Relational Database Service)          │
│  ├─ PostgreSQL (Multi-AZ)                   │
│  └─ Automated backups                       │
│                                             │
│  ElastiCache                                │
│  └─ Redis (Multi-AZ)                        │
│                                             │
│  MSK (Managed Kafka)                        │
│  └─ Kafka cluster (3 brokers)               │
│                                             │
│  S3 (Simple Storage Service)                │
│  ├─ Document storage (KYC docs)             │
│  └─ Backup storage                          │
│                                             │
│  CloudWatch                                 │
│  ├─ Log aggregation                         │
│  └─ Alarms and alerts                       │
│                                             │
│  ALB (Application Load Balancer)            │
│  └─ External traffic routing                │
│                                             │
│  Route53                                    │
│  └─ DNS management                          │
│                                             │
│  ACM (Certificate Manager)                  │
│  └─ SSL/TLS certificates                    │
│                                             │
│  Secrets Manager                            │
│  └─ Store sensitive configs                 │
│                                             │
└─────────────────────────────────────────────┘
```

---

## Project Structure

### Repository Organization

```
bankflow/
├── README.md
├── docker-compose.yml           # Local development
├── k8s/                         # Kubernetes manifests
│   ├── base/                    # Base configurations
│   ├── overlays/
│   │   ├── dev/
│   │   ├── staging/
│   │   └── prod/
│   └── terraform/               # EKS infrastructure
│
├── services/
│   ├── identity-service/        # Golang
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── handlers/
│   │   │   ├── models/
│   │   │   ├── repositories/
│   │   │   └── services/
│   │   ├── pkg/
│   │   ├── Dockerfile
│   │   └── go.mod
│   │
│   ├── customer-service/        # Java Spring Boot
│   │   ├── src/
│   │   │   └── main/
│   │   │       ├── java/
│   │   │       │   └── com/bankflow/customer/
│   │   │       │       ├── controller/
│   │   │       │       ├── service/
│   │   │       │       ├── repository/
│   │   │       │       ├── model/
│   │   │       │       └── config/
│   │   │       └── resources/
│   │   ├── Dockerfile
│   │   └── pom.xml
│   │
│   ├── account-service/         # Golang
│   │
