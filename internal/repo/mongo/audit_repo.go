package mongo

import (
	"context"
	"errors"

	"github.com/Veysel440/go-audit/internal/core"
	"github.com/Veysel440/go-audit/internal/service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repo struct {
	col *mongo.Collection
}

type Config struct {
	URI           string
	DB            string
	Collection    string
	RetentionDays int64
}

func New(ctx context.Context, cfg Config) (*Repo, error) {
	cl, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, err
	}
	col := cl.Database(cfg.DB).Collection(cfg.Collection)
	if err := ensureIndexes(ctx, col, cfg.RetentionDays); err != nil {
		return nil, err
	}
	return &Repo{col: col}, nil
}

func (r *Repo) Insert(ctx context.Context, a core.Audit) (core.Audit, error) {
	_, err := r.col.InsertOne(ctx, a)
	return a, err
}

func (r *Repo) InsertIdempotent(ctx context.Context, a core.Audit) (core.Audit, bool, error) {
	_, err := r.col.InsertOne(ctx, a)
	if err == nil {
		return a, true, nil
	}
	var we mongo.WriteException
	if errors.As(err, &we) {
		for _, e := range we.WriteErrors {
			if e.Code == 11000 {
				var out core.Audit
				if err := r.col.FindOne(ctx, bson.M{"idemHash": a.IdemHash}).Decode(&out); err != nil {
					return core.Audit{}, false, err
				}
				return out, false, nil
			}
		}
	}
	return core.Audit{}, false, err
}

func (r *Repo) Get(ctx context.Context, id string) (core.Audit, error) {
	var out core.Audit
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&out)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return core.Audit{}, service.ErrNotFound
	}
	return out, err
}

func (r *Repo) List(ctx context.Context, f service.ListFilter) ([]core.Audit, error) {
	q := bson.M{}
	if f.ActorID != "" {
		q["actorId"] = f.ActorID
	}
	if f.ActorType != "" {
		q["actorType"] = f.ActorType
	}
	if f.Action != "" {
		q["action"] = f.Action
	}
	if f.ResourceID != "" {
		q["resourceId"] = f.ResourceID
	}
	if f.ResourceType != "" {
		q["resourceType"] = f.ResourceType
	}
	if f.Since != nil || f.Until != nil {
		rng := bson.M{}
		if f.Since != nil {
			rng["$gte"] = f.Since.UTC()
		}
		if f.Until != nil {
			rng["$lte"] = f.Until.UTC()
		}
		q["createdAt"] = rng
	}
	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetLimit(f.Limit).
		SetSkip(f.Offset)

	cur, err := r.col.Find(ctx, q, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var list []core.Audit
	for cur.Next(ctx) {
		var a core.Audit
		if err := cur.Decode(&a); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, cur.Err()
}
