package workflow

// This file contains preset workflow configurations for common use cases.
// These presets can be used as-is or as starting points for custom workflows.

// TodoWorkflow provides a workflow configuration suitable for todo/task management applications.
// It includes a "completion" dimension with "pending" and "done" states, where:
// - "active" context shows only pending items (for position paths)
// - "all" context shows both pending and done items (for complete listing)
// - Bottom-up completion is enabled (when all children are done, parent becomes done)
var TodoWorkflow = WorkflowConfig{
	Dimensions: []StatusDimension{
		{
			Name:         "completion",
			Values:       []string{"pending", "done"},
			DefaultValue: "pending",
		},
	},
	Visibility: map[string][]VisibilityRule{
		"active": {
			{
				Context:   "active",
				Dimension: "completion",
				Include:   []string{"pending"},
			},
		},
		"all": {
			{
				Context:   "all",
				Dimension: "completion",
				Include:   []string{"pending", "done"},
			},
		},
	},
	Transitions: map[string][]TransitionRule{
		"completion": {
			{
				Dimension: "completion",
				From:      "pending",
				To:        []string{"done"},
			},
			{
				Dimension: "completion",
				From:      "done",
				To:        []string{"pending"},
			},
		},
	},
	AutoTransitions: []AutoTransitionRule{
		{
			Trigger:         "status_change",
			Condition:       "all_children_status_equals",
			ConditionValue:  "done",
			TargetDimension: "completion",
			Action:          "set_status",
			ActionValue:     "done",
		},
	},
}

// TodoWithPriorityWorkflow extends the basic todo workflow with priority management.
// It adds a "priority" dimension with low/medium/high values that are visible in all contexts.
var TodoWithPriorityWorkflow = WorkflowConfig{
	Dimensions: []StatusDimension{
		{
			Name:         "completion",
			Values:       []string{"pending", "done"},
			DefaultValue: "pending",
		},
		{
			Name:         "priority",
			Values:       []string{"low", "medium", "high"},
			DefaultValue: "medium",
		},
	},
	Visibility: map[string][]VisibilityRule{
		"active": {
			{
				Context:   "active",
				Dimension: "completion",
				Include:   []string{"pending"},
			},
		},
		"all": {
			{
				Context:   "all",
				Dimension: "completion",
				Include:   []string{"pending", "done"},
			},
		},
		"high_priority": {
			{
				Context:   "high_priority",
				Dimension: "completion",
				Include:   []string{"pending"},
			},
			{
				Context:   "high_priority",
				Dimension: "priority",
				Include:   []string{"high"},
			},
		},
	},
	Transitions: map[string][]TransitionRule{
		"completion": {
			{
				Dimension: "completion",
				From:      "pending",
				To:        []string{"done"},
			},
			{
				Dimension: "completion",
				From:      "done",
				To:        []string{"pending"},
			},
		},
		"priority": {
			{
				Dimension: "priority",
				From:      "low",
				To:        []string{"medium", "high"},
			},
			{
				Dimension: "priority",
				From:      "medium",
				To:        []string{"low", "high"},
			},
			{
				Dimension: "priority",
				From:      "high",
				To:        []string{"low", "medium"},
			},
		},
	},
	AutoTransitions: []AutoTransitionRule{
		{
			Trigger:         "status_change",
			Condition:       "all_children_status_equals",
			ConditionValue:  "done",
			TargetDimension: "completion",
			Action:          "set_status",
			ActionValue:     "done",
		},
	},
}

// CMSWorkflow provides a workflow configuration suitable for content management systems.
// It includes publication status and priority dimensions with context-aware visibility.
var CMSWorkflow = WorkflowConfig{
	Dimensions: []StatusDimension{
		{
			Name:         "publication",
			Values:       []string{"draft", "review", "published", "archived"},
			DefaultValue: "draft",
		},
		{
			Name:         "priority",
			Values:       []string{"low", "normal", "high", "urgent"},
			DefaultValue: "normal",
		},
	},
	Visibility: map[string][]VisibilityRule{
		"public": {
			{
				Context:   "public",
				Dimension: "publication",
				Include:   []string{"published"},
			},
		},
		"editorial": {
			{
				Context:   "editorial",
				Dimension: "publication",
				Include:   []string{"draft", "review", "published"},
			},
		},
		"admin": {
			{
				Context:   "admin",
				Dimension: "publication",
				Include:   []string{"draft", "review", "published", "archived"},
			},
		},
		"review_queue": {
			{
				Context:   "review_queue",
				Dimension: "publication",
				Include:   []string{"review"},
			},
		},
	},
	Transitions: map[string][]TransitionRule{
		"publication": {
			{
				Dimension: "publication",
				From:      "draft",
				To:        []string{"review", "archived"},
			},
			{
				Dimension: "publication",
				From:      "review",
				To:        []string{"draft", "published", "archived"},
			},
			{
				Dimension: "publication",
				From:      "published",
				To:        []string{"archived"},
			},
			{
				Dimension: "publication",
				From:      "archived",
				To:        []string{"draft"},
			},
		},
		"priority": {
			{
				Dimension: "priority",
				From:      "low",
				To:        []string{"normal", "high", "urgent"},
			},
			{
				Dimension: "priority",
				From:      "normal",
				To:        []string{"low", "high", "urgent"},
			},
			{
				Dimension: "priority",
				From:      "high",
				To:        []string{"low", "normal", "urgent"},
			},
			{
				Dimension: "priority",
				From:      "urgent",
				To:        []string{"low", "normal", "high"},
			},
		},
	},
	AutoTransitions: []AutoTransitionRule{
		// No auto-transitions for CMS - publication decisions are manual
	},
}

// IssueTrackerWorkflow provides a workflow configuration suitable for issue/bug tracking systems.
// It includes state management with typical issue lifecycle states.
var IssueTrackerWorkflow = WorkflowConfig{
	Dimensions: []StatusDimension{
		{
			Name:         "state",
			Values:       []string{"open", "in_progress", "resolved", "closed", "reopened"},
			DefaultValue: "open",
		},
		{
			Name:         "priority",
			Values:       []string{"trivial", "minor", "major", "critical", "blocker"},
			DefaultValue: "major",
		},
		{
			Name:         "type",
			Values:       []string{"bug", "feature", "task", "epic"},
			DefaultValue: "bug",
		},
	},
	Visibility: map[string][]VisibilityRule{
		"active": {
			{
				Context:   "active",
				Dimension: "state",
				Include:   []string{"open", "in_progress", "reopened"},
			},
		},
		"kanban_board": {
			{
				Context:   "kanban_board",
				Dimension: "state",
				Include:   []string{"open", "in_progress"},
			},
		},
		"resolved": {
			{
				Context:   "resolved",
				Dimension: "state",
				Include:   []string{"resolved"},
			},
		},
		"all": {
			{
				Context:   "all",
				Dimension: "state",
				Include:   []string{"open", "in_progress", "resolved", "closed", "reopened"},
			},
		},
		"critical_issues": {
			{
				Context:   "critical_issues",
				Dimension: "state",
				Include:   []string{"open", "in_progress", "reopened"},
			},
			{
				Context:   "critical_issues",
				Dimension: "priority",
				Include:   []string{"critical", "blocker"},
			},
		},
	},
	Transitions: map[string][]TransitionRule{
		"state": {
			{
				Dimension: "state",
				From:      "open",
				To:        []string{"in_progress", "resolved", "closed"},
			},
			{
				Dimension: "state",
				From:      "in_progress",
				To:        []string{"open", "resolved", "closed"},
			},
			{
				Dimension: "state",
				From:      "resolved",
				To:        []string{"closed", "reopened"},
			},
			{
				Dimension: "state",
				From:      "closed",
				To:        []string{"reopened"},
			},
			{
				Dimension: "state",
				From:      "reopened",
				To:        []string{"in_progress", "resolved", "closed"},
			},
		},
		"priority": {
			{
				Dimension: "priority",
				From:      "trivial",
				To:        []string{"minor", "major", "critical", "blocker"},
			},
			{
				Dimension: "priority",
				From:      "minor",
				To:        []string{"trivial", "major", "critical", "blocker"},
			},
			{
				Dimension: "priority",
				From:      "major",
				To:        []string{"trivial", "minor", "critical", "blocker"},
			},
			{
				Dimension: "priority",
				From:      "critical",
				To:        []string{"trivial", "minor", "major", "blocker"},
			},
			{
				Dimension: "priority",
				From:      "blocker",
				To:        []string{"trivial", "minor", "major", "critical"},
			},
		},
		"type": {
			{
				Dimension: "type",
				From:      "bug",
				To:        []string{"feature", "task", "epic"},
			},
			{
				Dimension: "type",
				From:      "feature",
				To:        []string{"bug", "task", "epic"},
			},
			{
				Dimension: "type",
				From:      "task",
				To:        []string{"bug", "feature", "epic"},
			},
			{
				Dimension: "type",
				From:      "epic",
				To:        []string{"bug", "feature", "task"},
			},
		},
	},
	AutoTransitions: []AutoTransitionRule{
		// Auto-close issues when all subtasks are resolved
		{
			Trigger:         "status_change",
			Condition:       "all_children_status_equals",
			ConditionValue:  "resolved",
			TargetDimension: "state",
			Action:          "set_status",
			ActionValue:     "resolved",
		},
	},
}

// KanbanWorkflow provides a simple workflow for kanban-style task management.
// It focuses on work-in-progress limits and flow through different stages.
var KanbanWorkflow = WorkflowConfig{
	Dimensions: []StatusDimension{
		{
			Name:         "stage",
			Values:       []string{"backlog", "todo", "in_progress", "review", "done"},
			DefaultValue: "backlog",
		},
		{
			Name:         "size",
			Values:       []string{"xs", "s", "m", "l", "xl"},
			DefaultValue: "m",
		},
	},
	Visibility: map[string][]VisibilityRule{
		"board": {
			{
				Context:   "board",
				Dimension: "stage",
				Include:   []string{"todo", "in_progress", "review", "done"},
			},
		},
		"backlog": {
			{
				Context:   "backlog",
				Dimension: "stage",
				Include:   []string{"backlog"},
			},
		},
		"active": {
			{
				Context:   "active",
				Dimension: "stage",
				Include:   []string{"todo", "in_progress", "review"},
			},
		},
		"wip": {
			{
				Context:   "wip",
				Dimension: "stage",
				Include:   []string{"in_progress"},
			},
		},
	},
	Transitions: map[string][]TransitionRule{
		"stage": {
			{
				Dimension: "stage",
				From:      "backlog",
				To:        []string{"todo"},
			},
			{
				Dimension: "stage",
				From:      "todo",
				To:        []string{"in_progress", "backlog"},
			},
			{
				Dimension: "stage",
				From:      "in_progress",
				To:        []string{"review", "todo"},
			},
			{
				Dimension: "stage",
				From:      "review",
				To:        []string{"done", "in_progress"},
			},
			{
				Dimension: "stage",
				From:      "done",
				To:        []string{"review"}, // Allow reopening
			},
		},
		"size": {
			// Size can be changed freely
			{
				Dimension: "size",
				From:      "xs",
				To:        []string{"s", "m", "l", "xl"},
			},
			{
				Dimension: "size",
				From:      "s",
				To:        []string{"xs", "m", "l", "xl"},
			},
			{
				Dimension: "size",
				From:      "m",
				To:        []string{"xs", "s", "l", "xl"},
			},
			{
				Dimension: "size",
				From:      "l",
				To:        []string{"xs", "s", "m", "xl"},
			},
			{
				Dimension: "size",
				From:      "xl",
				To:        []string{"xs", "s", "m", "l"},
			},
		},
	},
	AutoTransitions: []AutoTransitionRule{
		// No auto-transitions for kanban - progression is manual
	},
}