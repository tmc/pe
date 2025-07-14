#!/bin/bash
#
# Dependency Management Tools for PE Project
# Automated scripts for maintaining Go module dependencies
#

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_FILE="${PROJECT_ROOT}/logs/dependency-maintenance.log"
REPORT_DIR="${PROJECT_ROOT}/reports/dependencies"

# Ensure directories exist
mkdir -p "$(dirname "$LOG_FILE")" "$REPORT_DIR"

# Logging function
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Daily vulnerability check
daily_vuln_check() {
    log "Starting daily vulnerability check"
    print_status "$BLUE" "🔍 Running govulncheck..."
    
    cd "$PROJECT_ROOT"
    
    if govulncheck ./... > "$REPORT_DIR/vulncheck-$(date +%Y%m%d).txt" 2>&1; then
        print_status "$GREEN" "✅ No vulnerabilities found"
        log "Vulnerability check passed"
        return 0
    else
        print_status "$RED" "❌ Vulnerabilities detected!"
        log "ALERT: Vulnerabilities found - see report"
        cat "$REPORT_DIR/vulncheck-$(date +%Y%m%d).txt"
        return 1
    fi
}

# Weekly freshness report
weekly_freshness_report() {
    log "Generating weekly freshness report"
    print_status "$BLUE" "📊 Analyzing dependency freshness..."
    
    cd "$PROJECT_ROOT"
    local report_file="$REPORT_DIR/freshness-$(date +%Y%m%d).txt"
    
    {
        echo "=== DEPENDENCY FRESHNESS REPORT ==="
        echo "Generated: $(date)"
        echo "Project: $(basename "$PROJECT_ROOT")"
        echo ""
        
        echo "=== OUTDATED PACKAGES ==="
        go list -m -u all | grep '\[.*\]' || echo "All packages are up to date"
        echo ""
        
        echo "=== DEPENDENCY COUNT ==="
        echo "Total dependencies: $(go list -m all | wc -l)"
        echo "Direct dependencies: $(go mod graph | grep "^$(go list -m)" | wc -l)"
        echo "Outdated packages: $(go list -m -u all | grep '\[.*\]' | wc -l)"
        echo ""
        
        echo "=== SECURITY STATUS ==="
        if govulncheck ./... >/dev/null 2>&1; then
            echo "✅ No known vulnerabilities"
        else
            echo "❌ Vulnerabilities detected"
        fi
    } > "$report_file"
    
    print_status "$GREEN" "📋 Report saved to: $report_file"
    log "Weekly freshness report generated"
}

# Monthly full dependency audit
monthly_audit() {
    log "Starting monthly dependency audit"
    print_status "$BLUE" "🔍 Performing comprehensive dependency audit..."
    
    cd "$PROJECT_ROOT"
    local audit_file="$REPORT_DIR/audit-$(date +%Y%m).txt"
    
    {
        echo "=== COMPREHENSIVE DEPENDENCY AUDIT ==="
        echo "Generated: $(date)"
        echo "Go Version: $(go version)"
        echo ""
        
        echo "=== MODULE INFO ==="
        go list -m
        echo ""
        
        echo "=== ALL DEPENDENCIES ==="
        go list -m all
        echo ""
        
        echo "=== DEPENDENCY GRAPH ==="
        go mod graph | head -20
        if [[ $(go mod graph | wc -l) -gt 20 ]]; then
            echo "... (truncated, $(go mod graph | wc -l) total relationships)"
        fi
        echo ""
        
        echo "=== OUTDATED ANALYSIS ==="
        go list -m -u all | grep '\[.*\]' || echo "All packages are current"
        echo ""
        
        echo "=== LICENSE SCAN ==="
        echo "Note: Manual license review recommended"
        go list -m all | while read -r mod; do
            echo "- $mod"
        done
        echo ""
        
        echo "=== SECURITY SCAN ==="
        govulncheck ./... || true
        
    } > "$audit_file"
    
    print_status "$GREEN" "📋 Audit report saved to: $audit_file"
    log "Monthly audit completed"
}

# Update specific dependency with testing
update_dependency() {
    local dep_name="$1"
    local version="${2:-latest}"
    
    log "Updating dependency: $dep_name to $version"
    print_status "$YELLOW" "🔄 Updating $dep_name..."
    
    cd "$PROJECT_ROOT"
    
    # Backup current go.mod and go.sum
    cp go.mod "go.mod.backup-$(date +%Y%m%d-%H%M%S)"
    cp go.sum "go.sum.backup-$(date +%Y%m%d-%H%M%S)"
    
    # Update the dependency
    if [[ "$version" == "latest" ]]; then
        go get -u "$dep_name"
    else
        go get "$dep_name@$version"
    fi
    
    # Run tests to verify update
    if go test ./...; then
        print_status "$GREEN" "✅ Update successful and tests pass"
        go mod tidy
        log "Successfully updated $dep_name to $version"
    else
        print_status "$RED" "❌ Tests failed after update, reverting..."
        mv "go.mod.backup-$(date +%Y%m%d)"* go.mod
        mv "go.sum.backup-$(date +%Y%m%d)"* go.sum
        log "Reverted $dep_name update due to test failures"
        return 1
    fi
}

# Batch update golang.org/x packages
update_golang_x_packages() {
    log "Starting batch update of golang.org/x packages"
    print_status "$BLUE" "🔄 Updating golang.org/x packages..."
    
    local packages=(
        "golang.org/x/crypto"
        "golang.org/x/net"
        "golang.org/x/sys"
        "golang.org/x/term"
        "golang.org/x/tools"
        "golang.org/x/sync"
        "golang.org/x/mod"
    )
    
    local failed_updates=()
    
    for pkg in "${packages[@]}"; do
        if go list -m all | grep -q "^$pkg "; then
            print_status "$YELLOW" "Updating $pkg..."
            if update_dependency "$pkg"; then
                print_status "$GREEN" "✅ $pkg updated successfully"
            else
                failed_updates+=("$pkg")
                print_status "$RED" "❌ Failed to update $pkg"
            fi
        else
            print_status "$YELLOW" "⏭️  $pkg not in dependencies, skipping"
        fi
    done
    
    if [[ ${#failed_updates[@]} -eq 0 ]]; then
        print_status "$GREEN" "🎉 All golang.org/x packages updated successfully"
        log "Batch update of golang.org/x packages completed successfully"
    else
        print_status "$YELLOW" "⚠️  Some packages failed to update: ${failed_updates[*]}"
        log "Batch update completed with failures: ${failed_updates[*]}"
    fi
}

# Generate dependency health metrics
generate_health_metrics() {
    log "Generating dependency health metrics"
    print_status "$BLUE" "📊 Calculating dependency health metrics..."
    
    cd "$PROJECT_ROOT"
    local metrics_file="$REPORT_DIR/health-metrics-$(date +%Y%m%d).json"
    
    local total_deps=$(go list -m all | wc -l)
    local direct_deps=$(go mod graph | grep "^$(go list -m)" | wc -l)
    local outdated_deps=$(go list -m -u all | grep '\[.*\]' | wc -l)
    local security_status="unknown"
    
    if govulncheck ./... >/dev/null 2>&1; then
        security_status="clean"
    else
        security_status="vulnerabilities"
    fi
    
    cat > "$metrics_file" <<EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "project": "$(basename "$PROJECT_ROOT")",
  "go_version": "$(go version | cut -d' ' -f3)",
  "metrics": {
    "total_dependencies": $total_deps,
    "direct_dependencies": $direct_deps,
    "indirect_dependencies": $((total_deps - direct_deps)),
    "outdated_dependencies": $outdated_deps,
    "outdated_percentage": $(( outdated_deps * 100 / total_deps )),
    "security_status": "$security_status",
    "transitive_multiplier": $(echo "scale=2; $total_deps / $direct_deps" | bc -l)
  }
}
EOF
    
    print_status "$GREEN" "📋 Health metrics saved to: $metrics_file"
    log "Dependency health metrics generated"
}

# Main menu function
show_menu() {
    echo ""
    print_status "$BLUE" "🔧 PE Dependency Management Tools"
    echo ""
    echo "1) Daily vulnerability check"
    echo "2) Weekly freshness report"
    echo "3) Monthly full audit"
    echo "4) Update specific dependency"
    echo "5) Batch update golang.org/x packages"
    echo "6) Generate health metrics"
    echo "7) Update go.starlark.net"
    echo "8) Run all maintenance tasks"
    echo "9) Exit"
    echo ""
}

# Special handler for starlark update
update_starlark() {
    log "Updating go.starlark.net with special handling"
    print_status "$BLUE" "🌟 Updating go.starlark.net..."
    
    # This is a critical dependency, so extra care is needed
    cd "$PROJECT_ROOT"
    
    # Create comprehensive backup
    local backup_dir="backups/starlark-update-$(date +%Y%m%d-%H%M%S)"
    mkdir -p "$backup_dir"
    cp go.mod go.sum "$backup_dir/"
    
    # Update starlark
    if update_dependency "go.starlark.net"; then
        print_status "$GREEN" "✅ Starlark updated successfully"
        # Run extended tests for starlark functionality
        if go test ./ext/starlark/... ./cmd/pe/... -v; then
            print_status "$GREEN" "✅ Extended starlark tests pass"
            log "Starlark update completed successfully with extended testing"
        else
            print_status "$RED" "❌ Extended tests failed, manual review needed"
            log "Starlark update succeeded but extended tests failed"
        fi
    else
        print_status "$RED" "❌ Starlark update failed"
        log "Failed to update go.starlark.net"
    fi
}

# Run all maintenance tasks
run_all_maintenance() {
    log "Running all maintenance tasks"
    print_status "$BLUE" "🔄 Running comprehensive maintenance..."
    
    daily_vuln_check
    echo ""
    weekly_freshness_report
    echo ""
    generate_health_metrics
    echo ""
    
    print_status "$GREEN" "🎉 All maintenance tasks completed"
    log "Comprehensive maintenance run completed"
}

# Main script logic
main() {
    if [[ $# -eq 0 ]]; then
        # Interactive mode
        while true; do
            show_menu
            read -p "Choose an option: " choice
            case $choice in
                1) daily_vuln_check ;;
                2) weekly_freshness_report ;;
                3) monthly_audit ;;
                4) 
                    read -p "Enter dependency name: " dep_name
                    read -p "Enter version (or 'latest'): " version
                    update_dependency "$dep_name" "$version"
                    ;;
                5) update_golang_x_packages ;;
                6) generate_health_metrics ;;
                7) update_starlark ;;
                8) run_all_maintenance ;;
                9) 
                    print_status "$GREEN" "👋 Goodbye!"
                    exit 0
                    ;;
                *) print_status "$RED" "Invalid option" ;;
            esac
            echo ""
            read -p "Press Enter to continue..."
        done
    else
        # Command-line mode
        case "$1" in
            "daily-check"|"vuln-check") daily_vuln_check ;;
            "weekly-report"|"freshness") weekly_freshness_report ;;
            "monthly-audit"|"audit") monthly_audit ;;
            "update") 
                [[ $# -lt 2 ]] && { echo "Usage: $0 update <dependency> [version]"; exit 1; }
                update_dependency "$2" "${3:-latest}"
                ;;
            "update-golang-x") update_golang_x_packages ;;
            "health-metrics"|"metrics") generate_health_metrics ;;
            "update-starlark") update_starlark ;;
            "all"|"maintenance") run_all_maintenance ;;
            "help"|"--help")
                echo "PE Dependency Management Tools"
                echo ""
                echo "Usage: $0 [command]"
                echo ""
                echo "Commands:"
                echo "  daily-check        Run daily vulnerability check"
                echo "  weekly-report      Generate weekly freshness report"
                echo "  monthly-audit      Perform monthly dependency audit"
                echo "  update <dep> [ver] Update specific dependency"
                echo "  update-golang-x    Batch update golang.org/x packages"
                echo "  health-metrics     Generate dependency health metrics"
                echo "  update-starlark    Update go.starlark.net with special handling"
                echo "  all                Run all maintenance tasks"
                echo "  help               Show this help"
                echo ""
                echo "If no command is provided, interactive mode will start."
                ;;
            *) 
                print_status "$RED" "Unknown command: $1"
                echo "Use '$0 help' for usage information"
                exit 1
                ;;
        esac
    fi
}

# Execute main function with all arguments
main "$@"