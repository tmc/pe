#!/bin/bash
# Mock script to simulate the module registry for testing

# This would normally:
# 1. Create a root gist with registry information
# 2. Fork gists for each module
# 3. Update the registry when modules are pushed

echo "Setting up mock registry..."

# Create a mock root gist ID
export PE_ROOT_GIST_ID="mock-root-gist-12345"

# Create mock responses directory
mkdir -p .pe/test/mocks

# Create mock gist response for push
cat > .pe/test/mocks/gist-create.json << 'EOF'
{
  "id": "mock-gist-id",
  "html_url": "https://gist.github.com/mock-gist-id",
  "description": "Prompt module: hello",
  "public": false,
  "owner": {
    "login": "tmc"
  },
  "files": {
    "module.json": {
      "filename": "module.json",
      "content": "{\"name\":\"tmc/hello\"}"
    },
    "prompt.txt": {
      "filename": "prompt.txt", 
      "content": "say hello in a random language"
    }
  }
}
EOF

echo "Mock registry ready"