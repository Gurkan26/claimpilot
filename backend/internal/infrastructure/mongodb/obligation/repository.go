package obligation

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/domain/obligation/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const collectionName = "obligations"

// MongoRepository implements obligation.ObligationRepository using MongoDB.
type MongoRepository struct {
	collection *mongo.Collection
}

// NewMongoRepository creates a new MongoDB-backed obligation repository.
func NewMongoRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{
		collection: db.Collection(collectionName),
	}
}

func (r *MongoRepository) Create(ctx context.Context, obl *model.Obligation) error {
	if obl.ID.IsZero() {
		obl.ID = bson.NewObjectID()
	}
	obl.CreatedAt = time.Now().UTC()
	obl.UpdatedAt = obl.CreatedAt

	_, err := r.collection.InsertOne(ctx, obl)
	if err != nil {
		return fmt.Errorf("insert obligation: %w", err)
	}
	return nil
}

func (r *MongoRepository) FindByID(ctx context.Context, id bson.ObjectID) (*model.Obligation, error) {
	var obl model.Obligation
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&obl)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("obligation %s not found", id.Hex())
		}
		return nil, fmt.Errorf("find obligation: %w", err)
	}
	return &obl, nil
}

func (r *MongoRepository) FindByUserID(ctx context.Context, userID uuid.UUID, statusFilter []model.ObligationStatus, limit, offset int64) ([]*model.Obligation, error) {
	filter := bson.M{"user_id": userID}
	if len(statusFilter) > 0 {
		filter["status"] = bson.M{"$in": statusFilter}
	}

	opts := options.Find().
		SetSort(bson.M{"due_date": 1}).
		SetLimit(limit).
		SetSkip(offset)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("find obligations by user: %w", err)
	}
	defer cursor.Close(ctx)

	var obligations []*model.Obligation
	if err := cursor.All(ctx, &obligations); err != nil {
		return nil, fmt.Errorf("decode obligations: %w", err)
	}
	return obligations, nil
}

func (r *MongoRepository) FindByDocumentID(ctx context.Context, docID bson.ObjectID) ([]*model.Obligation, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"source_document_id": docID})
	if err != nil {
		return nil, fmt.Errorf("find obligations by document: %w", err)
	}
	defer cursor.Close(ctx)

	var obligations []*model.Obligation
	if err := cursor.All(ctx, &obligations); err != nil {
		return nil, fmt.Errorf("decode obligations: %w", err)
	}
	return obligations, nil
}

func (r *MongoRepository) FindUpcoming(ctx context.Context, userID uuid.UUID, daysAhead int) ([]*model.Obligation, error) {
	deadline := time.Now().Add(time.Duration(daysAhead) * 24 * time.Hour)

	filter := bson.M{
		"user_id":  userID,
		"due_date": bson.M{"$lte": deadline, "$gte": time.Now()},
		"status": bson.M{"$in": []model.ObligationStatus{
			model.ObligationStatusDetected,
			model.ObligationStatusPendingApproval,
			model.ObligationStatusInProgress,
		}},
	}

	opts := options.Find().SetSort(bson.M{"due_date": 1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("find upcoming obligations: %w", err)
	}
	defer cursor.Close(ctx)

	var obligations []*model.Obligation
	if err := cursor.All(ctx, &obligations); err != nil {
		return nil, fmt.Errorf("decode obligations: %w", err)
	}
	return obligations, nil
}

func (r *MongoRepository) UpdateStatus(ctx context.Context, id bson.ObjectID, status model.ObligationStatus) error {
	result, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"status":     status,
			"updated_at": time.Now().UTC(),
		}},
	)
	if err != nil {
		return fmt.Errorf("update obligation status: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("obligation %s not found", id.Hex())
	}
	return nil
}

func (r *MongoRepository) UpdateAgentOutput(ctx context.Context, id bson.ObjectID, output *model.AgentOutput) error {
	field := "analyst_output"
	if output.AgentType == "verifier" {
		field = "verifier_output"
	}

	result, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			field:        output,
			"updated_at": time.Now().UTC(),
		}},
	)
	if err != nil {
		return fmt.Errorf("update agent output: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("obligation %s not found", id.Hex())
	}
	return nil
}

func (r *MongoRepository) UpdateVerification(ctx context.Context, id bson.ObjectID, status model.ObligationStatus, riskLevel model.RiskLevel, output *model.AgentOutput) error {
	update := bson.M{
		"status":          status,
		"verifier_output": output,
		"updated_at":      time.Now().UTC(),
	}
	if riskLevel != "" {
		update["risk_level"] = riskLevel
	}

	result, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": update},
	)
	if err != nil {
		return fmt.Errorf("update obligation verification: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("obligation %s not found", id.Hex())
	}
	return nil
}

func (r *MongoRepository) SetApprovedAction(ctx context.Context, id bson.ObjectID, action *model.Action) error {
	result, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"approved_action": action,
			"status":          model.ObligationStatusInProgress,
			"updated_at":      time.Now().UTC(),
		}},
	)
	if err != nil {
		return fmt.Errorf("set approved action: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("obligation %s not found", id.Hex())
	}
	return nil
}

func (r *MongoRepository) LinkMarketplaceOpportunity(ctx context.Context, oblID, oppID bson.ObjectID) error {
	result, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": oblID},
		bson.M{"$set": bson.M{
			"marketplace_opportunity_id": oppID,
			"updated_at":                 time.Now().UTC(),
		}},
	)
	if err != nil {
		return fmt.Errorf("link marketplace opportunity: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("obligation %s not found", oblID.Hex())
	}
	return nil
}

func (r *MongoRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete obligation: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("obligation %s not found", id.Hex())
	}
	return nil
}

func (r *MongoRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	result, err := r.collection.DeleteMany(ctx, bson.M{"user_id": userID})
	if err != nil {
		return 0, fmt.Errorf("delete obligations by user: %w", err)
	}
	return result.DeletedCount, nil
}
