package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/Veysel440/go-audit/internal/core"
)

var (
	ErrNotFound    = errors.New("not_found")
	ErrValidation  = errors.New("validation")
	ErrRateLimited = errors.New("rate_limited")
)

// Repo defines the contract for data persistence
type Repo interface {
	Insert(ctx context.Context, a core.Audit) (core.Audit, error)
	InsertIdempotent(ctx context.Context, a core.Audit) (core.Audit, bool, error)
	Get(ctx context.Context, id string) (core.Audit, error)
	List(ctx context.Context, f ListFilter) ([]core.Audit, error)
	LastByResource(ctx context.Context, rt, rid string) (core.Audit, error)
	Ping(ctx context.Context) error
}

// Service contains business logic on top of Repo
type Service struct {
	repo Repo
}

// New returns a new Service
func New(repo Repo) *Service { return &Service{repo: repo} }

// Create inserts a single audit record with chain-hash support
func (s *Service) Create(ctx context.Context, in core.CreateAudit) (core.Audit, error) {
	if in.ActorID == "" || in.Action == "" || in.ResourceID == "" || in.ResourceType == "" {
		return core.Audit{}, ErrValidation
	}

	now := time.Now().UTC()

	// chain hash: link to previous
	var prev string
	if last, err := s.repo.LastByResource(ctx, in.ResourceType, in.ResourceID); err == nil {
		prev = last.ChainHash
	}

	type chainInput struct {
		ActorID, ActorType, Action, ResourceID, ResourceType string
		CreatedAt                                            time.Time
	}
	b, err := json.Marshal(chainInput{
		ActorID: in.ActorID, ActorType: in.ActorType, Action: in.Action,
		ResourceID: in.ResourceID, ResourceType: in.ResourceType, CreatedAt: now,
	})
	if err != nil {
		return core.Audit{}, err
	}

	sum := sha256.Sum256(append([]byte(prev), b...))
	chainHash := hex.EncodeToString(sum[:])

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
		ChainPrev:    prev,
		ChainHash:    chainHash,
		IdemHash:     in.IdemHash,
	}

	if in.IdemHash != "" {
		out, _, err := s.repo.InsertIdempotent(ctx, a)
		return out, err
	}
	return s.repo.Insert(ctx, a)
}

// CreateBatch inserts multiple audit records
func (s *Service) CreateBatch(ctx context.Context, items []core.CreateAudit) ([]core.Audit, []error) {
	out := make([]core.Audit, len(items))
	errs := make([]error, len(items))
	for i, it := range items {
		a, err := s.Create(ctx, it)
		out[i], errs[i] = a, err
	}
	return out, errs
}

// Get returns a single audit by ID
func (s *Service) Get(ctx context.Context, id string) (core.Audit, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) List(ctx context.Context, f ListFilter) ([]core.Audit, error) {
	f.Normalize()
	return s.repo.List(ctx, f)
}

// newID generates a random hex string
func newID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
