package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ensureIndexes(ctx context.Context, c *mongo.Collection, retentionDays int64) error {
	ttl := int32(1)
	_, err := c.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "createdAt", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(int32(retentionDays * 86400)),
		},
		{
			Keys:    bson.D{{Key: "actorId", Value: 1}, {Key: "createdAt", Value: -1}},
			Options: options.Index().SetName("actor_created"),
		},
		{
			Keys: bson.D{{Key: "resourceType", Value: 1}, {Key: "resourceId", Value: 1}, {Key: "createdAt", Value: -1}},
		},
		{
			Keys:    bson.D{{Key: "idemHash", Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
		{Keys: bson.D{{Key: "action", Value: 1}}},
		{Keys: bson.D{{Key: "ttl_marker", Value: ttl}}},
	})
	return err
}
