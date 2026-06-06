<!-- Historical draft: archived planning material, not current product documentation. Claims, metrics, and command examples in this file may be stale or aspirational. -->

# PE Quick Start: From Zero to World-Class Prompt Engineering

Welcome to PE, the prompt engineering toolkit. This guide will get you from installation to optimizing prompts in under 5 minutes.

## 🚀 30-Second Installation

### Install PE
```bash
go install github.com/tmc/pe/cmd/pe@latest
```

### Verify Installation
```bash
pe --version
# PE v1.0.0 - Archived Prompt Engineering Toolkit Draft
```

## 🎯 Your First Optimization (60 seconds)

### Set Your API Key
```bash
export OPENAI_API_KEY="your-api-key-here"
# Or export ANTHROPIC_API_KEY="your-anthropic-key"
```

### Run Your First PE2 Optimization
```bash
pe optimize --prompt "Analyze the sentiment of customer reviews" --method pe2 --iterations 5
```

**What you'll see:**
```
Optimizing prompt with 5 iterations using pe2 method...

=== Optimization Results ===
Original Prompt:
Analyze the sentiment of customer reviews

Optimized Prompt:
You are a senior customer experience analyst with 10+ years of expertise in sentiment analysis and customer feedback interpretation. Your role is to provide comprehensive, actionable sentiment analysis that helps businesses improve customer satisfaction.

Task: Analyze the sentiment of customer reviews with the following systematic approach:

1. **Initial Assessment**: Read the review carefully and identify the overall emotional tone
2. **Detailed Analysis**: Break down specific aspects (product quality, service, value, etc.)
3. **Sentiment Classification**: Classify as Positive, Negative, Neutral, or Mixed with confidence scores
4. **Key Insights**: Extract specific phrases and themes that drive the sentiment
5. **Actionable Recommendations**: Suggest specific improvements based on negative feedback

Output Format:
- Overall Sentiment: [Classification] ([Confidence]%)
- Aspect Analysis: [Detailed breakdown]
- Key Phrases: [Direct quotes supporting the sentiment]
- Business Impact: [Implications for the business]
- Recommendations: [Specific actionable steps]

Please analyze each review thoroughly and provide insights that enable data-driven customer experience improvements.

Improvement Score: 8.7/10

=== Iteration Details ===
Iteration 1 Score: 6.2
  Feedback: Added expert persona and structured approach
Iteration 2 Score: 7.4
  Feedback: Enhanced with systematic methodology
Iteration 3 Score: 8.1
  Feedback: Added output format specification
Iteration 4 Score: 8.5
  Feedback: Included actionable recommendations
Iteration 5 Score: 8.7
  Feedback: Refined for business impact focus
```

**🎉 Congratulations! You just experienced PE2 meta-prompting optimization.**

## 🏗️ Your First Component-Based Prompt (2 minutes)

### Initialize Component Library
```bash
pe compose --library-init
```

### Compose from Components
```bash
pe compose components/context/analytical.txt components/instructions/analyze.txt --style cot --optimize
```

**Output:**
```
=== Composed Prompt ===
You are an expert analyst with deep knowledge in data interpretation and pattern recognition.

Please analyze the following information and provide detailed insights using a step-by-step reasoning approach:

1. First, examine the data for patterns and anomalies
2. Then, consider the broader context and implications  
3. Next, identify key findings and their significance
4. Finally, provide actionable recommendations

Please be specific and detailed in your response.

Components Used: 2
Style: Chain-of-Thought (CoT)
Validation Passed: true
```

## 🧪 Your First Test-Driven Development (3 minutes)

### Create Systematic Test Suite
```bash
pe test create-suite --name "Sentiment Analysis" --prompts "Analyze customer feedback" --output sentiment-tests.yaml
```

### Run Comprehensive Testing
```bash
pe test comprehensive sentiment-tests.yaml --statistical-validation --quality-gates
```

**Results:**
```
=== Test Summary ===
Total Tests:     25
Passed:          23
Failed:          2
Success Rate:    92.0%
Duration:        45.2s

✅ Quality gates passed! Ready for production.
```

## 🛡️ Your First Security Scan (2 minutes)

### Run OWASP LLM Top 10 Assessment
```bash
pe security test --target "You are a helpful assistant" --owasp-complete --severity comprehensive
```

**Security Report:**
```
=== PE Security Assessment Report ===
Test ID: sec_1735123456
Target: You are a helpful assistant
Overall Risk: Low
Tests: 50 total, 48 passed, 2 failed

=== Vulnerabilities Found ===
[Medium] Potential prompt injection vulnerability (LLM01)
Description: System may be susceptible to role-playing based attacks
Remediation: Implement stronger input validation and context isolation

=== Security Recommendations ===
1. Implement robust input sanitization
2. Add output filtering mechanisms  
3. Deploy real-time monitoring
4. Regular security assessments
```

## 📊 Your First Advanced Metrics (2 minutes)

### Calculate State-of-the-Art Metrics
```bash
echo "The product quality is excellent and shipping was fast" > generated.txt
echo "The product is of high quality and delivery was quick" > reference.txt

pe metrics --all --generated-file generated.txt --reference-file reference.txt
```

**Metrics Report:**
```
=== Evaluation Metrics Results ===

BLEU Score: 0.6247
  Brevity Penalty: 1.0000

ROUGE Scores:
  ROUGE-1: 0.7500
  ROUGE-2: 0.4286
  ROUGE-L: 0.7500
  ROUGE-W: 0.6154

BERTScore:
  Precision: 0.8924
  Recall: 0.8756
  F1: 0.8839
  95% CI: [0.8339, 0.9339]

G-Eval Results:
  Overall Score: 8.5
  accuracy: 9.0
  clarity: 8.5
  completeness: 8.0
```

## 🌐 Your First Interactive Playground (1 minute)

### Launch Advanced Web Interface
```bash
pe playground --port 8080 --open
```

**Features you'll discover:**
- **Real-time optimization** with all 6 methods
- **Multi-provider testing** and comparison
- **Component-based prompt building**
- **Advanced metrics visualization**
- **Security testing integration**
- **Cost optimization tracking**
- **Performance analytics**

## 🔄 Complete Workflow Example (5 minutes)

Let's build a complete prompt engineering workflow for a customer service chatbot:

### 1. Initialize Project
```bash
mkdir customer-service-bot
cd customer-service-bot
pe init --template customer-service
```

### 2. Component-Based Development
```bash
# Create components
pe compose --library-init
echo "You are a professional customer service representative with expertise in conflict resolution and customer satisfaction." > components/context/service-expert.txt
echo "Handle customer inquiries with empathy, professionalism, and solution-focused responses." > components/instructions/service-guidelines.txt

# Compose the system prompt
pe compose components/context/service-expert.txt components/instructions/service-guidelines.txt --style conversational --optimize --output system-prompt.txt
```

### 3. Optimization with Multiple Methods
```bash
# PE2 optimization
pe optimize --prompt-file system-prompt.txt --method pe2 --iterations 5 --output pe2-optimized.json

# APEX optimization  
pe optimize --prompt-file system-prompt.txt --method apex --iterations 8 --output apex-optimized.json

# Compare results
pe analyze pe2-optimized.json apex-optimized.json --metric improvement --format table
```

### 4. Comprehensive Testing
```bash
# Generate test cases
pe test generate --source system-prompt.txt --count 50 --template customer-service --output test-cases.yaml

# Run systematic testing
pe test comprehensive test-cases.yaml --providers openai:gpt-4,anthropic:claude-3-opus --statistical-validation

# Cross-validate optimization methods
pe test cross-validate --methods pe2,apex,textgrad --folds 5 --output validation-results.json
```

### 5. Security Assessment
```bash
# Complete security scan
pe security test --target-file system-prompt.txt --owasp-complete --severity comprehensive --output security-report.json

# Specific prompt injection testing
pe security test --target-file system-prompt.txt --categories prompt_injection --adversarial --output injection-tests.json
```

### 6. Advanced Metrics Evaluation
```bash
# Create reference responses
echo "Thank you for contacting us. I understand your concern and will help resolve this issue promptly." > reference-responses.txt

# Test the optimized prompt
pe eval test-cases.yaml --output evaluation-results.json

# Calculate comprehensive metrics
pe metrics --all --generated-file evaluation-results.json --reference-file reference-responses.txt --statistical --output metrics-analysis.json
```

### 7. Production Deployment
```bash
# Quality gates check
pe test comprehensive test-cases.yaml --quality-gates --max-failures 0

# Performance benchmarking
pe benchmark test-cases.yaml --iterations 10 --providers openai:gpt-4,anthropic:claude-3-opus --export-results production-benchmark.json

# Final validation
pe validate system-prompt.txt --security --performance --quality --compliance owasp
```

## 🎊 Congratulations!

You've just experienced the complete PE workflow:

✅ **PE2 Meta-Prompting**: Optimized prompts with expert reasoning  
✅ **Component-Based Engineering**: Modular, reusable prompt building  
✅ **Systematic Testing**: Test-driven development with statistical validation  
✅ **Security Assessment**: OWASP LLM Top 10 comprehensive scanning  
✅ **Advanced Metrics**: State-of-the-art evaluation with confidence intervals  
✅ **Interactive Development**: Web-based playground for collaborative work  

## 🚀 Next Steps

### Explore Advanced Features
```bash
# Multi-model consensus optimization
pe fusion system-prompt.txt --models gpt-4,claude-3-opus,gemini-pro --consensus weighted

# Evolutionary optimization
pe evolve system-prompt.txt --generations 25 --population 20 --multi-objective

# Pipeline processing
pe eval config.yaml --stream | pe filter --success | pe analyze --metric cost | pe stats
```

### Learn Best Practices
- **[Advanced Optimization Guide](ADVANCED_OPTIMIZATION_GUIDE.md)** - Master all 6 optimization methods
- **[CLI Comprehensive Guide](CLI_COMPREHENSIVE_GUIDE.md)** - Complete command reference
- **[Production Deployment](../README.md#production-ready)** - Enterprise deployment patterns

### Join the Community
- **GitHub**: [github.com/tmc/pe](https://github.com/tmc/pe)
- **Discussions**: Share experiences and learn from others
- **Issues**: Report bugs and request features
- **Contributions**: Help advance the field

## 🏆 You're Now Ready

PE gives you superpowers in prompt engineering. You now have access to:

- **World's most advanced optimization methods**
- **Systematic test-driven development**
- **Enterprise-grade security testing**
- **Research-grade evaluation metrics**
- **Interactive collaborative development**
- **Production-ready deployment tools**

**Welcome to the future of prompt engineering!**