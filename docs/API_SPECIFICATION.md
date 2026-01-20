# API Specification - University Project Hub

## Base URL

```
Development: http://localhost:8080/api/v1
Production: https://api.university-hub.edu/api/v1
```

## Authentication

All authenticated endpoints require JWT token in header:

```
Authorization: Bearer <token>
```

## Standard Response Format

### Success Response

```json
{
  "success": true,
  "message": "Operation successful",
  "data": { ... }
}
```

### Error Response

```json
{
  "success": false,
  "message": "Error description",
  "error_code": "ERR_CODE",
  "errors": { ... },
  "timestamp": "2026-01-17T10:00:00Z"
}
```

## HTTP Status Codes

- `200 OK` - Success
- `201 Created` - Resource created
- `400 Bad Request` - Validation error
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - State conflict (e.g., already submitted)
- `422 Unprocessable Entity` - Business logic error
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

---

## Authentication Endpoints

### Register User

```http
POST /auth/register
```

**Request Body:**

```json
{
  "name": "John Doe",
  "email": "john.doe@university.edu",
  "password": "SecurePass123!",
  "role": "student",
  "university_id": 1,
  "department_id": 5,
  "student_id": "CS/2024/001"
}
```

**Validation Rules:**

- `name`: required, 2-100 characters
- `email`: required, valid email, unique
- `password`: required, min 8 chars, 1 uppercase, 1 number, 1 special
- `role`: required, enum [student, teacher, admin]
- `university_id`: required, must exist
- `department_id`: required, must exist
- `student_id`: required for students

**Response:** `201 Created`

```json
{
  "success": true,
  "message": "Registration successful",
  "data": {
    "user": {
      "id": 42,
      "name": "John Doe",
      "email": "john.doe@university.edu",
      "role": "student",
      "university": { "id": 1, "name": "ASTU" },
      "department": { "id": 5, "name": "Computer Science" }
    },
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_at": "2026-01-18T10:00:00Z"
  }
}
```

---

### Login

```http
POST /auth/login
```

**Request Body:**

```json
{
  "email": "john.doe@university.edu",
  "password": "SecurePass123!"
}
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "user": { ... },
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_at": "2026-01-18T10:00:00Z"
  }
}
```

**Rate Limit:** 5 requests per 15 minutes per IP

**Error Responses:**

- `401` - Invalid credentials
- `429` - Too many login attempts

---

### Get Current User

```http
GET /auth/me
Authorization: Bearer <token>
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "user": {
      "id": 42,
      "name": "John Doe",
      "email": "john.doe@university.edu",
      "role": "student",
      "university": { ... },
      "department": { ... },
      "profile_photo": "/storage/profiles/42.jpg"
    }
  }
}
```

---

## Team Endpoints

### Create Team

```http
POST /teams
Authorization: Bearer <token>
Role: student
```

**Request Body:**

```json
{
  "name": "AI Research Team",
  "department_id": 5,
  "advisor_id": 12,
  "member_ids": [43, 44, 45]
}
```

**Validation Rules:**

- `name`: required, 3-100 characters, unique per department
- `department_id`: required, must be creator's department
- `advisor_id`: required, must be teacher in same department
- `member_ids`: array, max 4 additional members (5 total including leader)

**Business Rules:**

- Creator automatically becomes leader
- All members must be students in same department
- All members receive invitation notifications
- Advisor receives approval request
- Team starts in `pending_advisor_approval` status

**Response:** `201 Created`

```json
{
  "success": true,
  "message": "Team created successfully",
  "data": {
    "team": {
      "id": 7,
      "name": "AI Research Team",
      "status": "pending_advisor_approval",
      "leader": {
        "id": 42,
        "name": "John Doe"
      },
      "members": [
        {
          "id": 43,
          "name": "Jane Smith",
          "invitation_status": "pending"
        }
      ],
      "advisor": {
        "id": 12,
        "name": "Dr. Ahmed"
      },
      "created_at": "2026-01-17T10:00:00Z"
    }
  }
}
```

---

### Get Team

```http
GET /teams/:id
Authorization: Bearer <token>
```

**Access:** Team members, advisor, admin

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "team": {
      "id": 7,
      "name": "AI Research Team",
      "status": "approved",
      "department": { ... },
      "leader": { ... },
      "members": [ ... ],
      "advisor": { ... },
      "proposal": {
        "id": 15,
        "status": "under_review"
      },
      "created_at": "2026-01-17T10:00:00Z"
    }
  }
}
```

---

### List My Teams

```http
GET /teams
Authorization: Bearer <token>
```

**Query Parameters:**

- `status` - Filter by status: pending_advisor_approval, approved, rejected
- `page` - Page number (default: 1)
- `limit` - Items per page (default: 20, max: 100)

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "teams": [ ... ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 45,
      "pages": 3
    }
  }
}
```

**Access Rules:**

- Students: Teams they're member of or created
- Teachers: Teams in their department or they advise
- Admins: All teams

---

### Respond to Team Invitation

```http
POST /teams/:team_id/invitations/respond
Authorization: Bearer <token>
Role: student
```

**Request Body:**

```json
{
  "response": "accept"
}
```

**Values:** `accept` or `reject`

**Business Rules:**

- Only invited member can respond
- Cannot respond twice
- Leader notified of response
- If all members accept AND advisor approves → team becomes "approved"

**Response:** `200 OK`

```json
{
  "success": true,
  "message": "Invitation accepted",
  "data": {
    "team_member": {
      "team_id": 7,
      "user_id": 43,
      "invitation_status": "accepted",
      "responded_at": "2026-01-17T10:30:00Z"
    }
  }
}
```

---

### Advisor Team Approval

```http
POST /teams/:id/advisor-response
Authorization: Bearer <token>
Role: teacher
```

**Request Body:**

```json
{
  "decision": "approve",
  "comment": "Strong team with complementary skills"
}
```

**Values:** `approve` or `reject`

**Validation:**

- `decision`: required, enum
- `comment`: required, min 10 characters

**Business Rules:**

- Only designated advisor can respond
- Cannot respond twice
- If approved: team can create proposals
- If rejected: team cannot proceed
- All members notified

**Response:** `200 OK`

```json
{
  "success": true,
  "message": "Team approved",
  "data": {
    "team": {
      "id": 7,
      "status": "approved",
      "advisor_response_comment": "Strong team...",
      "advisor_responded_at": "2026-01-17T11:00:00Z"
    }
  }
}
```

---

## Proposal Endpoints

### Create Proposal

```http
POST /proposals
Authorization: Bearer <token>
Role: student (team leader only)
```

**Request Body:**

```json
{
  "team_id": 7
}
```

**Business Rules:**

- Only team leader can create
- Team must be approved
- One proposal per team
- Starts in `draft` status

**Response:** `201 Created`

```json
{
  "success": true,
  "message": "Proposal created",
  "data": {
    "proposal": {
      "id": 15,
      "team_id": 7,
      "status": "draft",
      "current_version_id": null,
      "created_at": "2026-01-17T12:00:00Z"
    }
  }
}
```

**Error:** `409 Conflict` if team already has proposal

---

### Create Proposal Version

```http
POST /proposals/:id/versions
Authorization: Bearer <token>
Role: student (team leader only)
Content-Type: multipart/form-data
```

**Form Data:**

- `title` - string, required, 10-200 characters
- `objectives` - text, required, min 100 characters
- `methodology` - text, required, min 100 characters
- `expected_outcomes` - text, required, min 50 characters
- `file` - PDF file, required, max 10MB

**Business Rules:**

- Only when proposal status is `draft` or `revision_required`
- Auto-increment version number
- Store file with integrity hash
- Update proposal.current_version_id
- Capture IP address for audit

**Response:** `201 Created`

```json
{
  "success": true,
  "message": "Version created",
  "data": {
    "version": {
      "id": 42,
      "proposal_id": 15,
      "version_number": 1,
      "title": "AI-Powered Learning Platform",
      "objectives": "...",
      "methodology": "...",
      "expected_outcomes": "...",
      "file_url": "/api/v1/files/proposals/15/v1.pdf",
      "file_hash": "sha256:abc123...",
      "created_by": 42,
      "created_at": "2026-01-17T12:30:00Z"
    }
  }
}
```

**Errors:**

- `403` - Not team leader or proposal locked
- `422` - Invalid file type or size

---

### Submit Proposal

```http
POST /proposals/:id/submit
Authorization: Bearer <token>
Role: student (team leader only)
```

**Request Body:**

```json
{
  "acknowledgement": true
}
```

**Business Rules:**

- Only from `draft` status
- Must have at least one complete version
- Transitions to `submitted` status
- Version becomes immutable
- Advisor receives notification

**Response:** `200 OK`

```json
{
  "success": true,
  "message": "Proposal submitted successfully",
  "data": {
    "proposal": {
      "id": 15,
      "status": "submitted",
      "submitted_at": "2026-01-17T13:00:00Z",
      "current_version": { ... }
    }
  }
}
```

**Error Handling:**

- Idempotent: If already submitted, returns success
- `422` if missing required version data

---

### Get Proposal

```http
GET /proposals/:id
Authorization: Bearer <token>
```

**Access:** Team members, advisor, admin

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "proposal": {
      "id": 15,
      "team": {
        "id": 7,
        "name": "AI Research Team",
        "leader": { ... },
        "members": [ ... ]
      },
      "status": "under_review",
      "current_version": {
        "id": 42,
        "version_number": 1,
        "title": "AI-Powered Learning Platform",
        "objectives": "...",
        "methodology": "...",
        "expected_outcomes": "...",
        "file_url": "/api/v1/files/proposals/15/v1.pdf",
        "created_at": "2026-01-17T12:30:00Z"
      },
      "versions": [
        { "id": 42, "version_number": 1, ... }
      ],
      "feedback": [
        {
          "id": 8,
          "version_id": 42,
          "reviewer": { ... },
          "decision": "revise",
          "comment": "...",
          "created_at": "2026-01-17T14:00:00Z"
        }
      ],
      "submitted_at": "2026-01-17T13:00:00Z",
      "can_edit": false,
      "can_submit": false,
      "can_create_version": false
    }
  }
}
```

---

### List Proposals

```http
GET /proposals
Authorization: Bearer <token>
```

**Query Parameters:**

- `status` - Filter: draft, submitted, under_review, revision_required, approved, rejected
- `department_id` - Filter by department (teachers only)
- `page` - Page number
- `limit` - Items per page

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "proposals": [ ... ],
    "pagination": { ... }
  }
}
```

**Access Rules:**

- Students: Only their team's proposal
- Teachers: Proposals from their department or teams they advise
- Admins: All proposals

---

### Start Review

```http
POST /proposals/:id/start-review
Authorization: Bearer <token>
Role: teacher (advisor only)
```

**Business Rules:**

- Only from `submitted` status
- Only team advisor
- Transitions to `under_review`

**Response:** `200 OK`

```json
{
  "success": true,
  "message": "Review started",
  "data": {
    "proposal": {
      "id": 15,
      "status": "under_review"
    }
  }
}
```

---

### Submit Feedback

```http
POST /proposals/:id/feedback
Authorization: Bearer <token>
Role: teacher (advisor only)
```

**Request Body:**

```json
{
  "version_id": 42,
  "decision": "revise",
  "comment": "Methodology needs more detail on data collection. Include timeline."
}
```

**Validation:**

- `version_id`: required, must be current version
- `decision`: required, enum [approve, revise, reject]
- `comment`: required, min 20 characters

**Business Rules:**

- Only from `under_review` status
- Only team advisor
- Feedback is immutable once submitted
- Decision determines state transition:
  - `approve` → `approved` (creates project)
  - `revise` → `revision_required` (allows new version)
  - `reject` → `rejected` (terminal)

**Response:** `201 Created`

```json
{
  "success": true,
  "message": "Feedback submitted",
  "data": {
    "feedback": {
      "id": 8,
      "proposal_id": 15,
      "version_id": 42,
      "reviewer": { ... },
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

**Side Effects:**

- Notifications sent to team
- If approved: Project automatically created
- Audit log entry created

---

## AI Advisory Endpoints

### Analyze Proposal

```http
POST /ai/analyze-proposal
Authorization: Bearer <token>
Role: student (team member) or teacher
```

**Request Body:**

```json
{
  "proposal_id": 15,
  "version_id": 42
}
```

**Business Rules:**

- Can only analyze own team's proposals (students) or advised proposals (teachers)
- Results NOT stored in database
- NO impact on approval workflow
- Can be called multiple times

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "analysis": {
      "summary": "The proposal aims to develop an AI-powered adaptive learning platform using reinforcement learning...",
      "completeness_score": 85,
      "missing_sections": [
        "Risk analysis and mitigation strategies",
        "Detailed project timeline with milestones"
      ],
      "structure_issues": [
        "Objectives could be more specific with measurable outcomes",
        "Methodology lacks details on data collection process"
      ],
      "suggestions": [
        "Include specific success metrics (e.g., accuracy, user engagement)",
        "Add Gantt chart or timeline",
        "Define clear evaluation criteria"
      ],
      "analyzed_at": "2026-01-17T15:00:00Z"
    }
  }
}
```

**Critical:** This endpoint does NOT create any database records

---

### Check Similarity

```http
GET /ai/check-similarity
Authorization: Bearer <token>
Role: teacher
```

**Query Parameters:**

- `proposal_id` - required
- `threshold` - optional, float 0.0-1.0 (default: 0.3)

**Business Rules:**

- Only teachers can check
- Only checks against approved, public projects
- Provides context, NOT judgment
- Teacher decides if similarity is problematic

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "similar_projects": [
      {
        "project_id": 23,
        "title": "Adaptive Learning System Using ML",
        "department": "Computer Science",
        "year": 2024,
        "similarity_score": 0.52,
        "overlap_areas": [
          "Reinforcement learning for personalization",
          "User behavior tracking methodology"
        ]
      },
      {
        "project_id": 31,
        "title": "Personalized Education Platform",
        "department": "Software Engineering",
        "year": 2023,
        "similarity_score": 0.38,
        "overlap_areas": ["Adaptive content delivery"]
      }
    ],
    "checked_at": "2026-01-17T15:05:00Z"
  }
}
```

---

## Project Endpoints

### Get Project

```http
GET /projects/:id
Authorization: Bearer <token> (optional for public projects)
```

**Access Rules:**

- Private projects: Team members, advisor, admin only
- Public projects: Everyone (including unauthenticated)

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "project": {
      "id": 8,
      "proposal_id": 15,
      "team": {
        "name": "AI Research Team",
        "members": [ ... ]
      },
      "summary": "AI-Powered Learning Platform with adaptive algorithms",
      "department": { ... },
      "visibility": "public",
      "documentation": [
        {
          "id": 12,
          "document_type": "final_report",
          "file_url": "/api/v1/files/projects/8/final_report.pdf",
          "status": "approved",
          "submitted_at": "2026-01-17T16:00:00Z"
        }
      ],
      "reviews": [
        {
          "id": 45,
          "user": { "name": "Alice Johnson" },
          "rating": 5,
          "comment": "Excellent implementation",
          "created_at": "2026-01-18T10:00:00Z"
        }
      ],
      "average_rating": 4.5,
      "view_count": 234,
      "share_count": 12,
      "created_at": "2026-01-17T14:00:00Z",
      "published_at": "2026-01-17T18:00:00Z"
    }
  }
}
```

---

### List Public Projects

```http
GET /projects/public
```

**Query Parameters:**

- `department_id` - Filter by department
- `year` - Filter by year
- `search` - Full-text search in title and summary
- `sort` - Sort by: rating, date, views (default: rating)
- `page` - Page number
- `limit` - Items per page

**Response:** `200 OK`

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
        "year": 2025,
        "team_size": 4,
        "average_rating": 4.5,
        "view_count": 234,
        "thumbnail_url": "/api/v1/files/projects/8/thumbnail.jpg"
      }
    ],
    "pagination": { ... }
  }
}
```

---

### Upload Project Documentation

```http
POST /projects/:id/documentation
Authorization: Bearer <token>
Role: student (team member)
Content-Type: multipart/form-data
```

**Form Data:**

- `document_type` - enum: final_report, presentation, proposal_document, code_repository
- `file` - File (PDF for reports, PPTX for presentations)

**Validation:**

- final_report: PDF, max 20MB
- presentation: PPTX or PDF, max 50MB
- code_repository: ZIP, max 100MB

**Business Rules:**

- Only team members can upload
- Each document type can only be uploaded once
- Files are immutable once uploaded
- Advisor receives notification for review

**Response:** `201 Created`

```json
{
  "success": true,
  "message": "Documentation uploaded",
  "data": {
    "documentation": {
      "id": 12,
      "project_id": 8,
      "document_type": "final_report",
      "file_url": "/api/v1/files/projects/8/final_report.pdf",
      "file_hash": "sha256:def456...",
      "status": "pending",
      "submitted_by": 42,
      "submitted_at": "2026-01-17T16:00:00Z"
    }
  }
}
```

---

### Review Project Documentation

```http
POST /projects/:id/documentation/:doc_id/review
Authorization: Bearer <token>
Role: teacher (advisor)
```

**Request Body:**

```json
{
  "decision": "approve",
  "comment": "Well-documented and comprehensive"
}
```

**Values:** `approve` or `reject`

**Response:** `200 OK`

```json
{
  "success": true,
  "message": "Documentation approved",
  "data": {
    "documentation": {
      "id": 12,
      "status": "approved",
      "review_comment": "...",
      "reviewed_by": 12,
      "reviewed_at": "2026-01-17T17:00:00Z"
    }
  }
}
```

---

### Publish Project

```http
POST /projects/:id/publish
Authorization: Bearer <token>
Role: teacher (advisor)
```

**Request Body:**

```json
{
  "acknowledgement": true
}
```

**Business Rules:**

- Only advisor can publish
- All required documentation must be uploaded and approved
- Once published, project appears in public archive
- Cannot be unpublished (permanent)

**Response:** `200 OK`

```json
{
  "success": true,
  "message": "Project published",
  "data": {
    "project": {
      "id": 8,
      "visibility": "public",
      "published_at": "2026-01-17T18:00:00Z"
    }
  }
}
```

---

### Add Project Review

```http
POST /projects/:id/reviews
Authorization: Bearer <token>
```

**Request Body:**

```json
{
  "rating": 5,
  "comment": "Excellent implementation with innovative approach"
}
```

**Validation:**

- `rating`: required, integer 1-5
- `comment`: optional, max 500 characters

**Business Rules:**

- Only authenticated users
- Only public projects
- One review per user per project
- Does NOT affect academic evaluation

**Response:** `201 Created`

```json
{
  "success": true,
  "message": "Review submitted",
  "data": {
    "review": {
      "id": 45,
      "project_id": 8,
      "user": {
        "name": "Alice Johnson"
      },
      "rating": 5,
      "comment": "...",
      "created_at": "2026-01-18T10:00:00Z"
    },
    "updated_average": 4.6
  }
}
```

**Error:** `409 Conflict` if user already reviewed

---

## Notification Endpoints

### Get Notifications

```http
GET /notifications
Authorization: Bearer <token>
```

**Query Parameters:**

- `is_read` - Filter: true, false, or omit for all
- `page` - Page number
- `limit` - Items per page (max 50)

**Response:** `200 OK`

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
      },
      {
        "id": 122,
        "message": "New feedback on your proposal",
        "reference_type": "feedback",
        "reference_id": 8,
        "is_read": true,
        "created_at": "2026-01-17T13:30:00Z"
      }
    ],
    "unread_count": 3,
    "pagination": { ... }
  }
}
```

---

### Mark Notification as Read

```http
POST /notifications/:id/mark-read
Authorization: Bearer <token>
```

**Response:** `200 OK`

```json
{
  "success": true,
  "message": "Notification marked as read"
}
```

---

### Mark All as Read

```http
POST /notifications/mark-all-read
Authorization: Bearer <token>
```

**Response:** `200 OK`

```json
{
  "success": true,
  "message": "All notifications marked as read"
}
```

---

## File Access Endpoints

### Download Proposal Document

```http
GET /files/proposals/:proposal_id/:filename
Authorization: Bearer <token>
```

**Access:** Team members, advisor, admin

**Response:** File download with appropriate headers

```
Content-Type: application/pdf
Content-Disposition: inline; filename="proposal_v1.pdf"
```

---

### Download Project Document

```http
GET /files/projects/:project_id/:filename
Authorization: Bearer <token> (optional for public projects)
```

**Access:**

- Private: Team members, advisor, admin
- Public: Everyone

---

## Admin Endpoints

### List All Users

```http
GET /admin/users
Authorization: Bearer <token>
Role: admin
```

**Query Parameters:**

- `role` - Filter by role
- `department_id` - Filter by department
- `is_active` - Filter by active status
- `search` - Search by name or email
- `page` - Page number
- `limit` - Items per page

---

### Create User

```http
POST /admin/users
Authorization: Bearer <token>
Role: admin
```

**Use Case:** Bulk import or manual user creation

---

### Get Audit Logs

```http
GET /admin/audit-logs
Authorization: Bearer <token>
Role: admin
```

**Query Parameters:**

- `entity_type` - Filter: proposal, team, user, etc.
- `entity_id` - Specific entity
- `actor_id` - User who performed action
- `action` - Filter: create, submit, approve, etc.
- `from_date` - Start date (ISO 8601)
- `to_date` - End date (ISO 8601)
- `page` - Page number
- `limit` - Items per page

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "audit_logs": [
      {
        "id": 5821,
        "entity_type": "proposal",
        "entity_id": 15,
        "action": "submit",
        "actor": {
          "id": 42,
          "name": "John Doe",
          "role": "student"
        },
        "old_state": {
          "status": "draft"
        },
        "new_state": {
          "status": "submitted",
          "submitted_at": "2026-01-17T13:00:00Z"
        },
        "ip_address": "192.168.1.100",
        "timestamp": "2026-01-17T13:00:00Z"
      }
    ],
    "pagination": { ... }
  }
}
```

---

## Error Codes Reference

| Code           | HTTP Status | Description                             |
| -------------- | ----------- | --------------------------------------- |
| AUTH_001       | 401         | Missing or invalid authentication token |
| AUTH_002       | 403         | Insufficient permissions for action     |
| AUTH_003       | 401         | Invalid credentials                     |
| AUTH_004       | 429         | Too many login attempts                 |
| STATE_001      | 422         | Invalid proposal state transition       |
| STATE_002      | 409         | Resource in conflicting state           |
| TEAM_001       | 403         | User is not team leader                 |
| TEAM_002       | 422         | Team not approved                       |
| TEAM_003       | 409         | Invitation already responded            |
| PROPOSAL_001   | 422         | Proposal is locked (not editable)       |
| PROPOSAL_002   | 409         | Proposal already submitted              |
| PROPOSAL_003   | 404         | Proposal not found                      |
| VERSION_001    | 404         | Version not found                       |
| VERSION_002    | 422         | Cannot create version in current state  |
| FILE_001       | 422         | Invalid file type                       |
| FILE_002       | 422         | File too large                          |
| FILE_003       | 404         | File not found                          |
| VALIDATION_001 | 400         | Request validation failed               |

---

## Rate Limits

| Endpoint Pattern     | Limit        | Window     |
| -------------------- | ------------ | ---------- |
| /auth/login          | 5 requests   | 15 minutes |
| /auth/register       | 3 requests   | 1 hour     |
| /proposals/\*/submit | 3 requests   | 1 hour     |
| /ai/\*               | 10 requests  | 1 minute   |
| Default              | 100 requests | 15 minutes |

Rate limit headers included in all responses:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
X-RateLimit-Reset: 1705496400
```

---

## Webhooks (Future Enhancement)

System will support webhooks for external integrations:

- `proposal.submitted`
- `proposal.approved`
- `project.published`
- `team.created`

Configuration endpoint: `POST /admin/webhooks`
