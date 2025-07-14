#!/bin/bash
#
# Setup script for PE dependency monitoring system
# Initializes cron jobs and directory structure for automated dependency management
#

set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SCRIPT_DIR="$PROJECT_ROOT/scripts"
DEPENDENCY_TOOL="$SCRIPT_DIR/dependency-tools.sh"

print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Create directory structure
setup_directories() {
    print_status "$BLUE" "📁 Setting up directory structure..."
    
    mkdir -p "$PROJECT_ROOT"/{logs,reports/dependencies,backups}
    
    # Create .gitkeep files to preserve directory structure in git
    touch "$PROJECT_ROOT/logs/.gitkeep"
    touch "$PROJECT_ROOT/reports/.gitkeep"
    touch "$PROJECT_ROOT/backups/.gitkeep"
    
    print_status "$GREEN" "✅ Directory structure created"
}

# Setup cron jobs for automated monitoring
setup_cron_jobs() {
    print_status "$BLUE" "⏰ Setting up cron jobs..."
    
    # Create temporary cron file
    local temp_cron="/tmp/pe_dependency_cron"
    
    # Get existing cron jobs (excluding PE dependency ones)
    crontab -l 2>/dev/null | grep -v "# PE Dependency" > "$temp_cron" || true
    
    # Add PE dependency monitoring jobs
    cat >> "$temp_cron" <<EOF

# PE Dependency Monitoring Jobs
# Daily vulnerability check at 9 AM
0 9 * * * cd $PROJECT_ROOT && $DEPENDENCY_TOOL daily-check # PE Dependency

# Weekly freshness report on Mondays at 10 AM  
0 10 * * 1 cd $PROJECT_ROOT && $DEPENDENCY_TOOL weekly-report # PE Dependency

# Monthly audit on the 1st of each month at 11 AM
0 11 1 * * cd $PROJECT_ROOT && $DEPENDENCY_TOOL monthly-audit # PE Dependency

# Health metrics every 3 days at 2 PM
0 14 */3 * * cd $PROJECT_ROOT && $DEPENDENCY_TOOL health-metrics # PE Dependency
EOF
    
    # Install the new cron schedule
    if crontab "$temp_cron"; then
        print_status "$GREEN" "✅ Cron jobs installed successfully"
    else
        print_status "$YELLOW" "⚠️  Failed to install cron jobs (manual setup may be required)"
    fi
    
    # Clean up
    rm "$temp_cron"
}

# Create initial baseline reports
create_baseline() {
    print_status "$BLUE" "📊 Creating baseline reports..."
    
    cd "$PROJECT_ROOT"
    
    # Run initial dependency tools to create baseline
    "$DEPENDENCY_TOOL" health-metrics
    "$DEPENDENCY_TOOL" weekly-report
    "$DEPENDENCY_TOOL" daily-check
    
    print_status "$GREEN" "✅ Baseline reports created"
}

# Setup Git hooks for dependency checking
setup_git_hooks() {
    print_status "$BLUE" "🔗 Setting up Git hooks..."
    
    local hooks_dir="$PROJECT_ROOT/.git/hooks"
    local pre_commit_hook="$hooks_dir/pre-commit"
    
    # Create pre-commit hook that checks for dependency security
    cat > "$pre_commit_hook" <<'EOF'
#!/bin/bash
#
# Pre-commit hook for PE dependency security check
#

echo "🔍 Running pre-commit dependency security check..."

# Change to project root
cd "$(git rev-parse --show-toplevel)"

# Run vulnerability check
if command -v govulncheck >/dev/null 2>&1; then
    if ! govulncheck ./... >/dev/null 2>&1; then
        echo "❌ Security vulnerabilities detected in dependencies!"
        echo "Run 'govulncheck ./...' for details"
        echo "Consider running './scripts/dependency-tools.sh daily-check'"
        exit 1
    fi
    echo "✅ No security vulnerabilities in dependencies"
else
    echo "⚠️  govulncheck not available, skipping security check"
fi

echo "✅ Pre-commit dependency check passed"
EOF
    
    chmod +x "$pre_commit_hook"
    
    print_status "$GREEN" "✅ Git hooks configured"
}

# Create monitoring dashboard script
create_dashboard() {
    print_status "$BLUE" "📊 Creating monitoring dashboard..."
    
    cat > "$SCRIPT_DIR/dependency-dashboard.sh" <<'EOF'
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
EOF
    
    chmod +x "$SCRIPT_DIR/dependency-dashboard.sh"
    
    print_status "$GREEN" "✅ Monitoring dashboard created"
}

# Main setup function
main() {
    print_status "$BLUE" "🚀 Setting up PE dependency monitoring system..."
    echo ""
    
    setup_directories
    echo ""
    
    if [[ "${1:-}" != "--no-cron" ]]; then
        setup_cron_jobs
        echo ""
    else
        print_status "$YELLOW" "⏭️  Skipping cron setup (--no-cron specified)"
        echo ""
    fi
    
    create_baseline
    echo ""
    
    setup_git_hooks
    echo ""
    
    create_dashboard
    echo ""
    
    print_status "$GREEN" "🎉 Dependency monitoring system setup complete!"
    echo ""
    echo "Available tools:"
    echo "  • $DEPENDENCY_TOOL (main tool)"
    echo "  • $SCRIPT_DIR/dependency-dashboard.sh (dashboard)"
    echo ""
    echo "To view current status:"
    echo "  $SCRIPT_DIR/dependency-dashboard.sh"
    echo ""
    echo "To run maintenance:"
    echo "  $DEPENDENCY_TOOL all"
}

# Check if script is being run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi