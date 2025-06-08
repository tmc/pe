package templates

import "time"

// getBuiltinTemplates returns a collection of built-in templates
func getBuiltinTemplates() []*Template {
	return []*Template{
		{
			Name:        "summarization",
			Description: "Summarize text content into key points",
			Category:    "text-processing",
			Tags:        []string{"summary", "text", "analysis"},
			Author:      "PE Team",
			Version:     "1.0.0",
			License:     "MIT",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Prompt: `Summarize the following text into {{max_points}} key points:

{{text}}

Please provide a clear, concise summary that captures the main ideas.`,
			Variables: map[string]Variable{
				"text": {
					Name:        "text",
					Description: "Text to summarize",
					Type:        "string",
					Required:    true,
					Examples:    []string{"Long article content...", "Research paper abstract..."},
				},
				"max_points": {
					Name:        "max_points",
					Description: "Maximum number of summary points",
					Type:        "number",
					Required:    false,
					Default:     5,
					Examples:    []string{"3", "5", "10"},
					Validation: Validation{
						Min: 1,
						Max: 20,
					},
				},
			},
			Examples: []Example{
				{
					Name:        "Article Summary",
					Description: "Summarize a news article",
					Variables: map[string]interface{}{
						"text":       "Artificial intelligence has been making significant breakthroughs...",
						"max_points": 3,
					},
					Expected: "Key points about AI breakthroughs",
				},
			},
			Providers: []string{"openai:gpt-4", "anthropic:claude-3-sonnet", "openai:gpt-3.5-turbo"},
		},

		{
			Name:        "code-review",
			Description: "Perform a comprehensive code review",
			Category:    "development",
			Tags:        []string{"code", "review", "programming", "quality"},
			Author:      "PE Team",
			Version:     "1.0.0",
			License:     "MIT",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Prompt:      "Please review the following {{language}} code and provide feedback on:\n\n1. Code quality and style\n2. Potential bugs or issues\n3. Performance considerations\n4. Security concerns\n5. Suggestions for improvement\n\nCode:\n```{{language}}\n{{code}}\n```\n\nFocus areas: {{focus_areas}}",
			Variables: map[string]Variable{
				"code": {
					Name:        "code",
					Description: "Code to review",
					Type:        "string",
					Required:    true,
					Examples:    []string{"function example() { return 'hello'; }"},
				},
				"language": {
					Name:        "language",
					Description: "Programming language",
					Type:        "string",
					Required:    true,
					Examples:    []string{"javascript", "python", "go", "java"},
					Validation: Validation{
						Options: []string{"javascript", "python", "go", "java", "c++", "rust", "typescript"},
					},
				},
				"focus_areas": {
					Name:        "focus_areas",
					Description: "Specific areas to focus on",
					Type:        "string",
					Required:    false,
					Default:     "general code quality",
					Examples:    []string{"security", "performance", "maintainability"},
				},
			},
			Examples: []Example{
				{
					Name:        "JavaScript Function Review",
					Description: "Review a JavaScript function",
					Variables: map[string]interface{}{
						"code":        "function processData(data) { for(var i=0; i<data.length; i++) { console.log(data[i]); } }",
						"language":    "javascript",
						"focus_areas": "performance and modern JavaScript practices",
					},
					Expected: "Code review with suggestions for modern JavaScript",
				},
			},
			Providers: []string{"openai:gpt-4", "anthropic:claude-3-sonnet"},
		},

		{
			Name:        "creative-writing",
			Description: "Generate creative content based on prompts",
			Category:    "creative",
			Tags:        []string{"writing", "story", "creative", "fiction"},
			Author:      "PE Team",
			Version:     "1.0.0",
			License:     "MIT",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Prompt: `Write a {{genre}} {{content_type}} with the following elements:

Setting: {{setting}}
Main character: {{character}}
Theme: {{theme}}
Tone: {{tone}}
Length: {{length}}

Additional requirements: {{requirements}}`,
			Variables: map[string]Variable{
				"genre": {
					Name:        "genre",
					Description: "Genre of the story",
					Type:        "string",
					Required:    true,
					Examples:    []string{"science fiction", "fantasy", "mystery", "romance"},
					Validation: Validation{
						Options: []string{"science fiction", "fantasy", "mystery", "romance", "thriller", "horror", "comedy", "drama"},
					},
				},
				"content_type": {
					Name:        "content_type",
					Description: "Type of content to generate",
					Type:        "string",
					Required:    true,
					Examples:    []string{"short story", "poem", "dialogue"},
					Validation: Validation{
						Options: []string{"short story", "poem", "dialogue", "character description", "scene"},
					},
				},
				"setting": {
					Name:        "setting",
					Description: "Setting or location",
					Type:        "string",
					Required:    true,
					Examples:    []string{"futuristic city", "medieval castle", "modern office"},
				},
				"character": {
					Name:        "character",
					Description: "Main character description",
					Type:        "string",
					Required:    true,
					Examples:    []string{"a young detective", "an alien diplomat", "a struggling artist"},
				},
				"theme": {
					Name:        "theme",
					Description: "Central theme or message",
					Type:        "string",
					Required:    false,
					Default:     "personal growth",
					Examples:    []string{"redemption", "love conquers all", "technology vs humanity"},
				},
				"tone": {
					Name:        "tone",
					Description: "Tone of the writing",
					Type:        "string",
					Required:    false,
					Default:     "engaging",
					Examples:    []string{"dark", "humorous", "mysterious", "uplifting"},
				},
				"length": {
					Name:        "length",
					Description: "Desired length",
					Type:        "string",
					Required:    false,
					Default:     "500-800 words",
					Examples:    []string{"100 words", "500-800 words", "1000+ words"},
				},
				"requirements": {
					Name:        "requirements",
					Description: "Additional requirements or constraints",
					Type:        "string",
					Required:    false,
					Default:     "none",
					Examples:    []string{"include a twist ending", "use first person narrative", "avoid violence"},
				},
			},
			Examples: []Example{
				{
					Name:        "Sci-Fi Short Story",
					Description: "Generate a science fiction short story",
					Variables: map[string]interface{}{
						"genre":        "science fiction",
						"content_type": "short story",
						"setting":      "space station orbiting Mars",
						"character":    "a maintenance engineer with telepathic abilities",
						"theme":        "isolation and connection",
						"tone":         "mysterious",
						"length":       "800 words",
						"requirements": "include a revelation about the character's past",
					},
					Expected: "A compelling sci-fi story with the specified elements",
				},
			},
			Providers: []string{"openai:gpt-4", "anthropic:claude-3-sonnet", "anthropic:claude-3-opus"},
		},

		{
			Name:        "data-analysis",
			Description: "Analyze and interpret data sets",
			Category:    "analytics",
			Tags:        []string{"data", "analysis", "statistics", "insights"},
			Author:      "PE Team",
			Version:     "1.0.0",
			License:     "MIT",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Prompt: `Analyze the following {{data_type}} data and provide insights:

Data:
{{data}}

Please provide:
1. Summary statistics
2. Key patterns and trends
3. Notable anomalies or outliers
4. {{analysis_focus}}
5. Actionable recommendations

Format: {{output_format}}`,
			Variables: map[string]Variable{
				"data": {
					Name:        "data",
					Description: "Data to analyze",
					Type:        "string",
					Required:    true,
					Examples:    []string{"CSV data", "JSON dataset", "text with numbers"},
				},
				"data_type": {
					Name:        "data_type",
					Description: "Type of data being analyzed",
					Type:        "string",
					Required:    true,
					Examples:    []string{"sales", "user behavior", "financial", "survey"},
				},
				"analysis_focus": {
					Name:        "analysis_focus",
					Description: "Specific focus for the analysis",
					Type:        "string",
					Required:    false,
					Default:     "business implications",
					Examples:    []string{"predictive insights", "performance metrics", "risk assessment"},
				},
				"output_format": {
					Name:        "output_format",
					Description: "Preferred output format",
					Type:        "string",
					Required:    false,
					Default:     "structured report",
					Examples:    []string{"bullet points", "detailed report", "executive summary"},
					Validation: Validation{
						Options: []string{"bullet points", "detailed report", "executive summary", "structured report"},
					},
				},
			},
			Examples: []Example{
				{
					Name:        "Sales Data Analysis",
					Description: "Analyze quarterly sales data",
					Variables: map[string]interface{}{
						"data":           "Q1: $120k, Q2: $135k, Q3: $142k, Q4: $158k",
						"data_type":      "sales",
						"analysis_focus": "growth trends and forecasting",
						"output_format":  "executive summary",
					},
					Expected: "Analysis of sales trends with growth insights",
				},
			},
			Providers: []string{"openai:gpt-4", "anthropic:claude-3-sonnet"},
		},

		{
			Name:        "email-composer",
			Description: "Compose professional emails for various purposes",
			Category:    "communication",
			Tags:        []string{"email", "communication", "professional", "business"},
			Author:      "PE Team",
			Version:     "1.0.0",
			License:     "MIT",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Prompt: `Compose a {{email_type}} email with the following details:

To: {{recipient}}
Subject: {{subject}}
Purpose: {{purpose}}
Tone: {{tone}}
Key points to include: {{key_points}}

{{#if context}}
Context: {{context}}
{{/if}}

Please write a well-structured, professional email that is {{length}} in length.`,
			Variables: map[string]Variable{
				"email_type": {
					Name:        "email_type",
					Description: "Type of email",
					Type:        "string",
					Required:    true,
					Examples:    []string{"business inquiry", "follow-up", "request", "proposal"},
					Validation: Validation{
						Options: []string{"business inquiry", "follow-up", "request", "proposal", "apology", "invitation", "announcement"},
					},
				},
				"recipient": {
					Name:        "recipient",
					Description: "Email recipient",
					Type:        "string",
					Required:    true,
					Examples:    []string{"client", "colleague", "manager", "vendor"},
				},
				"subject": {
					Name:        "subject",
					Description: "Email subject line",
					Type:        "string",
					Required:    true,
					Examples:    []string{"Meeting Request", "Project Update", "Proposal Submission"},
				},
				"purpose": {
					Name:        "purpose",
					Description: "Main purpose of the email",
					Type:        "string",
					Required:    true,
					Examples:    []string{"schedule a meeting", "provide project update", "request information"},
				},
				"tone": {
					Name:        "tone",
					Description: "Email tone",
					Type:        "string",
					Required:    false,
					Default:     "professional",
					Examples:    []string{"formal", "professional", "friendly", "urgent"},
					Validation: Validation{
						Options: []string{"formal", "professional", "friendly", "urgent", "apologetic"},
					},
				},
				"key_points": {
					Name:        "key_points",
					Description: "Key points to include",
					Type:        "string",
					Required:    true,
					Examples:    []string{"project deadline, budget constraints", "meeting agenda items"},
				},
				"context": {
					Name:        "context",
					Description: "Additional context or background",
					Type:        "string",
					Required:    false,
					Examples:    []string{"previous conversation", "related project details"},
				},
				"length": {
					Name:        "length",
					Description: "Desired email length",
					Type:        "string",
					Required:    false,
					Default:     "concise",
					Examples:    []string{"brief", "concise", "detailed"},
					Validation: Validation{
						Options: []string{"brief", "concise", "detailed"},
					},
				},
			},
			Examples: []Example{
				{
					Name:        "Meeting Request",
					Description: "Request a meeting with a client",
					Variables: map[string]interface{}{
						"email_type": "business inquiry",
						"recipient":  "potential client",
						"subject":    "Partnership Discussion Meeting",
						"purpose":    "discuss potential partnership opportunities",
						"tone":       "professional",
						"key_points": "mutual benefits, timeline, next steps",
						"length":     "concise",
					},
					Expected: "Professional meeting request email",
				},
			},
			Providers: []string{"openai:gpt-4", "anthropic:claude-3-sonnet", "openai:gpt-3.5-turbo"},
		},

		{
			Name:        "learning-assistant",
			Description: "Create educational content and explanations",
			Category:    "education",
			Tags:        []string{"learning", "education", "teaching", "explanation"},
			Author:      "PE Team",
			Version:     "1.0.0",
			License:     "MIT",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Prompt: `Explain {{topic}} to someone with {{knowledge_level}} knowledge in {{subject_area}}.

Learning objectives:
{{learning_objectives}}

Please provide:
1. Clear, step-by-step explanation
2. Relevant examples
3. Key concepts to remember
4. {{assessment_type}}

Teaching style: {{teaching_style}}
Estimated time: {{time_estimate}}`,
			Variables: map[string]Variable{
				"topic": {
					Name:        "topic",
					Description: "Topic to explain",
					Type:        "string",
					Required:    true,
					Examples:    []string{"machine learning", "photosynthesis", "supply and demand"},
				},
				"subject_area": {
					Name:        "subject_area",
					Description: "Subject area or field",
					Type:        "string",
					Required:    true,
					Examples:    []string{"computer science", "biology", "economics", "physics"},
				},
				"knowledge_level": {
					Name:        "knowledge_level",
					Description: "Student's knowledge level",
					Type:        "string",
					Required:    true,
					Examples:    []string{"beginner", "intermediate", "advanced"},
					Validation: Validation{
						Options: []string{"beginner", "intermediate", "advanced"},
					},
				},
				"learning_objectives": {
					Name:        "learning_objectives",
					Description: "What the student should learn",
					Type:        "string",
					Required:    true,
					Examples:    []string{"understand basic concepts", "apply knowledge to solve problems"},
				},
				"teaching_style": {
					Name:        "teaching_style",
					Description: "Preferred teaching approach",
					Type:        "string",
					Required:    false,
					Default:     "interactive",
					Examples:    []string{"visual", "interactive", "hands-on", "theoretical"},
					Validation: Validation{
						Options: []string{"visual", "interactive", "hands-on", "theoretical", "storytelling"},
					},
				},
				"assessment_type": {
					Name:        "assessment_type",
					Description: "Type of assessment or practice",
					Type:        "string",
					Required:    false,
					Default:     "practice questions",
					Examples:    []string{"quiz questions", "practice exercises", "real-world scenarios"},
				},
				"time_estimate": {
					Name:        "time_estimate",
					Description: "Estimated learning time",
					Type:        "string",
					Required:    false,
					Default:     "15-20 minutes",
					Examples:    []string{"10 minutes", "30 minutes", "1 hour"},
				},
			},
			Examples: []Example{
				{
					Name:        "Beginner Programming",
					Description: "Explain variables to a programming beginner",
					Variables: map[string]interface{}{
						"topic":               "variables in programming",
						"subject_area":        "computer science",
						"knowledge_level":     "beginner",
						"learning_objectives": "understand what variables are and how to use them",
						"teaching_style":      "interactive",
						"assessment_type":     "simple coding exercises",
						"time_estimate":       "20 minutes",
					},
					Expected: "Clear explanation with examples and exercises",
				},
			},
			Providers: []string{"openai:gpt-4", "anthropic:claude-3-sonnet"},
		},
	}
}
