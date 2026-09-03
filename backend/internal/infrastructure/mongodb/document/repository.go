package document

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/domain/document/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const collectionName = "documents"

// MongoRepository implements document.DocumentRepository using MongoDB.
type MongoRepository struct {
	collection *mongo.Collection
}

// NewMongoRepository creates a new MongoDB-backed document repository.
func NewMongoRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{
		collection: db.Collection(collectionName),
	}
}

func (r *MongoRepository) Create(ctx context.Context, doc *model.Document) error {
	if doc.ID.IsZero() {
		doc.ID = bson.NewObjectID()
	}
	doc.CreatedAt = time.Now().UTC()
	doc.UpdatedAt = doc.CreatedAt

	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("insert document: %w", err)
	}
	return nil
}

func (r *MongoRepository) FindByID(ctx context.Context, id bson.ObjectID) (*model.Document, error) {
	var doc model.Document
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("document %s not found", id.Hex())
		}
		return nil, fmt.Errorf("find document: %w", err)
	}
	return &doc, nil
}

func (r *MongoRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int64) ([]*model.Document, error) {
	opts := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetLimit(limit).
		SetSkip(offset)

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, fmt.Errorf("find documents by user: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []*model.Document
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode documents: %w", err)
	}
	return docs, nil
}

func (r *MongoRepository) FindByOrganizationID(ctx context.Context, orgID uuid.UUID, limit, offset int64) ([]*model.Document, error) {
	opts := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetLimit(limit).
		SetSkip(offset)

	cursor, err := r.collection.Find(ctx, bson.M{"organization_id": orgID}, opts)
	if err != nil {
		return nil, fmt.Errorf("find documents by org: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []*model.Document
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode documents: %w", err)
	}
	return docs, nil
}

func (r *MongoRepository) UpdateStatus(ctx context.Context, id bson.ObjectID, status model.DocumentStatus) error {
	result, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"status":     status,
			"updated_at": time.Now().UTC(),
		}},
	)
	if err != nil {
		return fmt.Errorf("update document status: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("document %s not found", id.Hex())
	}
	return nil
}

func (r *MongoRepository) UpdateExtractionResult(ctx context.Context, id bson.ObjectID, result *model.ExtractionResult) error {
	updateResult, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"extraction_result": result,
			"updated_at":        time.Now().UTC(),
		}},
	)
	if err != nil {
		return fmt.Errorf("update extraction result: %w", err)
	}
	if updateResult.MatchedCount == 0 {
		return fmt.Errorf("document %s not found", id.Hex())
	}
	return nil
}

func (r *MongoRepository) UpdateContent(ctx context.Context, id bson.ObjectID, rawContent, redactedContent string) error {
	update := bson.M{
		"updated_at": time.Now().UTC(),
	}
	if rawContent != "" {
		update["raw_content"] = rawContent
	}
	if redactedContent != "" {
		update["redacted_content"] = redactedContent
	}

	updateResult, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": update},
	)
	if err != nil {
		return fmt.Errorf("update document content: %w", err)
	}
	if updateResult.MatchedCount == 0 {
		return fmt.Errorf("document %s not found", id.Hex())
	}
	return nil
}

func (r *MongoRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete document: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("document %s not found", id.Hex())
	}
	return nil
}

func (r *MongoRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	result, err := r.collection.DeleteMany(ctx, bson.M{"user_id": userID})
	if err != nil {
		return 0, fmt.Errorf("delete documents by user: %w", err)
	}
	return result.DeletedCount, nil
}
