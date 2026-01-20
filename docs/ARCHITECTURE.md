# University Project Hub - Backend Architecture

## 1. System Overview

### Core Philosophy

- **Human-First Decision Making**: AI advises, humans decide
- **Immutable History**: All changes are append-only, never overwritten
- **Audit-Friendly**: Every action is traceable with timestamp, actor, and context
- **Strict Role Separation**: Clear boundaries between Students, Teachers, Admins, Public, AI
- **Single-Leader Governance**: One leader per proposal controls submissions

### System Purpose

University Project Hub externalizes institutional academic memory by transforming informal, memory-based project evaluation into a transparent, versioned, auditable workflow where proposals follow strict lifecycles, every revision is preserved, every decision is accountable, and approved projects become long-term knowledge assets.

---

## 2. Architecture Layers

```
┌─────────────────────────────────────────────────────────────┐
│                     API Layer (Gin HTTP)                      │
│  ┌──────────┬──────────┬──────────┬──────────┬──────────┐  │
│  │  Auth    │  Teams   │ Proposals│ Projects │   AI     │  │
│  │ Handlers │ Handlers │ Handlers │ Handlers │ Handlers │  │
│  └──────────┴──────────┴──────────┴──────────┴──────────┘  │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                    Middleware Layer                          │
│  • Authentication (JWT)                                      │
│  • Authorization (RBAC)                                      │
│  • Audit Logging                                             │
│  • Request Validation                                        │
│  • Rate Limiting                                             │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                    Service Layer                             │
│  • Business Logic                                            │
│  • State Machine Enforcement                                 │
│  • Version Control Logic                                     │
│  • Event Publishing                                          │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                    Repository Layer                          │
│  • Data Access Abstraction                                   │
│  • Query Optimization                                        │
│  • Transaction Management                                    │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                    Data Layer                                │
│  • PostgreSQL (Primary data store)                           │
│  • Audit Log Tables (Write-only, immutable)                  │
│  • File Storage (S3/Local) - Documents                       │
└─────────────────────────────────────────────────────────────┘
```

---

## 3. Core Entities & Relationships

### Entity Relationship Diagram

```
┌──────────────┐
│ University   │
└──────┬───────┘
       │ 1:N
       ↓
┌──────────────┐      ┌──────────────┐
│ Department   │←────→│     User     │
└──────┬───────┘  N:1 └──────┬───────┘
       │ 1:N             │ N:1
       ↓                 ↓
┌──────────────┐      ┌──────────────┐
│    Team      │←────→│ TeamMember   │
└──────┬───────┘  1:N └──────────────┘
       │ 1:1
       ↓
┌──────────────┐
│  Proposal    │──────┐
└──────┬───────┘      │ 1:N
       │ 1:N          ↓
       │         ┌──────────────┐
       ├────────→│   Version    │
       │         └──────────────┘
       │ 1:N          ↑ N:1
       ↓              │
┌──────────────┐      │
│  Feedback    │──────┘
└──────────────┘

┌──────────────┐
│  Proposal    │
│ (Approved)   │
└──────┬───────┘
       │ 1:1
       ↓
┌──────────────┐      ┌────────────────┐
│   Project    │←────→│ Documentation  │
└──────┬───────┘  1:N └────────────────┘
       │ 1:N
       ↓
┌──────────────┐
│ ProjectReview│
└──────────────┘

┌──────────────┐
│     User     │
└──────┬───────┘
       │ 1:N
       ↓
┌──────────────┐
│ Notification │
└──────────────┘

┌──────────────┐
│ AuditLog     │ (All actions)
└──────────────┘
```

### Key Entity Definitions

#### Proposal (Stateful Entity)

```go
type Proposal struct {
    ID               uint
    TeamID           uint
    Status           ProposalStatus  // State machine
    CurrentVersionID uint            // Points to active version
    CreatedAt        time.Time
    UpdatedAt        time.Time
    SubmittedAt      *time.Time      // First submission timestamp
    ApprovedAt       *time.Time      // Approval timestamp
    ApprovedBy       *uint           // Teacher who approved

    // Relations
    Team             Team
    Versions         []ProposalVersion
    Feedback         []Feedback
    Project          *Project        // Only if approved
}
```

**State Machine**:

```
Draft → Submitted → UnderReview → RevisionRequired → Draft (new version)
                                → Approved → Project Created
                                → Rejected (terminal)
```

#### ProposalVersion (Immutable)

```go
type ProposalVersion struct {
    ID                uint
    ProposalID        uint
    Title             string
    Objectives        string
    Methodology       string
    ExpectedOutcomes  string
    FileURL           string          // Stored document
    VersionNumber     int             // Sequential
    CreatedBy         uint            // Always team leader
    CreatedAt         time.Time       // Immutable timestamp
    IsApprovedVersion bool            // True for approved version
    IPAddress         string          // Audit trail
    UserAgent         string          // Audit trail
}
```

**Immutability Rules**:

- Once created, NEVER updated
- Soft delete not allowed
- Hash verification for file integrity

#### Feedback (Decision Record)

```go
type Feedback struct {
    ID                uint
    ProposalID        uint
    ProposalVersionID uint            // Specific version reviewed
    ReviewerID        uint            // Teacher
    Decision          FeedbackDecision // approve, revise, reject
    Comment           string          // Required justification
    IsStructured      bool            // AI-suggested vs manual
    CreatedAt         time.Time
    IPAddress         string

    // Relations
    Reviewer          User
    Version           ProposalVersion
}

type FeedbackDecision string
const (
    DecisionApprove  FeedbackDecision = "approve"
    DecisionRevise   FeedbackDecision = "revise"
    DecisionReject   FeedbackDecision = "reject"
)
```

**Immutability**: Feedback records are never edited or deleted.

#### AuditLog (System-Wide Tracking)

```go
type AuditLog struct {
    ID           uint
    EntityType   string      // "proposal", "team", "user", etc.
    EntityID     uint
    Action       string      // "create", "submit", "approve", "reject"
    ActorID      uint        // Who performed action
    ActorRole    Role
    OldState     string      // JSON snapshot (if applicable)
    NewState     string      // JSON snapshot
    IPAddress    string
    UserAgent    string
    Timestamp    time.Time
    SessionID    string
}
```

**Write-Only**: Audit logs are append-only, never modified or deleted.

---

## 4. State Machine Logic

### Proposal State Transitions

```go
// Valid transitions map
var ProposalStateTransitions = map[ProposalStatus][]ProposalStatus{
    StatusDraft: {
        StatusSubmitted,  // Leader submits
    },
    StatusSubmitted: {
        StatusUnderReview,  // Teacher opens for review
    },
    StatusUnderReview: {
        StatusRevisionRequired,  // Teacher requests changes
        StatusApproved,          // Teacher approves
        StatusRejected,          // Teacher rejects
    },
    StatusRevisionRequired: {
        StatusDraft,  // System auto-transitions when new version created
    },
    StatusApproved: {
        // Terminal state
    },
    StatusRejected: {
        // Terminal state
    },
}

// State transition rules
type StateTransitionRule struct {
    FromState       ProposalStatus
    ToState         ProposalStatus
    AllowedRoles    []Role
    RequiredChecks  []func(*Proposal, *User) error
}

// Example rules
var transitionRules = []StateTransitionRule{
    {
        FromState:    StatusDraft,
        ToState:      StatusSubmitted,
        AllowedRoles: []Role{RoleStudent},
        RequiredChecks: []func(*Proposal, *User) error{
            checkIsTeamLeader,
            checkMinVersionRequirements,
            checkTeamApproved,
        },
    },
    {
        FromState:    StatusUnderReview,
        ToState:      StatusApproved,
        AllowedRoles: []Role{RoleTeacher},
        RequiredChecks: []func(*Proposal, *User) error{
            checkIsTeamAdvisor,
            checkVersionCompleteness,
        },
    },
}
```

### Version Locking Logic

```go
// Determines if proposal can be edited
func (p *Proposal) IsEditable() bool {
    return p.Status == StatusDraft
}

// Determines if new version can be created
func (p *Proposal) CanCreateNewVersion() bool {
    return p.Status == StatusRevisionRequired
}

// Determines if proposal can be submitted
func (p *Proposal) CanSubmit(user *User) bool {
    return p.Status == StatusDraft &&
           p.Team.LeaderID == user.ID &&
           p.Team.Status == TeamStatusApproved
}
```

---

## 5. Role-Based Access Control (RBAC)

### Role Definitions

```go
type Role string
const (
    RoleStudent  Role = "student"
    RoleTeacher  Role = "teacher"
    RoleAdmin    Role = "admin"
    RolePublic   Role = "public"
    RoleAI       Role = "ai_system"  // Special role for AI endpoints
)

type Permission string
const (
    // Proposal permissions
    PermProposalCreate     Permission = "proposal:create"
    PermProposalRead       Permission = "proposal:read"
    PermProposalUpdate     Permission = "proposal:update"
    PermProposalSubmit     Permission = "proposal:submit"
    PermProposalReview     Permission = "proposal:review"
    PermProposalApprove    Permission = "proposal:approve"

    // Project permissions
    PermProjectRead        Permission = "project:read"
    PermProjectPublish     Permission = "project:publish"

    // Admin permissions
    PermUserManage         Permission = "user:manage"
    PermSystemConfigure    Permission = "system:configure"
)

var RolePermissions = map[Role][]Permission{
    RoleStudent: {
        PermProposalCreate,
        PermProposalRead,
        PermProposalUpdate,  // Only if leader + draft state
        PermProposalSubmit,  // Only if leader
    },
    RoleTeacher: {
        PermProposalRead,    // Only for their department
        PermProposalReview,
        PermProposalApprove,
        PermProjectPublish,
    },
    RoleAdmin: {
        PermUserManage,
        PermSystemConfigure,
        PermProjectRead,     // All projects
    },
    RolePublic: {
        PermProjectRead,     // Only published projects
    },
}
```

### Context-Based Access Control

Beyond role, access is determined by:

1. **Ownership**: Is user the team leader?
2. **State**: Is the proposal in an editable state?
3. **Department**: Does teacher belong to proposal's department?
4. **Visibility**: Is project published?

```go
type AccessContext struct {
    User       *User
    Resource   interface{}  // Proposal, Project, etc.
    Action     Permission
}

func (ac *AccessContext) IsAuthorized() bool {
    // Check role-based permissions
    if !hasPermission(ac.User.Role, ac.Action) {
        return false
    }

    // Check context-specific rules
    switch resource := ac.Resource.(type) {
    case *Proposal:
        return ac.canAccessProposal(resource)
    case *Project:
        return ac.canAccessProject(resource)
    }

    return false
}

func (ac *AccessContext) canAccessProposal(p *Proposal) bool {
    switch ac.Action {
    case PermProposalUpdate, PermProposalSubmit:
        // Only team leader in draft state
        return p.Team.LeaderID == ac.User.ID && p.IsEditable()

    case PermProposalReview, PermProposalApprove:
        // Only advisor of the team
        return p.Team.AdvisorID == ac.User.ID

    case PermProposalRead:
        // Team members, advisor, or admins
        return p.Team.HasMember(ac.User.ID) ||
               p.Team.AdvisorID == ac.User.ID ||
               ac.User.Role == RoleAdmin
    }

    return false
}
```

---

## 6. API Design

### 6.1 Authentication & Authorization

#### POST /api/v1/auth/register

**Access**: Public
**Purpose**: User registration

**Request**:

```json
{
  "name": "John Doe",
  "email": "john.doe@university.edu",
  "password": "SecurePassword123!",
  "role": "student",
  "university_id": 1,
  "department_id": 5,
  "student_id": "CS/2024/001"
}
```

**Response**:

```json
{
  "success": true,
  "message": "Registration successful",
  "data": {
    "user": {
      "id": 42,
      "name": "John Doe",
      "email": "john.doe@university.edu",
      "role": "student"
    },
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

**Validation**:

- Email must be unique and valid format
- Password: min 8 chars, 1 uppercase, 1 number, 1 special
- Role must be valid enum
- University and department must exist
- Student ID required for students

**Audit**: Log registration with IP and timestamp

---

#### POST /api/v1/auth/login

**Access**: Public

**Request**:

```json
{
  "email": "john.doe@university.edu",
  "password": "SecurePassword123!"
}
```

**Response**:

```json
{
  "success": true,
  "data": {
    "user": {
      "id": 42,
      "name": "John Doe",
      "role": "student",
      "department": {...}
    },
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_at": "2026-01-18T12:00:00Z"
  }
}
```

**Security**:

- Rate limit: 5 attempts per 15 minutes per IP
- Log failed attempts
- Account lock after 5 failed attempts

---

### 6.2 Team Management

#### POST /api/v1/teams

**Access**: Students only
**Purpose**: Create a new team

**Request**:

```json
{
  "name": "AI Research Team",
  "department_id": 5,
  "advisor_id": 12,
  "member_ids": [43, 44, 45]
}
```

**Response**:

```json
{
  "success": true,
  "data": {
    "team": {
      "id": 7,
      "name": "AI Research Team",
      "status": "pending_advisor_approval",
      "leader": {...},
      "members": [...]
    }
  }
}
```

**Business Rules**:

- Creator automatically becomes leader
- All members must be students in same department
- Team size: 1-5 members
- Advisor must be a teacher in the department
- All members receive invitation notifications
- Advisor receives approval request notification

**State**: Team starts in `pending_advisor_approval`

**Audit**: Log team creation with all members

---

#### POST /api/v1/teams/:id/invitations/:invitation_id/respond

**Access**: Invited student only

**Request**:

```json
{
  "response": "accept" // or "reject"
}
```

**Response**:

```json
{
  "success": true,
  "message": "Invitation accepted",
  "data": {
    "team_member": {
      "team_id": 7,
      "user_id": 43,
      "status": "accepted"
    }
  }
}
```

**Business Rules**:

- Only invited user can respond
- Cannot respond twice
- Leader notified of response
- If all members accept + advisor approves → team becomes "approved"

---

#### POST /api/v1/teams/:id/advisor-response

**Access**: Designated advisor only

**Request**:

```json
{
  "decision": "approve", // or "reject"
  "comment": "Strong team composition for ML project"
}
```

**Response**:

```json
{
  "success": true,
  "message": "Team approved",
  "data": {
    "team": {
      "id": 7,
      "status": "approved"
    }
  }
}
```

**Business Rules**:

- Only designated advisor can respond
- If rejected, team cannot create proposals
- If approved, team can create proposals
- All team members notified

**Audit**: Log advisor decision with justification

---

### 6.3 Proposal Management (Core Workflow)

#### POST /api/v1/proposals

**Access**: Team leaders of approved teams
**Purpose**: Create initial draft proposal

**Request**:

```json
{
  "team_id": 7
}
```

**Response**:

```json
{
  "success": true,
  "data": {
    "proposal": {
      "id": 15,
      "team_id": 7,
      "status": "draft",
      "current_version_id": null,
      "created_at": "2026-01-17T10:00:00Z"
    }
  }
}
```

**Business Rules**:

- Only team leader can create
- Team must be approved
- One proposal per team (can check if existing)
- Starts in `draft` state

**Audit**: Log proposal creation

---

#### POST /api/v1/proposals/:id/versions

**Access**: Team leader only
**Purpose**: Create/update proposal version

**Request**:

```json
{
  "title": "AI-Powered Learning Platform",
  "objectives": "Develop adaptive learning system...",
  "methodology": "Agile development with ML integration...",
  "expected_outcomes": "Functional platform with 90% accuracy...",
  "file": "<multipart/form-data>"
}
```

**Response**:

```json
{
  "success": true,
  "data": {
    "version": {
      "id": 42,
      "proposal_id": 15,
      "version_number": 1,
      "title": "AI-Powered Learning Platform",
      "file_url": "/storage/proposals/15/v1.pdf",
      "created_by": 42,
      "created_at": "2026-01-17T10:30:00Z"
    }
  }
}
```

**Business Rules**:

- Only when status is `draft` or `revision_required`
- Auto-increment version number
- File must be PDF, max 10MB
- Store file with versioned path
- Update proposal.current_version_id
- Capture IP address and user agent

**Validation**:

- Title: required, 10-200 chars
- Objectives: required, min 100 chars
- Methodology: required, min 100 chars
- File: required, PDF only

**Audit**: Log version creation with file hash

---

#### POST /api/v1/proposals/:id/submit

**Access**: Team leader only
**Purpose**: Submit proposal for review

**Request**:

```json
{
  "acknowledgement": true
}
```

**Response**:

```json
{
  "success": true,
  "message": "Proposal submitted for review",
  "data": {
    "proposal": {
      "id": 15,
      "status": "submitted",
      "submitted_at": "2026-01-17T11:00:00Z"
    }
  }
}
```

**Business Rules**:

- Only from `draft` state
- Must have at least one complete version
- Transitions to `submitted` state
- Version becomes locked (immutable)
- Advisor receives notification
- Cannot edit after submission

**Edge Cases**:

- Prevent double submission (idempotent if already submitted)
- Check team still approved
- Validate version completeness

**Audit**: Log submission with timestamp and version ID

**Notifications**:

- Advisor: "New proposal submitted for review"
- Team members: "Proposal submitted by leader"

---

#### POST /api/v1/proposals/:id/start-review

**Access**: Team advisor only
**Purpose**: Mark proposal as under review

**Request**: Empty body

**Response**:

```json
{
  "success": true,
  "data": {
    "proposal": {
      "id": 15,
      "status": "under_review"
    }
  }
}
```

**Business Rules**:

- Only from `submitted` state
- Only advisor of the team
- Signals active review in progress

**Audit**: Log review start

---

#### POST /api/v1/proposals/:id/feedback

**Access**: Team advisor only
**Purpose**: Provide feedback and decision

**Request**:

```json
{
  "version_id": 42,
  "decision": "revise", // or "approve" or "reject"
  "comment": "Methodology section needs more detail on data collection. Expected outcomes should include specific metrics."
}
```

**Response**:

```json
{
  "success": true,
  "message": "Feedback submitted",
  "data": {
    "feedback": {
      "id": 8,
      "proposal_id": 15,
      "version_id": 42,
      "decision": "revise",
      "comment": "...",
      "created_at": "2026-01-17T14:00:00Z"
    },
    "proposal": {
      "id": 15,
      "status": "revision_required"
    }
  }
}
```

**Business Rules**:

- Only from `under_review` state
- Only advisor can provide feedback
- Comment required (min 20 chars)
- Decision determines state transition:
  - `approve` → `approved` (terminal, creates project)
  - `revise` → `revision_required` (unlocks editing)
  - `reject` → `rejected` (terminal)
- Feedback is immutable once submitted
- Version ID must match current version

**State Transitions**:

- **Approve**:
  - Set proposal.status = `approved`
  - Set proposal.approved_at = now
  - Set proposal.approved_by = reviewer_id
  - Set version.is_approved_version = true
  - Trigger project creation
  - Notify team: "Proposal approved!"
- **Revise**:
  - Set proposal.status = `revision_required`
  - Notify leader: "Revisions requested"
  - New version can be created
- **Reject**:
  - Set proposal.status = `rejected`
  - Terminal state, no further action
  - Notify team: "Proposal rejected"

**Audit**: Log feedback with decision, version ID, and timestamp

---

#### GET /api/v1/proposals/:id

**Access**: Team members, advisor, admin

**Response**:

```json
{
  "success": true,
  "data": {
    "proposal": {
      "id": 15,
      "team": {...},
      "status": "under_review",
      "current_version": {...},
      "versions": [...],
      "feedback": [...],
      "created_at": "...",
      "submitted_at": "...",
      "can_edit": false,
      "can_submit": false
    }
  }
}
```

**Business Rules**:

- Team members see full history
- Include state transition permissions
- Feedback sorted by creation date

---

#### GET /api/v1/proposals

**Access**: Role-dependent

**Query Parameters**:

```
?status=under_review&department_id=5&page=1&limit=20
```

**Response**:

```json
{
  "success": true,
  "data": {
    "proposals": [...],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 45
    }
  }
}
```

**Filtering Rules**:

- **Students**: Only their team's proposal
- **Teachers**: Proposals from their department or teams they advise
- **Admins**: All proposals

---

### 6.4 AI Advisory Endpoints

#### POST /api/v1/ai/analyze-proposal

**Access**: Students (team leader), Teachers
**Purpose**: Get AI suggestions BEFORE official submission

**Request**:

```json
{
  "proposal_id": 15,
  "version_id": 42
}
```

**Response**:

```json
{
  "success": true,
  "data": {
    "analysis": {
      "summary": "The proposal focuses on developing an adaptive AI learning platform...",
      "completeness_score": 85,
      "missing_sections": ["Risk analysis", "Timeline details"],
      "structure_issues": [
        "Objectives section could be more specific with measurable outcomes"
      ],
      "similarity_flags": [
        {
          "project_id": 23,
          "similarity_score": 0.42,
          "overlap_areas": ["Machine learning methodology"]
        }
      ],
      "suggestions": [
        "Include specific success metrics",
        "Add detailed project timeline"
      ]
    }
  }
}
```

**Critical Design Choices**:

1. **Non-Authoritative**: Results are suggestions, not decisions
2. **No Storage**: AI analysis not saved in database
3. **No Impact on Approval**: Teachers never see AI scores
4. **Transparent**: Students know this is AI-generated

**Business Rules**:

- Can only analyze own team's proposals (students) or advised proposals (teachers)
- Can be called multiple times (no limit)
- Does not change proposal state
- Does not create audit log (advisory only)

**Use Case**:

- Student creates version
- Student calls AI analysis
- Student improves based on suggestions
- Student submits when confident
- Teacher reviews WITHOUT seeing AI analysis

**Why This Matters**:

- Reduces low-quality submissions
- Reduces teacher workload
- Does NOT replace human judgment
- Does NOT create bias in evaluation

---

#### GET /api/v1/ai/check-similarity

**Access**: Teachers only
**Purpose**: Check similarity with existing projects (during review)

**Query Parameters**:

```
?proposal_id=15&threshold=0.3
```

**Response**:

```json
{
  "success": true,
  "data": {
    "similar_projects": [
      {
        "project_id": 23,
        "title": "Adaptive Learning System",
        "department": "Computer Science",
        "year": 2024,
        "similarity_score": 0.52,
        "overlap_summary": "Both use reinforcement learning for personalization"
      }
    ]
  }
}
```

**Business Rules**:

- Only teachers can check
- Only checks against approved, public projects
- Provides context, not judgment
- Teacher decides if similarity is problematic

---

### 6.5 Project Management (Post-Approval)

#### GET /api/v1/projects/:id

**Access**: Role-dependent

**Response**:

```json
{
  "success": true,
  "data": {
    "project": {
      "id": 8,
      "proposal_id": 15,
      "team": {...},
      "summary": "AI-Powered Learning Platform",
      "visibility": "private",
      "documentation": [...],
      "reviews": [...],
      "average_rating": 4.5,
      "created_at": "..."
    }
  }
}
```

**Access Rules**:

- **Private**: Team members, advisor, admin only
- **Published**: Everyone (including public users)

---

#### POST /api/v1/projects/:id/documentation

**Access**: Team members only
**Purpose**: Upload final documentation

**Request**:

```json
{
  "document_type": "final_report", // or "presentation", "code"
  "file": "<multipart/form-data>"
}
```

**Response**:

```json
{
  "success": true,
  "data": {
    "documentation": {
      "id": 12,
      "project_id": 8,
      "document_type": "final_report",
      "file_url": "/storage/projects/8/final_report.pdf",
      "status": "pending",
      "submitted_by": 42,
      "submitted_at": "2026-01-17T16:00:00Z"
    }
  }
}
```

**Business Rules**:

- Only team members
- Document types: final_report, presentation, proposal_document, code_repository
- Files are immutable once uploaded
- Each type can only be uploaded once
- Advisor receives notification for review

---

#### POST /api/v1/projects/:id/publish

**Access**: Team advisor only
**Purpose**: Make project publicly visible

**Request**:

```json
{
  "acknowledgement": true
}
```

**Response**:

```json
{
  "success": true,
  "message": "Project published",
  "data": {
    "project": {
      "id": 8,
      "visibility": "public"
    }
  }
}
```

**Business Rules**:

- Only advisor can publish
- All required documentation must be uploaded and approved
- Once published, project appears in public archive
- Cannot be unpublished (permanent)

---

#### GET /api/v1/projects/public

**Access**: Everyone (including non-authenticated)

**Query Parameters**:

```
?department_id=5&year=2025&search=machine learning&page=1
```

**Response**:

```json
{
  "success": true,
  "data": {
    "projects": [
      {
        "id": 8,
        "title": "AI-Powered Learning Platform",
        "summary": "...",
        "department": "Computer Science",
        "team_members": ["John Doe", "Jane Smith"],
        "year": 2025,
        "average_rating": 4.5,
        "view_count": 234
      }
    ],
    "pagination": {...}
  }
}
```

**Business Rules**:

- Only published projects
- Anonymized team member names (optional, configurable)
- Sorted by rating, date, or views

---

#### POST /api/v1/projects/:id/reviews

**Access**: Public users (authenticated)
**Purpose**: Rate and comment on public projects

**Request**:

```json
{
  "rating": 5,
  "comment": "Excellent implementation of ML concepts"
}
```

**Response**:

```json
{
  "success": true,
  "data": {
    "review": {
      "id": 45,
      "project_id": 8,
      "rating": 5,
      "comment": "...",
      "created_at": "..."
    }
  }
}
```

**Business Rules**:

- Only authenticated users
- One review per user per project
- Rating: 1-5 stars
- Comment optional
- Does NOT affect academic evaluation

---

### 6.6 Notification System

#### GET /api/v1/notifications

**Access**: Authenticated users

**Response**:

```json
{
  "success": true,
  "data": {
    "notifications": [
      {
        "id": 123,
        "message": "Your proposal has been approved!",
        "reference_type": "proposal",
        "reference_id": 15,
        "is_read": false,
        "created_at": "2026-01-17T14:00:00Z"
      }
    ],
    "unread_count": 3
  }
}
```

---

#### POST /api/v1/notifications/:id/mark-read

**Access**: Notification owner only

**Response**:

```json
{
  "success": true,
  "message": "Notification marked as read"
}
```

---

### 6.7 Admin Endpoints

#### POST /api/v1/admin/users

**Access**: Admin only
**Purpose**: Create user accounts (bulk import)

#### GET /api/v1/admin/audit-logs

**Access**: Admin only
**Purpose**: View system audit trail

**Query Parameters**:

```
?entity_type=proposal&entity_id=15&actor_id=42&from_date=2026-01-01
```

---

## 7. Database Schema Design

### Schema Principles

1. **Immutable Tables**: `proposal_versions`, `feedback`, `audit_logs` have no UPDATE queries
2. **Soft Deletes**: Use `deleted_at` for user-facing entities, never for audit data
3. **Timestamp Everything**: Every table has `created_at`, mutable tables have `updated_at`
4. **Foreign Key Constraints**: Enforce referential integrity
5. **Indexes**: On foreign keys, status fields, and query patterns

### Core Tables

```sql
-- Universities
CREATE TABLE universities (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Departments
CREATE TABLE departments (
    id SERIAL PRIMARY KEY,
    university_id INTEGER REFERENCES universities(id),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(university_id, name)
);

-- Users
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('student', 'teacher', 'admin', 'public')),
    university_id INTEGER REFERENCES universities(id),
    department_id INTEGER REFERENCES departments(id),
    student_id VARCHAR(50),
    profile_photo VARCHAR(500),
    is_active BOOLEAN DEFAULT TRUE,
    email_verified BOOLEAN DEFAULT FALSE,
    failed_login_attempts INTEGER DEFAULT 0,
    last_login_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_department ON users(department_id);

-- Teams
CREATE TABLE teams (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    department_id INTEGER REFERENCES departments(id),
    leader_id INTEGER REFERENCES users(id),  -- Team leader
    advisor_id INTEGER REFERENCES users(id),  -- Assigned teacher
    status VARCHAR(30) DEFAULT 'pending_advisor_approval'
           CHECK (status IN ('pending_advisor_approval', 'approved', 'rejected')),
    advisor_response_comment TEXT,
    advisor_responded_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_teams_status ON teams(status);
CREATE INDEX idx_teams_advisor ON teams(advisor_id);

-- Team Members (Many-to-Many with invitation tracking)
CREATE TABLE team_members (
    team_id INTEGER REFERENCES teams(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id),
    role VARCHAR(20) DEFAULT 'member' CHECK (role IN ('leader', 'member')),
    invitation_status VARCHAR(20) DEFAULT 'pending'
                      CHECK (invitation_status IN ('pending', 'accepted', 'rejected')),
    invited_at TIMESTAMP DEFAULT NOW(),
    responded_at TIMESTAMP,
    PRIMARY KEY (team_id, user_id)
);
CREATE INDEX idx_team_members_status ON team_members(invitation_status);

-- Proposals (Stateful entity)
CREATE TABLE proposals (
    id SERIAL PRIMARY KEY,
    team_id INTEGER UNIQUE REFERENCES teams(id),  -- One proposal per team
    status VARCHAR(30) DEFAULT 'draft'
           CHECK (status IN ('draft', 'submitted', 'under_review',
                            'revision_required', 'approved', 'rejected')),
    current_version_id INTEGER,  -- FK added later
    submitted_at TIMESTAMP,
    approved_at TIMESTAMP,
    approved_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_proposals_status ON proposals(status);
CREATE INDEX idx_proposals_team ON proposals(team_id);

-- Proposal Versions (IMMUTABLE)
CREATE TABLE proposal_versions (
    id SERIAL PRIMARY KEY,
    proposal_id INTEGER REFERENCES proposals(id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    objectives TEXT NOT NULL,
    methodology TEXT NOT NULL,
    expected_outcomes TEXT NOT NULL,
    file_url VARCHAR(500) NOT NULL,
    file_hash VARCHAR(64),  -- SHA-256 for integrity
    version_number INTEGER NOT NULL,
    is_approved_version BOOLEAN DEFAULT FALSE,
    created_by INTEGER REFERENCES users(id),  -- Always team leader
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(proposal_id, version_number)
);
CREATE INDEX idx_versions_proposal ON proposal_versions(proposal_id);
CREATE INDEX idx_versions_approved ON proposal_versions(is_approved_version);

-- Add foreign key constraint
ALTER TABLE proposals
ADD CONSTRAINT fk_current_version
FOREIGN KEY (current_version_id) REFERENCES proposal_versions(id);

-- Feedback (IMMUTABLE)
CREATE TABLE feedback (
    id SERIAL PRIMARY KEY,
    proposal_id INTEGER REFERENCES proposals(id),
    proposal_version_id INTEGER REFERENCES proposal_versions(id),
    reviewer_id INTEGER REFERENCES users(id),  -- Teacher
    decision VARCHAR(20) NOT NULL CHECK (decision IN ('approve', 'revise', 'reject')),
    comment TEXT NOT NULL,
    is_structured BOOLEAN DEFAULT FALSE,  -- AI-suggested vs manual
    ip_address INET,
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_feedback_proposal ON feedback(proposal_id);
CREATE INDEX idx_feedback_reviewer ON feedback(reviewer_id);
CREATE INDEX idx_feedback_decision ON feedback(decision);

-- Projects (Created from approved proposals)
CREATE TABLE projects (
    id SERIAL PRIMARY KEY,
    proposal_id INTEGER UNIQUE REFERENCES proposals(id),
    team_id INTEGER REFERENCES teams(id),
    summary TEXT NOT NULL,
    department_id INTEGER REFERENCES departments(id),
    approved_by INTEGER REFERENCES users(id),
    visibility VARCHAR(20) DEFAULT 'private' CHECK (visibility IN ('private', 'public')),
    share_count INTEGER DEFAULT 0,
    view_count INTEGER DEFAULT 0,
    published_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_projects_visibility ON projects(visibility);
CREATE INDEX idx_projects_department ON projects(department_id);

-- Project Documentation (IMMUTABLE once uploaded)
CREATE TABLE project_documentation (
    id SERIAL PRIMARY KEY,
    project_id INTEGER REFERENCES projects(id) ON DELETE CASCADE,
    document_type VARCHAR(30) NOT NULL
                  CHECK (document_type IN ('final_report', 'presentation',
                                          'proposal_document', 'code_repository')),
    file_url VARCHAR(500) NOT NULL,
    file_hash VARCHAR(64),
    status VARCHAR(20) DEFAULT 'pending'
           CHECK (status IN ('pending', 'approved', 'rejected')),
    review_comment TEXT,
    reviewed_by INTEGER REFERENCES users(id),
    reviewed_at TIMESTAMP,
    submitted_by INTEGER REFERENCES users(id),
    submitted_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(project_id, document_type)  -- One of each type per project
);
CREATE INDEX idx_documentation_project ON project_documentation(project_id);
CREATE INDEX idx_documentation_status ON project_documentation(status);

-- Project Reviews (Public ratings)
CREATE TABLE project_reviews (
    id SERIAL PRIMARY KEY,
    project_id INTEGER REFERENCES projects(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id),
    rating INTEGER CHECK (rating BETWEEN 1 AND 5),
    comment TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(project_id, user_id)  -- One review per user per project
);
CREATE INDEX idx_reviews_project ON project_reviews(project_id);

-- Notifications
CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    reference_type VARCHAR(50),  -- 'team', 'proposal', 'project', etc.
    reference_id INTEGER,
    message TEXT NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_notifications_user ON notifications(user_id);
CREATE INDEX idx_notifications_unread ON notifications(user_id, is_read);

-- Audit Logs (WRITE-ONLY, NEVER UPDATE/DELETE)
CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    entity_type VARCHAR(50) NOT NULL,
    entity_id INTEGER,
    action VARCHAR(50) NOT NULL,  -- 'create', 'submit', 'approve', etc.
    actor_id INTEGER REFERENCES users(id),
    actor_role VARCHAR(20),
    old_state JSONB,  -- Previous state snapshot
    new_state JSONB,  -- New state snapshot
    ip_address INET,
    user_agent TEXT,
    session_id VARCHAR(255),
    timestamp TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_audit_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_actor ON audit_logs(actor_id);
CREATE INDEX idx_audit_timestamp ON audit_logs(timestamp);
CREATE INDEX idx_audit_action ON audit_logs(action);
```

### Triggers for Audit Logging

```sql
-- Auto-audit trigger for proposals
CREATE OR REPLACE FUNCTION audit_proposal_changes()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO audit_logs (
        entity_type, entity_id, action, actor_id,
        old_state, new_state, timestamp
    ) VALUES (
        'proposal', NEW.id, TG_OP,
        NEW.updated_by,  -- Need to pass from application
        row_to_json(OLD),
        row_to_json(NEW),
        NOW()
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER proposal_audit_trigger
AFTER INSERT OR UPDATE ON proposals
FOR EACH ROW EXECUTE FUNCTION audit_proposal_changes();
```

---

## 8. Security Measures

### 8.1 Authentication Security

```go
// JWT Configuration
type JWTConfig struct {
    Secret           string
    ExpirationTime   time.Duration  // 24 hours
    RefreshTime      time.Duration  // 7 days
    Issuer           string
    Algorithm        string         // HS256
}

// Token Claims
type TokenClaims struct {
    UserID       uint   `json:"user_id"`
    Email        string `json:"email"`
    Role         Role   `json:"role"`
    DepartmentID uint   `json:"department_id"`
    jwt.RegisteredClaims
}

// Password Requirements
const (
    MinPasswordLength = 8
    RequireUppercase  = true
    RequireNumber     = true
    RequireSpecial    = true
)

// Password hashing
func HashPassword(password string) (string, error) {
    // Use bcrypt with cost 12
    return bcrypt.GenerateFromPassword([]byte(password), 12)
}
```

### 8.2 Input Validation

```go
// Validation middleware
func ValidateRequest(schema interface{}) gin.HandlerFunc {
    return func(c *gin.Context) {
        var input interface{}
        if err := c.ShouldBindJSON(&input); err != nil {
            response.Error(c, 400, "Invalid request format", err)
            c.Abort()
            return
        }

        // Use validator.v10
        validate := validator.New()
        if err := validate.Struct(input); err != nil {
            response.Error(c, 400, "Validation failed", err)
            c.Abort()
            return
        }

        c.Set("validated_input", input)
        c.Next()
    }
}

// Example validation struct
type CreateProposalVersionRequest struct {
    Title            string `json:"title" validate:"required,min=10,max=200"`
    Objectives       string `json:"objectives" validate:"required,min=100"`
    Methodology      string `json:"methodology" validate:"required,min=100"`
    ExpectedOutcomes string `json:"expected_outcomes" validate:"required,min=50"`
}
```

### 8.3 Rate Limiting

```go
// Rate limit configuration
var rateLimits = map[string]RateLimit{
    "/api/v1/auth/login": {
        Requests: 5,
        Window:   15 * time.Minute,
    },
    "/api/v1/proposals/:id/submit": {
        Requests: 3,
        Window:   1 * time.Hour,
    },
    "/api/v1/ai/*": {
        Requests: 10,
        Window:   1 * time.Minute,
    },
}

// Implementation using Redis
func RateLimitMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        key := fmt.Sprintf("ratelimit:%s:%s",
                          c.ClientIP(), c.FullPath())

        // Check Redis counter
        count, _ := redis.Incr(key)
        if count == 1 {
            redis.Expire(key, window)
        }

        if count > limit {
            response.Error(c, 429, "Rate limit exceeded", nil)
            c.Abort()
            return
        }

        c.Next()
    }
}
```

### 8.4 File Upload Security

```go
// File validation
type FileValidator struct {
    AllowedTypes []string  // ["application/pdf"]
    MaxSize      int64     // 10 MB
    ScanVirus    bool      // ClamAV integration
}

func ValidateFile(file *multipart.FileHeader) error {
    // Check file size
    if file.Size > 10*1024*1024 {
        return errors.New("file too large")
    }

    // Check MIME type (not just extension)
    buff := make([]byte, 512)
    f, _ := file.Open()
    f.Read(buff)
    mimeType := http.DetectContentType(buff)

    if mimeType != "application/pdf" {
        return errors.New("invalid file type")
    }

    // Compute hash for integrity
    hash := sha256.New()
    io.Copy(hash, f)

    return nil
}

// Secure file storage
func StoreFile(file *multipart.FileHeader, proposal *Proposal, version int) (string, error) {
    // Generate secure path
    filename := fmt.Sprintf("%d/v%d_%s.pdf",
                           proposal.ID, version, uuid.New().String())

    // Store with restricted permissions
    // Serve through authenticated endpoint, not direct access

    return filename, nil
}
```

### 8.5 SQL Injection Prevention

```go
// ALWAYS use parameterized queries with GORM
// NEVER concatenate user input into queries

// Safe
db.Where("email = ?", userInput).First(&user)

// UNSAFE - Never do this
db.Where(fmt.Sprintf("email = '%s'", userInput)).First(&user)
```

### 8.6 CORS Configuration

```go
func CORSMiddleware() gin.HandlerFunc {
    return cors.New(cors.Config{
        AllowOrigins:     []string{"https://university-hub.edu"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
        AllowHeaders:     []string{"Authorization", "Content-Type"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    })
}
```

---

## 9. Error Handling & Edge Cases

### 9.1 Error Response Format

```go
type ErrorResponse struct {
    Success   bool        `json:"success"`
    Message   string      `json:"message"`
    ErrorCode string      `json:"error_code"`
    Errors    interface{} `json:"errors,omitempty"`
    Timestamp time.Time   `json:"timestamp"`
}

// Error codes
const (
    ErrUnauthorized         = "AUTH_001"
    ErrForbidden            = "AUTH_002"
    ErrInvalidState         = "STATE_001"
    ErrNotTeamLeader        = "TEAM_001"
    ErrProposalLocked       = "PROPOSAL_001"
    ErrDuplicateSubmission  = "PROPOSAL_002"
    ErrVersionNotFound      = "VERSION_001"
)
```

### 9.2 Critical Edge Cases

#### Multiple Submission Attempts

```go
func (s *ProposalService) Submit(proposalID uint, userID uint) error {
    // Check if already submitted (idempotent)
    if proposal.Status != StatusDraft {
        if proposal.Status == StatusSubmitted {
            return nil  // Already submitted, no-op
        }
        return ErrInvalidState
    }

    // Use transaction with row locking
    tx := s.db.Begin()
    defer tx.Rollback()

    var proposal Proposal
    tx.Clauses(clause.Locking{Strength: "UPDATE"}).
       First(&proposal, proposalID)

    // Check again after lock
    if proposal.Status != StatusDraft {
        return ErrInvalidState
    }

    // Proceed with submission
    proposal.Status = StatusSubmitted
    tx.Save(&proposal)
    tx.Commit()

    return nil
}
```

#### Late Revision After Approval

```go
// Prevent version creation if not in valid state
func (s *ProposalService) CreateVersion(...) error {
    if !proposal.IsEditable() && !proposal.CanCreateNewVersion() {
        return errors.New("cannot create version: proposal is locked")
    }

    // Additional check: team still approved
    if proposal.Team.Status != TeamStatusApproved {
        return errors.New("team no longer approved")
    }

    return nil
}
```

#### Concurrent Feedback Submission

```go
// Use database constraint to prevent multiple approvals
ALTER TABLE feedback
ADD CONSTRAINT one_approval_per_proposal
UNIQUE (proposal_id, decision)
WHERE decision = 'approve';

// Service layer check
func (s *ProposalService) SubmitFeedback(...) error {
    // Check if already decided
    var existingFeedback Feedback
    err := s.db.Where("proposal_id = ? AND decision IN ?",
                      proposalID, []string{"approve", "reject"}).
              First(&existingFeedback).Error

    if err == nil {
        return errors.New("proposal already decided")
    }

    // Proceed
}
```

#### AI Advisory Conflicts

```go
// AI results are non-authoritative
// If AI says "reject" but teacher approves, teacher wins
// Solution: AI endpoints are completely separate from decision flow

func AnalyzeProposal(proposalID uint) AIAnalysis {
    // This function NEVER writes to proposal table
    // This function NEVER changes proposal state
    // Results are returned, not stored

    return AIAnalysis{
        Suggestions: [...],
        Timestamp:   time.Now(),
    }
}
```

#### Team Member Leaves After Approval

```go
// Business decision: Once proposal approved, team composition is locked
// Implementation: Soft delete user but keep in team_members for history

func (s *UserService) DeactivateUser(userID uint) error {
    // Don't remove from team_members
    // Just mark user as inactive
    db.Model(&User{}).Where("id = ?", userID).
       Update("is_active", false)

    // Proposals and projects remain intact
    return nil
}
```

---

## 10. Design Justifications

### 10.1 Why Immutable Versions?

**Problem**: Students could claim "we didn't write that" or "you saw old version"

**Solution**: Once created, versions are never modified

- Database-level: No UPDATE permissions on `proposal_versions`
- Application-level: Repository layer has no Update() method
- Audit-level: File hash stored, tampering detectable

**Benefits**:

- Academic integrity
- Dispute resolution
- Clear timeline of changes
- Teacher confidence in what they're reviewing

---

### 10.2 Why Leader-Only Submission?

**Problem**: Multiple team members could submit different versions simultaneously

**Solution**: Only team leader can submit

- Clear ownership
- Single point of responsibility
- Prevents conflicts
- Mirrors real academic accountability

**Implementation**:

```go
func (ac *AccessContext) canSubmitProposal(p *Proposal) bool {
    return p.Team.LeaderID == ac.User.ID
}
```

**Benefits**:

- No confusion about "official" version
- Leader accountable for quality
- Members collaborate but don't conflict

---

### 10.3 Why AI Has No Authority?

**Problem**: Automated decisions lack nuance, create liability, reduce learning

**Solution**: AI only suggests, never decides

- No AI results stored in database
- Teachers never see AI scores
- Students use AI as tool, not oracle

**Ethical Justification**:

- Preserves human judgment
- Prevents algorithmic bias
- Maintains educational value
- Reduces legal liability

**Technical Implementation**:

- AI endpoints separate from workflow
- No foreign keys from proposals to AI results
- AI analysis ephemeral (not persisted)

---

### 10.4 Why Strict State Machine?

**Problem**: Ad-hoc workflows create confusion, missed steps, disputes

**Solution**: Explicit state transitions with enforcement

- Every transition validated
- Role-specific permissions per state
- Impossible to skip steps

**Example**:

```
Can't go from Draft → Approved directly
Must go Draft → Submitted → UnderReview → Approved
```

**Benefits**:

- Process transparency
- Audit compliance
- Fair treatment (everyone follows same path)
- Reduced errors

---

### 10.5 Why Separate Project from Proposal?

**Problem**: Proposals are workflow entities, projects are outputs

**Solution**: Project created only when proposal approved

- Proposal: Draft → Review → Approval
- Project: Documentation → Publication → Archive

**Why This Matters**:

- Proposals are internal (academic governance)
- Projects are external (knowledge base)
- Different access controls
- Different lifecycles

---

### 10.6 Why Comprehensive Audit Logs?

**Problem**: Disputes about "what happened when"

**Solution**: Every action logged with full context

- Actor, timestamp, IP, session
- Before/after state snapshots
- Immutable (write-only)

**Legal/Academic Justification**:

- Accountability
- Dispute resolution
- Compliance (accreditation audits)
- Security investigation

**Performance Consideration**:

- Async logging to avoid blocking
- Separate table with partitioning
- Retention policy (7 years standard for academic records)

---

### 10.7 Why Department-Scoped Access?

**Problem**: Teachers shouldn't see all proposals across university

**Solution**: Teachers access only their department's proposals

- Enforced at query level
- Indexed for performance
- Prevents information leakage

**Implementation**:

```go
// Teacher query
db.Where("department_id = ?", teacher.DepartmentID).Find(&proposals)

// Admin query
db.Find(&proposals)  // No restriction
```

---

## 11. Scalability Considerations

### 11.1 Multi-University Support

**Current Design**: Single instance per university

**Future**: Multi-tenant architecture

```go
type Config struct {
    Mode string  // "single" or "multi-tenant"
}

// Add to all queries
db.Where("university_id = ?", currentUniversityID)

// Or use separate databases per university
```

### 11.2 File Storage

**Current**: Local filesystem

**Production**: S3-compatible storage

```go
type FileStorage interface {
    Upload(file io.Reader, path string) (string, error)
    Download(path string) (io.Reader, error)
    Delete(path string) error
}

// Implementations: LocalStorage, S3Storage, MinIOStorage
```

### 11.3 Caching Strategy

**Redis for**:

- User sessions
- Rate limiting
- Notification counts
- Public project lists

**Don't cache**:

- Proposal states (critical accuracy)
- Audit logs
- Versions

### 11.4 Database Optimization

**Indexes Created**:

- Foreign keys (automatic joins)
- Status fields (frequent filtering)
- Composite: (user_id, is_read) for notifications

**Query Optimization**:

- Eager loading for relations: `db.Preload("Team.Members")`
- Pagination for large lists
- Archived proposals moved to separate table after 2 years

---

## 12. Future Extensibility

### Modular Design Enables:

1. **Presentation Module**: Upload and evaluate final presentations
2. **Plagiarism Detection**: Integration with Turnitin-like services
3. **Timeline Tracking**: Gantt charts for project progress
4. **Collaborative Editing**: Real-time document co-authoring
5. **External Review**: Industry experts as guest reviewers
6. **Analytics Dashboard**: Proposal trends, approval rates, topic clustering

### Extension Points:

```go
// Plugin architecture for custom validators
type ProposalValidator interface {
    Validate(proposal *Proposal, version *ProposalVersion) error
}

// Register custom validators
RegisterValidator(&PlagiarismValidator{})
RegisterValidator(&MinimumWordCountValidator{})

// Event hooks for integrations
OnProposalSubmitted(func(p *Proposal) {
    // Send to external system
    analytics.Track("proposal_submitted", p.ID)
})
```

---

## 13. Implementation Roadmap

### Phase 1: Core Foundation (Weeks 1-2)

- [ ] Database schema implementation
- [ ] User authentication (JWT)
- [ ] Basic RBAC middleware
- [ ] Audit logging infrastructure

### Phase 2: Team Management (Week 3)

- [ ] Team creation
- [ ] Member invitations
- [ ] Advisor approval workflow

### Phase 3: Proposal Workflow (Weeks 4-6)

- [ ] Proposal creation
- [ ] Version management (immutable)
- [ ] State machine implementation
- [ ] Submission logic
- [ ] Teacher feedback

### Phase 4: Project Management (Week 7)

- [ ] Project creation on approval
- [ ] Documentation upload
- [ ] Publication workflow

### Phase 5: AI Integration (Week 8)

- [ ] Proposal summarization
- [ ] Completeness checking
- [ ] Similarity detection

### Phase 6: Public Archive (Week 9)

- [ ] Public project listing
- [ ] Reviews and ratings
- [ ] Search and filtering

### Phase 7: Polish & Security (Week 10)

- [ ] Comprehensive testing
- [ ] Security audit
- [ ] Performance optimization
- [ ] Documentation

---

## Summary

This architecture delivers:

✅ **Human-First Governance**: AI advises, humans decide  
✅ **Immutable History**: Versions and feedback never modified  
✅ **Strict Process**: State machine prevents shortcuts  
✅ **Full Auditability**: Every action logged  
✅ **Role Separation**: Clear boundaries between actors  
✅ **Academic Integrity**: Leader-only submission, locked versions  
✅ **Scalability**: Multi-university ready  
✅ **Security**: RBAC, input validation, audit logs  
✅ **Extensibility**: Modular design for future features

The system transforms informal, memory-based evaluation into transparent, versioned, auditable workflow where humans maintain authority and the platform becomes institutional memory.
