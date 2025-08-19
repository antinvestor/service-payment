#!/bin/bash
# Test script for Dependabot setup validation
# This script verifies the Dependabot configuration and CI setup

set -e

echo "🔍 Validating Dependabot setup..."

# Check if we're in the right directory
if [ ! -f ".github/dependabot.yml" ]; then
    echo "❌ Error: .github/dependabot.yml not found. Are you in the repository root?"
    exit 1
fi

echo "✅ Dependabot configuration file found"

# Validate YAML syntax
echo "🔧 Validating YAML syntax..."
if command -v python3 &> /dev/null; then
    python3 -c "import yaml; yaml.safe_load(open('.github/dependabot.yml'))" && echo "✅ dependabot.yml syntax is valid"
    python3 -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yml'))" && echo "✅ ci.yml syntax is valid"
    python3 -c "import yaml; yaml.safe_load(open('.github/workflows/dependabot-auto-merge.yml'))" && echo "✅ dependabot-auto-merge.yml syntax is valid"
else
    echo "⚠️ Python3 not available, skipping YAML validation"
fi

# Check Go modules exist
echo "🐹 Checking Go modules..."
if [ -f "apps/default/go.mod" ] && [ -f "apps/integrations/jenga-api/go.mod" ]; then
    echo "✅ Both Go modules found"
else
    echo "❌ Error: Go modules not found"
    exit 1
fi

# Test Go module builds
echo "🔨 Testing Go module builds..."
(cd apps/default && go build ./... && echo "✅ Default app builds successfully")
(cd apps/integrations/jenga-api && go build ./... && echo "✅ Jenga API builds successfully")

# Check for available updates
echo "📦 Checking for available dependency updates..."
echo "Default module updates:"
(cd apps/default && go list -m -u all | grep '\[.*\]' | head -5) || echo "No updates found"
echo "Jenga API module updates:"
(cd apps/integrations/jenga-api && go list -m -u all | grep '\[.*\]' | head -5) || echo "No updates found"

echo ""
echo "🎉 Dependabot setup validation complete!"
echo ""
echo "📋 Summary of configuration:"
echo "- Dependabot monitors 2 Go modules + GitHub Actions"
echo "- Weekly updates for Go dependencies"  
echo "- Auto-merge for minor/patch updates after CI passes"
echo "- Manual review required for major updates"
echo "- Comprehensive CI pipeline tests both modules"
echo ""
echo "🚀 Ready for automated dependency management!"