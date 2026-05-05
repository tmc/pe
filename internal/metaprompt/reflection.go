package metaprompt

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/tmc/pe/internal/llm"
)

// ReflectionEngine performs meta-analysis of prompt engineering processes
type ReflectionEngine struct {
	llm Generator
}

// NewReflectionEngine creates a new reflection engine
func NewReflectionEngine(llmProvider Generator) *ReflectionEngine {
	return &ReflectionEngine{
		llm: llmProvider,
	}
}

// SuccessPattern represents a successful prompt engineering pattern
type SuccessPattern struct {
	PatternID     string             `json:"pattern_id"`
	PatternName   string             `json:"pattern_name"`
	Description   string             `json:"description"`
	Context       string             `json:"context"`       // When this pattern works well
	Techniques    []string           `json:"techniques"`    // Specific techniques used
	Effectiveness float64            `json:"effectiveness"` // Success rate (0-1)
	Examples      []string           `json:"examples"`      // Example successful applications
	Conditions    []string           `json:"conditions"`    // Conditions for success
	Metrics       map[string]float64 `json:"metrics"`       // Performance metrics
}

// StrategyRecommendation provides guidance on tool/technique selection
type StrategyRecommendation struct {
	Scenario       string   `json:"scenario"`        // Type of prompt engineering task
	Recommended    []string `json:"recommended"`     // Recommended tools/techniques
	NotRecommended []string `json:"not_recommended"` // Tools to avoid
	Rationale      string   `json:"rationale"`       // Why this strategy works
	Confidence     float64  `json:"confidence"`      // Confidence in recommendation
	Alternatives   []string `json:"alternatives"`    // Alternative approaches
}

// KnowledgeItem represents distilled prompt engineering knowledge
type KnowledgeItem struct {
	ItemID       string   `json:"item_id"`
	Category     string   `json:"category"` // "principle", "technique", "pattern", "pitfall"
	Title        string   `json:"title"`
	Content      string   `json:"content"`
	Importance   float64  `json:"importance"`   // How important this knowledge is (0-1)
	Confidence   float64  `json:"confidence"`   // Confidence in this knowledge (0-1)
	Sources      []string `json:"sources"`      // Where this knowledge came from
	Applications []string `json:"applications"` // Where to apply this knowledge
}

// WorkflowInsight represents improvements to the PE workflow
type WorkflowInsight struct {
	InsightID      string   `json:"insight_id"`
	InsightType    string   `json:"insight_type"` // "efficiency", "quality", "process"
	Description    string   `json:"description"`
	CurrentState   string   `json:"current_state"`  // How things currently work
	ProposedState  string   `json:"proposed_state"` // How they could work better
	Benefits       []string `json:"benefits"`       // Expected improvements
	Implementation string   `json:"implementation"` // How to implement this insight
	Priority       string   `json:"priority"`       // "low", "medium", "high"
}

// ReflectionResult contains the complete meta-analysis
type ReflectionResult struct {
	SessionSummary          SessionSummary           `json:"session_summary"`
	SuccessPatterns         []SuccessPattern         `json:"success_patterns"`
	StrategyRecommendations []StrategyRecommendation `json:"strategy_recommendations"`
	KnowledgeBase           []KnowledgeItem          `json:"knowledge_base"`
	WorkflowInsights        []WorkflowInsight        `json:"workflow_insights"`
	OverallEffectiveness    float64                  `json:"overall_effectiveness"`
	LearningsExtracted      int                      `json:"learnings_extracted"`
	Recommendations         []string                 `json:"recommendations"`
	NextSteps               []string                 `json:"next_steps"`
}

// SessionSummary provides high-level session analysis
type SessionSummary struct {
	TotalPrompts       int           `json:"total_prompts"`
	TotalOptimizations int           `json:"total_optimizations"`
	AvgImprovement     float64       `json:"avg_improvement"`
	TotalDuration      time.Duration `json:"total_duration"`
	ToolsUsed          []string      `json:"tools_used"`
	MostEffectiveTool  string        `json:"most_effective_tool"`
	CommonPatterns     []string      `json:"common_patterns"`
	KeyChallenges      []string      `json:"key_challenges"`
}

// PerformReflection conducts comprehensive meta-analysis
func (re *ReflectionEngine) PerformReflection(ctx context.Context, sessionData SessionData) (*ReflectionResult, error) {
	result := &ReflectionResult{}

	// Generate session summary
	summary := re.generateSessionSummary(sessionData)
	result.SessionSummary = summary

	// Mine success patterns
	patterns, err := re.mineSuccessPatterns(ctx, sessionData)
	if err != nil {
		return nil, fmt.Errorf("failed to mine success patterns: %w", err)
	}
	result.SuccessPatterns = patterns

	// Generate strategy recommendations
	strategies, err := re.generateStrategyRecommendations(ctx, sessionData, patterns)
	if err != nil {
		return nil, fmt.Errorf("failed to generate strategy recommendations: %w", err)
	}
	result.StrategyRecommendations = strategies

	// Extract knowledge items
	knowledge, err := re.extractKnowledge(ctx, sessionData, patterns)
	if err != nil {
		return nil, fmt.Errorf("failed to extract knowledge: %w", err)
	}
	result.KnowledgeBase = knowledge

	// Generate workflow insights
	insights, err := re.generateWorkflowInsights(ctx, sessionData, summary)
	if err != nil {
		return nil, fmt.Errorf("failed to generate workflow insights: %w", err)
	}
	result.WorkflowInsights = insights

	// Calculate overall metrics
	result.OverallEffectiveness = re.calculateOverallEffectiveness(sessionData)
	result.LearningsExtracted = len(patterns) + len(knowledge) + len(insights)

	// Generate final recommendations
	recommendations, err := re.generateFinalRecommendations(ctx, result)
	if err == nil {
		result.Recommendations = recommendations
	}

	// Generate next steps
	nextSteps, err := re.generateNextSteps(ctx, result)
	if err == nil {
		result.NextSteps = nextSteps
	}

	return result, nil
}

// SessionData represents historical prompt engineering session data
type SessionData struct {
	Sessions []OptimizationSession `json:"sessions"`
}

// OptimizationSession represents a single prompt optimization session
type OptimizationSession struct {
	SessionID     string                 `json:"session_id"`
	Timestamp     time.Time              `json:"timestamp"`
	InitialPrompt string                 `json:"initial_prompt"`
	FinalPrompt   string                 `json:"final_prompt"`
	ToolsUsed     []string               `json:"tools_used"`
	Iterations    []IterationResult      `json:"iterations"`
	FinalScore    float64                `json:"final_score"`
	Duration      time.Duration          `json:"duration"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// generateSessionSummary creates high-level session analysis
func (re *ReflectionEngine) generateSessionSummary(sessionData SessionData) SessionSummary {
	summary := SessionSummary{
		TotalPrompts:   len(sessionData.Sessions),
		ToolsUsed:      []string{},
		CommonPatterns: []string{},
		KeyChallenges:  []string{},
	}

	if len(sessionData.Sessions) == 0 {
		return summary
	}

	// Calculate averages and aggregates
	var totalImprovement float64
	var totalDuration time.Duration
	var totalOptimizations int
	toolUsage := make(map[string]int)

	for _, session := range sessionData.Sessions {
		totalOptimizations += len(session.Iterations)
		totalDuration += session.Duration

		if len(session.Iterations) > 0 {
			firstScore := session.Iterations[0].Score
			lastScore := session.Iterations[len(session.Iterations)-1].Score
			totalImprovement += (lastScore - firstScore)
		}

		for _, tool := range session.ToolsUsed {
			toolUsage[tool]++
		}
	}

	summary.TotalOptimizations = totalOptimizations
	summary.AvgImprovement = totalImprovement / float64(len(sessionData.Sessions))
	summary.TotalDuration = totalDuration

	// Find most effective tool
	var mostUsedTool string
	var maxUsage int
	for tool, count := range toolUsage {
		summary.ToolsUsed = append(summary.ToolsUsed, tool)
		if count > maxUsage {
			maxUsage = count
			mostUsedTool = tool
		}
	}
	summary.MostEffectiveTool = mostUsedTool

	return summary
}

// mineSuccessPatterns identifies successful prompt engineering patterns
func (re *ReflectionEngine) mineSuccessPatterns(ctx context.Context, sessionData SessionData) ([]SuccessPattern, error) {
	patternPrompt := fmt.Sprintf(`You are an expert prompt engineering researcher. Analyze these optimization sessions to identify successful patterns.

SESSION DATA:
%s

TASK: Identify 3-5 patterns that consistently led to successful prompt improvements. For each pattern:
1. Describe the specific technique or approach
2. Identify the contexts where it works best
3. Quantify its effectiveness based on the data
4. Provide concrete examples from the sessions

Focus on patterns that:
- Appear across multiple sessions
- Show consistent improvement results
- Are actionable and replicable
- Address common prompt engineering challenges

FORMAT your response as JSON:
{
  "success_patterns": [
    {
      "pattern_id": "iterative_refinement",
      "pattern_name": "Progressive Specification",
      "description": "Gradually adding specific constraints and format requirements",
      "context": "When initial prompts are too broad or ambiguous",
      "techniques": ["constraint addition", "format specification", "example inclusion"],
      "effectiveness": 0.85,
      "examples": ["Session X showed 40%% improvement", "Session Y improved clarity"],
      "conditions": ["Initial score < 6", "Ambiguous requirements", "Multiple valid interpretations"],
      "metrics": {"avg_improvement": 0.35, "success_rate": 0.85}
    }
  ]
}`, re.formatSessionData(sessionData))

	options := llm.GenerateOptions{
		Temperature: floatPtr(0.3),
		MaxTokens:   intPtr(1500),
	}

	resp, err := re.llm.Generate(ctx, patternPrompt, options)
	if err != nil {
		return nil, err
	}

	var result struct {
		SuccessPatterns []SuccessPattern `json:"success_patterns"`
	}

	if err := json.Unmarshal([]byte(resp.Text), &result); err != nil {
		return re.generateDefaultPatterns(), nil
	}

	return result.SuccessPatterns, nil
}

// generateStrategyRecommendations creates tool selection guidance
func (re *ReflectionEngine) generateStrategyRecommendations(ctx context.Context, sessionData SessionData, patterns []SuccessPattern) ([]StrategyRecommendation, error) {
	strategyPrompt := fmt.Sprintf(`Based on the session data and success patterns, generate strategy recommendations for different prompt engineering scenarios.

SUCCESS PATTERNS:
%s

SESSION ANALYSIS:
%s

TASK: Create 4-6 strategy recommendations for different scenarios. Each should specify:
1. When to use which tools/techniques
2. What to avoid in each scenario
3. Clear rationale for the recommendations
4. Alternative approaches

Consider scenarios like:
- Brand new prompt creation
- Existing prompt optimization  
- Error-prone prompt fixing
- Performance tuning
- Multi-step workflow optimization

FORMAT your response as JSON:
{
  "strategy_recommendations": [
    {
      "scenario": "New prompt creation from scratch",
      "recommended": ["analysis", "iterative_refinement", "validation"],
      "not_recommended": ["complex_optimization", "advanced_textgrad"],
      "rationale": "Start simple and build complexity gradually",
      "confidence": 0.9,
      "alternatives": ["template-based approach", "example-driven development"]
    }
  ]
}`, re.formatSuccessPatterns(patterns), re.formatSessionSummary(sessionData))

	options := llm.GenerateOptions{
		Temperature: floatPtr(0.3),
		MaxTokens:   intPtr(1200),
	}

	resp, err := re.llm.Generate(ctx, strategyPrompt, options)
	if err != nil {
		return nil, err
	}

	var result struct {
		StrategyRecommendations []StrategyRecommendation `json:"strategy_recommendations"`
	}

	if err := json.Unmarshal([]byte(resp.Text), &result); err != nil {
		return re.generateDefaultStrategies(), nil
	}

	return result.StrategyRecommendations, nil
}

// extractKnowledge distills key prompt engineering insights
func (re *ReflectionEngine) extractKnowledge(ctx context.Context, sessionData SessionData, patterns []SuccessPattern) ([]KnowledgeItem, error) {
	knowledgePrompt := fmt.Sprintf(`Extract key prompt engineering knowledge from this analysis. Distill the most important insights, principles, and techniques.

SUCCESS PATTERNS:
%s

SESSION INSIGHTS:
%s

TASK: Extract 5-8 key knowledge items covering:
1. Core principles that drive success
2. Specific techniques that work consistently  
3. Common patterns to recognize and apply
4. Important pitfalls to avoid

Each knowledge item should be:
- Actionable and specific
- Backed by evidence from the data
- Broadly applicable across scenarios
- Clear about when and how to apply it

FORMAT your response as JSON:
{
  "knowledge_items": [
    {
      "item_id": "specificity_principle",
      "category": "principle",
      "title": "Specificity Drives Performance",
      "content": "Adding specific constraints and format requirements consistently improves prompt performance",
      "importance": 0.9,
      "confidence": 0.85,
      "sources": ["pattern analysis", "session data"],
      "applications": ["initial prompt design", "optimization iterations", "error fixing"]
    }
  ]
}`, re.formatSuccessPatterns(patterns), re.formatSessionAnalysis(sessionData))

	options := llm.GenerateOptions{
		Temperature: floatPtr(0.2),
		MaxTokens:   intPtr(1200),
	}

	resp, err := re.llm.Generate(ctx, knowledgePrompt, options)
	if err != nil {
		return nil, err
	}

	var result struct {
		KnowledgeItems []KnowledgeItem `json:"knowledge_items"`
	}

	if err := json.Unmarshal([]byte(resp.Text), &result); err != nil {
		return re.generateDefaultKnowledge(), nil
	}

	return result.KnowledgeItems, nil
}

// generateWorkflowInsights identifies process improvements
func (re *ReflectionEngine) generateWorkflowInsights(ctx context.Context, sessionData SessionData, summary SessionSummary) ([]WorkflowInsight, error) {
	workflowPrompt := fmt.Sprintf(`Analyze the prompt engineering workflow and identify improvement opportunities.

WORKFLOW ANALYSIS:
- Total sessions: %d
- Average improvement: %.2f
- Most used tool: %s
- Average duration: %v

TASK: Identify 3-5 workflow improvements focusing on:
1. Efficiency gains (faster optimization)
2. Quality improvements (better results)
3. Process streamlining (smoother workflow)
4. Tool integration opportunities

Each insight should specify:
- Current workflow issues
- Proposed improvements
- Expected benefits
- Implementation approach

FORMAT your response as JSON:
{
  "workflow_insights": [
    {
      "insight_id": "early_analysis",
      "insight_type": "efficiency",
      "description": "Perform analysis stage earlier in the process",
      "current_state": "Analysis happens ad-hoc during optimization",
      "proposed_state": "Standardized analysis as first step in all optimizations",
      "benefits": ["Faster convergence", "Better initial direction", "Reduced iterations"],
      "implementation": "Add analysis stage to default workflow templates",
      "priority": "high"
    }
  ]
}`, summary.TotalPrompts, summary.AvgImprovement, summary.MostEffectiveTool, summary.TotalDuration)

	options := llm.GenerateOptions{
		Temperature: floatPtr(0.3),
		MaxTokens:   intPtr(1000),
	}

	resp, err := re.llm.Generate(ctx, workflowPrompt, options)
	if err != nil {
		return nil, err
	}

	var result struct {
		WorkflowInsights []WorkflowInsight `json:"workflow_insights"`
	}

	if err := json.Unmarshal([]byte(resp.Text), &result); err != nil {
		return re.generateDefaultInsights(), nil
	}

	return result.WorkflowInsights, nil
}

// calculateOverallEffectiveness computes session effectiveness metrics
func (re *ReflectionEngine) calculateOverallEffectiveness(sessionData SessionData) float64 {
	if len(sessionData.Sessions) == 0 {
		return 0.0
	}

	var totalEffectiveness float64
	successfulSessions := 0

	for _, session := range sessionData.Sessions {
		if len(session.Iterations) > 0 {
			firstScore := session.Iterations[0].Score
			lastScore := session.Iterations[len(session.Iterations)-1].Score
			improvement := (lastScore - firstScore) / 10.0 // Normalize to 0-1

			if improvement > 0 {
				totalEffectiveness += improvement
				successfulSessions++
			}
		}
	}

	if successfulSessions == 0 {
		return 0.0
	}

	return totalEffectiveness / float64(successfulSessions)
}

// generateFinalRecommendations creates overall guidance
func (re *ReflectionEngine) generateFinalRecommendations(ctx context.Context, result *ReflectionResult) ([]string, error) {
	recPrompt := fmt.Sprintf(`Based on this comprehensive reflection analysis, provide final recommendations for improving prompt engineering practice.

ANALYSIS SUMMARY:
- Overall Effectiveness: %.2f
- Success Patterns Found: %d
- Knowledge Items Extracted: %d
- Workflow Insights: %d

KEY PATTERNS:
%s

TASK: Provide 5-7 final recommendations that synthesize all insights into actionable guidance.

FORMAT:
1. [Recommendation 1]
2. [Recommendation 2]
3. [Recommendation 3] 
4. [Recommendation 4]
5. [Recommendation 5]
6. [Recommendation 6]
7. [Recommendation 7]`,
		result.OverallEffectiveness,
		len(result.SuccessPatterns),
		len(result.KnowledgeBase),
		len(result.WorkflowInsights),
		re.formatTopPatterns(result.SuccessPatterns))

	options := llm.GenerateOptions{
		Temperature: floatPtr(0.3),
		MaxTokens:   intPtr(600),
	}

	resp, err := re.llm.Generate(ctx, recPrompt, options)
	if err != nil {
		return nil, err
	}

	return re.parseRecommendations(resp.Text), nil
}

// generateNextSteps creates actionable next steps
func (re *ReflectionEngine) generateNextSteps(ctx context.Context, result *ReflectionResult) ([]string, error) {
	nextStepsPrompt := fmt.Sprintf(`Based on this reflection analysis, suggest specific next steps for continuing to improve prompt engineering capabilities.

CURRENT STATE:
- Effectiveness: %.2f
- Areas for improvement identified: %d
- High-priority insights: %d

TASK: Suggest 3-5 specific, actionable next steps for immediate implementation.

FORMAT:
1. [Next step 1]
2. [Next step 2]
3. [Next step 3]
4. [Next step 4]
5. [Next step 5]`,
		result.OverallEffectiveness,
		len(result.WorkflowInsights),
		re.countHighPriorityInsights(result.WorkflowInsights))

	options := llm.GenerateOptions{
		Temperature: floatPtr(0.3),
		MaxTokens:   intPtr(400),
	}

	resp, err := re.llm.Generate(ctx, nextStepsPrompt, options)
	if err != nil {
		return nil, err
	}

	return re.parseRecommendations(resp.Text), nil
}

// Helper functions
func (re *ReflectionEngine) formatSessionData(sessionData SessionData) string {
	if len(sessionData.Sessions) == 0 {
		return "No session data available"
	}

	var formatted strings.Builder
	for i, session := range sessionData.Sessions {
		if i >= 5 { // Limit to 5 sessions for analysis
			break
		}
		formatted.WriteString(fmt.Sprintf("Session %d: %d iterations, %.1f final score, %v duration\n",
			i+1, len(session.Iterations), session.FinalScore, session.Duration))
	}
	return formatted.String()
}

func (re *ReflectionEngine) formatSessionSummary(sessionData SessionData) string {
	summary := re.generateSessionSummary(sessionData)
	return fmt.Sprintf("Sessions: %d, Avg Improvement: %.2f, Tools: %v",
		summary.TotalPrompts, summary.AvgImprovement, summary.ToolsUsed)
}

func (re *ReflectionEngine) formatSuccessPatterns(patterns []SuccessPattern) string {
	var formatted strings.Builder
	for _, pattern := range patterns {
		formatted.WriteString(fmt.Sprintf("- %s (%.1f%% effective): %s\n",
			pattern.PatternName, pattern.Effectiveness*100, pattern.Description))
	}
	return formatted.String()
}

func (re *ReflectionEngine) formatSessionAnalysis(sessionData SessionData) string {
	return fmt.Sprintf("%d sessions analyzed with insights on tool usage and optimization patterns",
		len(sessionData.Sessions))
}

func (re *ReflectionEngine) formatTopPatterns(patterns []SuccessPattern) string {
	// Sort by effectiveness
	sorted := make([]SuccessPattern, len(patterns))
	copy(sorted, patterns)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Effectiveness > sorted[j].Effectiveness
	})

	var formatted strings.Builder
	for i, pattern := range sorted {
		if i >= 3 { // Top 3 patterns
			break
		}
		formatted.WriteString(fmt.Sprintf("- %s: %s\n", pattern.PatternName, pattern.Description))
	}
	return formatted.String()
}

func (re *ReflectionEngine) countHighPriorityInsights(insights []WorkflowInsight) int {
	count := 0
	for _, insight := range insights {
		if insight.Priority == "high" {
			count++
		}
	}
	return count
}

func (re *ReflectionEngine) parseRecommendations(text string) []string {
	lines := strings.Split(text, "\n")
	var recommendations []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "1.") || strings.HasPrefix(line, "2.") ||
			strings.HasPrefix(line, "3.") || strings.HasPrefix(line, "4.") ||
			strings.HasPrefix(line, "5.") || strings.HasPrefix(line, "6.") ||
			strings.HasPrefix(line, "7.") {
			rec := strings.TrimSpace(line[2:])
			if rec != "" {
				recommendations = append(recommendations, rec)
			}
		}
	}

	return recommendations
}

// Default fallback generators
func (re *ReflectionEngine) generateDefaultPatterns() []SuccessPattern {
	return []SuccessPattern{
		{
			PatternID:     "iterative_improvement",
			PatternName:   "Iterative Refinement",
			Description:   "Gradual improvement through multiple optimization cycles",
			Context:       "Most prompt optimization scenarios",
			Techniques:    []string{"incremental changes", "feedback incorporation"},
			Effectiveness: 0.75,
			Examples:      []string{"Consistent improvement across sessions"},
		},
	}
}

func (re *ReflectionEngine) generateDefaultStrategies() []StrategyRecommendation {
	return []StrategyRecommendation{
		{
			Scenario:       "General prompt optimization",
			Recommended:    []string{"analysis", "iterative_refinement"},
			NotRecommended: []string{"complex_methods_initially"},
			Rationale:      "Start simple and build complexity",
			Confidence:     0.8,
		},
	}
}

func (re *ReflectionEngine) generateDefaultKnowledge() []KnowledgeItem {
	return []KnowledgeItem{
		{
			ItemID:       "clarity_principle",
			Category:     "principle",
			Title:        "Clarity Improves Performance",
			Content:      "Clear, specific instructions lead to better outputs",
			Importance:   0.9,
			Confidence:   0.8,
			Applications: []string{"all prompt scenarios"},
		},
	}
}

func (re *ReflectionEngine) generateDefaultInsights() []WorkflowInsight {
	return []WorkflowInsight{
		{
			InsightID:     "systematic_approach",
			InsightType:   "process",
			Description:   "Use systematic optimization approaches",
			CurrentState:  "Ad-hoc optimization",
			ProposedState: "Structured optimization workflow",
			Benefits:      []string{"Better results", "Faster optimization"},
			Priority:      "medium",
		},
	}
}
