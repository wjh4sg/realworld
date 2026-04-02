param(
    [string]$BaseUrl = $env:API_BASE_URL
)

if ([string]::IsNullOrWhiteSpace($BaseUrl)) {
    $BaseUrl = "http://localhost:18080/api"
}

$postmanCollection = ".\api\Conduit.postman_collection.json"
$testEnvironment = ".\api\test-env.json"
$reportDir = ".\.reports\newman"
$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$reportFile = Join-Path $reportDir "api-test-report-$timestamp.html"

Write-Host "========================================" -ForegroundColor Green
Write-Host "API Automated Test Script (Windows)" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host "Base URL: $BaseUrl" -ForegroundColor Cyan

if (-not (Test-Path $postmanCollection)) {
    Write-Host "Error: Postman collection file not found at $postmanCollection" -ForegroundColor Red
    exit 1
}

if (-not (Test-Path $testEnvironment)) {
    Write-Host "Error: Test environment file not found at $testEnvironment" -ForegroundColor Red
    exit 1
}

if (-not (Get-Command newman -ErrorAction SilentlyContinue)) {
    Write-Host "Error: newman is not installed or not in PATH" -ForegroundColor Red
    exit 1
}

if (-not (Test-Path $reportDir)) {
    New-Item -ItemType Directory -Path $reportDir -Force | Out-Null
}

Write-Host "Checking API availability..." -ForegroundColor Cyan
try {
    $response = Invoke-WebRequest -Uri "$BaseUrl/articles" -Method GET -TimeoutSec 5 -UseBasicParsing -ErrorAction Stop
    Write-Host "API server is reachable (status $($response.StatusCode))" -ForegroundColor Green
} catch {
    Write-Host "Warning: API server may not be reachable yet: $($_.Exception.Message)" -ForegroundColor Yellow
}

Write-Host "Starting Newman tests..." -ForegroundColor Cyan

try {
    newman run $postmanCollection `
        --environment $testEnvironment `
        --env-var "APIURL=$BaseUrl" `
        --reporters cli,html `
        --reporter-html-export $reportFile `
        --timeout 60000

    Write-Host "Test execution completed" -ForegroundColor Green
    Write-Host "Report: $reportFile" -ForegroundColor Green
} catch {
    Write-Host "Error: Test execution failed - $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

Write-Host "========================================" -ForegroundColor Green
