package audit

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/domain/audit/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const collectionName = "audit_log"

// MongoAgentAuditRepository implements audit.AgentAuditRepository using MongoDB.
type MongoAgentAuditRepository struct {
	collection *mongo.Collection
}

// NewMongoAgentAuditRepository creates a new MongoDB-backed agent audit repository.
func NewMongoAgentAuditRepository(db *mongo.Database) *MongoAgentAuditRepository {
	return &MongoAgentAuditRepository{
		collection: db.Collection(collectionName),
	}
}

func (r *MongoAgentAuditRepository) Create(ctx context.Context, entry *model.AgentAuditEntry) error {
	if entry.ID.IsZero() {
		entry.ID = bson.NewObjectID()
	}
	entry.CreatedAt = time.Now().UTC()

	_, err := r.collection.InsertOne(ctx, entry)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}

func (r *MongoAgentAuditRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int64) ([]*model.AgentAuditEntry, error) {
	opts := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetLimit(limit).
		SetSkip(offset)

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, fmt.Errorf("find audit logs by user: %w", err)
	}
	defer cursor.Close(ctx)

	var list []*model.AgentAuditEntry
	if err := cursor.All(ctx, &list); err != nil {
		return nil, fmt.Errorf("decode audit logs: %w", err)
	}
	return list, nil
}

func (r *MongoAgentAuditRepository) FindByObligationID(ctx context.Context, oblID bson.ObjectID) ([]*model.AgentAuditEntry, error) {
	opts := options.Find().SetSort(bson.M{"created_at": 1})

	cursor, err := r.collection.Find(ctx, bson.M{"obligation_id": oblID}, opts)
	if err != nil {
		return nil, fmt.Errorf("find audit logs by obligation: %w", err)
	}
	defer cursor.Close(ctx)

	var list []*model.AgentAuditEntry
	if err := cursor.All(ctx, &list); err != nil {
		return nil, fmt.Errorf("decode audit logs: %w", err)
	}
	return list, nil
}

func (r *MongoAgentAuditRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	result, err := r.collection.DeleteMany(ctx, bson.M{"user_id": userID})
	if err != nil {
		return 0, fmt.Errorf("delete audit logs by user: %w", err)
	}
	return result.DeletedCount, nil
}
