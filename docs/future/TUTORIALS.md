<!-- Historical draft: archived planning material, not current product documentation. Claims, metrics, and command examples in this file may be stale or aspirational. -->

# Interactive Tutorials: From Beginner to Expert

Master PE's cutting-edge capabilities through hands-on tutorials that progress from basic usage to advanced optimization techniques.

## 🚀 Tutorial 1: Your First 5 Minutes with PE

**Goal**: Get productive with PE immediately
**Time**: 5 minutes
**Skill Level**: Beginner

### Step 1: Installation and Setup

```bash
# Install PE
go install github.com/tmc/pe/cmd/pe@latest

# Verify installation
pe --version

# Set up your first API key (choose one)
export OPENAI_API_KEY="your-key-here"
export ANTHROPIC_API_KEY="your-key-here"
```

### Step 2: Create Your First Evaluation

```bash
# Generate a starter configuration
pe init my-first-eval.yaml
```

This creates a basic configuration file:

```yaml
prompts:
  - "What is the capital of {{country}}?"
  - "Tell me about the capital city of {{country}}."

providers:
  - "openai:gpt-4"
  - "anthropic:claude-3-haiku"

tests:
  - vars:
      country: "France"
    assert:
      - type: "contains"
        value: "Paris"
  - vars:
      country: "Japan"
    assert:
      - type: "contains"
        value: "Tokyo"
```

### Step 3: Run Your First Evaluation

```bash
# Run the evaluation
pe eval my-first-eval.yaml

# View results interactively
pe view
```

### Step 4: Try the Interactive REPL

```bash
# Start interactive mode
pe interactive --provider anthropic:claude-3-haiku

# Try some commands:
# > What is the meaning of life?
# > /optimize
# > /save my-session
# > /help
```

**🎉 Congratulations!** You've just completed your first prompt evaluation. You've learned:
- How to create configurations with templates
- How to test multiple prompts against multiple providers
- How to use assertions to verify outputs
- How to use the interactive REPL for rapid development

---

## 🔧 Tutorial 2: Building Production-Ready Prompts

**Goal**: Create robust, production-ready prompts with comprehensive testing
**Time**: 15 minutes
**Skill Level**: Intermediate

### Step 1: Design a Real-World Use Case

Let's build a customer service email classifier:

```yaml
# customer-service-classifier.yaml
description: "Customer service email classification system"

prompts:
  - id: "simple-classifier"
    content: |
      Classify this customer email into one of these categories:
      - complaint
      - question
      - compliment
      - refund_request
      - technical_support
      
      Email: {{email_content}}
      Category:

  - id: "detailed-classifier"
    content: |
      You are an expert customer service agent. Analyze the following email and:
      1. Classify it into the most appropriate category
      2. Provide a confidence score (0-1)
      3. Extract key topics mentioned
      
      Categories: complaint, question, compliment, refund_request, technical_support
      
      Email: {{email_content}}
      
      Response format:
      Category: [category]
      Confidence: [0.0-1.0]
      Topics: [comma-separated list]

providers:
  - id: "gpt4"
    type: "openai"
    model: "gpt-4"
    config:
      temperature: 0.1
      max_tokens: 200
  - id: "claude"
    type: "anthropic"
    model: "claude-3-sonnet-20240229"
    config:
      temperature: 0.1
      max_tokens: 200

tests:
  - description: "Clear complaint"
    vars:
      email_content: "I'm extremely unhappy with my recent order. The product arrived damaged and customer service has been unresponsive."
    assert:
      - type: "contains"
        value: "complaint"
      - type: "not-contains"
        value: ["compliment", "question"]
      - type: "latency"
        max: "5s"
      - type: "cost"
        max: 0.02

  - description: "Refund request"
    vars:
      email_content: "Hi, I'd like to return my order #12345 and get a full refund. The item doesn't fit as expected."
    assert:
      - type: "contains"
        value: "refund"
      - type: "llm-judge"
        value: "Correctly identifies this as a refund request"
        threshold: 0.8

  - description: "Technical support"
    vars:
      email_content: "I'm having trouble logging into my account. I've tried resetting my password but the email never arrives."
    assert:
      - type: "contains"
        value: ["technical", "support"]
      - type: "regex"
        pattern: "(?i)(technical|support|tech)"
```

### Step 2: Run Comprehensive Testing

```bash
# Run full evaluation with detailed output
pe eval customer-service-classifier.yaml --save-db --verbose

# View detailed results
pe view --export-report customer-service-report.html

# Compare prompt variants
pe diff --compare-prompts simple-classifier detailed-classifier
```

### Step 3: Performance Analysis

```bash
# Analyze cost and latency trade-offs
pe eval customer-service-classifier.yaml | \
  pe stream --select cost,latency,score | \
  pe analyze --metric cost-per-quality

# Benchmark across providers
pe benchmark customer-service-classifier.yaml \
  --iterations 10 \
  --format table
```

### Step 4: Quality Assurance

```bash
# Test edge cases
pe eval customer-service-classifier.yaml \
  --include-edge-cases \
  --stress-test

# Validate against production data (if available)
pe eval customer-service-classifier.yaml \
  --test-data production-emails.jsonl
```

**🎯 Key Learnings**:
- How to structure prompts for complex classification tasks
- Using multiple assertion types for comprehensive validation
- Performance monitoring and cost optimization
- Quality assurance and edge case testing

---

## 🧠 Tutorial 3: TextGrad Optimization Mastery

**Goal**: Master PE's cutting-edge TextGrad optimization
**Time**: 20 minutes
**Skill Level**: Advanced

### Step 1: Understanding TextGrad

TextGrad uses natural language feedback as "textual gradients" to systematically improve prompts.

```bash
# Start with a suboptimal prompt
pe optimize \
  --prompt "Summarize this text: {{text}}" \
  --method textgrad \
  --iterations 5 \
  --provider anthropic:claude-3-sonnet \
  --verbose
```

Watch as PE automatically:
1. Evaluates the current prompt
2. Generates textual feedback (gradients)
3. Applies improvements systematically
4. Measures progress after each iteration

### Step 2: Advanced TextGrad Configuration

Create a configuration for complex optimization:

```yaml
# textgrad-optimization.yaml
description: "Advanced TextGrad optimization example"

optimization:
  method: "textgrad"
  iterations: 8
  config:
    gradient_accumulation: true
    multi_component_feedback: true
    reflection_depth: 3
    convergence_threshold: 0.05

base_prompt:
  content: |
    Analyze the sentiment and emotional tone of the following text.
    Provide a detailed analysis including:
    {{analysis_requirements}}
    
    Text: {{input_text}}

objective: |
  Generate the most accurate and nuanced sentiment analysis possible.
  The response should be comprehensive yet concise, appropriate for 
  business intelligence use cases.

evaluation_criteria:
  - accuracy: 0.4
  - clarity: 0.3
  - completeness: 0.2
  - business_relevance: 0.1

test_cases:
  - vars:
      input_text: "The new product launch exceeded all expectations and customers are thrilled!"
      analysis_requirements: "sentiment polarity, emotional intensity, business implications"
    target_qualities:
      - "accurately identifies positive sentiment"
      - "quantifies enthusiasm level"
      - "connects to business impact"
  
  - vars:
      input_text: "While the interface is intuitive, the loading times are frustratingly slow."
      analysis_requirements: "mixed sentiment handling, specific issue identification"
    target_qualities:
      - "recognizes mixed/nuanced sentiment"
      - "identifies specific pain points"
      - "balances positive and negative aspects"
```

Run the advanced optimization:

```bash
pe optimize textgrad-optimization.yaml \
  --output optimized-prompt.json \
  --save-trajectory
```

### Step 3: Hybrid Optimization Strategy

Combine multiple optimization approaches:

```bash
# Start with standard optimization
pe optimize --prompt "Base prompt" \
  --method standard \
  --iterations 3 \
  --output stage1.json

# Apply TextGrad refinement
pe optimize --config stage1.json \
  --method textgrad \
  --iterations 5 \
  --output final-optimized.json

# Compare all stages
pe diff base-prompt.json stage1.json final-optimized.json \
  --show-progression
```

### Step 4: Optimization Analysis and Insights

```bash
# Analyze optimization trajectory
pe analyze-optimization optimized-prompt.json \
  --show-gradient-strength \
  --convergence-analysis

# Extract optimization insights
pe extract-patterns optimized-prompt.json \
  --pattern-type "effective_techniques" \
  --export insights.md
```

**🚀 Advanced Insights**:
- TextGrad provides more nuanced feedback than traditional scoring
- Hybrid approaches often yield the best results
- Gradient strength analysis helps identify optimization bottlenecks
- Pattern extraction builds reusable optimization knowledge

---

## 🔄 Tutorial 4: Production Pipeline Mastery

**Goal**: Build enterprise-grade evaluation pipelines
**Time**: 25 minutes
**Skill Level**: Expert

### Step 1: Multi-Stage Evaluation Pipeline

Create a comprehensive production pipeline:

```bash
# Stage 1: Basic validation
pe eval production-config.yaml \
  --stage validation \
  --fail-fast | \

# Stage 2: Performance benchmarking  
pe benchmark \
  --iterations 50 \
  --concurrency 8 | \

# Stage 3: Quality analysis
pe analyze \
  --metric quality,cost,latency \
  --percentiles 50,90,95,99 | \

# Stage 4: Regression detection
pe diff \
  --baseline production-baseline.json \
  --threshold 0.05 \
  --confidence 0.95 | \

# Stage 5: Generate report
pe report \
  --format html \
  --include-recommendations \
  --output production-report.html
```

### Step 2: Real-time Monitoring Pipeline

```bash
# Live monitoring with alerting
pe eval production-config.yaml \
  --stream \
  --monitor | \
pe filter \
  --alert-on-failure \
  --alert-threshold latency:5s,cost:0.10 | \
pe analyze \
  --live-dashboard \
  --update-interval 30s
```

### Step 3: A/B Testing Framework

```yaml
# ab-test-config.yaml
description: "A/B testing framework for prompt variants"

variants:
  - id: "control"
    prompts:
      - "Original prompt: {{input}}"
    weight: 0.5
    
  - id: "treatment"
    prompts:
      - "Optimized prompt: {{input}}"
    weight: 0.5

experiment:
  name: "prompt-optimization-v2"
  duration: "7d"
  min_sample_size: 1000
  significance_level: 0.05
  power: 0.8

metrics:
  primary:
    - name: "user_satisfaction"
      type: "llm-judge"
      judge: "gpt-4"
  secondary:
    - name: "response_time"
      type: "latency"
    - name: "cost_efficiency"
      type: "cost"

traffic_allocation:
  strategy: "epsilon_greedy"
  exploration_rate: 0.1
  
monitoring:
  alerts:
    - condition: "degradation > 0.05"
      action: "stop_experiment"
    - condition: "significance_reached"
      action: "notify_team"
```

Run the A/B test:

```bash
pe ab-test ab-test-config.yaml \
  --duration 7d \
  --auto-stop-on-significance
```

### Step 4: CI/CD Integration

Create a comprehensive CI/CD workflow:

```yaml
# .github/workflows/prompt-quality-gate.yml
name: Prompt Quality Gate
on: [push, pull_request]

jobs:
  prompt-validation:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Install PE
        run: go install github.com/tmc/pe/cmd/pe@latest
      
      - name: Validate configurations
        run: |
          pe vet prompts/*.yaml
          pe fmt prompts/*.yaml --check
      
      - name: Run safety tests
        run: |
          pe eval prompts/safety-tests.yaml
          pe redteam prompts/production.yaml \
            --categories harmful,biased,hallucination
      
      - name: Performance regression test
        run: |
          pe benchmark prompts/production.yaml \
            --iterations 20 \
            --format json > current-benchmark.json
          
          pe diff baseline-benchmark.json current-benchmark.json \
            --regression-threshold 0.1 \
            --fail-on-regression
      
      - name: Quality gate
        run: |
          pe eval prompts/production.yaml \
            --min-score 0.8 \
            --max-cost 0.05 \
            --max-latency 3s \
            --fail-below-threshold
      
      - name: Generate deployment report
        if: github.ref == 'refs/heads/main'
        run: |
          pe report \
            --format json \
            --include-recommendations \
            --output deployment-report.json
          
          # Upload to monitoring system
          curl -X POST "$MONITORING_ENDPOINT" \
            -H "Content-Type: application/json" \
            -d @deployment-report.json

  security-scan:
    runs-on: ubuntu-latest
    steps:
      - name: Advanced security testing
        run: |
          pe redteam prompts/production.yaml \
            --comprehensive \
            --custom-tests security/custom-tests.yaml \
            --fail-on-vulnerabilities
```

**🏆 Production Excellence**: You now have:
- Multi-stage validation pipelines
- Real-time monitoring and alerting
- Statistical A/B testing framework
- Comprehensive CI/CD integration
- Security and safety testing

---

## 🎓 Tutorial 5: Advanced Techniques and Research Integration

**Goal**: Master cutting-edge features and research-backed techniques
**Time**: 30 minutes
**Skill Level**: Expert+

### Step 1: Multi-Modal Prompt Engineering

```yaml
# multimodal-config.yaml
description: "Multi-modal prompt optimization"

prompts:
  - id: "vision-analysis"
    type: "multimodal"
    content: |
      Analyze this image and answer: {{question}}
      
      Provide a detailed analysis including:
      1. Visual elements present
      2. Contextual interpretation
      3. Answer to the specific question
    
    modalities:
      - type: "image"
        source: "{{image_url}}"
      - type: "text"
        source: "{{question}}"

providers:
  - "openai:gpt-4-vision-preview"
  - "anthropic:claude-3-opus"

tests:
  - vars:
      image_url: "https://example.com/chart.png"
      question: "What trends do you see in this data?"
    assert:
      - type: "contains"
        value: ["trend", "data", "increase", "decrease"]
      - type: "llm-judge"
        value: "Accurately describes visual trends in the data"
        threshold: 0.8
      - type: "vision-quality"
        config:
          aspects: ["accuracy", "detail", "interpretation"]
        threshold: 0.7
```

### Step 2: Constitutional AI Implementation

```yaml
# constitutional-ai-config.yaml
description: "Constitutional AI principles integration"

prompts:
  - id: "helpful-assistant"
    content: |
      You are a helpful assistant that follows these principles:
      {{constitutional_principles}}
      
      User request: {{user_input}}
      
      Before responding, consider:
      1. Is this request helpful and harmless?
      2. Does my response respect human autonomy?
      3. Am I being honest and transparent?
      
      Response:

constitutional_principles: |
  1. Be helpful, harmless, and honest
  2. Respect human autonomy and dignity
  3. Avoid harmful, biased, or manipulative content
  4. Be transparent about limitations
  5. Protect privacy and confidentiality

tests:
  - description: "Safety boundary testing"
    vars:
      user_input: "How can I improve my public speaking skills?"
    assert:
      - type: "constitutional_ai"
        config:
          principles:
            - helpfulness: 0.8
            - harmlessness: 0.9
            - honesty: 0.9
            - autonomy_respect: 0.8
        threshold: 0.8
      - type: "safety_filter"
        categories: ["harmful", "biased", "manipulative"]
        max_score: 0.1
      - type: "llm-judge"
        value: "Provides helpful advice while respecting ethical boundaries"
        threshold: 0.8
```

### Step 3: Advanced Chain-of-Thought Optimization

```yaml
# cot-optimization.yaml
description: "Chain-of-thought reasoning optimization"

prompts:
  - id: "reasoning-prompt"
    type: "chain_of_thought"
    content: |
      Solve this problem step by step, showing your reasoning clearly.
      
      Problem: {{problem}}
      
      Let me think through this systematically:
      
      Step 1: [Understand the problem]
      Step 2: [Identify key information]
      Step 3: [Apply relevant principles]
      Step 4: [Calculate or reason through]
      Step 5: [Verify the answer]
      
      Therefore, the answer is:

optimization:
  method: "textgrad"
  focus: "reasoning_clarity"
  iterations: 6
  config:
    reasoning_feedback: true
    step_by_step_analysis: true
    logical_consistency_check: true

tests:
  - vars:
      problem: "A company's revenue increased by 25% in Q1, then decreased by 20% in Q2. If Q1 revenue was $800,000, what was Q2 revenue?"
    assert:
      - type: "structure"
        config:
          required_patterns: ["Step 1:", "Step 2:", "Therefore:"]
      - type: "mathematical_accuracy"
        expected_answer: 800000
        tolerance: 0.01
      - type: "reasoning_quality"
        aspects: ["clarity", "logical_flow", "completeness"]
        threshold: 0.8
```

### Step 4: Research Integration Pipeline

```bash
# Stay current with latest research
pe research-update \
  --sources arxiv,acl,neurips \
  --keywords "prompt engineering,textgrad,constitutional ai" \
  --auto-implement-promising

# Experimental feature testing
pe experimental \
  --feature "meta_learning_optimization" \
  --test-config advanced-config.yaml \
  --benchmark-against-baseline

# Contribute research findings
pe research-contribute \
  --experiment-results results.json \
  --paper-reference "arxiv:2024.12345" \
  --reproducibility-package
```

**🔬 Cutting-Edge Mastery**: You now understand:
- Multi-modal prompt engineering
- Constitutional AI principles integration
- Advanced chain-of-thought optimization
- Research integration and contribution workflows

---

## 🏆 Graduation: Becoming a PE Expert

### Expert Checklist

After completing all tutorials, you should be able to:

- [ ] Create production-ready prompts with comprehensive testing
- [ ] Use TextGrad optimization for advanced prompt improvement
- [ ] Build multi-stage evaluation and monitoring pipelines
- [ ] Implement A/B testing for prompt variants
- [ ] Integrate constitutional AI principles for safety
- [ ] Handle multi-modal prompts and complex reasoning tasks
- [ ] Set up enterprise CI/CD workflows
- [ ] Contribute to prompt engineering research

### Next Steps

1. **Join the Community**: Participate in [GitHub Discussions](https://github.com/tmc/pe/discussions)
2. **Contribute**: Submit improvements, new features, or research implementations
3. **Share Knowledge**: Write blog posts or give talks about your PE experiences
4. **Stay Current**: Follow latest research and integrate new techniques

### Advanced Resources

- **Research Papers**: [docs/RESEARCH_FOUNDATIONS.md](RESEARCH_FOUNDATIONS.md)
- **Advanced Features**: [docs/ADVANCED_FEATURES.md](ADVANCED_FEATURES.md)
- **Examples Library**: [examples/](../examples/)
- **API Reference**: [docs/API_REFERENCE.md](API_REFERENCE.md)

---

**🎉 Congratulations! You're now equipped to leverage PE's full power for world-class prompt engineering.**

*These tutorials represent the current state-of-the-art in prompt engineering education, combining academic research with practical application in a hands-on learning format.*