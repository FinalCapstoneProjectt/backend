# Test script for University Project Hub API
# Run server first: go run cmd/server/main.go

$baseUrl = "http://localhost:8080/api/v1"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Testing University Project Hub API" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

# 1. Health Check
Write-Host "1. Health Check..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method Get
    Write-Host "✓ Server is running" -ForegroundColor Green
    $response | ConvertTo-Json
} catch {
    Write-Host "✗ Server is not running. Start it first!" -ForegroundColor Red
    exit
}

# 2. Register Admin (if not exists)
Write-Host "`n2. Register Admin User..." -ForegroundColor Yellow
$adminBody = @{
    name = "Admin User"
    email = "admin@university.edu"
    password = "admin123"
    role = "admin"
} | ConvertTo-Json

try {
    $admin = Invoke-RestMethod -Uri "$baseUrl/auth/register" -Method Post -Body $adminBody -ContentType "application/json"
    Write-Host "✓ Admin registered" -ForegroundColor Green
} catch {
    Write-Host "! Admin might already exist (that's ok)" -ForegroundColor Yellow
}

# 3. Login as Admin
Write-Host "`n3. Login as Admin..." -ForegroundColor Yellow
$loginBody = @{
    email = "admin@university.edu"
    password = "admin123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
    $token = $loginResponse.data.token
    Write-Host "✓ Login successful" -ForegroundColor Green
    Write-Host "Token: $token"
    
    $headers = @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }
} catch {
    Write-Host "✗ Login failed: $_" -ForegroundColor Red
    exit
}

# 4. Create University
Write-Host "`n4. Create University..." -ForegroundColor Yellow
$universityBody = @{
    name = "Tech University"
    academic_year = "2025-2026"
    project_period = "September 2025 - June 2026"
    visibility_rule = "private"
    ai_checker_enabled = $true
} | ConvertTo-Json

try {
    $university = Invoke-RestMethod -Uri "$baseUrl/universities" -Method Post -Body $universityBody -Headers $headers
    $universityId = $university.data.id
    Write-Host "✓ University created (ID: $universityId)" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed: $_" -ForegroundColor Red
}

# 5. List Universities
Write-Host "`n5. List Universities..." -ForegroundColor Yellow
try {
    $universities = Invoke-RestMethod -Uri "$baseUrl/universities" -Method Get -Headers $headers
    Write-Host "✓ Found $($universities.data.Count) universities" -ForegroundColor Green
    $universities.data | Format-Table -Property id, name, academic_year
} catch {
    Write-Host "✗ Failed: $_" -ForegroundColor Red
}

# 6. Create Department
Write-Host "`n6. Create Department..." -ForegroundColor Yellow
$deptBody = @{
    name = "Computer Science Engineering"
    code = "CSE"
    university_id = $universityId
} | ConvertTo-Json

try {
    $department = Invoke-RestMethod -Uri "$baseUrl/departments" -Method Post -Body $deptBody -Headers $headers
    $departmentId = $department.data.id
    Write-Host "✓ Department created (ID: $departmentId)" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed: $_" -ForegroundColor Red
}

# 7. List Departments
Write-Host "`n7. List Departments..." -ForegroundColor Yellow
try {
    $departments = Invoke-RestMethod -Uri "$baseUrl/departments" -Method Get -Headers $headers
    Write-Host "✓ Found $($departments.data.Count) departments" -ForegroundColor Green
    $departments.data | Format-Table -Property id, name, code
} catch {
    Write-Host "✗ Failed: $_" -ForegroundColor Red
}

# 8. Create Teacher
Write-Host "`n8. Create Teacher..." -ForegroundColor Yellow
$teacherBody = @{
    name = "Dr. John Smith"
    email = "john.smith@university.edu"
    password = "teacher123"
    university_id = $universityId
    department_id = $departmentId
} | ConvertTo-Json

try {
    $teacher = Invoke-RestMethod -Uri "$baseUrl/admin/users/teacher" -Method Post -Body $teacherBody -Headers $headers
    $teacherId = $teacher.data.id
    Write-Host "✓ Teacher created (ID: $teacherId)" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed: $_" -ForegroundColor Red
}

# 9. Create Student
Write-Host "`n9. Create Student..." -ForegroundColor Yellow
$studentBody = @{
    name = "Alice Johnson"
    email = "alice.johnson@student.edu"
    password = "student123"
    student_id = "STU2025001"
    university_id = $universityId
    department_id = $departmentId
} | ConvertTo-Json

try {
    $student = Invoke-RestMethod -Uri "$baseUrl/admin/users/student" -Method Post -Body $studentBody -Headers $headers
    $studentId = $student.data.id
    Write-Host "✓ Student created (ID: $studentId)" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed: $_" -ForegroundColor Red
}

# 10. List All Users
Write-Host "`n10. List All Users..." -ForegroundColor Yellow
try {
    $users = Invoke-RestMethod -Uri "$baseUrl/admin/users" -Method Get -Headers $headers
    Write-Host "✓ Found $($users.data.Count) users" -ForegroundColor Green
    $users.data | Format-Table -Property id, name, email, role, is_active
} catch {
    Write-Host "✗ Failed: $_" -ForegroundColor Red
}

# 11. Get Swagger Documentation
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "✓ All tests completed!" -ForegroundColor Green
Write-Host "View Swagger UI at: http://localhost:8080/swagger/index.html" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
