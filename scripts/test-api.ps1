# Test script for University Project Hub API
$baseUrl = "http://localhost:8080"

# Test 1: Health check
Write-Host "`n========== Test 1: Health Check ==========" -ForegroundColor Cyan
$response = Invoke-RestMethod -Uri "$baseUrl/health" -Method Get
Write-Host "Health status:" ($response | ConvertTo-Json -Depth 3)

# Test 2: Register a new user
Write-Host "`n========== Test 2: Register User ==========" -ForegroundColor Cyan
$registerData = @{
    name = "John Doe"
    email = "john.doe@astu.edu.et"
    password = "password123"
    role = "student"
    university_id = 1
    department_id = 1
} | ConvertTo-Json

try {
    $registerResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/register" `
        -Method Post `
        -Body $registerData `
        -ContentType "application/json"
    Write-Host "Register response:" ($registerResponse | ConvertTo-Json -Depth 3)
} catch {
    Write-Host "Register error: $($_.Exception.Message)" -ForegroundColor Red
    $responseContent = $_.ErrorDetails.Message
    Write-Host "Error details: $responseContent"
}

# Test 3: Login
Write-Host "`n========== Test 3: Login ==========" -ForegroundColor Cyan
$loginData = @{
    email = "john.doe@astu.edu.et"
    password = "password123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/login" `
        -Method Post `
        -Body $loginData `
        -ContentType "application/json"
    Write-Host "Login response:" ($loginResponse | ConvertTo-Json -Depth 3)
    
    $token = $loginResponse.data.token
    Write-Host "`nToken obtained: $token" -ForegroundColor Green

    # Test 4: Get profile (protected route)
    Write-Host "`n========== Test 4: Get Profile (Protected) ==========" -ForegroundColor Cyan
    $headers = @{
        Authorization = "Bearer $token"
    }
    $profileResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/profile" `
        -Method Get `
        -Headers $headers
    Write-Host "Profile response:" ($profileResponse | ConvertTo-Json -Depth 3)

    # Test 5: Access protected teams endpoint
    Write-Host "`n========== Test 5: Access Teams (Protected) ==========" -ForegroundColor Cyan
    $teamsResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/teams" `
        -Method Get `
        -Headers $headers
    Write-Host "Teams response:" ($teamsResponse | ConvertTo-Json -Depth 3)

} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    $responseContent = $_.ErrorDetails.Message
    Write-Host "Error details: $responseContent"
}

Write-Host "`n========== Tests Complete ==========" -ForegroundColor Green
