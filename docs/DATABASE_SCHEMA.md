# Database Schema - University Project Hub

## Schema Overview

This database schema implements an immutable-first design where critical academic records (versions, feedback, audit logs) are write-only and never modified or deleted.

## Design Principles

1. **Immutability**: Core academic records never updated after creation
2. **Auditability**: Every change logged with full context
3. **Referential Integrity**: Foreign key constraints enforced
4. **Performance**: Strategic indexing on query patterns
5. **Soft Deletes**: User-facing entities use `deleted_at` instead of hard deletes

---

## Complete SQL Schema

```sql
-- ============================================================================
-- UNIVERSITY PROJECT HUB DATABASE SCHEMA
-- Version: 1.0
-- Last Updated: 2026-01-17
-- ============================================================================

-- Enable UUID extension for generating unique identifiers
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable pgcrypto for password hashing
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ============================================================================
-- INSTITUTIONAL HIERARCHY
-- ============================================================================

-- Universities
CREATE TABLE universities (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    website VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_universities_code ON universities(code);

-- Departments
CREATE TABLE departments (
    id SERIAL PRIMARY KEY,
    university_id INTEGER NOT NULL REFERENCES universities(id) ON DELETE RESTRICT,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(university_id, code),
    UNIQUE(university_id, name)
);

CREATE INDEX idx_departments_university ON departments(university_id);
CREATE INDEX idx_departments_code ON departments(code);

COMMENT ON TABLE departments IS 'Academic departments within universities';

-- ============================================================================
-- USER MANAGEMENT
-- ============================================================================

-- Users (Students, Teachers, Admins, Public)
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('student', 'teacher', 'admin', 'public')),

    -- Institutional associations
    university_id INTEGER REFERENCES universities(id) ON DELETE RESTRICT,
    department_id INTEGER REFERENCES departments(id) ON DELETE RESTRICT,

    -- Student-specific
    student_id VARCHAR(50),

    -- Profile
    profile_photo VARCHAR(500),
    bio TEXT,

    -- Account status
    is_active BOOLEAN DEFAULT TRUE,
    email_verified BOOLEAN DEFAULT FALSE,
    email_verification_token VARCHAR(255),
    email_verified_at TIMESTAMP,

    -- Security
    failed_login_attempts INTEGER DEFAULT 0,
    account_locked_until TIMESTAMP,
    password_reset_token VARCHAR(255),
    password_reset_expires TIMESTAMP,

    -- Activity tracking
    last_login_at TIMESTAMP,
    last_login_ip INET,

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,

    -- Constraints
    CHECK (role != 'student' OR student_id IS NOT NULL),
    CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_department ON users(department_id);
CREATE INDEX idx_users_university ON users(university_id);
CREATE INDEX idx_users_student_id ON users(student_id) WHERE student_id IS NOT NULL;
CREATE INDEX idx_users_active ON users(is_active) WHERE is_active = TRUE;

COMMENT ON TABLE users IS 'All system users: students, teachers, admins, and public users';
COMMENT ON COLUMN users.role IS 'User role: student, teacher, admin, or public';
COMMENT ON COLUMN users.failed_login_attempts IS 'Counter for failed login attempts (reset on successful login)';

-- ============================================================================
-- TEAM MANAGEMENT
-- ============================================================================

-- Teams
CREATE TABLE teams (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    department_id INTEGER NOT NULL REFERENCES departments(id) ON DELETE RESTRICT,
    leader_id INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    advisor_id INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

    -- Status workflow
    status VARCHAR(30) DEFAULT 'pending_advisor_approval' NOT NULL
           CHECK (status IN ('pending_advisor_approval', 'approved', 'rejected')),

    -- Advisor response
    advisor_response_comment TEXT,
    advisor_responded_at TIMESTAMP,

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,

    -- Constraints
    UNIQUE(department_id, name),
    CHECK (leader_id != advisor_id)
);

CREATE INDEX idx_teams_status ON teams(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_teams_leader ON teams(leader_id);
CREATE INDEX idx_teams_advisor ON teams(advisor_id);
CREATE INDEX idx_teams_department ON teams(department_id);

COMMENT ON TABLE teams IS 'Student project teams with leader, members, and assigned advisor';
COMMENT ON COLUMN teams.status IS 'Team approval status: pending_advisor_approval, approved, or rejected';

-- Team Members (Many-to-Many with Invitation Tracking)
CREATE TABLE team_members (
    team_id INTEGER NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

    -- Member role
    role VARCHAR(20) DEFAULT 'member' NOT NULL
         CHECK (role IN ('leader', 'member')),

    -- Invitation workflow
    invitation_status VARCHAR(20) DEFAULT 'pending' NOT NULL
                      CHECK (invitation_status IN ('pending', 'accepted', 'rejected')),
    invited_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    responded_at TIMESTAMP,

    PRIMARY KEY (team_id, user_id)
);

CREATE INDEX idx_team_members_user ON team_members(user_id);
CREATE INDEX idx_team_members_status ON team_members(invitation_status);

COMMENT ON TABLE team_members IS 'Team membership with invitation tracking';
COMMENT ON COLUMN team_members.invitation_status IS 'Invitation status: pending, accepted, or rejected';

-- ============================================================================
-- PROPOSAL WORKFLOW (Core Academic Process)
-- ============================================================================

-- Proposals (Stateful Entity)
CREATE TABLE proposals (
    id SERIAL PRIMARY KEY,
    team_id INTEGER UNIQUE NOT NULL REFERENCES teams(id) ON DELETE RESTRICT,

    -- State machine
    status VARCHAR(30) DEFAULT 'draft' NOT NULL
           CHECK (status IN ('draft', 'submitted', 'under_review',
                            'revision_required', 'approved', 'rejected')),

    -- Version tracking
    current_version_id INTEGER,  -- FK constraint added after proposal_versions table

    -- Submission tracking
    submitted_at TIMESTAMP,
    submission_count INTEGER DEFAULT 0,  -- Track number of submissions (draft -> submitted cycles)

    -- Approval tracking
    approved_at TIMESTAMP,
    approved_by INTEGER REFERENCES users(id) ON DELETE RESTRICT,

    -- Rejection tracking
    rejected_at TIMESTAMP,
    rejected_by INTEGER REFERENCES users(id) ON DELETE RESTRICT,
    rejection_reason TEXT,

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,

    -- Constraints
    CHECK (status != 'submitted' OR submitted_at IS NOT NULL),
    CHECK (status != 'approved' OR (approved_at IS NOT NULL AND approved_by IS NOT NULL)),
    CHECK (status != 'rejected' OR (rejected_at IS NOT NULL AND rejected_by IS NOT NULL))
);

CREATE INDEX idx_proposals_status ON proposals(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_proposals_team ON proposals(team_id);
CREATE INDEX idx_proposals_approved ON proposals(status) WHERE status = 'approved';

COMMENT ON TABLE proposals IS 'Project proposals - stateful entities that follow strict workflow';
COMMENT ON COLUMN proposals.status IS 'Proposal state: draft, submitted, under_review, revision_required, approved, or rejected';
COMMENT ON COLUMN proposals.current_version_id IS 'Points to the active/latest version';

-- Proposal Versions (IMMUTABLE - Write-Only)
CREATE TABLE proposal_versions (
    id SERIAL PRIMARY KEY,
    proposal_id INTEGER NOT NULL REFERENCES proposals(id) ON DELETE CASCADE,

    -- Content (immutable)
    title VARCHAR(500) NOT NULL,
    objectives TEXT NOT NULL,
    methodology TEXT NOT NULL,
    expected_outcomes TEXT NOT NULL,

    -- File storage
    file_url VARCHAR(500) NOT NULL,
    file_hash VARCHAR(64) NOT NULL,  -- SHA-256 for integrity verification
    file_size_bytes BIGINT NOT NULL,

    -- Version metadata
    version_number INTEGER NOT NULL,
    is_approved_version BOOLEAN DEFAULT FALSE,

    -- Audit trail (who, when, where)
    created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    ip_address INET NOT NULL,
    user_agent TEXT,
    session_id VARCHAR(255),

    -- Timestamp (immutable)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,

    -- Constraints
    UNIQUE(proposal_id, version_number),
    CHECK (version_number > 0),
    CHECK (file_size_bytes > 0 AND file_size_bytes <= 10485760),  -- Max 10MB
    CHECK (char_length(title) >= 10 AND char_length(title) <= 500),
    CHECK (char_length(objectives) >= 100),
    CHECK (char_length(methodology) >= 100),
    CHECK (char_length(expected_outcomes) >= 50)
);

CREATE INDEX idx_versions_proposal ON proposal_versions(proposal_id);
CREATE INDEX idx_versions_approved ON proposal_versions(is_approved_version) WHERE is_approved_version = TRUE;
CREATE INDEX idx_versions_created_by ON proposal_versions(created_by);
CREATE INDEX idx_versions_created_at ON proposal_versions(created_at);

COMMENT ON TABLE proposal_versions IS 'IMMUTABLE proposal versions - never updated or deleted';
COMMENT ON COLUMN proposal_versions.file_hash IS 'SHA-256 hash for file integrity verification';
COMMENT ON COLUMN proposal_versions.is_approved_version IS 'True for the version that was approved';

-- Add foreign key constraint from proposals to proposal_versions
ALTER TABLE proposals
ADD CONSTRAINT fk_current_version
FOREIGN KEY (current_version_id) REFERENCES proposal_versions(id) ON DELETE SET NULL;

-- Feedback (IMMUTABLE - Write-Only)
CREATE TABLE feedback (
    id SERIAL PRIMARY KEY,
    proposal_id INTEGER NOT NULL REFERENCES proposals(id) ON DELETE CASCADE,
    proposal_version_id INTEGER NOT NULL REFERENCES proposal_versions(id) ON DELETE CASCADE,
    reviewer_id INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

    -- Decision (immutable)
    decision VARCHAR(20) NOT NULL CHECK (decision IN ('approve', 'revise', 'reject')),
    comment TEXT NOT NULL CHECK (char_length(comment) >= 20),

    -- Structured feedback flags
    is_structured BOOLEAN DEFAULT FALSE,  -- AI-suggested vs manual feedback

    -- Audit trail
    ip_address INET NOT NULL,
    user_agent TEXT,
    session_id VARCHAR(255),

    -- Timestamp (immutable)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,

    -- Constraints
    CHECK (decision != 'approve' OR char_length(comment) >= 20)
);

CREATE INDEX idx_feedback_proposal ON feedback(proposal_id);
CREATE INDEX idx_feedback_version ON feedback(proposal_version_id);
CREATE INDEX idx_feedback_reviewer ON feedback(reviewer_id);
CREATE INDEX idx_feedback_decision ON feedback(decision);
CREATE INDEX idx_feedback_created_at ON feedback(created_at);

-- Prevent multiple terminal decisions (approve/reject) on same proposal
CREATE UNIQUE INDEX idx_feedback_one_approval
ON feedback(proposal_id)
WHERE decision = 'approve';

CREATE UNIQUE INDEX idx_feedback_one_rejection
ON feedback(proposal_id)
WHERE decision = 'reject';

COMMENT ON TABLE feedback IS 'IMMUTABLE feedback records - never updated or deleted';
COMMENT ON COLUMN feedback.decision IS 'Teacher decision: approve, revise, or reject';
COMMENT ON COLUMN feedback.is_structured IS 'True if based on AI suggestions';

-- ============================================================================
-- PROJECT MANAGEMENT (Approved Proposals)
-- ============================================================================

-- Projects (Created from Approved Proposals)
CREATE TABLE projects (
    id SERIAL PRIMARY KEY,
    proposal_id INTEGER UNIQUE NOT NULL REFERENCES proposals(id) ON DELETE RESTRICT,
    team_id INTEGER NOT NULL REFERENCES teams(id) ON DELETE RESTRICT,
    department_id INTEGER NOT NULL REFERENCES departments(id) ON DELETE RESTRICT,

    -- Content
    summary TEXT NOT NULL,
    keywords VARCHAR(255)[],  -- Array of keywords for search

    -- Approval metadata
    approved_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

    -- Visibility control
    visibility VARCHAR(20) DEFAULT 'private' NOT NULL
               CHECK (visibility IN ('private', 'public')),
    published_at TIMESTAMP,
    published_by INTEGER REFERENCES users(id) ON DELETE RESTRICT,

    -- Engagement metrics
    view_count INTEGER DEFAULT 0,
    share_count INTEGER DEFAULT 0,
    download_count INTEGER DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,

    -- Constraints
    CHECK (visibility != 'public' OR (published_at IS NOT NULL AND published_by IS NOT NULL)),
    CHECK (view_count >= 0 AND share_count >= 0 AND download_count >= 0)
);

CREATE INDEX idx_projects_visibility ON projects(visibility) WHERE deleted_at IS NULL;
CREATE INDEX idx_projects_department ON projects(department_id);
CREATE INDEX idx_projects_team ON projects(team_id);
CREATE INDEX idx_projects_published ON projects(published_at) WHERE visibility = 'public';
CREATE INDEX idx_projects_keywords ON projects USING GIN(keywords);

COMMENT ON TABLE projects IS 'Approved proposals that have become formal projects';
COMMENT ON COLUMN projects.visibility IS 'private (team/advisor only) or public (in archive)';

-- Project Documentation (Immutable Once Uploaded)
CREATE TABLE project_documentation (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,

    -- Document type
    document_type VARCHAR(30) NOT NULL
                  CHECK (document_type IN ('final_report', 'presentation',
                                          'proposal_document', 'code_repository')),

    -- File storage
    file_url VARCHAR(500) NOT NULL,
    file_hash VARCHAR(64) NOT NULL,  -- SHA-256 for integrity
    file_size_bytes BIGINT NOT NULL,
    file_mime_type VARCHAR(100) NOT NULL,

    -- Review status
    status VARCHAR(20) DEFAULT 'pending' NOT NULL
           CHECK (status IN ('pending', 'approved', 'rejected')),
    review_comment TEXT,
    reviewed_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMP,

    -- Submission metadata
    submitted_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,

    -- Constraints
    UNIQUE(project_id, document_type),  -- One of each type per project
    CHECK (status != 'approved' OR (reviewed_by IS NOT NULL AND reviewed_at IS NOT NULL)),
    CHECK (status != 'rejected' OR (reviewed_by IS NOT NULL AND reviewed_at IS NOT NULL))
);

CREATE INDEX idx_documentation_project ON project_documentation(project_id);
CREATE INDEX idx_documentation_status ON project_documentation(status);
CREATE INDEX idx_documentation_type ON project_documentation(document_type);

COMMENT ON TABLE project_documentation IS 'Project documentation files (reports, presentations, code)';
COMMENT ON COLUMN project_documentation.document_type IS 'Type: final_report, presentation, proposal_document, or code_repository';

-- Project Reviews (Public Ratings/Comments)
CREATE TABLE project_reviews (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Review content
    rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment TEXT CHECK (comment IS NULL OR char_length(comment) <= 500),

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,

    -- Constraints
    UNIQUE(project_id, user_id)  -- One review per user per project
);

CREATE INDEX idx_reviews_project ON project_reviews(project_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_reviews_user ON project_reviews(user_id);
CREATE INDEX idx_reviews_rating ON project_reviews(rating);

COMMENT ON TABLE project_reviews IS 'Public user reviews and ratings for published projects';
COMMENT ON COLUMN project_reviews.rating IS 'Rating from 1 (poor) to 5 (excellent)';

-- ============================================================================
-- NOTIFICATION SYSTEM
-- ============================================================================

-- Notifications
CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Reference to source entity
    reference_type VARCHAR(50) NOT NULL,  -- 'team', 'proposal', 'project', 'feedback', etc.
    reference_id INTEGER NOT NULL,

    -- Notification content
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    action_url VARCHAR(500),  -- Deep link to relevant page

    -- Status
    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP,

    -- Priority (for future use)
    priority VARCHAR(20) DEFAULT 'normal' CHECK (priority IN ('low', 'normal', 'high', 'urgent')),

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Constraints
    CHECK (reference_type IN ('team', 'proposal', 'feedback', 'project', 'documentation', 'review', 'system'))
);

CREATE INDEX idx_notifications_user ON notifications(user_id);
CREATE INDEX idx_notifications_unread ON notifications(user_id, is_read) WHERE is_read = FALSE;
CREATE INDEX idx_notifications_type ON notifications(reference_type, reference_id);
CREATE INDEX idx_notifications_created ON notifications(created_at);

COMMENT ON TABLE notifications IS 'User notifications for all system events';
COMMENT ON COLUMN notifications.reference_type IS 'Type of entity that triggered notification';

-- ============================================================================
-- AUDIT LOGGING (WRITE-ONLY, NEVER UPDATE/DELETE)
-- ============================================================================

-- Audit Logs
CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,

    -- Entity information
    entity_type VARCHAR(50) NOT NULL,
    entity_id INTEGER,

    -- Action
    action VARCHAR(50) NOT NULL,  -- 'create', 'update', 'delete', 'submit', 'approve', etc.

    -- Actor information
    actor_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    actor_role VARCHAR(20),
    actor_email VARCHAR(255),  -- Denormalized for audit retention

    -- State snapshots (JSONB for flexibility)
    old_state JSONB,
    new_state JSONB,
    changes JSONB,  -- Specific fields that changed

    -- Request metadata
    ip_address INET NOT NULL,
    user_agent TEXT,
    request_id VARCHAR(255),
    session_id VARCHAR(255),

    -- Timestamp (immutable)
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,

    -- Additional context
    metadata JSONB,  -- Flexible field for additional context

    -- Constraints
    CHECK (entity_type IN ('user', 'team', 'team_member', 'proposal', 'proposal_version',
                          'feedback', 'project', 'documentation', 'review', 'notification'))
);

CREATE INDEX idx_audit_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_actor ON audit_logs(actor_id);
CREATE INDEX idx_audit_action ON audit_logs(action);
CREATE INDEX idx_audit_timestamp ON audit_logs(timestamp);
CREATE INDEX idx_audit_session ON audit_logs(session_id) WHERE session_id IS NOT NULL;

-- Partition audit_logs by month for better performance (PostgreSQL 10+)
-- Uncomment if using partitioning
-- ALTER TABLE audit_logs PARTITION BY RANGE (timestamp);

COMMENT ON TABLE audit_logs IS 'IMMUTABLE audit trail - write-only, never updated or deleted';
COMMENT ON COLUMN audit_logs.old_state IS 'JSON snapshot of entity before change';
COMMENT ON COLUMN audit_logs.new_state IS 'JSON snapshot of entity after change';
COMMENT ON COLUMN audit_logs.changes IS 'JSON object showing specific fields changed';

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Trigger function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply trigger to tables with updated_at
CREATE TRIGGER update_universities_updated_at BEFORE UPDATE ON universities
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_departments_updated_at BEFORE UPDATE ON departments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_teams_updated_at BEFORE UPDATE ON teams
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_proposals_updated_at BEFORE UPDATE ON proposals
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_projects_updated_at BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_project_reviews_updated_at BEFORE UPDATE ON project_reviews
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Trigger to prevent UPDATE/DELETE on immutable tables
CREATE OR REPLACE FUNCTION prevent_modification()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'UPDATE' THEN
        RAISE EXCEPTION 'Updates not allowed on immutable table %', TG_TABLE_NAME;
    ELSIF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION 'Deletes not allowed on immutable table %', TG_TABLE_NAME;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Apply immutability trigger to write-only tables
CREATE TRIGGER prevent_proposal_versions_modification
    BEFORE UPDATE OR DELETE ON proposal_versions
    FOR EACH ROW EXECUTE FUNCTION prevent_modification();

CREATE TRIGGER prevent_feedback_modification
    BEFORE UPDATE OR DELETE ON feedback
    FOR EACH ROW EXECUTE FUNCTION prevent_modification();

CREATE TRIGGER prevent_audit_logs_modification
    BEFORE UPDATE OR DELETE ON audit_logs
    FOR EACH ROW EXECUTE FUNCTION prevent_modification();

COMMENT ON FUNCTION prevent_modification() IS 'Prevents updates and deletes on immutable tables';

-- ============================================================================
-- INITIAL DATA / SEED DATA
-- ============================================================================

-- Insert default university
INSERT INTO universities (id, name, code, website)
VALUES (1, 'Adama Science and Technology University', 'ASTU', 'https://www.astu.edu.et')
ON CONFLICT (id) DO NOTHING;

-- Insert sample departments
INSERT INTO departments (university_id, name, code) VALUES
(1, 'Computer Science', 'CS'),
(1, 'Software Engineering', 'SE'),
(1, 'Information Technology', 'IT'),
(1, 'Electrical Engineering', 'EE'),
(1, 'Mechanical Engineering', 'ME')
ON CONFLICT (university_id, code) DO NOTHING;

-- Insert default admin user (password: Admin@123)
INSERT INTO users (name, email, password_hash, role, university_id, is_active, email_verified)
VALUES (
    'System Administrator',
    'admin@astu.edu.et',
    crypt('Admin@123', gen_salt('bf', 12)),
    'admin',
    1,
    TRUE,
    TRUE
)
ON CONFLICT (email) DO NOTHING;

-- ============================================================================
-- VIEWS FOR COMMON QUERIES
-- ============================================================================

-- View: Active proposals with team and advisor information
CREATE OR REPLACE VIEW v_active_proposals AS
SELECT
    p.id,
    p.status,
    p.submitted_at,
    p.created_at,
    t.id AS team_id,
    t.name AS team_name,
    d.id AS department_id,
    d.name AS department_name,
    u_leader.id AS leader_id,
    u_leader.name AS leader_name,
    u_leader.email AS leader_email,
    u_advisor.id AS advisor_id,
    u_advisor.name AS advisor_name,
    u_advisor.email AS advisor_email,
    pv.id AS current_version_id,
    pv.title AS current_title,
    pv.version_number
FROM proposals p
JOIN teams t ON p.team_id = t.id
JOIN departments d ON t.department_id = d.id
JOIN users u_leader ON t.leader_id = u_leader.id
JOIN users u_advisor ON t.advisor_id = u_advisor.id
LEFT JOIN proposal_versions pv ON p.current_version_id = pv.id
WHERE p.deleted_at IS NULL
  AND t.deleted_at IS NULL;

COMMENT ON VIEW v_active_proposals IS 'Active proposals with team and advisor details';

-- View: Public projects with ratings
CREATE OR REPLACE VIEW v_public_projects AS
SELECT
    p.id,
    p.summary,
    p.keywords,
    p.published_at,
    p.view_count,
    p.share_count,
    t.name AS team_name,
    d.name AS department_name,
    u.name AS university_name,
    pv.title,
    COALESCE(AVG(pr.rating), 0) AS average_rating,
    COUNT(pr.id) AS review_count
FROM projects p
JOIN teams t ON p.team_id = t.id
JOIN departments d ON t.department_id = d.id
JOIN universities u ON d.university_id = u.id
JOIN proposals prop ON p.proposal_id = prop.id
JOIN proposal_versions pv ON prop.current_version_id = pv.id AND pv.is_approved_version = TRUE
LEFT JOIN project_reviews pr ON p.id = pr.project_id AND pr.deleted_at IS NULL
WHERE p.visibility = 'public'
  AND p.deleted_at IS NULL
GROUP BY p.id, p.summary, p.keywords, p.published_at, p.view_count, p.share_count,
         t.name, d.name, u.name, pv.title;

COMMENT ON VIEW v_public_projects IS 'Public projects with average ratings for archive';

-- ============================================================================
-- FUNCTIONS FOR COMMON OPERATIONS
-- ============================================================================

-- Function to get unread notification count for a user
CREATE OR REPLACE FUNCTION get_unread_notification_count(p_user_id INTEGER)
RETURNS INTEGER AS $$
BEGIN
    RETURN (
        SELECT COUNT(*)
        FROM notifications
        WHERE user_id = p_user_id
          AND is_read = FALSE
    );
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_unread_notification_count IS 'Returns count of unread notifications for a user';

-- Function to check if user is team leader
CREATE OR REPLACE FUNCTION is_team_leader(p_user_id INTEGER, p_team_id INTEGER)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1
        FROM teams
        WHERE id = p_team_id
          AND leader_id = p_user_id
          AND deleted_at IS NULL
    );
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION is_team_leader IS 'Checks if user is the leader of specified team';

-- Function to get proposal state permissions
CREATE OR REPLACE FUNCTION get_proposal_permissions(p_proposal_id INTEGER, p_user_id INTEGER)
RETURNS TABLE (
    can_edit BOOLEAN,
    can_submit BOOLEAN,
    can_review BOOLEAN,
    can_approve BOOLEAN
) AS $$
DECLARE
    v_status VARCHAR(30);
    v_team_id INTEGER;
    v_leader_id INTEGER;
    v_advisor_id INTEGER;
BEGIN
    SELECT p.status, p.team_id, t.leader_id, t.advisor_id
    INTO v_status, v_team_id, v_leader_id, v_advisor_id
    FROM proposals p
    JOIN teams t ON p.team_id = t.id
    WHERE p.id = p_proposal_id;

    RETURN QUERY SELECT
        (v_status = 'draft' AND p_user_id = v_leader_id)::BOOLEAN,
        (v_status = 'draft' AND p_user_id = v_leader_id)::BOOLEAN,
        (v_status IN ('submitted', 'under_review') AND p_user_id = v_advisor_id)::BOOLEAN,
        (v_status = 'under_review' AND p_user_id = v_advisor_id)::BOOLEAN;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_proposal_permissions IS 'Returns user permissions for a proposal based on state and role';

-- ============================================================================
-- DATABASE MAINTENANCE
-- ============================================================================

-- Analyze tables for query optimization
ANALYZE universities;
ANALYZE departments;
ANALYZE users;
ANALYZE teams;
ANALYZE team_members;
ANALYZE proposals;
ANALYZE proposal_versions;
ANALYZE feedback;
ANALYZE projects;
ANALYZE project_documentation;
ANALYZE project_reviews;
ANALYZE notifications;
ANALYZE audit_logs;

-- ============================================================================
-- BACKUP AND RETENTION POLICIES (Documentation)
-- ============================================================================

/*
BACKUP STRATEGY:
1. Full backup daily at 2 AM
2. Transaction log backup every 4 hours
3. Retention: 30 days for daily backups, 1 year for monthly backups

AUDIT LOG RETENTION:
- Keep all audit logs for 7 years (standard academic retention)
- Consider archiving to separate storage after 2 years

SOFT-DELETED RECORDS:
- User-facing entities use soft delete (deleted_at)
- Permanently delete after 90 days (compliance period)
- Immutable tables (versions, feedback, audit_logs) NEVER deleted
*/

-- ============================================================================
-- END OF SCHEMA
-- ============================================================================
```

## Schema Validation Queries

```sql
-- Verify immutable tables have triggers
SELECT
    tgname AS trigger_name,
    tgrelid::regclass AS table_name
FROM pg_trigger
WHERE tgname LIKE 'prevent%modification';

-- Check foreign key constraints
SELECT
    tc.table_name,
    kcu.column_name,
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name
FROM information_schema.table_constraints AS tc
JOIN information_schema.key_column_usage AS kcu
    ON tc.constraint_name = kcu.constraint_name
JOIN information_schema.constraint_column_usage AS ccu
    ON ccu.constraint_name = tc.constraint_name
WHERE tc.constraint_type = 'FOREIGN KEY'
ORDER BY tc.table_name, kcu.column_name;

-- Check indexes
SELECT
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE schemaname = 'public'
ORDER BY tablename, indexname;

-- Verify check constraints
SELECT
    tc.table_name,
    cc.check_clause
FROM information_schema.table_constraints tc
JOIN information_schema.check_constraints cc
    ON tc.constraint_name = cc.constraint_name
WHERE tc.constraint_type = 'CHECK'
ORDER BY tc.table_name;
```

## Performance Optimization Notes

### Indexing Strategy

1. **Foreign Keys**: All foreign keys indexed (automatic with GORM)
2. **Status Fields**: Indexed for filtering (proposals.status, teams.status)
3. **Composite Indexes**: (user_id, is_read) for notification queries
4. **Partial Indexes**: WHERE deleted_at IS NULL for soft-deleted tables
5. **GIN Indexes**: For array columns (projects.keywords) and JSONB

### Query Optimization

1. **Use Views**: Pre-computed joins for common queries
2. **Eager Loading**: Load relations in application layer
3. **Pagination**: LIMIT/OFFSET for large result sets
4. **Covering Indexes**: Include frequently queried columns

### Partitioning Considerations

For high-volume installations, consider partitioning:

- `audit_logs` by month (range partitioning)
- `notifications` by month
- `project_reviews` by project_id (hash partitioning)

## Migration Strategy

1. **Version Control**: Use migration tool (golang-migrate, Goose)
2. **Rollback Plan**: Every migration must have down() function
3. **Zero-Downtime**: Add columns first, backfill, then add constraints
4. **Testing**: Run migrations on staging before production

## Security Considerations

1. **Row-Level Security** (RLS): Can be enabled for multi-tenancy
2. **Encrypted Columns**: Consider encrypting sensitive data (passwords already bcrypt)
3. **Audit Access**: Log all queries to audit_logs table
4. **Backup Encryption**: Encrypt backups at rest

This schema implements academic integrity through immutability, provides comprehensive audit trails, enforces workflow constraints, and scales to multi-university deployment.
