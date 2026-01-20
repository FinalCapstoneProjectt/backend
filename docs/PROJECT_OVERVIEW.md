# Project Overview - University Project Hub

## 🎯 What This System Does

University Project Hub is an **academic governance platform** that transforms how universities manage final-year project proposals from initial idea to published archive.

### The Problem It Solves

**Before**: Informal, memory-based evaluation

- Teachers remember past topics in their heads
- No standard process for reviews
- Lost history of changes
- Disputes about "what was submitted"
- Duplicate projects slip through

**After**: Transparent, auditable, workflow-driven process

- System remembers everything (institutional memory)
- Standard workflow for all proposals
- Complete version history preserved
- Every decision traceable
- Similarity detection prevents duplication

---

## 👥 User Roles & What They Do

### 🎓 Students

- Form teams (2-5 members)
- Create proposal drafts
- **Team Leader**: Only one who can submit proposals
- Respond to teacher feedback
- Upload final documentation
- Cannot skip any steps

### 👨‍🏫 Teachers (Advisors)

- Approve teams
- Review submitted proposals
- Provide structured feedback (approve/revise/reject)
- Approve final documentation
- Publish projects to public archive
- Only people who make academic decisions

### 🔧 Admins

- Manage users and departments
- View audit logs
- Configure system settings
- Cannot touch academic content (governance only)

### 🌐 Public Users

- Browse published projects
- Rate and comment on projects
- Learn from past work
- Zero influence on academic evaluation

### 🤖 AI System

- Analyzes proposals for completeness
- Flags similar past projects
- Suggests improvements
- **NEVER makes decisions** (advisory only)
- **NEVER stores results** (ephemeral analysis)

---

## 📋 Core Workflow

### Step-by-Step Process

```
1. TEAM FORMATION
   Student creates team → Members invited → Advisor approves

2. PROPOSAL CREATION
   Leader creates draft → Adds content → Can get AI suggestions

3. SUBMISSION
   Leader submits → Locks version (immutable) → Advisor notified

4. REVIEW
   Advisor opens review → Evaluates proposal

5. DECISION
   ├─ Approve → Project created → Documentation phase
   ├─ Revise → New version allowed → Back to draft
   └─ Reject → Terminal (cannot continue)

6. PROJECT DOCUMENTATION
   Team uploads reports → Advisor reviews → Approves

7. PUBLICATION
   Advisor publishes → Appears in public archive → Permanent
```

### State Machine Visualization

```
          [Draft]
             ↓ (leader submits)
        [Submitted]
             ↓ (advisor opens)
      [Under Review]
             ↓ (advisor decides)
    ┌────────┼────────┐
    ↓        ↓        ↓
[Approved] [Revise] [Rejected]
    ↓        ↓         ↓
[Project] [Draft]  [End]
          (v2)
```

---

## 🔐 Security & Access Control

### Who Can Do What

| Action               | Student        | Teacher         | Admin | Public |
| -------------------- | -------------- | --------------- | ----- | ------ |
| Create team          | Leader only    | ❌              | ❌    | ❌     |
| Submit proposal      | Leader only    | ❌              | ❌    | ❌     |
| Review proposal      | ❌             | Advisor only    | ❌    | ❌     |
| Approve proposal     | ❌             | Advisor only    | ❌    | ❌     |
| View proposal        | Team + Advisor | Dept. proposals | All   | ❌     |
| Edit draft           | Leader only    | ❌              | ❌    | ❌     |
| Edit submitted       | ❌             | ❌              | ❌    | ❌     |
| View public projects | ✅             | ✅              | ✅    | ✅     |
| Rate projects        | ✅             | ✅              | ❌    | ✅     |
| View audit logs      | ❌             | ❌              | ✅    | ❌     |

### State-Based Locking

| Proposal State    | Can Edit         | Can Submit | Can Review |
| ----------------- | ---------------- | ---------- | ---------- |
| Draft             | Leader           | Leader     | ❌         |
| Submitted         | ❌               | ❌         | Advisor    |
| Under Review      | ❌               | ❌         | Advisor    |
| Revision Required | Leader (new ver) | Leader     | ❌         |
| Approved          | ❌               | ❌         | ❌         |
| Rejected          | ❌               | ❌         | ❌         |

---

## 🔒 Immutability & Audit

### What Never Changes

1. **Proposal Versions**: Once created, cannot be edited or deleted
2. **Feedback Records**: Teacher decisions are permanent
3. **Audit Logs**: Every action logged forever

### What Gets Logged

Every action captures:

- **Who**: User ID, role, email
- **What**: Action type (create, submit, approve, etc.)
- **When**: Precise timestamp
- **Where**: IP address, user agent
- **State**: Before and after snapshots (JSON)

### Why This Matters

- **Academic Integrity**: No "we didn't write that" disputes
- **Accountability**: Teachers' decisions are on record
- **Compliance**: Meets accreditation audit requirements
- **Dispute Resolution**: Complete history for investigations

---

## 🤖 AI Integration Philosophy

### What AI Does

✅ Summarizes proposal content  
✅ Checks for missing sections  
✅ Flags similar past projects  
✅ Suggests improvements

### What AI NEVER Does

❌ Makes approval/rejection decisions  
❌ Stores analysis results  
❌ Influences teacher evaluation  
❌ Has access to proposal history

### Why This Design

1. **Ethical**: Humans maintain authority over academic decisions
2. **Transparent**: Students know they're getting AI suggestions
3. **Non-Biased**: AI cannot create systematic bias (no stored preferences)
4. **Educative**: AI helps students improve, doesn't judge them

---

## 📊 Database Design Highlights

### Immutable Tables (Write-Only)

- `proposal_versions` - **Cannot UPDATE or DELETE**
- `feedback` - **Cannot UPDATE or DELETE**
- `audit_logs` - **Cannot UPDATE or DELETE**

Enforced by database triggers that reject modification attempts.

### Soft Delete Tables

- `users`, `teams`, `proposals`, `projects` - Use `deleted_at` timestamp
- Allows data recovery and historical queries

### Versioning Strategy

```
Proposal (stateful entity)
  ├─ Version 1 (draft submitted)
  ├─ Version 2 (after revision)
  └─ Version 3 (approved version)
```

Each version is completely independent, immutable snapshot.

---

## 🎓 Academic Benefits

### For Students

- Clear process (no confusion about "what next")
- Transparent evaluation (know exactly what teacher said)
- AI guidance (improve before submission)
- Historical reference (learn from past projects)

### For Teachers

- Reduced workload (fewer low-quality submissions)
- Consistent process (same workflow for everyone)
- Similarity detection (catch duplicates easily)
- Audit protection (decisions are documented)

### For University

- Institutional memory (system remembers, not just people)
- Quality assurance (standard process enforced)
- Accreditation compliance (complete audit trail)
- Knowledge base (published projects become library)

---

## 🚀 Technical Highlights

### Clean Architecture

```
HTTP Layer (Gin)
     ↓
Service Layer (Business Logic)
     ↓
Repository Layer (Data Access)
     ↓
Database (PostgreSQL)
```

### Key Design Patterns

- **State Machine**: Enforces proposal workflow
- **Repository Pattern**: Abstract data access
- **Dependency Injection**: Testable, modular code
- **Middleware Pipeline**: Auth → RBAC → Audit
- **Immutable Data**: Append-only for critical records

### Performance Features

- Indexed queries on status, department, dates
- Eager loading for complex queries
- Pagination for large result sets
- Connection pooling for database
- Potential for caching (Redis) on read-heavy endpoints

---

## 📈 Scalability

### Current Design Supports

- **Multi-Department**: Single university, multiple departments
- **High Volume**: Hundreds of proposals per semester
- **Concurrent Users**: 1000+ simultaneous users
- **File Storage**: 10MB per document, unlimited proposals
- **Audit Retention**: 7 years of logs (standard academic requirement)

### Future Extensions

- **Multi-University**: Add tenant isolation
- **Real-Time Collaboration**: WebSockets for live editing
- **Advanced Analytics**: Dashboard for trends
- **External Reviewers**: Industry experts as guest advisors
- **Plagiarism Integration**: Turnitin-like services

---

## 🎯 What Makes This Different

**Most systems**: Store documents  
**This system**: **Enforces academic process**

**Key Differentiators**:

1. **Process-First**: Workflow is enforced, not optional
2. **Immutable History**: Academic integrity by design
3. **Audit-Native**: Every action logged from day one
4. **Human-Controlled**: AI suggests, humans decide
5. **Institutional Memory**: System externalizes knowledge

---

## 📝 Implementation Status

### ✅ Completed

- Architecture design
- Database schema
- API specification
- Implementation guide
- Documentation

### 🚧 To Implement

- Authentication system
- Team management
- Proposal workflow
- State machine enforcement
- AI integration
- Notification system
- Admin dashboard

### 📅 Timeline

- **Week 1-2**: Foundation (auth, models, middleware)
- **Week 3-4**: Core workflow (teams, proposals)
- **Week 5-6**: Reviews and projects
- **Week 7-8**: AI integration and polish
- **Week 9-10**: Testing and deployment

---

## 🎓 Learning Outcomes

This project demonstrates:

### System Design

- Clean architecture principles
- State machine implementation
- RBAC and security
- Immutable data patterns

### Software Engineering

- RESTful API design
- Database normalization
- Transaction management
- Error handling

### Academic Domain

- Workflow automation
- Academic governance
- Institutional knowledge management
- Audit compliance

### Ethical Computing

- Human-first AI design
- Transparent decision-making
- Privacy and data protection
- Accessibility considerations

---

## 📚 Quick Links

- **[Architecture](docs/ARCHITECTURE.md)**: Complete system design
- **[API Docs](docs/API_SPECIFICATION.md)**: All endpoints with examples
- **[Database](docs/DATABASE_SCHEMA.md)**: Full SQL schema
- **[Implementation](docs/IMPLEMENTATION_GUIDE.md)**: Step-by-step guide

---

**This is not just a file storage system. It's an academic governance platform that brings transparency, accountability, and institutional memory to university project management.**
