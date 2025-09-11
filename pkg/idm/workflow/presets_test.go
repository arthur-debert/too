package workflow

import (
	"testing"
)

func TestPresetConfigurations_Validate(t *testing.T) {
	presets := map[string]WorkflowConfig{
		"TodoWorkflow":             TodoWorkflow,
		"TodoWithPriorityWorkflow": TodoWithPriorityWorkflow,
		"CMSWorkflow":              CMSWorkflow,
		"IssueTrackerWorkflow":     IssueTrackerWorkflow,
		"KanbanWorkflow":           KanbanWorkflow,
	}

	for name, config := range presets {
		t.Run(name, func(t *testing.T) {
			if err := config.Validate(); err != nil {
				t.Errorf("Preset %s validation failed: %v", name, err)
			}
		})
	}
}

func TestTodoWorkflow_BasicFunctionality(t *testing.T) {
	config := TodoWorkflow

	t.Run("has completion dimension", func(t *testing.T) {
		dim := config.GetDimension("completion")
		if dim == nil {
			t.Fatal("Missing completion dimension")
		}
		if dim.DefaultValue != "pending" {
			t.Errorf("Expected default 'pending', got %q", dim.DefaultValue)
		}
	})

	t.Run("active context shows only pending", func(t *testing.T) {
		rules := config.Visibility["active"]
		if len(rules) == 0 {
			t.Fatal("No rules for active context")
		}

		rule := rules[0]
		if rule.Context != "active" || rule.Dimension != "completion" {
			t.Error("Invalid active context rule")
		}

		// Test visibility
		pendingStatuses := map[string]string{"completion": "pending"}
		doneStatuses := map[string]string{"completion": "done"}

		if !rule.Matches("active", pendingStatuses) {
			t.Error("Pending items should be visible in active context")
		}
		if rule.Matches("active", doneStatuses) {
			t.Error("Done items should not be visible in active context")
		}
	})

	t.Run("all context shows both pending and done", func(t *testing.T) {
		rules := config.Visibility["all"]
		if len(rules) == 0 {
			t.Fatal("No rules for all context")
		}

		rule := rules[0]
		pendingStatuses := map[string]string{"completion": "pending"}
		doneStatuses := map[string]string{"completion": "done"}

		if !rule.Matches("all", pendingStatuses) {
			t.Error("Pending items should be visible in all context")
		}
		if !rule.Matches("all", doneStatuses) {
			t.Error("Done items should be visible in all context")
		}
	})

	t.Run("valid transitions", func(t *testing.T) {
		transitionRules := config.Transitions["completion"]
		if len(transitionRules) != 2 {
			t.Fatalf("Expected 2 transition rules, got %d", len(transitionRules))
		}

		// Find pending->done rule
		var pendingToDone *TransitionRule
		for i, rule := range transitionRules {
			if rule.From == "pending" {
				pendingToDone = &transitionRules[i]
				break
			}
		}

		if pendingToDone == nil {
			t.Fatal("Missing pending->done transition rule")
		}

		if !pendingToDone.CanTransition("pending", "done") {
			t.Error("Should allow pending->done transition")
		}
		if pendingToDone.CanTransition("pending", "invalid") {
			t.Error("Should not allow pending->invalid transition")
		}
	})

	t.Run("has auto-transition rule", func(t *testing.T) {
		if len(config.AutoTransitions) != 1 {
			t.Fatalf("Expected 1 auto-transition rule, got %d", len(config.AutoTransitions))
		}

		rule := config.AutoTransitions[0]
		if rule.Condition != "all_children_status_equals" || rule.ConditionValue != "done" {
			t.Error("Invalid auto-transition rule condition")
		}
		if rule.Action != "set_status" || rule.ActionValue != "done" {
			t.Error("Invalid auto-transition rule action")
		}
	})
}

func TestTodoWithPriorityWorkflow_ExtendedFeatures(t *testing.T) {
	config := TodoWithPriorityWorkflow

	t.Run("has both completion and priority dimensions", func(t *testing.T) {
		completion := config.GetDimension("completion")
		priority := config.GetDimension("priority")

		if completion == nil {
			t.Error("Missing completion dimension")
		}
		if priority == nil {
			t.Error("Missing priority dimension")
		}

		if priority.DefaultValue != "medium" {
			t.Errorf("Expected priority default 'medium', got %q", priority.DefaultValue)
		}
	})

	t.Run("high_priority context filters correctly", func(t *testing.T) {
		rules := config.Visibility["high_priority"]
		if len(rules) != 2 {
			t.Fatalf("Expected 2 rules for high_priority context, got %d", len(rules))
		}

		// Test item that matches both rules (pending + high priority)
		matchingStatuses := map[string]string{
			"completion": "pending",
			"priority":   "high",
		}

		visible := true
		for _, rule := range rules {
			if !rule.Matches("high_priority", matchingStatuses) {
				visible = false
				break
			}
		}
		if !visible {
			t.Error("High priority pending item should be visible in high_priority context")
		}

		// Test item that doesn't match (done or low priority)
		nonMatchingStatuses := map[string]string{
			"completion": "done",
			"priority":   "high",
		}

		visible = true
		for _, rule := range rules {
			if !rule.Matches("high_priority", nonMatchingStatuses) {
				visible = false
				break
			}
		}
		if visible {
			t.Error("Done high priority item should not be visible in high_priority context")
		}
	})
}

func TestCMSWorkflow_PublicationStates(t *testing.T) {
	config := CMSWorkflow

	t.Run("has publication and priority dimensions", func(t *testing.T) {
		pub := config.GetDimension("publication")
		priority := config.GetDimension("priority")

		if pub == nil {
			t.Error("Missing publication dimension")
		}
		if priority == nil {
			t.Error("Missing priority dimension")
		}

		expectedPubValues := []string{"draft", "review", "published", "archived"}
		if len(pub.Values) != len(expectedPubValues) {
			t.Errorf("Expected %d publication values, got %d", len(expectedPubValues), len(pub.Values))
		}
	})

	t.Run("public context shows only published", func(t *testing.T) {
		rules := config.Visibility["public"]
		if len(rules) != 1 {
			t.Fatalf("Expected 1 rule for public context, got %d", len(rules))
		}

		rule := rules[0]
		publishedStatuses := map[string]string{"publication": "published"}
		draftStatuses := map[string]string{"publication": "draft"}

		if !rule.Matches("public", publishedStatuses) {
			t.Error("Published items should be visible in public context")
		}
		if rule.Matches("public", draftStatuses) {
			t.Error("Draft items should not be visible in public context")
		}
	})

	t.Run("editorial workflow transitions", func(t *testing.T) {
		transitionRules := config.Transitions["publication"]

		// Find draft->review rule
		var draftToReview *TransitionRule
		for i, rule := range transitionRules {
			if rule.From == "draft" {
				draftToReview = &transitionRules[i]
				break
			}
		}

		if draftToReview == nil {
			t.Fatal("Missing draft transition rule")
		}

		if !draftToReview.CanTransition("draft", "review") {
			t.Error("Should allow draft->review transition")
		}
		if draftToReview.CanTransition("draft", "published") {
			t.Error("Should not allow direct draft->published transition")
		}
	})
}

func TestIssueTrackerWorkflow_StateTransitions(t *testing.T) {
	config := IssueTrackerWorkflow

	t.Run("has state, priority, and type dimensions", func(t *testing.T) {
		state := config.GetDimension("state")
		priority := config.GetDimension("priority")
		issueType := config.GetDimension("type")

		if state == nil {
			t.Error("Missing state dimension")
		}
		if priority == nil {
			t.Error("Missing priority dimension")
		}
		if issueType == nil {
			t.Error("Missing type dimension")
		}

		if state.DefaultValue != "open" {
			t.Errorf("Expected state default 'open', got %q", state.DefaultValue)
		}
	})

	t.Run("critical_issues context filters correctly", func(t *testing.T) {
		rules := config.Visibility["critical_issues"]
		if len(rules) != 2 {
			t.Fatalf("Expected 2 rules for critical_issues context, got %d", len(rules))
		}

		// Test critical issue
		criticalStatuses := map[string]string{
			"state":    "open",
			"priority": "critical",
		}

		visible := true
		for _, rule := range rules {
			if !rule.Matches("critical_issues", criticalStatuses) {
				visible = false
				break
			}
		}
		if !visible {
			t.Error("Critical open issue should be visible")
		}

		// Test non-critical issue
		minorStatuses := map[string]string{
			"state":    "open",
			"priority": "minor",
		}

		visible = true
		for _, rule := range rules {
			if !rule.Matches("critical_issues", minorStatuses) {
				visible = false
				break
			}
		}
		if visible {
			t.Error("Minor issue should not be visible in critical_issues context")
		}
	})

	t.Run("state transition flow", func(t *testing.T) {
		transitionRules := config.Transitions["state"]

		// Find open state rule
		var openRule *TransitionRule
		for i, rule := range transitionRules {
			if rule.From == "open" {
				openRule = &transitionRules[i]
				break
			}
		}

		if openRule == nil {
			t.Fatal("Missing open state transition rule")
		}

		// Test valid transitions from open
		validTransitions := []string{"in_progress", "resolved", "closed"}
		for _, to := range validTransitions {
			if !openRule.CanTransition("open", to) {
				t.Errorf("Should allow open->%s transition", to)
			}
		}

		// Test invalid transition
		if openRule.CanTransition("open", "reopened") {
			t.Error("Should not allow open->reopened transition")
		}
	})
}

func TestKanbanWorkflow_StageProgression(t *testing.T) {
	config := KanbanWorkflow

	t.Run("has stage and size dimensions", func(t *testing.T) {
		stage := config.GetDimension("stage")
		size := config.GetDimension("size")

		if stage == nil {
			t.Error("Missing stage dimension")
		}
		if size == nil {
			t.Error("Missing size dimension")
		}

		expectedStages := []string{"backlog", "todo", "in_progress", "review", "done"}
		if len(stage.Values) != len(expectedStages) {
			t.Errorf("Expected %d stages, got %d", len(expectedStages), len(stage.Values))
		}
	})

	t.Run("board context excludes backlog", func(t *testing.T) {
		rules := config.Visibility["board"]
		if len(rules) != 1 {
			t.Fatalf("Expected 1 rule for board context, got %d", len(rules))
		}

		rule := rules[0]
		backlogStatuses := map[string]string{"stage": "backlog"}
		todoStatuses := map[string]string{"stage": "todo"}

		if rule.Matches("board", backlogStatuses) {
			t.Error("Backlog items should not be visible on board")
		}
		if !rule.Matches("board", todoStatuses) {
			t.Error("Todo items should be visible on board")
		}
	})

	t.Run("stage progression rules", func(t *testing.T) {
		transitionRules := config.Transitions["stage"]

		// Find backlog rule
		var backlogRule *TransitionRule
		for i, rule := range transitionRules {
			if rule.From == "backlog" {
				backlogRule = &transitionRules[i]
				break
			}
		}

		if backlogRule == nil {
			t.Fatal("Missing backlog transition rule")
		}

		// Backlog should only go to todo
		if !backlogRule.CanTransition("backlog", "todo") {
			t.Error("Should allow backlog->todo transition")
		}
		if backlogRule.CanTransition("backlog", "in_progress") {
			t.Error("Should not allow direct backlog->in_progress transition")
		}
	})

	t.Run("wip context shows only in-progress", func(t *testing.T) {
		rules := config.Visibility["wip"]
		if len(rules) != 1 {
			t.Fatalf("Expected 1 rule for wip context, got %d", len(rules))
		}

		rule := rules[0]
		wipStatuses := map[string]string{"stage": "in_progress"}
		todoStatuses := map[string]string{"stage": "todo"}

		if !rule.Matches("wip", wipStatuses) {
			t.Error("In-progress items should be visible in wip context")
		}
		if rule.Matches("wip", todoStatuses) {
			t.Error("Todo items should not be visible in wip context")
		}
	})
}