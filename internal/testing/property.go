package testing

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tmc/pe/internal/llm"
)

// PropertyTest represents a property-based test
type PropertyTest struct {
	Name        string                 `yaml:"name"`
	Property    string                 `yaml:"property"`
	Generator   string                 `yaml:"generator"`
	Constraint  string                 `yaml:"constraint"`
	Iterations  int                    `yaml:"iterations"`
	Options     map[string]interface{} `yaml:"options"`
}

// PropertyTestResult contains the result of a property test
type PropertyTestResult struct {
	Name          string        `json:"name"`
	Property      string        `json:"property"`
	Passed        bool          `json:"passed"`
	Iterations    int           `json:"iterations"`
	Failures      []TestFailure `json:"failures"`
	TotalDuration time.Duration `json:"total_duration"`
	AverageLatency time.Duration `json:"average_latency"`
}

// TestFailure represents a failed test case
type TestFailure struct {
	Iteration int           `json:"iteration"`
	Input     string        `json:"input"`
	Output    string        `json:"output"`
	Reason    string        `json:"reason"`
	Latency   time.Duration `json:"latency"`
}

// PropertyTester handles property-based testing
type PropertyTester struct {
	provider llm.Provider
	options  llm.GenerateOptions
}

// NewPropertyTester creates a new property tester
func NewPropertyTester(provider llm.Provider, options llm.GenerateOptions) *PropertyTester {
	return &PropertyTester{
		provider: provider,
		options:  options,
	}
}

// RunPropertyTest executes a property-based test
func (pt *PropertyTester) RunPropertyTest(ctx context.Context, test PropertyTest) (*PropertyTestResult, error) {
	result := &PropertyTestResult{
		Name:       test.Name,
		Property:   test.Property,
		Iterations: test.Iterations,
		Failures:   []TestFailure{},
	}

	startTime := time.Now()
	var totalLatency time.Duration

	// Run iterations
	for i := 0; i < test.Iterations; i++ {
		// Generate test input
		input, err := pt.generateInput(test.Generator, test.Options)
		if err != nil {
			return nil, fmt.Errorf("error generating input for iteration %d: %v", i, err)
		}

		// Execute with provider
		iterStart := time.Now()
		response, err := pt.provider.Generate(ctx, input, pt.options)
		iterLatency := time.Since(iterStart)
		totalLatency += iterLatency

		if err != nil {
			result.Failures = append(result.Failures, TestFailure{
				Iteration: i + 1,
				Input:     input,
				Output:    "",
				Reason:    fmt.Sprintf("Provider error: %v", err),
				Latency:   iterLatency,
			})
			continue
		}

		// Check property constraint
		if !pt.checkConstraint(test.Constraint, input, response.Text, response) {
			result.Failures = append(result.Failures, TestFailure{
				Iteration: i + 1,
				Input:     input,
				Output:    response.Text,
				Reason:    fmt.Sprintf("Property violation: %s", test.Constraint),
				Latency:   iterLatency,
			})
		}
	}

	result.TotalDuration = time.Since(startTime)
	result.AverageLatency = totalLatency / time.Duration(test.Iterations)
	result.Passed = len(result.Failures) == 0

	return result, nil
}

// generateInput generates test input based on the generator specification
func (pt *PropertyTester) generateInput(generator string, options map[string]interface{}) (string, error) {
	switch generator {
	case "random_text":
		return pt.generateRandomText(options)
	case "random_question":
		return pt.generateRandomQuestion(options)
	case "random_code":
		return pt.generateRandomCode(options)
	case "random_json":
		return pt.generateRandomJSON(options)
	case "sentences":
		return pt.generateSentences(options)
	case "numbers":
		return pt.generateNumbers(options)
	default:
		return "", fmt.Errorf("unknown generator: %s", generator)
	}
}

// generateRandomText generates random text
func (pt *PropertyTester) generateRandomText(options map[string]interface{}) (string, error) {
	minLength := getIntOption(options, "min_length", 10)
	maxLength := getIntOption(options, "max_length", 100)
	
	words := []string{
		"the", "quick", "brown", "fox", "jumps", "over", "lazy", "dog",
		"artificial", "intelligence", "machine", "learning", "computer", "science",
		"algorithm", "data", "analysis", "programming", "software", "development",
		"technology", "innovation", "research", "experiment", "hypothesis", "theory",
	}
	
	length := rand.Intn(maxLength-minLength+1) + minLength
	var result []string
	
	for i := 0; i < length; i++ {
		result = append(result, words[rand.Intn(len(words))])
	}
	
	return strings.Join(result, " "), nil
}

// generateRandomQuestion generates random questions
func (pt *PropertyTester) generateRandomQuestion(options map[string]interface{}) (string, error) {
	questions := []string{
		"What is %s?",
		"How does %s work?",
		"Why is %s important?",
		"When was %s invented?",
		"Where is %s used?",
		"Who created %s?",
		"Can you explain %s?",
		"What are the benefits of %s?",
	}
	
	topics := []string{
		"artificial intelligence", "machine learning", "blockchain", "quantum computing",
		"cloud computing", "cybersecurity", "data science", "robotics", "IoT",
		"virtual reality", "augmented reality", "natural language processing",
	}
	
	questionTemplate := questions[rand.Intn(len(questions))]
	topic := topics[rand.Intn(len(topics))]
	
	return fmt.Sprintf(questionTemplate, topic), nil
}

// generateRandomCode generates random code snippets
func (pt *PropertyTester) generateRandomCode(options map[string]interface{}) (string, error) {
	language := getStringOption(options, "language", "python")
	
	templates := map[string][]string{
		"python": {
			"def %s():\n    return %s",
			"for i in range(%d):\n    print(%s)",
			"class %s:\n    def __init__(self):\n        self.%s = %s",
			"if %s:\n    %s\nelse:\n    %s",
		},
		"javascript": {
			"function %s() {\n    return %s;\n}",
			"for (let i = 0; i < %d; i++) {\n    console.log(%s);\n}",
			"class %s {\n    constructor() {\n        this.%s = %s;\n    }\n}",
		},
	}
	
	if langTemplates, exists := templates[language]; exists {
		template := langTemplates[rand.Intn(len(langTemplates))]
		return pt.fillCodeTemplate(template), nil
	}
	
	return fmt.Sprintf("# Example %s code\nprint('Hello, World!')", language), nil
}

// generateRandomJSON generates random JSON structures
func (pt *PropertyTester) generateRandomJSON(options map[string]interface{}) (string, error) {
	depth := getIntOption(options, "depth", 2)
	return pt.generateJSONObject(depth), nil
}

// generateSentences generates random sentences
func (pt *PropertyTester) generateSentences(options map[string]interface{}) (string, error) {
	count := getIntOption(options, "count", 3)
	
	sentences := []string{
		"The sun rises in the east and sets in the west.",
		"Technology continues to evolve at a rapid pace.",
		"Climate change is one of the most pressing issues of our time.",
		"Education plays a crucial role in personal development.",
		"Innovation drives progress in many industries.",
	}
	
	var result []string
	for i := 0; i < count; i++ {
		result = append(result, sentences[rand.Intn(len(sentences))])
	}
	
	return strings.Join(result, " "), nil
}

// generateNumbers generates random numbers
func (pt *PropertyTester) generateNumbers(options map[string]interface{}) (string, error) {
	min := getIntOption(options, "min", 1)
	max := getIntOption(options, "max", 100)
	count := getIntOption(options, "count", 1)
	
	var numbers []string
	for i := 0; i < count; i++ {
		num := rand.Intn(max-min+1) + min
		numbers = append(numbers, strconv.Itoa(num))
	}
	
	return strings.Join(numbers, ", "), nil
}

// fillCodeTemplate fills in a code template with random values
func (pt *PropertyTester) fillCodeTemplate(template string) string {
	// Simple template filling - in a real implementation this would be more sophisticated
	functionNames := []string{"calculate", "process", "handle", "execute", "run", "compute"}
	variableNames := []string{"data", "value", "result", "item", "element", "object"}
	values := []string{"42", "'hello'", "True", "None", "[]", "{}"}
	
	result := template
	result = strings.ReplaceAll(result, "%s", functionNames[rand.Intn(len(functionNames))])
	result = strings.ReplaceAll(result, "%d", strconv.Itoa(rand.Intn(10)+1))
	
	// Replace remaining %s with variables or values
	for strings.Contains(result, "%s") {
		if rand.Float32() < 0.5 {
			result = strings.Replace(result, "%s", variableNames[rand.Intn(len(variableNames))], 1)
		} else {
			result = strings.Replace(result, "%s", values[rand.Intn(len(values))], 1)
		}
	}
	
	return result
}

// generateJSONObject generates a random JSON object
func (pt *PropertyTester) generateJSONObject(depth int) string {
	if depth <= 0 {
		return `"value"`
	}
	
	keys := []string{"name", "id", "value", "data", "config", "settings", "info"}
	values := []string{`"string"`, `42`, `true`, `false`, `null`}
	
	var pairs []string
	numPairs := rand.Intn(3) + 1
	
	for i := 0; i < numPairs; i++ {
		key := keys[rand.Intn(len(keys))]
		var value string
		
		if rand.Float32() < 0.3 && depth > 1 {
			// Nested object
			value = pt.generateJSONObject(depth - 1)
		} else {
			value = values[rand.Intn(len(values))]
		}
		
		pairs = append(pairs, fmt.Sprintf(`"%s": %s`, key, value))
	}
	
	return fmt.Sprintf(`{%s}`, strings.Join(pairs, ", "))
}

// checkConstraint evaluates a constraint against the input and output
func (pt *PropertyTester) checkConstraint(constraint, input, output string, response *llm.GenerateResponse) bool {
	// Simple constraint evaluation - this could be much more sophisticated
	
	// Length constraints
	if strings.Contains(constraint, "response.length >") {
		re := regexp.MustCompile(`response\.length > (\d+)`)
		matches := re.FindStringSubmatch(constraint)
		if len(matches) > 1 {
			if threshold, err := strconv.Atoi(matches[1]); err == nil {
				return len(output) > threshold
			}
		}
	}
	
	if strings.Contains(constraint, "response.length <") {
		re := regexp.MustCompile(`response\.length < (\d+)`)
		matches := re.FindStringSubmatch(constraint)
		if len(matches) > 1 {
			if threshold, err := strconv.Atoi(matches[1]); err == nil {
				return len(output) < threshold
			}
		}
	}
	
	// Content constraints
	if strings.Contains(constraint, "response.contains") {
		re := regexp.MustCompile(`response\.contains\("([^"]+)"\)`)
		matches := re.FindStringSubmatch(constraint)
		if len(matches) > 1 {
			return strings.Contains(strings.ToLower(output), strings.ToLower(matches[1]))
		}
	}
	
	// Token constraints
	if strings.Contains(constraint, "response.tokens <") {
		re := regexp.MustCompile(`response\.tokens < (\d+)`)
		matches := re.FindStringSubmatch(constraint)
		if len(matches) > 1 {
			if threshold, err := strconv.Atoi(matches[1]); err == nil {
				return response.CompletionTokens < threshold
			}
		}
	}
	
	// Latency constraints
	if strings.Contains(constraint, "response.latency <") {
		re := regexp.MustCompile(`response\.latency < (\d+)`)
		matches := re.FindStringSubmatch(constraint)
		if len(matches) > 1 {
			if threshold, err := strconv.Atoi(matches[1]); err == nil {
				return response.Latency < time.Duration(threshold)*time.Millisecond
			}
		}
	}
	
	// Relation constraints
	if strings.Contains(constraint, "response.length > prompt.length") {
		inputLength := len(input)
		outputLength := len(output)
		if strings.Contains(constraint, "* 0.5") {
			return outputLength > inputLength/2
		}
		return outputLength > inputLength
	}
	
	// Default: assume constraint is met if we can't parse it
	return true
}

// Helper functions
func getIntOption(options map[string]interface{}, key string, defaultValue int) int {
	if value, exists := options[key]; exists {
		switch v := value.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return defaultValue
}

func getStringOption(options map[string]interface{}, key, defaultValue string) string {
	if value, exists := options[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}