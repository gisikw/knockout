package main

import (
	"strings"
	"testing"
)

func TestValidatePlanQuestions(t *testing.T) {
	tests := []struct {
		name      string
		questions []PlanQuestion
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "empty slice is valid",
			questions: []PlanQuestion{},
			wantErr:   false,
		},
		{
			name: "valid question with minimal fields",
			questions: []PlanQuestion{
				{
					ID:       "q1",
					Question: "Test question?",
					Options: []QuestionOption{
						{Label: "Option A", Value: "a"},
						{Label: "Option B", Value: "b"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid question with all fields",
			questions: []PlanQuestion{
				{
					ID:       "q1",
					Question: "Test question?",
					Context:  "Some context",
					Options: []QuestionOption{
						{Label: "Option A", Value: "a", Description: "Description A"},
						{Label: "Option B", Value: "b", Description: "Description B"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple valid questions",
			questions: []PlanQuestion{
				{
					ID:       "q1",
					Question: "First question?",
					Options: []QuestionOption{
						{Label: "Option A", Value: "a"},
					},
				},
				{
					ID:       "q2",
					Question: "Second question?",
					Options: []QuestionOption{
						{Label: "Option X", Value: "x"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing id",
			questions: []PlanQuestion{
				{
					Question: "Test question?",
					Options: []QuestionOption{
						{Label: "Option A", Value: "a"},
					},
				},
			},
			wantErr: true,
			errMsg:  "missing required field 'id'",
		},
		{
			name: "missing question",
			questions: []PlanQuestion{
				{
					ID: "q1",
					Options: []QuestionOption{
						{Label: "Option A", Value: "a"},
					},
				},
			},
			wantErr: true,
			errMsg:  "missing required field 'question'",
		},
		{
			name: "missing options",
			questions: []PlanQuestion{
				{
					ID:       "q1",
					Question: "Test question?",
					Options:  []QuestionOption{},
				},
			},
			wantErr: true,
			errMsg:  "missing required field 'options'",
		},
		{
			name: "option missing label",
			questions: []PlanQuestion{
				{
					ID:       "q1",
					Question: "Test question?",
					Options: []QuestionOption{
						{Value: "a"},
					},
				},
			},
			wantErr: true,
			errMsg:  "missing required field 'label'",
		},
		{
			name: "option missing value",
			questions: []PlanQuestion{
				{
					ID:       "q1",
					Question: "Test question?",
					Options: []QuestionOption{
						{Label: "Option A"},
					},
				},
			},
			wantErr: true,
			errMsg:  "missing required field 'value'",
		},
		{
			name: "second question invalid",
			questions: []PlanQuestion{
				{
					ID:       "q1",
					Question: "Valid question?",
					Options: []QuestionOption{
						{Label: "Option A", Value: "a"},
					},
				},
				{
					ID: "q2",
					// Missing question field
					Options: []QuestionOption{
						{Label: "Option B", Value: "b"},
					},
				},
			},
			wantErr: true,
			errMsg:  "missing required field 'question'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePlanQuestions(tt.questions)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePlanQuestions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidatePlanQuestions() error = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}
