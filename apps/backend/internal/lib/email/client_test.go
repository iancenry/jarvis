package email

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iancenry/jarvis/internal/model/todo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientRenderEmailUsesEmbeddedTemplates(t *testing.T) {
	client := &Client{}

	testCases := []struct {
		name         string
		templateName Template
		data         map[string]any
		wantContains string
	}{
		{
			name:         "welcome",
			templateName: TemplateWelcome,
			data: map[string]any{
				"UserFirstName": "Ian",
			},
			wantContains: "Welcome to JARVIS",
		},
		{
			name:         "due date reminder",
			templateName: TemplateDueDateReminder,
			data: map[string]any{
				"TodoTitle":    "Ship release",
				"TodoID":       uuid.New().String(),
				"DueDate":      time.Now().Format(time.RFC822),
				"DaysUntilDue": 1,
			},
			wantContains: "Ship release",
		},
		{
			name:         "overdue notification",
			templateName: TemplateOverdueNotification,
			data: map[string]any{
				"TodoTitle":   "Fix bug",
				"TodoID":      uuid.New().String(),
				"DueDate":     time.Now().Format(time.RFC822),
				"DaysOverdue": 3,
			},
			wantContains: "Overdue Todo",
		},
		{
			name:         "weekly report",
			templateName: TemplateWeeklyReport,
			data: map[string]any{
				"WeekStart":      "January 1, 2026",
				"WeekEnd":        "January 7, 2026",
				"CompletedCount": 2,
				"ActiveCount":    4,
				"OverdueCount":   1,
				"CompletedTodos": []todo.PopulatedTodo{},
				"OverdueTodos":   []todo.PopulatedTodo{},
				"HasCompleted":   true,
				"HasOverdue":     true,
			},
			wantContains: "Weekly Report",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, err := client.renderEmail(tc.templateName, tc.data)

			require.NoError(t, err)
			assert.Contains(t, body, tc.wantContains)
		})
	}
}
