# University Project Hub - Backend

> A workflow-driven academic governance system that controls how student project ideas are proposed, reviewed, revised, approved, documented, and archived.

## 🎯 Project Vision

University Project Hub replaces informal, memory-based evaluation with a **transparent, versioned, auditable process** where:

- Humans make decisions, AI only advises
- Every proposal follows a strict lifecycle
- Every revision is preserved
- Every decision is accountable
- Approved projects become institutional knowledge

## 🏗️ Architecture

This is a **Clean Architecture** Go backend with:

- **Layered Structure**: Handler → Service → Repository
- **Immutable Academic Records**: Versions and feedback never modified
- **State Machine Workflows**: Enforced proposal lifecycles
- **Comprehensive Audit Trails**: Every action logged
- **Role-Based Access Control**: Students, Teachers, Admins, Public, AI

### Tech Stack

- **Language**: Go 1.23+
- **Framework**: Gin (HTTP web framework)
- **ORM**: GORM
- **Database**: PostgreSQL 14+
- **Authentication**: JWT (golang-jwt)
- **Config**: Viper
- **Validation**: validator/v10

## 📁 Project Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── config/
│   └── config.go                   # Configuration management
├── docs/
│   ├── ARCHITECTURE.md             # Complete system architecture
│   ├── API_SPECIFICATION.md        # Detailed API documentation
│   ├── DATABASE_SCHEMA.md          # Database schema with SQL
│   └── IMPLEMENTATION_GUIDE.md     # Step-by-step implementation
├── internal/
│   ├── ai_checker/                 # AI advisory (non-authoritative)
│   ├── app/
│   │   ├── bootstrap.go            # App initialization
│   │   ├── middlewares.go          # Auth, RBAC, audit logging
│   │   └── router.go               # Route definitions
│   ├── auth/                       # Authentication & JWT
│   ├── departments/                # Department management
│   ├── documentation/              # Project documentation
│   ├── domain/
│   │   └── models.go               # Domain entities
│   ├── feedback/                   # Teacher feedback
│   ├── files/                      # File upload & validation
│   ├── notifications/              # Notification system
│   ├── projects/                   # Approved projects
│   ├── proposal_versions/          # Immutable versions
│   ├── proposals/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── state_machine.go        # Workflow enforcement
│   ├── reviews/                    # Public project reviews
│   ├── teams/                      # Team management
│   ├── universities/               # University data
│   └── users/                      # User management
├── pkg/
│   ├── database/
│   │   └── postgres.go             # Database connection
│   ├── enums/
│   │   └── enums.go                # Type-safe constants
│   ├── errors/
│   │   └── errors.go               # Custom error handling
│   ├── response/
│   │   └── http.go                 # Standardized responses
│   └── utils/                      # Shared utilities
├── scripts/                        # Deployment & utility scripts
├── .env.example                    # Environment variables template
├── go.mod                          # Go dependencies
└── README.md                       # This file
```

## 🚀 Getting Started

### Prerequisites

- Go 1.23 or higher
- PostgreSQL 14 or higher
- Git

### Installation

1. **Clone the repository**

```bash
git clone <repository-url>
cd backend
```

2. **Install dependencies**

```bash
go mod download
```

3. **Set up database**

```bash
# Create database
createdb university_hub

# Run schema (see docs/DATABASE_SCHEMA.md)
psql university_hub < docs/schema.sql
```

4. **Configure environment**

```bash
cp .env.example .env
# Edit .env with your settings
```

5. **Run the application**

```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`

### Quick Test

```bash
# Health check
curl http://localhost:8080/health

# Expected response:
# {"success":true,"message":"System is healthy","data":null}
```

## 📚 Documentation

| Document                                                    | Description                                                            |
| ----------------------------------------------------------- | ---------------------------------------------------------------------- |
| [**ARCHITECTURE.md**](docs/ARCHITECTURE.md)                 | Complete system design, entities, state machines, RBAC, justifications |
| [**API_SPECIFICATION.md**](docs/API_SPECIFICATION.md)       | All endpoints, request/response formats, examples                      |
| [**DATABASE_SCHEMA.md**](docs/DATABASE_SCHEMA.md)           | Full SQL schema, indexes, constraints, triggers                        |
| [**IMPLEMENTATION_GUIDE.md**](docs/IMPLEMENTATION_GUIDE.md) | Step-by-step development guide                                         |

## 🔑 Core Concepts

### Proposal Workflow

```
Draft → Submitted → Under Review → {Approved | Revision Required | Rejected}
                                      ↓            ↓
                                   Project    New Version
                                              (back to Draft)
```

### Key Principles

1. **Immutability**: Once submitted, versions are never edited (append-only)
2. **Single Leader**: Only team leader can submit proposals
3. **Human Authority**: Teachers make all decisions; AI only suggests
4. **Audit Trail**: Every action logged with actor, timestamp, IP
5. **State Enforcement**: Invalid transitions are blocked at service layer

### User Roles

| Role        | Capabilities                                                |
| ----------- | ----------------------------------------------------------- |
| **Student** | Create teams, work on proposals, submit (leader only)       |
| **Teacher** | Review proposals, provide feedback, approve/reject          |
| **Admin**   | Manage users, departments, view audit logs                  |
| **Public**  | View published projects, rate and comment                   |
| **AI**      | Analyze proposals, suggest improvements (non-authoritative) |

## 🔐 Security Features

- **JWT Authentication**: Secure token-based auth
- **RBAC**: Role-based access control on all endpoints
- **Input Validation**: Request validation with validator/v10
- **Rate Limiting**: Prevent abuse on auth and AI endpoints
- **Audit Logging**: Immutable audit trail for all actions
- **File Validation**: MIME type checking, size limits, hash verification
- **Password Security**: Bcrypt hashing with salt

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/proposals/...

# Integration tests
go test -tags=integration ./...
```

## 📊 API Examples

### Register User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john.doe@university.edu",
    "password": "SecurePass123!",
    "role": "student",
    "university_id": 1,
    "department_id": 5,
    "student_id": "CS/2024/001"
  }'
```

### Create Team (Authenticated)

```bash
curl -X POST http://localhost:8080/api/v1/teams \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "AI Research Team",
    "department_id": 5,
    "advisor_id": 12,
    "member_ids": [43, 44, 45]
  }'
```

### Submit Proposal

```bash
curl -X POST http://localhost:8080/api/v1/proposals/15/submit \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"acknowledgement": true}'
```

See [API_SPECIFICATION.md](docs/API_SPECIFICATION.md) for complete API documentation.

## 🔄 Development Workflow

1. **Create Feature Branch**

```bash
git checkout -b feature/team-management
```

2. **Implement Following Clean Architecture**
   - Start with domain models
   - Implement repository (data layer)
   - Implement service (business logic)
   - Implement handler (HTTP layer)
   - Add tests

3. **Test Thoroughly**

```bash
go test ./internal/teams/...
```

4. **Commit with Meaningful Messages**

```bash
git commit -m "feat: implement team creation with member invitations"
```

5. **Create Pull Request**

## 🐛 Troubleshooting

### Database Connection Issues

```bash
# Check PostgreSQL is running
pg_isready

# Verify connection string
psql "host=localhost port=5432 user=postgres dbname=university_hub"
```

### JWT Token Errors

- Ensure `JWT_SECRET` is set in .env
- Check token expiration (24 hours by default)
- Verify Authorization header format: `Bearer <token>`

### Migration Failures

```bash
# Drop and recreate database
dropdb university_hub
createdb university_hub
psql university_hub < docs/schema.sql
```

## 🚀 Deployment

### Docker Deployment

```bash
# Build image
docker build -t university-hub-backend .

# Run with docker-compose
docker-compose up -d
```

### Production Checklist

- [ ] Set strong `JWT_SECRET`
- [ ] Enable SSL for database connections
- [ ] Configure CORS for frontend domain
- [ ] Set up backup strategy for database
- [ ] Enable audit log retention policy
- [ ] Configure rate limiting
- [ ] Set up monitoring and logging
- [ ] Enable HTTPS

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Open a Pull Request

### Code Style

- Follow Go conventions (gofmt, golint)
- Write meaningful commit messages
- Add tests for new features
- Update documentation

## 📝 License

[Specify your license here]

## 👥 Team

- **Project Lead**: [Your Name]
- **Backend Team**: [Team Members]
- **Advisor**: [Advisor Name]

## 📧 Contact

For questions or support, contact [your-email@university.edu]

---

## 🎓 Academic Context

This is a final-year capstone project for [University Name], Department of Computer Science, demonstrating:

- **System Design**: Clean architecture, state machines, RBAC
- **Software Engineering**: Modular design, testing, documentation
- **Academic Governance**: Workflow automation, audit trails
- **Ethical AI**: Non-authoritative AI, human-first decisions
- **Database Design**: Immutability, referential integrity, audit logging

**Key Differentiator**: Unlike document management systems, this platform enforces academic process, preserves decision history, and externalizes institutional memory while keeping humans in control.

---

**Built with ❤️ using Go and Clean Architecture principles**
