package evaluator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tmc/pe/internal/llm"
)

// Rubric-based model-graded assertions ported from promptfoo.
//
// agent-rubric and search-rubric are, in promptfoo, llm-rubric prompts that
// REQUIRE a grader with agentic tool use (read/search/shell) or web search
// respectively. pe has no such grader, so it runs them as a plain rubric
// against the configured judge: an unmodified config validates and produces a
// score, but the score reflects the judge's own knowledge, not tool-gathered
// evidence. This degradation is deliberate and documented so configs run
// rather than erroring; it is analogous to pe honoring "similar:cosine" as a
// plain similarity check.

// evaluateRubric grades the output against the rubric in assertion.Value using
// the local judge, shared by agent-rubric and search-rubric.
func (ae *AssertionEvaluator) evaluateRubric(ctx context.Context, assertion Assertion, output string) (*AssertionResult, error) {
	rubric, ok := assertion.Value.(string)
	if !ok || strings.TrimSpace(rubric) == "" {
		return &AssertionResult{
			Type:    assertion.Type,
			Passed:  false,
			Score:   0.0,
			Message: fmt.Sprintf("%s assertion requires a rubric string value", assertion.Type),
		}, nil
	}

	judgePrompt := fmt.Sprintf(`You are an expert evaluator. Grade the output strictly against the rubric.

RUBRIC:
%s

OUTPUT TO EVALUATE:
%s

TASK:
Decide how well the output satisfies the rubric and assign a score from 0 to 10.

FORMAT (exactly):
SCORE: [0-10]
REASONING: [brief explanation]`, rubric, output)

	return ae.gradeWithJudge(ctx, assertion, judgePrompt, string(assertion.Type), map[string]interface{}{"rubric": rubric, "degraded": "tool-use not enforced"})
}

// evaluateSkillUsed asks the judge whether the agent used the required skill or
// tool described in assertion.Value, given the output and (when available) the
// recorded tool-call trajectory. promptfoo grades this with a model judge.
func (ae *AssertionEvaluator) evaluateSkillUsed(ctx context.Context, assertion Assertion, output string, metadata map[string]interface{}) (*AssertionResult, error) {
	criteria, ok := assertion.Value.(string)
	if !ok || strings.TrimSpace(criteria) == "" {
		return &AssertionResult{
			Type:    assertion.Type,
			Passed:  false,
			Score:   0.0,
			Message: "skill-used assertion requires the expected skill/tool as a string value",
		}, nil
	}
	judgePrompt := fmt.Sprintf(`You are evaluating whether an AI agent used a required skill or tool.

REQUIRED SKILL/TOOL:
%s

AGENT OUTPUT:
%s
%s
TASK:
Decide whether the agent actually used the required skill or tool. Score 10 if
it clearly did, 0 if it clearly did not.

FORMAT (exactly):
SCORE: [0-10]
REASONING: [brief explanation]`, criteria, output, trajectorySummary(metadata))

	return ae.gradeWithJudge(ctx, assertion, judgePrompt, "skill-used", map[string]interface{}{"skill": criteria})
}

// evaluateTrajectoryGoalSuccess asks the judge whether the agent achieved the
// goal in assertion.Value, given the output and the recorded trajectory.
func (ae *AssertionEvaluator) evaluateTrajectoryGoalSuccess(ctx context.Context, assertion Assertion, output string, metadata map[string]interface{}) (*AssertionResult, error) {
	goal, ok := assertion.Value.(string)
	if !ok || strings.TrimSpace(goal) == "" {
		return &AssertionResult{
			Type:    assertion.Type,
			Passed:  false,
			Score:   0.0,
			Message: "trajectory:goal-success requires the goal as a string value",
		}, nil
	}
	judgePrompt := fmt.Sprintf(`You are evaluating whether an AI agent achieved its goal.

GOAL:
%s

FINAL OUTPUT:
%s
%s
TASK:
Decide whether the agent achieved the goal. Score 10 for full success, 0 for
failure, and partial credit in between.

FORMAT (exactly):
SCORE: [0-10]
REASONING: [brief explanation]`, goal, output, trajectorySummary(metadata))

	return ae.gradeWithJudge(ctx, assertion, judgePrompt, "trajectory:goal-success", map[string]interface{}{"goal": goal})
}

// trajectorySummary renders the recorded tool-call trajectory (if any) into a
// block for inclusion in a judge prompt. It returns an empty string when no
// trajectory is available.
func trajectorySummary(metadata map[string]interface{}) string {
	steps, ok := trajectorySteps("", metadata)
	if !ok || len(steps) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\nTOOL-CALL TRAJECTORY:\n")
	for i, s := range steps {
		fmt.Fprintf(&b, "%d. %s\n", i+1, s.Name)
	}
	return b.String()
}

// conversationMessage is a single {role, content} entry in the message list
// submitted to the conversation-relevance judge.
type conversationMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// conversationVerdict is the judge's per-window JSON reply.
type conversationVerdict struct {
	Verdict string `json:"verdict"`
	Reason  string `json:"reason,omitempty"`
}

const conversationRelevanceWindowSize = 5

// evaluateConversationRelevance ports promptfoo's conversation-relevance
// (DeepEval): it builds an alternating user/assistant message list from the
// _conversation var (a list of {input, output} turns), and for each assistant
// turn asks the judge whether that assistant message is relevant given a
// sliding window of up to the preceding 5 messages. The score is the fraction
// of assistant turns judged relevant; it passes when score >= threshold
// (default 0, matching promptfoo). When _conversation is absent it degrades to
// a single turn from the rendered input and the output.
func (ae *AssertionEvaluator) evaluateConversationRelevance(ctx context.Context, assertion Assertion, output string, metadata map[string]interface{}) (*AssertionResult, error) {
	messages := buildConversationMessages(metadata, output)
	if len(messages) == 0 {
		return &AssertionResult{
			Type:    assertion.Type,
			Passed:  false,
			Score:   0.0,
			Message: "conversation-relevance has no messages to evaluate",
		}, nil
	}

	judgeProvider, err := ae.llmJudgeProvider(assertion)
	if err != nil {
		return nil, err
	}

	total, relevant := 0, 0
	var lastReason string
	for i, msg := range messages {
		if msg.Role != "assistant" {
			continue
		}
		total++
		lo := i - conversationRelevanceWindowSize + 1
		if lo < 0 {
			lo = 0
		}
		window := messages[lo : i+1]
		verdict, reason := ae.conversationVerdict(ctx, judgeProvider, window)
		if verdict {
			relevant++
		} else if reason != "" {
			lastReason = reason
		}
	}

	if total == 0 {
		return &AssertionResult{
			Type:    assertion.Type,
			Passed:  false,
			Score:   0.0,
			Message: "conversation-relevance found no assistant turns to evaluate",
		}, nil
	}

	score := float64(relevant) / float64(total)
	threshold := 0.0
	if assertion.Threshold != nil {
		threshold = *assertion.Threshold
	}
	passed := score >= threshold
	message := fmt.Sprintf("conversation-relevance %d/%d windows relevant (score %.3f, threshold %.2f)", relevant, total, score, threshold)
	if !passed && lastReason != "" {
		message += " - " + lastReason
	}
	return &AssertionResult{
		Type:    assertion.Type,
		Passed:  passed,
		Score:   score,
		Message: message,
		Metadata: map[string]interface{}{
			"method":   "conversation-relevance",
			"windows":  total,
			"relevant": relevant,
		},
	}, nil
}

// conversationVerdict asks the judge whether the last assistant message in the
// window is relevant, returning the boolean verdict and an optional reason. It
// mirrors promptfoo's verdict prompt and tolerates judges that wrap the JSON in
// prose by extracting the first JSON object.
func (ae *AssertionEvaluator) conversationVerdict(ctx context.Context, judge llm.Provider, window []conversationMessage) (bool, string) {
	windowJSON, err := json.MarshalIndent(window, "", "  ")
	if err != nil {
		return false, ""
	}
	prompt := fmt.Sprintf(`Based on the given list of message exchanges between a user and an LLM, generate a JSON object to indicate whether the LAST `+"`assistant`"+` message is relevant to context in messages. The JSON will have 2 fields: 'verdict' and 'reason'. The 'verdict' key should STRICTLY be either 'yes' or 'no', which states whether the last `+"`assistant`"+` message is relevant according to the context in messages. Provide a 'reason' ONLY if the answer is 'no'. You MUST USE the previous messages (if any) provided in the list of messages to make an informed judgement on relevancy.

You MUST ONLY provide a verdict for the LAST message on the list but MUST USE context from the previous messages. You DON'T have to provide a reason if the answer is 'yes'. ONLY provide a 'no' answer if the LLM response is COMPLETELY irrelevant to the message input. Vague LLM responses to vague inputs, such as greetings DOES NOT count as irrelevancies!

Messages: %s

JSON:`, string(windowJSON))

	response, err := judge.Generate(ctx, prompt, llm.GenerateOptions{})
	if err != nil {
		return false, ""
	}
	v := parseConversationVerdict(response.Text)
	return strings.EqualFold(strings.TrimSpace(v.Verdict), "yes"), strings.TrimSpace(v.Reason)
}

// parseConversationVerdict extracts the {verdict, reason} object from a judge
// reply, tolerating leading/trailing prose around the JSON.
func parseConversationVerdict(text string) conversationVerdict {
	var v conversationVerdict
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start >= 0 && end > start {
		if err := json.Unmarshal([]byte(text[start:end+1]), &v); err == nil {
			return v
		}
	}
	// Fall back to a substring check so a bare "yes"/"no" still scores.
	if strings.Contains(strings.ToLower(text), "\"verdict\": \"yes\"") || strings.EqualFold(strings.TrimSpace(text), "yes") {
		v.Verdict = "yes"
	}
	return v
}

// buildConversationMessages converts the _conversation var into an alternating
// user/assistant message list. Each turn contributes a user message (its input)
// and an assistant message (its output). When _conversation is missing or
// empty, it degrades to a single turn from metadata["input"] and the output.
func buildConversationMessages(metadata map[string]interface{}, output string) []conversationMessage {
	var messages []conversationMessage
	if metadata != nil {
		if turns, ok := metadata["_conversation"].([]interface{}); ok {
			for _, t := range turns {
				turn, ok := t.(map[string]interface{})
				if !ok {
					continue
				}
				if in, ok := turn["input"]; ok {
					messages = append(messages, conversationMessage{Role: "user", Content: stringifyTurn(in)})
				}
				if out, ok := turn["output"]; ok {
					messages = append(messages, conversationMessage{Role: "assistant", Content: stringifyTurn(out)})
				}
			}
		}
	}
	if len(messages) == 0 {
		messages = []conversationMessage{
			{Role: "user", Content: metaString(metadata, "input")},
			{Role: "assistant", Content: output},
		}
	}
	return messages
}

// stringifyTurn renders a conversation turn value; objects are JSON-encoded as
// promptfoo does, scalars are formatted directly.
func stringifyTurn(v interface{}) string {
	switch s := v.(type) {
	case string:
		return s
	default:
		if b, err := json.Marshal(v); err == nil {
			return string(b)
		}
		return fmt.Sprintf("%v", v)
	}
}
