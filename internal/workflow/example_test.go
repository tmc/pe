package workflow_test

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/pe/internal/workflow"
)

// upperCaller answers every call with the uppercased prompt.
type upperCaller struct{}

func (upperCaller) Call(_ context.Context, call workflow.Call) (workflow.Result, error) {
	return workflow.Result{Content: strings.ToUpper(call.Prompt)}, nil
}

func Example() {
	script := `
phase("Review")
findings = parallel([plan(lens + ": check the diff") for lens in ["bugs", "perf"]])
result = generate("summarize: " + ", ".join(findings))
`
	outcome, err := workflow.Run(context.Background(), "review.star", []byte(script), workflow.Options{
		Caller: upperCaller{},
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(outcome.Value)
	fmt.Println("calls:", len(outcome.Trace.Calls))
	// Output:
	// SUMMARIZE: BUGS: CHECK THE DIFF, PERF: CHECK THE DIFF
	// calls: 3
}
