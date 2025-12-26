package metaprompt

import (
	"testing"
	"time"
)

func TestNewReflectionEngine(t *testing.T) {
	provider := &mockProvider{}
	engine := NewReflectionEngine(provider)

	if engine == nil {
		t.Fatal("NewReflectionEngine returned nil")
	}
	if engine.llm == nil {
		t.Error("ReflectionEngine has nil llm provider")
	}
}

func TestGenerateSessionSummary(t *testing.T) {
	tests := []struct {
		name        string
		sessionData SessionData
		check       func(SessionSummary) bool
	}{
		{
			name:        "empty sessions",
			sessionData: SessionData{Sessions: []OptimizationSession{}},
			check: func(s SessionSummary) bool {
				return s.TotalPrompts == 0
			},
		},
		{
			name: "single session",
			sessionData: SessionData{
				Sessions: []OptimizationSession{
					{
						SessionID: "s1",
						Iterations: []IterationResult{
							{Score: 5.0},
							{Score: 8.0},
						},
						ToolsUsed: []string{"analysis", "optimization"},
						Duration:  time.Minute,
					},
				},
			},
			check: func(s SessionSummary) bool {
				return s.TotalPrompts == 1 &&
					s.TotalOptimizations == 2 &&
					s.AvgImprovement == 3.0
			},
		},
		{
			name: "multiple sessions",
			sessionData: SessionData{
				Sessions: []OptimizationSession{
					{
						Iterations: []IterationResult{{Score: 4.0}, {Score: 7.0}},
						ToolsUsed:  []string{"tool1"},
						Duration:   time.Minute,
					},
					{
						Iterations: []IterationResult{{Score: 5.0}, {Score: 9.0}},
						ToolsUsed:  []string{"tool1", "tool2"},
						Duration:   2 * time.Minute,
					},
				},
			},
			check: func(s SessionSummary) bool {
				return s.TotalPrompts == 2 &&
					s.TotalOptimizations == 4 &&
					len(s.ToolsUsed) >= 1
			},
		},
	}

	provider := &mockProvider{}
	engine := NewReflectionEngine(provider)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.generateSessionSummary(tt.sessionData)
			if !tt.check(result) {
				t.Errorf("generateSessionSummary() check failed: %+v", result)
			}
		})
	}
}

func TestCalculateOverallEffectiveness(t *testing.T) {
	tests := []struct {
		name        string
		sessionData SessionData
		wantMin     float64
		wantMax     float64
	}{
		{
			name:        "empty sessions",
			sessionData: SessionData{Sessions: []OptimizationSession{}},
			wantMin:     0.0,
			wantMax:     0.0,
		},
		{
			name: "no improvement",
			sessionData: SessionData{
				Sessions: []OptimizationSession{
					{
						Iterations: []IterationResult{
							{Score: 5.0},
							{Score: 4.0},
						},
					},
				},
			},
			wantMin: 0.0,
			wantMax: 0.0,
		},
		{
			name: "with improvements",
			sessionData: SessionData{
				Sessions: []OptimizationSession{
					{
						Iterations: []IterationResult{
							{Score: 5.0},
							{Score: 8.0},
						},
					},
				},
			},
			wantMin: 0.0,
			wantMax: 1.0,
		},
	}

	provider := &mockProvider{}
	engine := NewReflectionEngine(provider)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.calculateOverallEffectiveness(tt.sessionData)
			if result < tt.wantMin || result > tt.wantMax {
				t.Errorf("calculateOverallEffectiveness() = %v, want between %v and %v", result, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestFormatSessionData(t *testing.T) {
	tests := []struct {
		name        string
		sessionData SessionData
		check       func(string) bool
	}{
		{
			name:        "empty",
			sessionData: SessionData{},
			check: func(s string) bool {
				return s == "No session data available"
			},
		},
		{
			name: "with sessions",
			sessionData: SessionData{
				Sessions: []OptimizationSession{
					{
						Iterations: []IterationResult{{}, {}},
						FinalScore: 8.5,
						Duration:   time.Minute,
					},
				},
			},
			check: func(s string) bool {
				return len(s) > 0 && s != "No session data available"
			},
		},
	}

	provider := &mockProvider{}
	engine := NewReflectionEngine(provider)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.formatSessionData(tt.sessionData)
			if !tt.check(result) {
				t.Errorf("formatSessionData() check failed: %s", result)
			}
		})
	}
}

func TestFormatSessionSummary(t *testing.T) {
	provider := &mockProvider{}
	engine := NewReflectionEngine(provider)

	sessionData := SessionData{
		Sessions: []OptimizationSession{
			{
				Iterations: []IterationResult{{Score: 5.0}, {Score: 8.0}},
				ToolsUsed:  []string{"tool1"},
			},
		},
	}

	result := engine.formatSessionSummary(sessionData)
	if len(result) == 0 {
		t.Error("formatSessionSummary() returned empty string")
	}
}

func TestFormatSuccessPatterns(t *testing.T) {
	provider := &mockProvider{}
	engine := NewReflectionEngine(provider)

	patterns := []SuccessPattern{
		{
			PatternName:   "Test Pattern",
			Effectiveness: 0.85,
			Description:   "A test pattern",
		},
	}

	result := engine.formatSuccessPatterns(patterns)
	if len(result) == 0 {
		t.Error("formatSuccessPatterns() returned empty string")
	}
}

func TestFormatTopPatterns(t *testing.T) {
	provider := &mockProvider{}
	engine := NewReflectionEngine(provider)

	patterns := []SuccessPattern{
		{PatternName: "Low", Effectiveness: 0.5},
		{PatternName: "High", Effectiveness: 0.9},
		{PatternName: "Medium", Effectiveness: 0.7},
	}

	result := engine.formatTopPatterns(patterns)
	if len(result) == 0 {
		t.Error("formatTopPatterns() returned empty string")
	}
}

func TestCountHighPriorityInsights(t *testing.T) {
	tests := []struct {
		name     string
		insights []WorkflowInsight
		want     int
	}{
		{
			name:     "empty",
			insights: []WorkflowInsight{},
			want:     0,
		},
		{
			name: "no high priority",
			insights: []WorkflowInsight{
				{Priority: "low"},
				{Priority: "medium"},
			},
			want: 0,
		},
		{
			name: "with high priority",
			insights: []WorkflowInsight{
				{Priority: "high"},
				{Priority: "low"},
				{Priority: "high"},
			},
			want: 2,
		},
	}

	provider := &mockProvider{}
	engine := NewReflectionEngine(provider)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.countHighPriorityInsights(tt.insights)
			if result != tt.want {
				t.Errorf("countHighPriorityInsights() = %d, want %d", result, tt.want)
			}
		})
	}
}

func TestParseRecommendations(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name: "numbered list",
			input: `1. First recommendation
2. Second recommendation
3. Third recommendation`,
			want: 3,
		},
		{
			name:  "no recommendations",
			input: "Some text without numbers",
			want:  0,
		},
		{
			name: "all seven recommendations",
			input: `1. One
2. Two
3. Three
4. Four
5. Five
6. Six
7. Seven`,
			want: 7,
		},
	}

	provider := &mockProvider{}
	engine := NewReflectionEngine(provider)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.parseRecommendations(tt.input)
			if len(result) != tt.want {
				t.Errorf("parseRecommendations() returned %d, want %d", len(result), tt.want)
			}
		})
	}
}

func TestGenerateDefaultPatterns(t *testing.T) {
	provider := &mockProvider{}
	engine := NewReflectionEngine(provider)

	result := engine.generateDefaultPatterns()
	if len(result) == 0 {
		t.Error("generateDefaultPatterns() returned empty slice")
	}
	if result[0].PatternID == "" {
		t.Error("Default pattern has empty PatternID")
	}
}

func TestGenerateDefaultStrategies(t *testing.T) {
	provider := &mockProvider{}
	engine := NewReflectionEngine(provider)

	result := engine.generateDefaultStrategies()
	if len(result) == 0 {
		t.Error("generateDefaultStrategies() returned empty slice")
	}
	if result[0].Scenario == "" {
		t.Error("Default strategy has empty Scenario")
	}
}

func TestGenerateDefaultKnowledge(t *testing.T) {
	provider := &mockProvider{}
	engine := NewReflectionEngine(provider)

	result := engine.generateDefaultKnowledge()
	if len(result) == 0 {
		t.Error("generateDefaultKnowledge() returned empty slice")
	}
	if result[0].ItemID == "" {
		t.Error("Default knowledge item has empty ItemID")
	}
}

func TestGenerateDefaultInsights(t *testing.T) {
	provider := &mockProvider{}
	engine := NewReflectionEngine(provider)

	result := engine.generateDefaultInsights()
	if len(result) == 0 {
		t.Error("generateDefaultInsights() returned empty slice")
	}
	if result[0].InsightID == "" {
		t.Error("Default insight has empty InsightID")
	}
}

func TestSuccessPatternStruct(t *testing.T) {
	pattern := SuccessPattern{
		PatternID:     "p1",
		PatternName:   "Test Pattern",
		Description:   "Description",
		Context:       "When testing",
		Techniques:    []string{"tech1", "tech2"},
		Effectiveness: 0.85,
		Examples:      []string{"example1"},
		Conditions:    []string{"condition1"},
		Metrics:       map[string]float64{"accuracy": 0.9},
	}

	if pattern.Effectiveness < 0 || pattern.Effectiveness > 1 {
		t.Error("Effectiveness should be between 0 and 1")
	}
}

func TestStrategyRecommendationStruct(t *testing.T) {
	strategy := StrategyRecommendation{
		Scenario:       "Test scenario",
		Recommended:    []string{"tool1", "tool2"},
		NotRecommended: []string{"tool3"},
		Rationale:      "Because reasons",
		Confidence:     0.9,
		Alternatives:   []string{"alt1"},
	}

	if strategy.Confidence < 0 || strategy.Confidence > 1 {
		t.Error("Confidence should be between 0 and 1")
	}
}

func TestKnowledgeItemStruct(t *testing.T) {
	item := KnowledgeItem{
		ItemID:       "k1",
		Category:     "principle",
		Title:        "Test Principle",
		Content:      "Content here",
		Importance:   0.9,
		Confidence:   0.85,
		Sources:      []string{"source1"},
		Applications: []string{"app1"},
	}

	if item.Category != "principle" {
		t.Error("Category mismatch")
	}
}

func TestWorkflowInsightStruct(t *testing.T) {
	insight := WorkflowInsight{
		InsightID:      "w1",
		InsightType:    "efficiency",
		Description:    "Test insight",
		CurrentState:   "Current",
		ProposedState:  "Proposed",
		Benefits:       []string{"benefit1"},
		Implementation: "How to implement",
		Priority:       "high",
	}

	if insight.Priority != "high" && insight.Priority != "medium" && insight.Priority != "low" {
		t.Error("Invalid Priority value")
	}
}

func TestReflectionResultStruct(t *testing.T) {
	result := ReflectionResult{
		SessionSummary: SessionSummary{
			TotalPrompts: 10,
		},
		SuccessPatterns:         []SuccessPattern{{}},
		StrategyRecommendations: []StrategyRecommendation{{}},
		KnowledgeBase:           []KnowledgeItem{{}},
		WorkflowInsights:        []WorkflowInsight{{}},
		OverallEffectiveness:    0.85,
		LearningsExtracted:      5,
		Recommendations:         []string{"rec1"},
		NextSteps:               []string{"step1"},
	}

	if result.LearningsExtracted != 5 {
		t.Error("LearningsExtracted mismatch")
	}
}

func TestSessionSummaryStruct(t *testing.T) {
	summary := SessionSummary{
		TotalPrompts:       10,
		TotalOptimizations: 25,
		AvgImprovement:     2.5,
		TotalDuration:      time.Hour,
		ToolsUsed:          []string{"tool1"},
		MostEffectiveTool:  "tool1",
		CommonPatterns:     []string{"pattern1"},
		KeyChallenges:      []string{"challenge1"},
	}

	if summary.TotalOptimizations < summary.TotalPrompts {
		t.Error("TotalOptimizations should typically be >= TotalPrompts")
	}
}

func TestOptimizationSessionStruct(t *testing.T) {
	session := OptimizationSession{
		SessionID:     "session_1",
		Timestamp:     time.Now(),
		InitialPrompt: "initial",
		FinalPrompt:   "final",
		ToolsUsed:     []string{"tool1"},
		Iterations:    []IterationResult{{Score: 8.0}},
		FinalScore:    8.0,
		Duration:      time.Minute,
		Metadata:      map[string]interface{}{"key": "value"},
	}

	if session.FinalScore < 0 || session.FinalScore > 10 {
		t.Error("FinalScore should be between 0 and 10")
	}
}
