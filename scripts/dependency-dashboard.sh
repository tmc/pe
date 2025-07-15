#!/bin/bash
#
# Dependency Monitoring Dashboard
# Displays current status of all dependency metrics
#

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPORTS_DIR="$PROJECT_ROOT/reports/dependencies"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

print_header() {
    echo -e "${BOLD}${BLUE}$1${NC}"
    echo -e "${BLUE}$(printf '=%.0s' {1..60})${NC}"
}

print_metric() {
    local label="$1"
    local value="$2"
    local status="${3:-}"
    
    printf "%-25s: " "$label"
    if [[ "$status" == "good" ]]; then
        echo -e "${GREEN}$value${NC}"
    elif [[ "$status" == "warning" ]]; then
        echo -e "${YELLOW}$value${NC}"
    elif [[ "$status" == "error" ]]; then
        echo -e "${RED}$value${NC}"
    else
        echo "$value"
    fi
}

main() {
    clear
    
    print_header "PE Dependency Monitoring Dashboard"
    echo ""
    
    cd "$PROJECT_ROOT"
    
    # Current Status
    print_header "Current Status"
    
    # Security status
    if govulncheck ./... >/dev/null 2>&1; then
        print_metric "Security Status" "✅ No vulnerabilities" "good"
    else
        print_metric "Security Status" "❌ Vulnerabilities found" "error"
    fi
    
    # Dependency counts
    local total_deps=$(go list -m all | wc -l)
    local outdated_deps=$(go list -m -u all | grep '\[.*\]' | wc -l)
    
    print_metric "Total Dependencies" "$total_deps"
    print_metric "Outdated Dependencies" "$outdated_deps" "$([ $outdated_deps -eq 0 ] && echo 'good' || echo 'warning')"
    print_metric "Go Version" "$(go version | cut -d' ' -f3)"
    
    echo ""
    
    # Recent Reports
    print_header "Recent Reports"
    
    if [[ -d "$REPORTS_DIR" ]]; then
        echo "📊 Latest health metrics:"
        ls -la "$REPORTS_DIR"/health-metrics-*.json 2>/dev/null | tail -1 | awk '{print "   " $9 " (" $6 " " $7 " " $8 ")"}'
        
        echo "📋 Latest freshness report:"
        ls -la "$REPORTS_DIR"/freshness-*.txt 2>/dev/null | tail -1 | awk '{print "   " $9 " (" $6 " " $7 " " $8 ")"}'
        
        echo "🔍 Latest vulnerability check:"
        ls -la "$REPORTS_DIR"/vulncheck-*.txt 2>/dev/null | tail -1 | awk '{print "   " $9 " (" $6 " " $7 " " $8 ")"}'
    else
        echo "⚠️  No reports directory found"
    fi
    
    echo ""
    
    # Quick Actions
    print_header "Quick Actions"
    echo "1. Run vulnerability check: ./scripts/dependency-tools.sh daily-check"
    echo "2. Generate health metrics: ./scripts/dependency-tools.sh health-metrics"
    echo "3. Update golang.org/x: ./scripts/dependency-tools.sh update-golang-x"
    echo "4. Full maintenance: ./scripts/dependency-tools.sh all"
    echo ""
    
    # Outdated packages summary
    if [[ $outdated_deps -gt 0 ]]; then
        print_header "Outdated Packages (Top 5)"
        go list -m -u all | grep '\[.*\]' | head -5
        echo ""
    fi
}

main "$@"
