package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/talesmud/talesmud/pkg/entities/audit"
	"github.com/talesmud/talesmud/pkg/repository"
)

// AuditActor is who performed a write.
type AuditActor struct {
	UserID string
	Name   string
}

// AuditRecord is one change to append to the log.
type AuditRecord struct {
	Actor      AuditActor
	Action     string
	EntityType string
	EntityID   string
	Before     json.RawMessage
	After      json.RawMessage
	Source     string
	Undoable   bool
	Summary    string
	Inverse    *audit.Inverse
}

// AuditService records creator and ops writes and undoes creator rows.
type AuditService interface {
	Record(in AuditRecord) (*audit.Entry, error)
	Get(id string) (*audit.Entry, error)
	List(entityType, entityID string, limit int) ([]*audit.Entry, error)
	LoadRaw(entityType, id string) (json.RawMessage, error)
	// UndoCRUD restores a creator row from its before JSON.
	// Ops rows are refused here; the game applies those inverses.
	UndoCRUD(id string, actor AuditActor) (*audit.Entry, error)
	MarkUndone(id, undoneBy string) error
}

type auditService struct {
	repo repository.AuditRepository
}

// NewAuditService creates an audit service.
func NewAuditService(repo repository.AuditRepository) AuditService {
	return &auditService{repo: repo}
}

// AuditError is a refusal with an HTTP status.
type AuditError struct {
	Status int
	Msg    string
}

func (e *AuditError) Error() string {
	if e == nil {
		return ""
	}
	return e.Msg
}

func (e *AuditError) StatusCode() int {
	if e == nil {
		return 0
	}
	return e.Status
}

func auditErr(status int, msg string) error {
	return &AuditError{Status: status, Msg: msg}
}

func (s *auditService) Record(in AuditRecord) (*audit.Entry, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("audit log is not available")
	}
	name := in.Actor.Name
	if name == "" {
		name = in.Actor.UserID
	}
	if name == "" {
		name = "unknown"
	}
	source := in.Source
	if source == "" {
		source = audit.SourceCreator
	}
	entry := &audit.Entry{
		ID:          uuid.New().String(),
		Time:        time.Now().UTC(),
		ActorUserID: in.Actor.UserID,
		ActorName:   name,
		Action:      in.Action,
		EntityType:  in.EntityType,
		EntityID:    in.EntityID,
		Before:      compactRaw(in.Before),
		After:       compactRaw(in.After),
		Source:      source,
		Undoable:    in.Undoable,
		Summary:     in.Summary,
		Inverse:     in.Inverse,
	}
	if err := s.repo.Insert(entry); err != nil {
		return nil, err
	}
	return entry, nil
}

func (s *auditService) Get(id string) (*audit.Entry, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("audit log is not available")
	}
	return s.repo.FindByID(id)
}

func (s *auditService) List(entityType, entityID string, limit int) ([]*audit.Entry, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("audit log is not available")
	}
	return s.repo.List(entityType, entityID, limit)
}

func (s *auditService) LoadRaw(entityType, id string) (json.RawMessage, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("audit log is not available")
	}
	return s.repo.LoadRaw(entityType, id)
}

func (s *auditService) MarkUndone(id, undoneBy string) error {
	entry, err := s.Get(id)
	if err != nil {
		return err
	}
	entry.UndoneBy = undoneBy
	return s.repo.Update(entry)
}

// UndoCRUD restores the before snapshot when the entity still matches After.
func (s *auditService) UndoCRUD(id string, actor AuditActor) (*audit.Entry, error) {
	entry, err := s.Get(id)
	if err != nil {
		return nil, auditErr(404, "audit entry not found")
	}
	if entry.UndoneBy != "" {
		return nil, auditErr(400, "this change was already undone")
	}
	if !entry.Undoable {
		return nil, auditErr(400, "this change cannot be undone")
	}
	if entry.Source == audit.SourceOps {
		return nil, auditErr(400, "this ops change is undone through its inverse")
	}
	current, err := s.repo.LoadRaw(entry.EntityType, entry.EntityID)
	if err != nil {
		return nil, err
	}
	if canonicalRaw(current) != canonicalRaw(entry.After) {
		return nil, auditErr(409, "the entity changed since this change; undo was refused")
	}
	if isEmptyRaw(entry.Before) {
		if err := s.repo.DeleteRaw(entry.EntityType, entry.EntityID); err != nil {
			return nil, err
		}
	} else if err := s.repo.SaveRaw(entry.EntityType, entry.EntityID, entry.Before); err != nil {
		return nil, err
	}
	restored, err := s.repo.LoadRaw(entry.EntityType, entry.EntityID)
	if err != nil {
		return nil, err
	}
	undo, err := s.Record(AuditRecord{
		Actor:      actor,
		Action:     "undo",
		EntityType: entry.EntityType,
		EntityID:   entry.EntityID,
		Before:     current,
		After:      restored,
		Source:     audit.SourceCreator,
		Undoable:   true,
		Summary:    "undo " + entry.Action + " " + entry.EntityType + " " + entry.EntityID,
	})
	if err != nil {
		return nil, err
	}
	if err := s.MarkUndone(entry.ID, undo.ID); err != nil {
		return nil, err
	}
	return undo, nil
}

func isEmptyRaw(raw json.RawMessage) bool {
	trimmed := bytes.TrimSpace(raw)
	return len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null"))
}

func compactRaw(raw json.RawMessage) json.RawMessage {
	if isEmptyRaw(raw) {
		return nil
	}
	var buf bytes.Buffer
	if err := json.Compact(&buf, raw); err != nil {
		return raw
	}
	return buf.Bytes()
}

func canonicalRaw(raw json.RawMessage) string {
	if isEmptyRaw(raw) {
		return ""
	}
	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		return string(raw)
	}
	b, err := json.Marshal(v)
	if err != nil {
		return string(raw)
	}
	return string(b)
}
