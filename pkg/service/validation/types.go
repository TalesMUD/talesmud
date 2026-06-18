package validation

type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

type Issue struct {
	Severity   Severity `json:"severity"`
	Code       string   `json:"code"`
	EntityType string   `json:"entityType"`
	EntityID   string   `json:"entityId,omitempty"`
	Field      string   `json:"field,omitempty"`
	RefType    string   `json:"refType,omitempty"`
	RefID      string   `json:"refId,omitempty"`
	Message    string   `json:"message"`
}

type Result struct {
	Valid    bool    `json:"valid"`
	Errors   int     `json:"errors"`
	Warnings int     `json:"warnings"`
	Issues   []Issue `json:"issues"`
}

func NewResult() Result {
	return Result{Valid: true, Issues: []Issue{}}
}

func (r *Result) Add(issue Issue) {
	r.Issues = append(r.Issues, issue)
	switch issue.Severity {
	case SeverityError:
		r.Errors++
	case SeverityWarning:
		r.Warnings++
	}
	r.Valid = r.Errors == 0
}

func (r *Result) Merge(other Result) {
	for _, issue := range other.Issues {
		r.Add(issue)
	}
}

func Error(code, entityType, entityID, field, refType, refID, message string) Issue {
	return Issue{
		Severity:   SeverityError,
		Code:       code,
		EntityType: entityType,
		EntityID:   entityID,
		Field:      field,
		RefType:    refType,
		RefID:      refID,
		Message:    message,
	}
}

func Warning(code, entityType, entityID, field, refType, refID, message string) Issue {
	return Issue{
		Severity:   SeverityWarning,
		Code:       code,
		EntityType: entityType,
		EntityID:   entityID,
		Field:      field,
		RefType:    refType,
		RefID:      refID,
		Message:    message,
	}
}
