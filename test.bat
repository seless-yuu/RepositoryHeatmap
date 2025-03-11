@echo off
rem Repository Heatmap Test Script for Windows
rem This script is used to run tests on Windows

setlocal enabledelayedexpansion

echo =========================================================
echo        Repository Heatmap Test Script
echo =========================================================
echo.

rem List of directories to test
set TEST_DIRS=^
  "./internal/utils" ^
  "./internal/heatmap"

rem Initialize result variable
set FAILED=0

echo Running tests...
echo.

rem Run tests for each directory
for %%d in (%TEST_DIRS%) do (
  echo Running tests: %%d
  go test -v %%d
  
  if !errorlevel! neq 0 (
    echo Test failed: %%d
    set FAILED=1
  ) else (
    echo Test succeeded: %%d
  )
  echo.
)

rem Generate coverage report
echo Generating coverage report...
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
echo Coverage report generated at coverage.html
echo.

rem Display results
echo =========================================================
if !FAILED! equ 0 (
  echo All tests passed!
) else (
  echo Some tests failed
)
echo =========================================================

exit /b !FAILED! 