package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/Veysel440/go-audit/internal/core"
)

var (
	ErrNotFound    = errors.New("not_found")
	ErrValidation  = errors.New("validation")
	ErrRateLimited = errors.New("rate_limited")
)

type Repo interface {
	Insert(ctx context.Context, a core.Audit) (core.Audit, error)
	InsertIdempotent(ctx context.Context, a core.Audit) (core.Audit, bool, error) // created?, duplicate?
	Get(ctx context.Context, id string) (core.Audit, error)
	List(ctx context.Context, f ListFilter) ([]core.Audit, error)
}

type Service struct{ repo Repo }

func New(repo Repo) *Service { return &Service{repo: repo} }

func (s *Service) Create(ctx context.Context, in core.CreateAudit) (core.Audit, error) {
	if in.ActorID == "" || in.Action == "" || in.ResourceID == "" || in.ResourceType == "" {
		return core.Audit{}, ErrValidation
	}
	now := time.Now().UTC()
	a := core.Audit{
		ID:           newID(),
		ActorID:      in.ActorID,
		ActorType:    in.ActorType,
		Action:       in.Action,
		ResourceID:   in.ResourceID,
		ResourceType: in.ResourceType,
		IP:           in.IP,
		UA:           in.UA,
		Metadata:     in.Metadata,
		CreatedAt:    now,
		IdemHash:     in.IdemHash,
	}
	if in.IdemHash != "" {
		out, _, err := s.repo.InsertIdempotent(ctx, a)
		return out, err
	}
	return s.repo.Insert(ctx, a)
}

func (s *Service) Get(ctx context.Context, id string) (core.Audit, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) List(ctx context.Context, f ListFilter) ([]core.Audit, error) {
	if f.Limit <= 0 || f.Limit > 1000 {
		f.Limit = 100
	}
	return s.repo.List(ctx, f)
}

func newID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
