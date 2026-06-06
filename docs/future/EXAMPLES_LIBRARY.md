<!-- Historical draft: archived planning material, not current product documentation. Claims, metrics, and command examples in this file may be stale or aspirational. -->

# Interactive Examples Library

A comprehensive collection of real-world examples demonstrating PE's capabilities across different industries and use cases. Each example includes complete configurations, explanations, and optimization strategies.

## 🎯 Quick Start Examples

### Example 1: Simple Q&A System
**Use Case**: Basic question-answering with accuracy validation

```yaml
# examples/qa-system.yaml
description: "Simple Q&A system with accuracy testing"

prompts:
  - "Answer this question concisely: {{question}}"
  - |
    You are a helpful assistant. Answer the following question 
    accurately and concisely:
    
    Question: {{question}}
    Answer:

providers:
  - "openai:gpt-3.5-turbo"
  - "anthropic:claude-3-haiku"

tests:
  - description: "Geography question"
    vars:
      question: "What is the capital of Japan?"
    assert:
      - type: "contains"
        value: "Tokyo"
      - type: "length"
        max: 100
      - type: "latency"
        max: "3s"

  - description: "Math question"
    vars:
      question: "What is 15 * 8?"
    assert:
      - type: "contains"
        value: "120"
      - type: "not-contains"
        value: ["approximately", "about"]
```

**Run it**:
```bash
pe eval examples/qa-system.yaml
pe view  # See interactive results
```

---

## 🏢 Business Applications

### Example 2: Customer Service Email Classification
**Use Case**: Automatically categorize customer emails

```yaml
# examples/email-classifier.yaml
description: "Customer service email classification system"

prompts:
  - id: "basic-classifier"
    content: |
      Classify this customer email into one category:
      Categories: complaint, question, compliment, refund_request, technical_support, billing
      
      Email: {{email_content}}
      Category:

  - id: "detailed-classifier"
    content: |
      You are an expert customer service agent. Analyze this email and provide:
      
      1. Primary category: complaint, question, compliment, refund_request, technical_support, billing
      2. Urgency level: low, medium, high, critical
      3. Key topics mentioned
      4. Recommended action
      
      Email: {{email_content}}
      
      Analysis:
      Category: 
      Urgency: 
      Topics: 
      Recommended action:

providers:
  - id: "fast"
    type: "openai"
    model: "gpt-3.5-turbo"
    config:
      temperature: 0.1
      max_tokens: 200
  - id: "accurate"
    type: "openai"
    model: "gpt-4"
    config:
      temperature: 0.1
      max_tokens: 300

tests:
  - description: "Billing complaint"
    vars:
      email_content: |
        Subject: Charged twice for the same order
        
        Hi, I've been charged twice for order #12345. My credit card 
        shows two charges of $99.99 on the same day. Please fix this 
        immediately and refund the duplicate charge.
    assert:
      - type: "contains"
        value: ["billing", "complaint"]
      - type: "llm-judge"
        value: "Correctly identifies this as a billing complaint with high urgency"
        threshold: 0.8
      - type: "cost"
        max: 0.02

  - description: "Technical support question"
    vars:
      email_content: |
        Subject: Login issues
        
        I'm having trouble logging into my account. I've tried 
        resetting my password but the reset email never arrives. 
        Can you help?
    assert:
      - type: "contains"
        value: ["technical_support", "technical"]
      - type: "regex"
        pattern: "(?i)(technical|support|tech)"
      - type: "latency"
        max: "5s"

optimization:
  enabled: true
  method: "textgrad"
  iterations: 3
  target_metric: "accuracy"
```

**Advanced usage**:
```bash
# Run with optimization
pe eval examples/email-classifier.yaml --optimize

# Benchmark different models
pe benchmark examples/email-classifier.yaml --compare-providers

# Stream processing for real-time classification
echo "Customer email content" | pe ask --config examples/email-classifier.yaml
```

### Example 3: Content Generation with Quality Control
**Use Case**: Generate marketing copy with quality validation

```yaml
# examples/content-generation.yaml
description: "Marketing content generation with quality control"

prompts:
  - id: "product-description"
    content: |
      Write a compelling product description for:
      
      Product: {{product_name}}
      Category: {{category}}
      Key features: {{features}}
      Target audience: {{audience}}
      Tone: {{tone}}
      Length: {{length}} words
      
      The description should:
      1. Highlight key benefits
      2. Address the target audience directly
      3. Include a call-to-action
      4. Be SEO-friendly
      
      Product Description:

  - id: "blog-post"
    content: |
      Write a {{length}}-word blog post about {{topic}} for {{audience}}.
      
      Requirements:
      - Engaging introduction
      - 3-4 main points with examples
      - Conclusion with takeaways
      - Tone: {{tone}}
      - Include relevant keywords: {{keywords}}
      
      Blog Post:

providers:
  - "openai:gpt-4"
  - "anthropic:claude-3-sonnet"

tests:
  - description: "SaaS product description"
    vars:
      product_name: "CloudSync Pro"
      category: "File synchronization software"
      features: "Real-time sync, 256-bit encryption, team collaboration"
      audience: "Small business owners"
      tone: "professional yet approachable"
      length: 150
    assert:
      - type: "length"
        min: 120
        max: 180
      - type: "contains"
        value: ["CloudSync Pro", "sync", "business"]
      - type: "readability"
        min_grade_level: 8
        max_grade_level: 12
      - type: "sentiment"
        value: "positive"
      - type: "llm-judge"
        value: |
          Rate this product description (1-5) on:
          1. Clarity and appeal (25%)
          2. Feature highlighting (25%)
          3. Audience targeting (25%)
          4. Call-to-action presence (25%)
        threshold: 0.8

  - description: "Technical blog post"
    vars:
      topic: "API security best practices"
      audience: "Software developers"
      tone: "technical but accessible"
      length: 800
      keywords: "API security, authentication, encryption, best practices"
    assert:
      - type: "word_count"
        min: 700
        max: 900
      - type: "contains"
        value: ["API", "security", "authentication"]
      - type: "structure"
        config:
          required_sections: ["introduction", "conclusion"]
      - type: "technical_accuracy"
        domain: "cybersecurity"
        threshold: 0.8
      - type: "seo_score"
        keywords: ["API security", "authentication", "encryption"]
        threshold: 0.7

metrics:
  - name: "engagement_score"
    type: "llm-graded"
    judge: "gpt-4"
    criteria: "Rate how engaging and compelling this content is (1-10)"
  - name: "brand_alignment"
    type: "custom"
    script: "./brand-alignment-check.py"
```

**Advanced workflow**:
```bash
# Generate and optimize content
pe eval examples/content-generation.yaml --save-db
pe optimize --config examples/content-generation.yaml --method hybrid

# Quality assurance pipeline
pe eval examples/content-generation.yaml | \
  pe filter --min-score 0.8 | \
  pe analyze --metric engagement_score
```

---

## 🤖 AI/ML Applications

### Example 4: Model Output Comparison
**Use Case**: Compare different AI models for specific tasks

```yaml
# examples/model-comparison.yaml
description: "Comprehensive model comparison across multiple dimensions"

prompts:
  - |
    Summarize this text in exactly 2 sentences:
    
    {{text}}

providers:
  - id: "gpt4"
    type: "openai"
    model: "gpt-4"
    config:
      temperature: 0.1
  - id: "gpt35"
    type: "openai"
    model: "gpt-3.5-turbo"
    config:
      temperature: 0.1
  - id: "claude"
    type: "anthropic"
    model: "claude-3-sonnet-20240229"
    config:
      temperature: 0.1

tests:
  - description: "Technical article summary"
    vars:
      text: |
        Artificial intelligence has evolved significantly over the past decade, 
        with transformer architectures revolutionizing natural language processing. 
        Large language models like GPT-4 and Claude demonstrate remarkable 
        capabilities in understanding context, generating coherent text, and 
        performing complex reasoning tasks. However, these models also present 
        challenges including computational costs, potential biases, and the 
        need for responsible deployment practices. As AI systems become more 
        sophisticated, the importance of alignment with human values and 
        robust evaluation frameworks becomes increasingly critical.
    assert:
      - type: "sentence_count"
        value: 2
      - type: "length"
        min: 50
        max: 200
      - type: "contains"
        value: ["AI", "language models"]
      - type: "coherence"
        threshold: 0.8
      - type: "factuality"
        threshold: 0.9

analysis:
  metrics:
    - "accuracy"
    - "cost_per_token"
    - "latency"
    - "coherence"
  
  comparison_dimensions:
    - dimension: "cost_effectiveness"
      formula: "score / cost"
    - dimension: "speed_quality_ratio"
      formula: "score / latency"
```

**Run comparison**:
```bash
# Comprehensive model comparison
pe benchmark examples/model-comparison.yaml --iterations 20
pe diff --compare-providers --metric cost_effectiveness
```

### Example 5: Code Generation and Validation
**Use Case**: Generate and validate code snippets

```yaml
# examples/code-generation.yaml
description: "Code generation with comprehensive validation"

prompts:
  - id: "function-generator"
    content: |
      Generate a {{language}} function that {{task}}.
      
      Requirements:
      - Function name: {{function_name}}
      - Include proper error handling
      - Add comprehensive docstring/comments
      - Follow {{language}} best practices
      - Include type hints (if applicable)
      
      Code:

  - id: "class-generator"
    content: |
      Create a {{language}} class called {{class_name}} that {{description}}.
      
      Requirements:
      - Proper initialization method
      - Include all necessary methods
      - Add docstrings for all methods
      - Follow {{language}} naming conventions
      - Include example usage
      
      Class definition:

providers:
  - "openai:gpt-4"
  - "anthropic:claude-3-opus"

tests:
  - description: "Python factorial function"
    vars:
      language: "Python"
      task: "calculates the factorial of a number"
      function_name: "factorial"
    assert:
      - type: "contains"
        value: ["def factorial", "return"]
      - type: "code_syntax"
        language: "python"
      - type: "code_security"
        rules: ["no-eval", "no-exec", "no-dangerous-imports"]
      - type: "code_quality"
        aspects: ["readability", "efficiency", "maintainability"]
        threshold: 0.8
      - type: "llm-judge"
        value: |
          Evaluate this Python function:
          1. Correctness (40%)
          2. Error handling (20%)
          3. Documentation quality (20%)
          4. Code style (20%)
        threshold: 0.8

  - description: "JavaScript data validation class"
    vars:
      language: "JavaScript"
      class_name: "DataValidator"
      description: "validates user input data with customizable rules"
    assert:
      - type: "contains"
        value: ["class DataValidator", "constructor", "validate"]
      - type: "code_syntax"
        language: "javascript"
      - type: "code_patterns"
        required: ["error_handling", "method_documentation"]
      - type: "security_scan"
        rules: ["no-eval", "no-function-constructor", "no-dangerous-regex"]

validation:
  execution_tests:
    enabled: true
    timeout: "10s"
    test_cases:
      - input: 5
        expected_pattern: "120"
      - input: 0
        expected_pattern: "1"
```

**Code validation workflow**:
```bash
# Generate and validate code
pe eval examples/code-generation.yaml --execute-tests

# Security scanning
pe eval examples/code-generation.yaml --security-scan --strict

# Performance testing
pe benchmark examples/code-generation.yaml --metric code_quality
```

---

## 📚 Educational Applications

### Example 6: Educational Content Creation
**Use Case**: Create educational materials with learning objectives

```yaml
# examples/educational-content.yaml
description: "Educational content creation with learning assessment"

prompts:
  - id: "lesson-explanation"
    content: |
      Create a clear explanation of {{concept}} for {{grade_level}} students.
      
      Structure:
      1. Simple definition
      2. Real-world examples (2-3)
      3. Key points to remember
      4. Common misconceptions to avoid
      
      Learning objectives:
      - Students will understand {{learning_objective_1}}
      - Students will be able to {{learning_objective_2}}
      
      Style: {{teaching_style}}
      Length: {{length}} words
      
      Explanation:

  - id: "quiz-generator"
    content: |
      Generate {{num_questions}} multiple choice questions about {{topic}} 
      for {{grade_level}} students.
      
      Requirements:
      - Each question should have 4 options (A, B, C, D)
      - Only one correct answer per question
      - Include explanation for correct answer
      - Vary difficulty levels
      - Avoid trick questions
      
      Questions:

providers:
  - "openai:gpt-4"
  - "anthropic:claude-3-sonnet"

tests:
  - description: "Photosynthesis explanation for middle school"
    vars:
      concept: "photosynthesis"
      grade_level: "7th grade"
      learning_objective_1: "the basic process of photosynthesis"
      learning_objective_2: "explain why photosynthesis is important for life"
      teaching_style: "engaging and interactive"
      length: 300
    assert:
      - type: "word_count"
        min: 250
        max: 350
      - type: "readability"
        min_grade_level: 6
        max_grade_level: 8
      - type: "contains"
        value: ["photosynthesis", "plants", "sunlight", "oxygen"]
      - type: "educational_quality"
        aspects: ["clarity", "engagement", "accuracy", "age_appropriateness"]
        threshold: 0.8
      - type: "structure"
        config:
          required_sections: ["definition", "examples", "key points"]

  - description: "Basic algebra quiz"
    vars:
      num_questions: 5
      topic: "solving linear equations"
      grade_level: "9th grade"
    assert:
      - type: "question_count"
        value: 5
      - type: "contains"
        value: ["A)", "B)", "C)", "D)"]
      - type: "educational_validity"
        criteria: "age-appropriate difficulty and clear explanations"
        threshold: 0.8
      - type: "math_accuracy"
        domain: "algebra"
        threshold: 0.95

assessment:
  rubric:
    - criterion: "Content Accuracy"
      weight: 0.3
      description: "Information is factually correct and current"
    - criterion: "Age Appropriateness"
      weight: 0.25
      description: "Language and concepts suitable for target grade level"
    - criterion: "Engagement"
      weight: 0.25
      description: "Content is interesting and motivating for students"
    - criterion: "Clarity"
      weight: 0.2
      description: "Explanations are clear and easy to understand"
```

---

## 🏥 Healthcare Applications

### Example 7: Medical Information Extraction
**Use Case**: Extract structured information from medical texts

```yaml
# examples/medical-extraction.yaml
description: "Medical information extraction with validation"

prompts:
  - id: "symptom-extractor"
    content: |
      Extract the following information from this medical note:
      
      Patient Note: {{patient_note}}
      
      Please extract:
      1. Primary symptoms
      2. Duration of symptoms
      3. Severity (mild/moderate/severe)
      4. Relevant medical history mentioned
      5. Medications mentioned
      
      Format as JSON:
      {
        "symptoms": [],
        "duration": "",
        "severity": "",
        "medical_history": [],
        "medications": []
      }

  - id: "clinical-summary"
    content: |
      Summarize this clinical case for medical professionals:
      
      {{clinical_text}}
      
      Provide:
      1. Chief complaint
      2. History of present illness (brief)
      3. Assessment
      4. Plan
      
      Use standard medical terminology and format.

providers:
  - id: "medical-llm"
    type: "openai"
    model: "gpt-4"
    config:
      temperature: 0.1
      max_tokens: 500

tests:
  - description: "Chest pain case"
    vars:
      patient_note: |
        45-year-old male presents with acute onset chest pain that started 
        2 hours ago. Pain is described as crushing, 8/10 severity, radiating 
        to left arm. Patient has history of hypertension and diabetes. 
        Currently taking lisinopril and metformin. No known allergies.
    assert:
      - type: "json_valid"
      - type: "json_schema"
        schema: |
          {
            "type": "object",
            "required": ["symptoms", "duration", "severity", "medical_history", "medications"],
            "properties": {
              "symptoms": {"type": "array"},
              "duration": {"type": "string"},
              "severity": {"type": "string"},
              "medical_history": {"type": "array"},
              "medications": {"type": "array"}
            }
          }
      - type: "contains"
        value: ["chest pain", "lisinopril", "metformin"]
      - type: "medical_accuracy"
        domain: "cardiology"
        threshold: 0.9

safety:
  disclaimers:
    - "For educational and research purposes only"
    - "Not intended for actual medical diagnosis"
    - "Results should be reviewed by qualified medical professionals"
  
  validation:
    - type: "medical_terminology"
      verify_accuracy: true
    - type: "clinical_guidelines"
      reference: "ICD-10, SNOMED CT"
```

**Important**: This example is for educational purposes only and should not be used for actual medical diagnosis.

---

## 🌐 Multilingual Applications

### Example 8: Translation Quality Assessment
**Use Case**: Evaluate translation quality across languages

```yaml
# examples/translation-quality.yaml
description: "Translation quality assessment with cultural awareness"

prompts:
  - id: "translator"
    content: |
      Translate the following {{source_language}} text to {{target_language}}.
      Maintain the original meaning, tone, and cultural context.
      
      Text: {{text}}
      
      Translation:

  - id: "cultural-adapter"
    content: |
      Adapt this text for {{target_culture}} audience while maintaining 
      the core message:
      
      Original ({{source_language}}): {{text}}
      
      Consider:
      - Cultural norms and values
      - Local expressions and idioms
      - Appropriate level of formality
      
      Culturally adapted {{target_language}} version:

providers:
  - "openai:gpt-4"
  - "anthropic:claude-3-opus"

tests:
  - description: "Business email translation"
    vars:
      source_language: "English"
      target_language: "Japanese"
      target_culture: "Japanese business"
      text: |
        Dear Mr. Johnson,
        
        I hope this email finds you well. I wanted to follow up on our 
        meeting last week regarding the marketing proposal. Could we 
        schedule a call this Friday to discuss the next steps?
        
        Best regards,
        Sarah
    assert:
      - type: "contains"
        value: ["Johnson", "marketing", "Friday"]
      - type: "translation_quality"
        metrics: ["accuracy", "fluency", "cultural_appropriateness"]
        threshold: 0.8
      - type: "formality_level"
        expected: "formal"
        language: "japanese"
      - type: "cultural_sensitivity"
        culture: "japanese_business"
        threshold: 0.9

  - description: "Marketing slogan localization"
    vars:
      source_language: "English"
      target_language: "Spanish"
      target_culture: "Latin American"
      text: "Just Do It - Push your limits"
    assert:
      - type: "translation_quality"
        metrics: ["creativity", "impact", "cultural_relevance"]
        threshold: 0.8
      - type: "brand_consistency"
        original_tone: "motivational"
        threshold: 0.8
      - type: "market_appropriateness"
        region: "latin_america"
        threshold: 0.9

localization:
  quality_gates:
    - metric: "linguistic_accuracy"
      threshold: 0.9
    - metric: "cultural_sensitivity"
      threshold: 0.85
    - metric: "brand_alignment"
      threshold: 0.8
```

---

## 🚀 Advanced Optimization Examples

### Example 9: Multi-Objective Optimization
**Use Case**: Optimize prompts for multiple competing objectives

```yaml
# examples/multi-objective-optimization.yaml
description: "Multi-objective prompt optimization (quality vs cost vs speed)"

prompts:
  - id: "base-prompt"
    content: "Explain {{concept}} in simple terms"

optimization:
  method: "textgrad"
  iterations: 10
  objectives:
    - name: "quality"
      weight: 0.5
      metric: "llm_judge_score"
    - name: "cost"
      weight: 0.3
      metric: "cost_per_response"
      direction: "minimize"
    - name: "speed"
      weight: 0.2
      metric: "latency"
      direction: "minimize"

providers:
  - "openai:gpt-4"      # High quality, expensive, slower
  - "openai:gpt-3.5-turbo"  # Good quality, cheaper, faster
  - "anthropic:claude-3-haiku"  # Good quality, cheap, very fast

tests:
  - vars:
      concept: "quantum computing"
    assert:
      - type: "llm-judge"
        value: "Clear, accurate explanation suitable for general audience"
        threshold: 0.7
      - type: "readability"
        min: 0.6
      - type: "cost"
        max: 0.05
      - type: "latency"
        max: "10s"

pareto_analysis:
  enabled: true
  generate_frontier: true
  visualize: true
```

### Example 10: A/B Testing Framework
**Use Case**: Scientific comparison of prompt variants

```yaml
# examples/ab-testing.yaml
description: "A/B testing framework for prompt optimization"

experiment:
  name: "customer_support_prompt_optimization"
  hypothesis: "More empathetic language improves customer satisfaction scores"
  duration: "7d"
  min_sample_size: 100
  significance_level: 0.05

variants:
  - id: "control"
    name: "Standard response"
    allocation: 0.5
    prompt: |
      Respond to this customer inquiry: {{inquiry}}
      
      Be helpful and professional.
      
  - id: "treatment"
    name: "Empathetic response"
    allocation: 0.5
    prompt: |
      Respond to this customer inquiry with empathy and understanding: {{inquiry}}
      
      Guidelines:
      - Acknowledge their feelings
      - Show genuine concern
      - Offer helpful solutions
      - Use warm, human language

providers:
  - "anthropic:claude-3-sonnet"

tests:
  - vars:
      inquiry: "I'm frustrated because my order arrived damaged and I need it for a gift tomorrow."
    metrics:
      primary:
        - name: "customer_satisfaction"
          type: "llm-judge"
          judge: "gpt-4"
          criteria: "Rate customer satisfaction with this response (1-10)"
      secondary:
        - name: "empathy_score"
          type: "sentiment"
          aspect: "empathy"
        - name: "helpfulness"
          type: "llm-judge"
          criteria: "How helpful is this response (1-10)"

analysis:
  statistical_tests:
    - "t_test"
    - "mann_whitney_u"
    - "chi_square"
  
  effect_size:
    - "cohens_d"
    - "cliff_delta"
  
  confidence_intervals: [90, 95, 99]
```

**Run A/B test**:
```bash
# Start A/B test
pe ab-test examples/ab-testing.yaml --duration 7d

# Monitor progress
pe ab-test status customer_support_prompt_optimization

# Analyze results
pe ab-test analyze customer_support_prompt_optimization --report
```

---

## 🎮 Interactive Tutorials

### Tutorial 1: Building Your First Chatbot
**Interactive step-by-step guide**

```bash
# Step 1: Initialize chatbot project
pe tutorial start chatbot-basics

# Step 2: Create personality
pe tutorial next --define-personality

# Step 3: Add conversation flow
pe tutorial next --conversation-flow

# Step 4: Test and optimize
pe tutorial next --optimize

# Complete tutorial
pe tutorial complete --generate-report
```

### Tutorial 2: Advanced Optimization Workshop
**Hands-on TextGrad mastery**

```bash
# Interactive optimization workshop
pe workshop textgrad-mastery --guided

# Practice scenarios
pe workshop scenario --type "customer_service"
pe workshop scenario --type "content_generation"
pe workshop scenario --type "code_generation"

# Certification test
pe workshop test --level advanced
```

---

## 📊 Analytics and Reporting Examples

### Example 11: Comprehensive Analytics Dashboard
**Use Case**: Monitor prompt performance across all dimensions

```yaml
# examples/analytics-dashboard.yaml
description: "Comprehensive analytics and monitoring"

monitoring:
  enabled: true
  update_interval: "30s"
  metrics:
    - "accuracy_score"
    - "cost_per_request"
    - "latency_p95"
    - "error_rate"
    - "user_satisfaction"

dashboards:
  - name: "Production Overview"
    widgets:
      - type: "line_chart"
        metric: "accuracy_score"
        timeframe: "24h"
      - type: "gauge"
        metric: "cost_per_request"
        threshold: 0.05
      - type: "histogram"
        metric: "latency"
        bins: 20

alerts:
  - name: "High Error Rate"
    condition: "error_rate > 0.05"
    action: "slack_notification"
    channel: "#prompt-engineering"
  
  - name: "Cost Spike"
    condition: "cost_per_request > 0.10"
    action: "email"
    recipients: ["team@company.com"]

reporting:
  schedule: "daily"
  format: "html"
  include:
    - performance_summary
    - cost_analysis
    - quality_trends
    - recommendations
```

---

## 🔧 Custom Integration Examples

### Example 12: Slack Bot Integration
**Use Case**: Deploy PE as a Slack bot

```bash
# Deploy PE as Slack bot
pe deploy slack-bot \
  --config examples/slack-bot-config.yaml \
  --workspace your-workspace

# Configuration
cat > slack-bot-config.yaml << EOF
bot:
  name: "PE Assistant"
  description: "Prompt engineering assistant"
  
commands:
  - name: "optimize"
    description: "Optimize a prompt"
    usage: "/pe-optimize Your prompt here"
    
  - name: "evaluate"
    description: "Evaluate prompt quality"
    usage: "/pe-eval Your prompt here"

providers:
  - "openai:gpt-3.5-turbo"

features:
  - interactive_optimization
  - real_time_feedback
  - team_collaboration
EOF
```

### Example 13: API Service Deployment
**Use Case**: Deploy PE as a REST API service

```bash
# Deploy as API service
pe serve --port 8080 --config api-config.yaml

# API endpoints automatically generated:
# POST /v1/evaluate
# POST /v1/optimize  
# GET /v1/health
# GET /v1/metrics
```

---

## 🎯 Industry-Specific Examples

### Finance: Risk Assessment
```yaml
# examples/finance-risk.yaml
description: "Financial risk assessment prompt optimization"

prompts:
  - |
    Analyze the financial risk of this investment proposal:
    {{proposal}}
    
    Consider: market conditions, regulatory compliance, historical performance
    
    Risk Assessment:

tests:
  - vars:
      proposal: "Tech startup Series A investment, $2M round"
    assert:
      - type: "contains"
        value: ["risk", "market", "compliance"]
      - type: "financial_accuracy"
        domain: "investment_analysis"
        threshold: 0.9
```

### Legal: Contract Analysis
```yaml
# examples/legal-contract.yaml
description: "Legal contract clause analysis"

prompts:
  - |
    Identify potential issues in this contract clause:
    {{clause}}
    
    Focus on: ambiguity, enforceability, compliance
    
    Analysis:

safety:
  disclaimers:
    - "Not legal advice"
    - "Consult qualified attorney"
```

---

## 🚀 Getting Started with Examples

### Quick Start
```bash
# Clone examples repository
git clone https://github.com/tmc/pe-examples.git
cd pe-examples

# Run any example
pe eval examples/qa-system.yaml

# Browse interactive catalog
pe examples browse

# Search examples by category
pe examples search --category "customer-service"
```

### Contributing Examples
```bash
# Submit your example
pe examples submit my-example.yaml \
  --category "healthcare" \
  --description "Medical text analysis"

# Validate example
pe examples validate my-example.yaml
```

---

**🎉 This examples library demonstrates PE's versatility across industries and use cases. Each example is production-ready and includes optimization strategies, validation techniques, and best practices.**

**📝 New examples are added regularly. Star the repository and watch for updates!**