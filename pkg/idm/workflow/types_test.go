package workflow

import (
	"testing"
)

func TestStatusDimension_Validate(t *testing.T) {
	tests := []struct {
		name      string
		dimension StatusDimension
		wantErr   bool
	}{
		{
			name: "valid dimension",
			dimension: StatusDimension{
				Name:         "completion",
				Values:       []string{"pending", "done"},
				DefaultValue: "pending",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			dimension: StatusDimension{
				Name:   "",
				Values: []string{"pending", "done"},
			},
			wantErr: true,
		},
		{
			name: "empty values",
			dimension: StatusDimension{
				Name:   "completion",
				Values: []string{},
			},
			wantErr: true,
		},
		{
			name: "invalid default value",
			dimension: StatusDimension{
				Name:         "completion",
				Values:       []string{"pending", "done"},
				DefaultValue: "invalid",
			},
			wantErr: true,
		},
		{
			name: "no default value",
			dimension: StatusDimension{
				Name:   "completion",
				Values: []string{"pending", "done"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dimension.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("StatusDimension.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStatusDimension_HasValue(t *testing.T) {
	dimension := StatusDimension{
		Name:   "completion",
		Values: []string{"pending", "done"},
	}

	tests := []struct {
		value string
		want  bool
	}{
		{"pending", true},
		{"done", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			if got := dimension.HasValue(tt.value); got != tt.want {
				t.Errorf("StatusDimension.HasValue(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestVisibilityRule_Matches(t *testing.T) {
	rule := VisibilityRule{
		Context:   "active",
		Dimension: "completion",
		Include:   []string{"pending"},
		Exclude:   []string{"archived"},
	}

	tests := []struct {
		name     string
		context  string
		statuses map[string]string
		want     bool
	}{
		{
			name:    "matches context and included status",
			context: "active",
			statuses: map[string]string{
				"completion": "pending",
			},
			want: true,
		},
		{
			name:    "wrong context",
			context: "all",
			statuses: map[string]string{
				"completion": "pending",
			},
			want: false,
		},
		{
			name:    "excluded status",
			context: "active",
			statuses: map[string]string{
				"completion": "archived",
			},
			want: false,
		},
		{
			name:    "not included status",
			context: "active",
			statuses: map[string]string{
				"completion": "done",
			},
			want: false,
		},
		{
			name:    "missing dimension",
			context: "active",
			statuses: map[string]string{
				"priority": "high",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rule.Matches(tt.context, tt.statuses); got != tt.want {
				t.Errorf("VisibilityRule.Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVisibilityRule_MatchesWithEmptyInclude(t *testing.T) {
	rule := VisibilityRule{
		Context:   "admin",
		Dimension: "completion",
		Include:   []string{}, // Empty include means include all
		Exclude:   []string{"deleted"},
	}

	tests := []struct {
		name     string
		statuses map[string]string
		want     bool
	}{
		{
			name: "any status except excluded",
			statuses: map[string]string{
				"completion": "pending",
			},
			want: true,
		},
		{
			name: "excluded status",
			statuses: map[string]string{
				"completion": "deleted",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rule.Matches("admin", tt.statuses); got != tt.want {
				t.Errorf("VisibilityRule.Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransitionRule_CanTransition(t *testing.T) {
	rule := TransitionRule{
		Dimension: "completion",
		From:      "pending",
		To:        []string{"done", "cancelled"},
	}

	tests := []struct {
		name string
		from string
		to   string
		want bool
	}{
		{"valid transition to done", "pending", "done", true},
		{"valid transition to cancelled", "pending", "cancelled", true},
		{"invalid from state", "done", "pending", false},
		{"invalid to state", "pending", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rule.CanTransition(tt.from, tt.to); got != tt.want {
				t.Errorf("TransitionRule.CanTransition(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestWorkflowConfig_Validate(t *testing.T) {
	validConfig := WorkflowConfig{
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
		},
		Transitions: map[string][]TransitionRule{
			"completion": {
				{
					Dimension: "completion",
					From:      "pending",
					To:        []string{"done"},
				},
			},
		},
	}

	tests := []struct {
		name    string
		config  WorkflowConfig
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  validConfig,
			wantErr: false,
		},
		{
			name: "duplicate dimension names",
			config: WorkflowConfig{
				Dimensions: []StatusDimension{
					{Name: "completion", Values: []string{"pending", "done"}},
					{Name: "completion", Values: []string{"low", "high"}},
				},
			},
			wantErr: true,
		},
		{
			name: "visibility rule references unknown dimension",
			config: WorkflowConfig{
				Dimensions: []StatusDimension{
					{Name: "completion", Values: []string{"pending", "done"}},
				},
				Visibility: map[string][]VisibilityRule{
					"active": {
						{
							Context:   "active",
							Dimension: "unknown",
							Include:   []string{"pending"},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "transition rule references unknown dimension",
			config: WorkflowConfig{
				Dimensions: []StatusDimension{
					{Name: "completion", Values: []string{"pending", "done"}},
				},
				Transitions: map[string][]TransitionRule{
					"unknown": {
						{
							Dimension: "unknown",
							From:      "pending",
							To:        []string{"done"},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "transition rule with invalid from value",
			config: WorkflowConfig{
				Dimensions: []StatusDimension{
					{Name: "completion", Values: []string{"pending", "done"}},
				},
				Transitions: map[string][]TransitionRule{
					"completion": {
						{
							Dimension: "completion",
							From:      "invalid",
							To:        []string{"done"},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "transition rule with invalid to value",
			config: WorkflowConfig{
				Dimensions: []StatusDimension{
					{Name: "completion", Values: []string{"pending", "done"}},
				},
				Transitions: map[string][]TransitionRule{
					"completion": {
						{
							Dimension: "completion",
							From:      "pending",
							To:        []string{"invalid"},
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("WorkflowConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWorkflowConfig_GetDimension(t *testing.T) {
	config := WorkflowConfig{
		Dimensions: []StatusDimension{
			{Name: "completion", Values: []string{"pending", "done"}},
			{Name: "priority", Values: []string{"low", "high"}},
		},
	}

	tests := []struct {
		name string
		want *StatusDimension
	}{
		{"completion", &StatusDimension{Name: "completion", Values: []string{"pending", "done"}}},
		{"priority", &StatusDimension{Name: "priority", Values: []string{"low", "high"}}},
		{"unknown", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.GetDimension(tt.name)
			if tt.want == nil {
				if got != nil {
					t.Errorf("WorkflowConfig.GetDimension(%q) = %v, want nil", tt.name, got)
				}
			} else {
				if got == nil {
					t.Errorf("WorkflowConfig.GetDimension(%q) = nil, want %v", tt.name, tt.want)
				} else if got.Name != tt.want.Name {
					t.Errorf("WorkflowConfig.GetDimension(%q).Name = %q, want %q", tt.name, got.Name, tt.want.Name)
				}
			}
		})
	}
}

func TestStatusMetrics(t *testing.T) {
	metrics := NewStatusMetrics()

	// Add some test items
	metrics.AddItem(map[string]string{
		"completion": "pending",
		"priority":   "high",
	})
	metrics.AddItem(map[string]string{
		"completion": "done",
		"priority":   "low",
	})
	metrics.AddItem(map[string]string{
		"completion": "pending",
		"priority":   "high",
	})

	// Test total count
	if metrics.Total != 3 {
		t.Errorf("metrics.Total = %d, want 3", metrics.Total)
	}

	// Test dimension counts
	tests := []struct {
		dimension string
		value     string
		want      int
	}{
		{"completion", "pending", 2},
		{"completion", "done", 1},
		{"priority", "high", 2},
		{"priority", "low", 1},
		{"unknown", "value", 0},
	}

	for _, tt := range tests {
		got := metrics.GetCount(tt.dimension, tt.value)
		if got != tt.want {
			t.Errorf("metrics.GetCount(%q, %q) = %d, want %d", tt.dimension, tt.value, got, tt.want)
		}
	}

	// Test context counts
	metrics.SetContextCount("active", 2)
	metrics.SetContextCount("all", 3)

	if got := metrics.GetContextCount("active"); got != 2 {
		t.Errorf("metrics.GetContextCount(\"active\") = %d, want 2", got)
	}
	if got := metrics.GetContextCount("all"); got != 3 {
		t.Errorf("metrics.GetContextCount(\"all\") = %d, want 3", got)
	}
	if got := metrics.GetContextCount("unknown"); got != 0 {
		t.Errorf("metrics.GetContextCount(\"unknown\") = %d, want 0", got)
	}
}