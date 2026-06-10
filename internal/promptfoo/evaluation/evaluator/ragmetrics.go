package evaluator

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/pe/internal/llm"
)

// metaString reads a string value from the assertion metadata map.
func metaString(metadata map[string]interface{}, key string) string {
	if metadata == nil {
		return ""
	}
	if v, ok := metadata[key]; ok {
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
	return ""
}

// gradeWithJudge runs a model-graded rubric prompt and turns the judge's
// SCORE (0-10) into a normalized AssertionResult, sharing the parsing and
// thresholding used by llm-judge and g-eval. method labels the metric in the
// result metadata.
func (ae *AssertionEvaluator) gradeWithJudge(ctx context.Context, assertion Assertion, judgePrompt, method string, extra map[string]interface{}) (*AssertionResult, error) {
	judgeProvider, err := ae.llmJudgeProvider(assertion)
	if err != nil {
		return nil, err
	}

	response, err := judgeProvider.Generate(ctx, judgePrompt, llm.GenerateOptions{})
	if err != nil {
		return &AssertionResult{
			Type:    assertion.Type,
			Passed:  false,
			Score:   0.0,
			Message: fmt.Sprintf("Failed to evaluate %s: %v", method, err),
		}, nil
	}

	score, reasoning := ae.parseLLMJudgeResponse(response.Text)
	normalizedScore := score / 10.0

	threshold := 0.7
	if assertion.Threshold != nil {
		threshold = *assertion.Threshold
	}
	passed := normalizedScore >= threshold

	meta := map[string]interface{}{
		"method":    method,
		"reasoning": reasoning,
	}
	for k, v := range extra {
		meta[k] = v
	}
	return &AssertionResult{
		Type:     assertion.Type,
		Passed:   passed,
		Score:    normalizedScore,
		Actual:   score,
		Message:  fmt.Sprintf("%s score: %.1f/10 - %s", method, score, reasoning),
		Metadata: meta,
	}, nil
}

// evaluateAnswerRelevance grades how well the output answers the original
// question (the rendered prompt, supplied as metadata["input"]). This mirrors
// promptfoo's answer-relevance assertion and is a real model call.
func (ae *AssertionEvaluator) evaluateAnswerRelevance(ctx context.Context, assertion Assertion, output string, metadata map[string]interface{}) (*AssertionResult, error) {
	question := metaString(metadata, "input")
	if question == "" {
		return &AssertionResult{
			Type:    assertion.Type,
			Passed:  false,
			Score:   0.0,
			Message: "answer-relevance requires the question; none was available (no input in context)",
		}, nil
	}

	judgePrompt := fmt.Sprintf(`You are evaluating answer relevance.

QUESTION:
%s

ANSWER:
%s

TASK:
Rate how directly and completely the answer addresses the question, ignoring
factual correctness. A fully on-topic, complete answer scores 10; an off-topic
or evasive answer scores 0.

FORMAT (exactly):
SCORE: [0-10]
REASONING: [brief explanation]`, question, output)

	return ae.gradeWithJudge(ctx, assertion, judgePrompt, "answer-relevance", map[string]interface{}{"question": question})
}

// evaluateContextMetric grades the RAG context-* family (faithfulness, recall,
// relevance) using metadata["context"] and, for some metrics, the question.
// The rubric varies by assertion type; all share the judge scoring path.
func (ae *AssertionEvaluator) evaluateContextMetric(ctx context.Context, assertion Assertion, output string, metadata map[string]interface{}) (*AssertionResult, error) {
	contextText := metaString(metadata, "context")
	if contextText == "" {
		return &AssertionResult{
			Type:    assertion.Type,
			Passed:  false,
			Score:   0.0,
			Message: fmt.Sprintf("%s requires a 'context' var; none was provided", assertion.Type),
		}, nil
	}
	question := metaString(metadata, "input")

	var task string
	switch assertion.Type {
	case AssertionContextFaithfulness:
		task = "Rate how faithful the answer is to the context: every claim in the answer should be supported by the context. Unsupported or contradicted claims lower the score."
	case AssertionContextRecall:
		task = "Rate how well the answer recalls the information present in the context that is needed to answer the question. Missing relevant facts lower the score."
	case AssertionContextRelevance:
		task = "Rate how relevant the provided context is to the question. Off-topic or unused context lowers the score."
	default:
		task = "Rate the answer against the context."
	}

	judgePrompt := fmt.Sprintf(`You are evaluating a retrieval-augmented answer.

QUESTION:
%s

CONTEXT:
%s

ANSWER:
%s

TASK:
%s
A perfect result scores 10; a poor result scores 0.

FORMAT (exactly):
SCORE: [0-10]
REASONING: [brief explanation]`, question, contextText, output, task)

	return ae.gradeWithJudge(ctx, assertion, judgePrompt, string(assertion.Type), nil)
}
