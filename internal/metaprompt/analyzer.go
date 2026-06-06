package metaprompt

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tmc/pe/internal/llm"
)

// TextGradAnalyzer analyzes prompt-response gradients using attention patterns
type TextGradAnalyzer struct {
	llm Generator
}

// NewTextGradAnalyzer creates a new TextGrad analyzer
func NewTextGradAnalyzer(llmProvider Generator) *TextGradAnalyzer {
	return &TextGradAnalyzer{
		llm: llmProvider,
	}
}

// AttentionFlow represents token relationship mapping
type AttentionFlow struct {
	SourceToken string            `json:"source_token"`
	TargetToken string            `json:"target_token"`
	Weight      float64           `json:"weight"`
	Metadata    map[string]string `json:"metadata"`
}

// SemanticDrift represents concept preservation analysis
type SemanticDrift struct {
	ConceptID     string  `json:"concept_id"`
	OriginalValue string  `json:"original_value"`
	CurrentValue  string  `json:"current_value"`
	DriftScore    float64 `json:"drift_score"`
	DriftType     string  `json:"drift_type"` // "semantic", "syntactic", "contextual"
}

// CoherenceMetrics represents coherence scoring
type CoherenceMetrics struct {
	LocalCoherence  float64 `json:"local_coherence"`
	GlobalCoherence float64 `json:"global_coherence"`
	LogicalFlow     float64 `json:"logical_flow"`
	Consistency     float64 `json:"consistency"`
}

// AnalysisResult contains the complete analysis
type AnalysisResult struct {
	AttentionFlows    []AttentionFlow  `json:"attention_flows"`
	SemanticDrifts    []SemanticDrift  `json:"semantic_drifts"`
	CoherenceMetrics  CoherenceMetrics `json:"coherence_metrics"`
	GradientStrength  float64          `json:"gradient_strength"`
	OptimizationHints []string         `json:"optimization_hints"`
}

// AnalyzePromptGradients performs comprehensive gradient analysis
func (tga *TextGradAnalyzer) AnalyzePromptGradients(ctx context.Context, prompt, response string) (*AnalysisResult, error) {
	result := &AnalysisResult{}

	// Analyze attention flows
	flows, err := tga.analyzeAttentionFlows(ctx, prompt, response)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze attention flows: %w", err)
	}
	result.AttentionFlows = flows

	// Detect semantic drift
	drifts, err := tga.detectSemanticDrift(ctx, prompt, response)
	if err != nil {
		return nil, fmt.Errorf("failed to detect semantic drift: %w", err)
	}
	result.SemanticDrifts = drifts

	// Calculate coherence metrics
	metrics, err := tga.calculateCoherenceMetrics(ctx, prompt, response)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate coherence metrics: %w", err)
	}
	result.CoherenceMetrics = metrics

	// Compute gradient strength
	result.GradientStrength = tga.computeGradientStrength(flows, drifts, metrics)

	// Generate optimization hints
	hints, err := tga.generateOptimizationHints(ctx, result)
	if err != nil {
		return nil, fmt.Errorf("failed to generate optimization hints: %w", err)
	}
	result.OptimizationHints = hints

	return result, nil
}

// analyzeAttentionFlows maps token relationships
func (tga *TextGradAnalyzer) analyzeAttentionFlows(ctx context.Context, prompt, response string) ([]AttentionFlow, error) {
	analysisPrompt := fmt.Sprintf(`You are an expert in transformer attention analysis. Analyze the token relationships between this prompt and response.

PROMPT:
%s

RESPONSE:
%s

TASK: Identify the 10 most important token relationships that show how the prompt influenced the response. Focus on:
1. Key concept transfers
2. Instruction-to-output mappings
3. Context preservation patterns
4. Constraint adherence flows

FORMAT your response as JSON:
{
  "attention_flows": [
    {
      "source_token": "token from prompt",
      "target_token": "token in response", 
      "weight": 0.85,
      "metadata": {"type": "concept_transfer", "importance": "high"}
    }
  ]
}`, prompt, response)

	options := llm.GenerateOptions{
		Temperature: floatPtr(0.2),
		MaxTokens:   intPtr(1000),
	}

	resp, err := tga.llm.Generate(ctx, analysisPrompt, options)
	if err != nil {
		return nil, err
	}

	var result struct {
		AttentionFlows []AttentionFlow `json:"attention_flows"`
	}

	if err := json.Unmarshal([]byte(resp.Text), &result); err != nil {
		return nil, fmt.Errorf("parse attention flows: %w", err)
	}

	return result.AttentionFlows, nil
}

// detectSemanticDrift identifies concept preservation issues
func (tga *TextGradAnalyzer) detectSemanticDrift(ctx context.Context, prompt, response string) ([]SemanticDrift, error) {
	driftPrompt := fmt.Sprintf(`You are an expert in semantic analysis. Identify semantic drift between the prompt intent and response output.

PROMPT:
%s

RESPONSE:
%s

TASK: Analyze for semantic drift in these categories:
1. Concept preservation - Are key concepts maintained?
2. Intent alignment - Does response match prompt intent?
3. Constraint adherence - Are specified constraints followed?
4. Contextual consistency - Is context properly maintained?

FORMAT your response as JSON:
{
  "semantic_drifts": [
    {
      "concept_id": "main_task",
      "original_value": "summarize the text",
      "current_value": "analyzed sentiment instead",
      "drift_score": 0.7,
      "drift_type": "semantic"
    }
  ]
}`, prompt, response)

	options := llm.GenerateOptions{
		Temperature: floatPtr(0.2),
		MaxTokens:   intPtr(800),
	}

	resp, err := tga.llm.Generate(ctx, driftPrompt, options)
	if err != nil {
		return nil, err
	}

	var result struct {
		SemanticDrifts []SemanticDrift `json:"semantic_drifts"`
	}

	if err := json.Unmarshal([]byte(resp.Text), &result); err != nil {
		return nil, fmt.Errorf("parse semantic drifts: %w", err)
	}

	return result.SemanticDrifts, nil
}

// calculateCoherenceMetrics computes coherence scores
func (tga *TextGradAnalyzer) calculateCoherenceMetrics(ctx context.Context, prompt, response string) (CoherenceMetrics, error) {
	coherencePrompt := fmt.Sprintf(`You are an expert in text coherence analysis. Evaluate the coherence between this prompt and response.

PROMPT:
%s

RESPONSE:
%s

TASK: Rate coherence on these dimensions (0.0-1.0):
1. Local Coherence - Sentence-to-sentence flow
2. Global Coherence - Overall structure and organization  
3. Logical Flow - Reasoning and argument progression
4. Consistency - Terminology and style consistency

FORMAT your response as JSON:
{
  "local_coherence": 0.85,
  "global_coherence": 0.78,
  "logical_flow": 0.82,
  "consistency": 0.90
}`, prompt, response)

	options := llm.GenerateOptions{
		Temperature: floatPtr(0.1),
		MaxTokens:   intPtr(300),
	}

	resp, err := tga.llm.Generate(ctx, coherencePrompt, options)
	if err != nil {
		return CoherenceMetrics{}, err
	}

	var metrics CoherenceMetrics
	if err := json.Unmarshal([]byte(resp.Text), &metrics); err != nil {
		return CoherenceMetrics{}, fmt.Errorf("parse coherence metrics: %w", err)
	}

	return metrics, nil
}

// computeGradientStrength calculates overall gradient strength
func (tga *TextGradAnalyzer) computeGradientStrength(flows []AttentionFlow, drifts []SemanticDrift, metrics CoherenceMetrics) float64 {
	// Weight factors
	attentionWeight := 0.3
	driftWeight := 0.4
	coherenceWeight := 0.3

	// Calculate attention strength (average of top flows)
	var attentionStrength float64
	if len(flows) > 0 {
		var total float64
		for _, flow := range flows {
			total += flow.Weight
		}
		attentionStrength = total / float64(len(flows))
	}

	// Calculate drift penalty (higher drift = lower strength)
	var driftPenalty float64
	if len(drifts) > 0 {
		var totalDrift float64
		for _, drift := range drifts {
			totalDrift += drift.DriftScore
		}
		driftPenalty = totalDrift / float64(len(drifts))
	}

	// Calculate coherence strength
	coherenceStrength := (metrics.LocalCoherence + metrics.GlobalCoherence +
		metrics.LogicalFlow + metrics.Consistency) / 4.0

	// Combine factors
	gradientStrength := (attentionStrength * attentionWeight) +
		((1.0 - driftPenalty) * driftWeight) +
		(coherenceStrength * coherenceWeight)

	return gradientStrength
}

// generateOptimizationHints creates actionable improvement suggestions
func (tga *TextGradAnalyzer) generateOptimizationHints(ctx context.Context, analysis *AnalysisResult) ([]string, error) {
	hintsPrompt := fmt.Sprintf(`Based on this TextGrad analysis, provide 5 specific optimization hints for improving the prompt.

ANALYSIS SUMMARY:
- Gradient Strength: %.2f
- Average Coherence: %.2f
- Semantic Drifts Found: %d
- Attention Flows Analyzed: %d

KEY ISSUES:
%s

TASK: Generate specific, actionable optimization hints that address the identified issues.

FORMAT:
1. [Specific hint 1]
2. [Specific hint 2] 
3. [Specific hint 3]
4. [Specific hint 4]
5. [Specific hint 5]`,
		analysis.GradientStrength,
		(analysis.CoherenceMetrics.LocalCoherence+analysis.CoherenceMetrics.GlobalCoherence+
			analysis.CoherenceMetrics.LogicalFlow+analysis.CoherenceMetrics.Consistency)/4.0,
		len(analysis.SemanticDrifts),
		len(analysis.AttentionFlows),
		tga.summarizeIssues(analysis))

	options := llm.GenerateOptions{
		Temperature: floatPtr(0.3),
		MaxTokens:   intPtr(500),
	}

	resp, err := tga.llm.Generate(ctx, hintsPrompt, options)
	if err != nil {
		return nil, err
	}

	return tga.parseOptimizationHints(resp.Text), nil
}

func (tga *TextGradAnalyzer) parseOptimizationHints(text string) []string {
	lines := strings.Split(text, "\n")
	var hints []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "1.") || strings.HasPrefix(line, "2.") ||
			strings.HasPrefix(line, "3.") || strings.HasPrefix(line, "4.") ||
			strings.HasPrefix(line, "5.") {
			hint := strings.TrimSpace(line[2:])
			if hint != "" {
				hints = append(hints, hint)
			}
		}
	}

	return hints
}

func (tga *TextGradAnalyzer) summarizeIssues(analysis *AnalysisResult) string {
	var issues []string

	if analysis.GradientStrength < 0.6 {
		issues = append(issues, "Low gradient strength indicates weak prompt-response alignment")
	}

	if len(analysis.SemanticDrifts) > 0 {
		issues = append(issues, fmt.Sprintf("Found %d semantic drift patterns", len(analysis.SemanticDrifts)))
	}

	avgCoherence := (analysis.CoherenceMetrics.LocalCoherence + analysis.CoherenceMetrics.GlobalCoherence +
		analysis.CoherenceMetrics.LogicalFlow + analysis.CoherenceMetrics.Consistency) / 4.0
	if avgCoherence < 0.7 {
		issues = append(issues, "Below-average coherence scores")
	}

	if len(issues) == 0 {
		return "No significant issues identified"
	}

	return strings.Join(issues, "; ")
}

// Helper functions
func floatPtr(f float64) *float64 { return &f }
func intPtr(i int) *int           { return &i }
