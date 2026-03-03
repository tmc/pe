# Troubleshooting Guide

This comprehensive guide helps you diagnose and resolve common issues with PE. From setup problems to advanced optimization challenges, we've got you covered.

## 🚨 Quick Fixes (Most Common Issues)

### Issue: PE command not found
```bash
# Problem: "pe: command not found"
# Solution: Install or update PE
go install github.com/tmc/pe/cmd/pe@latest

# Verify installation
pe --version

# If still not found, check your Go bin path
echo $GOPATH/bin
export PATH=$PATH:$GOPATH/bin
```

### Issue: API key not working
```bash
# Problem: "API key invalid" or "unauthorized"
# Solution: Check API key setup

# For OpenAI
export OPENAI_API_KEY="sk-your-key-here"
echo $OPENAI_API_KEY  # Verify it's set

# For Anthropic  
export ANTHROPIC_API_KEY="your-key-here"
echo $ANTHROPIC_API_KEY  # Verify it's set

# Test API connectivity
pe ask --provider openai:gpt-3.5-turbo "Hello, world!"
```

### Issue: Configuration file errors
```bash
# Problem: YAML/JSON parsing errors
# Solution: Validate configuration

# Check syntax
pe vet config.yaml

# Format and fix common issues
pe fmt config.yaml --write

# Validate against schema
pe validate config.yaml --strict
```

### Issue: Slow performance
```bash
# Problem: Evaluations taking too long
# Solution: Optimize concurrency and caching

# Increase concurrency
pe eval config.yaml --max-concurrency 10

# Enable caching
pe eval config.yaml --cache --cache-ttl 3600

# Use faster models for development
pe eval config.yaml --provider openai:gpt-3.5-turbo
```

---

## 🔍 Diagnostic Tools

### System Health Check
```bash
# Run comprehensive system check
pe doctor

# Expected output:
✓ PE version: v1.0.0
✓ Go version: go1.21.0
✓ API keys configured: OpenAI, Anthropic
✓ Network connectivity: OK
✓ Cache directory: /home/user/.pe/cache
✓ Database: /home/user/.pe/pe.db
⚠ Warning: High API usage (80% of daily limit)
```

### Configuration Validation
```bash
# Validate configuration files
pe vet config.yaml --verbose

# Check for common issues
pe lint config.yaml

# Test configuration without execution
pe eval config.yaml --dry-run
```

### Performance Profiling
```bash
# Profile evaluation performance
pe profile eval config.yaml

# Memory usage analysis
pe profile --memory eval config.yaml

# Network latency analysis
pe profile --network eval config.yaml
```

---

## 📋 Configuration Issues

### YAML Formatting Problems

**Problem**: YAML syntax errors
```yaml
# ❌ Common mistakes
prompts:
- "What is {{variable}" # Missing closing brace
providers:
  - openai:gpt-4       # Missing quotes
tests:
  - vars:
    country: France    # Inconsistent indentation
```

**Solution**: Use proper YAML syntax
```yaml
# ✅ Correct format
prompts:
  - "What is {{variable}}?"
providers:
  - "openai:gpt-4"
tests:
  - vars:
      country: "France"
```

**Quick fix**:
```bash
pe fmt config.yaml --write  # Auto-format
pe vet config.yaml          # Check for issues
```

### Provider Configuration Issues

**Problem**: Provider authentication failures
```yaml
# ❌ Common issues
providers:
  - type: "openai"
    api_key: "${OPENAI_API_KEY}"  # Wrong syntax
  - openai:gpt-4
    temperature: "0.7"           # Wrong type
```

**Solution**: Correct provider configuration
```yaml
# ✅ Correct format
providers:
  - id: "openai-gpt4"
    type: "openai"
    model: "gpt-4"
    config:
      temperature: 0.7
      max_tokens: 500
  - "openai:gpt-4"  # Shorthand format
```

### Variable Template Problems

**Problem**: Template variable errors
```yaml
# ❌ Issues
prompts:
  - "What is {{Country}}?"    # Case mismatch
tests:
  - vars:
      country: "France"       # Different case
```

**Solution**: Consistent variable naming
```yaml
# ✅ Correct
prompts:
  - "What is {{country}}?"
tests:
  - vars:
      country: "France"
```

---

## 🌐 API and Network Issues

### Rate Limiting
**Problem**: Rate limit exceeded errors

**Symptoms**:
```
Error: Rate limit exceeded (429)
Retry-After: 60 seconds
```

**Solution**:
```bash
# Reduce concurrency
pe eval config.yaml --max-concurrency 2

# Add delays between requests
pe eval config.yaml --delay 1s

# Use exponential backoff
pe eval config.yaml --retry-strategy exponential

# Monitor rate limits
pe eval config.yaml --monitor-limits
```

### Network Connectivity
**Problem**: Network timeouts or connection errors

**Diagnosis**:
```bash
# Test connectivity
pe doctor --network

# Check specific provider
pe test-connection --provider openai:gpt-4

# Verify DNS resolution
nslookup api.openai.com
nslookup api.anthropic.com
```

**Solution**:
```bash
# Increase timeout
pe eval config.yaml --timeout 60s

# Use proxy if needed
export HTTP_PROXY="http://proxy.company.com:8080"
export HTTPS_PROXY="http://proxy.company.com:8080"

# Configure retry policy
pe eval config.yaml --retry-attempts 3 --retry-delay 5s
```

### SSL/TLS Issues
**Problem**: Certificate verification errors

**Solution**:
```bash
# Update certificates
sudo apt-get update && sudo apt-get install ca-certificates

# For macOS
brew install ca-certificates

# Temporarily disable SSL verification (not recommended for production)
pe eval config.yaml --insecure-skip-verify
```

---

## 🧠 Optimization Issues

### TextGrad Optimization Problems

**Problem**: TextGrad optimization not converging
```bash
# Symptoms
TextGrad iteration 5/10: No improvement detected
Gradient strength: 0.02 (very low)
```

**Solution**:
```bash
# Increase iteration count
pe optimize --method textgrad --iterations 15

# Adjust optimization parameters
pe optimize --method textgrad --config '{
  "gradient_threshold": 0.01,
  "learning_rate": 0.8,
  "exploration_factor": 0.2
}'

# Try hybrid approach
pe optimize --method hybrid --iterations 10
```

**Advanced debugging**:
```bash
# Analyze gradient quality
pe optimize --method textgrad --debug-gradients

# Export optimization trajectory
pe optimize --method textgrad --save-trajectory trajectory.json

# Visualize optimization progress
pe analyze-optimization trajectory.json --plot
```

### Performance Optimization Issues

**Problem**: Slow evaluation speed

**Diagnosis**:
```bash
# Profile performance
pe profile eval config.yaml

# Common bottlenecks:
# - Too many concurrent requests
# - Large prompt/response sizes
# - Complex assertions
# - Network latency
```

**Solutions**:
```bash
# Optimize concurrency
pe eval config.yaml --max-concurrency 8  # Adjust based on rate limits

# Use caching
pe eval config.yaml --cache --cache-ttl 3600

# Simplify assertions for development
pe eval config-dev.yaml  # Use simpler assertions

# Use faster models for iteration
pe eval config.yaml --provider openai:gpt-3.5-turbo
```

---

## 🔧 Database and Storage Issues

### Database Corruption
**Problem**: Database errors or corruption

**Symptoms**:
```
Error: database disk image is malformed
Error: unable to open database
```

**Solution**:
```bash
# Backup existing database
cp ~/.pe/pe.db ~/.pe/pe.db.backup

# Repair database
pe repair-db

# If repair fails, reset database
pe reset-db --confirm

# Restore from backup if needed
pe restore-db ~/.pe/pe.db.backup
```

### Storage Space Issues
**Problem**: Disk space errors

**Diagnosis**:
```bash
# Check PE storage usage
pe storage-info

# Output:
Cache directory: ~/.pe/cache (2.3 GB)
Database: ~/.pe/pe.db (450 MB)
Logs: ~/.pe/logs (120 MB)
Total: 2.87 GB
```

**Solution**:
```bash
# Clean cache
pe clean-cache --older-than 7d

# Clean old evaluations
pe clean-evaluations --keep-last 50

# Compact database
pe vacuum-db

# Configure storage limits
pe config set cache.max_size 1GB
pe config set db.max_size 500MB
```

---

## 🐛 Debugging Advanced Issues

### Memory Issues
**Problem**: Out of memory errors during large evaluations

**Diagnosis**:
```bash
# Monitor memory usage
pe profile --memory eval large-config.yaml

# Check system resources
free -h
top -p $(pgrep pe)
```

**Solution**:
```bash
# Reduce batch size
pe eval config.yaml --batch-size 10

# Enable streaming mode
pe eval config.yaml --stream

# Increase system memory or use swap
# Process in smaller chunks
pe eval config.yaml --chunk-size 100
```

### Assertion Failures
**Problem**: Unexpected assertion failures

**Debugging**:
```bash
# Run with detailed output
pe eval config.yaml --verbose --debug-assertions

# Test individual assertions
pe test-assertion --type "contains" --value "Paris" --text "The capital is Paris"

# Analyze assertion patterns
pe analyze-failures config.yaml --group-by assertion_type
```

### Complex Configuration Issues
**Problem**: Complex multi-file configurations not working

**Solution**:
```bash
# Validate entire configuration tree
pe vet config.yaml --recursive

# Test configuration loading
pe config-test config.yaml --trace-loading

# Debug variable resolution
pe debug-vars config.yaml --show-resolution
```

---

## 📊 Monitoring and Logging

### Enable Debug Logging
```bash
# Enable verbose logging
export PE_LOG_LEVEL=debug
pe eval config.yaml

# Log to file
pe eval config.yaml --log-file evaluation.log

# Real-time log monitoring
tail -f ~/.pe/logs/pe.log
```

### Performance Monitoring
```bash
# Enable performance metrics
pe eval config.yaml --metrics --metrics-interval 10s

# Export metrics to external system
pe eval config.yaml --export-metrics prometheus.txt

# Real-time dashboard
pe dashboard --port 8080
```

---

## 🚑 Emergency Procedures

### Complete Reset
**When everything is broken**:

```bash
# 1. Backup important data
pe backup --output pe-backup.tar.gz

# 2. Reset configuration
pe reset-config --confirm

# 3. Reset database
pe reset-db --confirm

# 4. Clear cache
pe clean-cache --all

# 5. Reinstall PE
go install github.com/tmc/pe/cmd/pe@latest

# 6. Verify installation
pe doctor
```

### Recovery from Backup
```bash
# Restore from backup
pe restore --input pe-backup.tar.gz

# Verify restoration
pe verify-restore
```

---

## 📞 Getting Help

### Community Support
- **GitHub Issues**: [Report bugs and request features](https://github.com/tmc/pe/issues)
- **Discussions**: [Ask questions and share tips](https://github.com/tmc/pe/discussions)
- **Discord**: [Real-time community chat](https://discord.gg/pe-toolkit)

### Professional Support
- **Enterprise Support**: Available for commercial users
- **Consulting**: Custom implementation and optimization services
- **Training**: Professional workshops and certification programs

### Diagnostic Information for Support

When reporting issues, include:

```bash
# Generate diagnostic report
pe diagnose --output diagnostic-report.json

# Include PE version
pe --version

# Include system information
pe system-info

# Include recent logs
pe logs --last 100 --level error
```

### Before Contacting Support

1. **Check this troubleshooting guide**
2. **Search existing GitHub issues**
3. **Try the diagnostic tools**
4. **Prepare a minimal reproduction case**
5. **Gather diagnostic information**

---

## 🔄 Version-Specific Issues

### Upgrading from v0.x to v1.x
```bash
# Check compatibility
pe migrate-check

# Backup before upgrade
pe backup --output pre-upgrade-backup.tar.gz

# Upgrade configurations
pe migrate-config --from v0.x --to v1.x

# Test after upgrade
pe test-migration
```

### Beta/Experimental Features
```bash
# Enable experimental features
pe config set experimental.enabled true

# Inspect available experimental/prototype commands
pe experimental --help
pe exp --help
```

---

**📝 Note**: This troubleshooting guide is continuously updated. For the latest version, check the [online documentation](https://github.com/tmc/pe/docs/TROUBLESHOOTING.md).

**🤝 Contributing**: Help improve this guide by [submitting issues](https://github.com/tmc/pe/issues) or [pull requests](https://github.com/tmc/pe/pulls) with additional troubleshooting scenarios.
