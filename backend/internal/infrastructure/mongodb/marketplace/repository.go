package marketplace

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/domain/marketplace/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// MongoRepository implements marketplace.MarketplaceRepository using MongoDB.
type MongoRepository struct {
	opportunities *mongo.Collection
	bids          *mongo.Collection
	transactions  *mongo.Collection
}

// NewMongoRepository creates a new MongoDB-backed marketplace repository.
func NewMongoRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{
		opportunities: db.Collection("marketplace_opportunities"),
		bids:          db.Collection("bids"),
		transactions:  db.Collection("transactions"),
	}
}

// --- Opportunities ---

func (r *MongoRepository) CreateOpportunity(ctx context.Context, opp *model.MarketplaceOpportunity) error {
	if opp.ID.IsZero() {
		opp.ID = bson.NewObjectID()
	}
	opp.CreatedAt = time.Now().UTC()
	opp.UpdatedAt = opp.CreatedAt

	_, err := r.opportunities.InsertOne(ctx, opp)
	if err != nil {
		return fmt.Errorf("insert opportunity: %w", err)
	}
	return nil
}

func (r *MongoRepository) FindOpportunityByID(ctx context.Context, id bson.ObjectID) (*model.MarketplaceOpportunity, error) {
	var opp model.MarketplaceOpportunity
	err := r.opportunities.FindOne(ctx, bson.M{"_id": id}).Decode(&opp)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("opportunity %s not found", id.Hex())
		}
		return nil, fmt.Errorf("find opportunity: %w", err)
	}
	return &opp, nil
}

func (r *MongoRepository) FindOpportunitiesByUserID(ctx context.Context, userID uuid.UUID, statusFilter []model.OpportunityStatus, limit, offset int64) ([]*model.MarketplaceOpportunity, error) {
	filter := bson.M{"user_id": userID}
	if len(statusFilter) > 0 {
		filter["status"] = bson.M{"$in": statusFilter}
	}

	opts := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetLimit(limit).
		SetSkip(offset)

	cursor, err := r.opportunities.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("find opportunities: %w", err)
	}
	defer cursor.Close(ctx)

	var opps []*model.MarketplaceOpportunity
	if err := cursor.All(ctx, &opps); err != nil {
		return nil, fmt.Errorf("decode opportunities: %w", err)
	}
	return opps, nil
}

func (r *MongoRepository) UpdateOpportunityStatus(ctx context.Context, id bson.ObjectID, status model.OpportunityStatus) error {
	result, err := r.opportunities.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"status":     status,
			"updated_at": time.Now().UTC(),
		}},
	)
	if err != nil {
		return fmt.Errorf("update opportunity status: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("opportunity %s not found", id.Hex())
	}
	return nil
}

// --- Bids ---

func (r *MongoRepository) CreateBid(ctx context.Context, bid *model.Bid) error {
	if bid.ID.IsZero() {
		bid.ID = bson.NewObjectID()
	}
	bid.CreatedAt = time.Now().UTC()

	_, err := r.bids.InsertOne(ctx, bid)
	if err != nil {
		return fmt.Errorf("insert bid: %w", err)
	}

	// Increment bid count on the opportunity
	_, _ = r.opportunities.UpdateOne(ctx,
		bson.M{"_id": bid.OpportunityID},
		bson.M{
			"$inc": bson.M{"bid_count": 1},
			"$set": bson.M{"updated_at": time.Now().UTC()},
		},
	)

	return nil
}

func (r *MongoRepository) FindBidByID(ctx context.Context, id bson.ObjectID) (*model.Bid, error) {
	var bid model.Bid
	err := r.bids.FindOne(ctx, bson.M{"_id": id}).Decode(&bid)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("bid %s not found", id.Hex())
		}
		return nil, fmt.Errorf("find bid: %w", err)
	}
	return &bid, nil
}

func (r *MongoRepository) FindBidsByOpportunityID(ctx context.Context, oppID bson.ObjectID) ([]*model.Bid, error) {
	opts := options.Find().SetSort(bson.M{"price.amount": 1}) // cheapest first

	cursor, err := r.bids.Find(ctx, bson.M{"opportunity_id": oppID}, opts)
	if err != nil {
		return nil, fmt.Errorf("find bids: %w", err)
	}
	defer cursor.Close(ctx)

	var bids []*model.Bid
	if err := cursor.All(ctx, &bids); err != nil {
		return nil, fmt.Errorf("decode bids: %w", err)
	}
	return bids, nil
}

func (r *MongoRepository) UpdateBidStatus(ctx context.Context, id bson.ObjectID, status model.BidStatus) error {
	result, err := r.bids.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"status": status}},
	)
	if err != nil {
		return fmt.Errorf("update bid status: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("bid %s not found", id.Hex())
	}
	return nil
}

func (r *MongoRepository) UpdateBidRiskScore(ctx context.Context, id bson.ObjectID, score float64, notes string) error {
	result, err := r.bids.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"risk_score": score,
			"risk_notes": notes,
		}},
	)
	if err != nil {
		return fmt.Errorf("update bid risk score: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("bid %s not found", id.Hex())
	}
	return nil
}

// --- Transactions ---

func (r *MongoRepository) CreateTransaction(ctx context.Context, txn *model.Transaction) error {
	if txn.ID.IsZero() {
		txn.ID = bson.NewObjectID()
	}
	txn.CreatedAt = time.Now().UTC()
	txn.UpdatedAt = txn.CreatedAt

	_, err := r.transactions.InsertOne(ctx, txn)
	if err != nil {
		return fmt.Errorf("insert transaction: %w", err)
	}
	return nil
}

func (r *MongoRepository) FindTransactionByID(ctx context.Context, id bson.ObjectID) (*model.Transaction, error) {
	var txn model.Transaction
	err := r.transactions.FindOne(ctx, bson.M{"_id": id}).Decode(&txn)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("transaction %s not found", id.Hex())
		}
		return nil, fmt.Errorf("find transaction: %w", err)
	}
	return &txn, nil
}

func (r *MongoRepository) FindTransactionsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int64) ([]*model.Transaction, error) {
	opts := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetLimit(limit).
		SetSkip(offset)

	cursor, err := r.transactions.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, fmt.Errorf("find transactions: %w", err)
	}
	defer cursor.Close(ctx)

	var txns []*model.Transaction
	if err := cursor.All(ctx, &txns); err != nil {
		return nil, fmt.Errorf("decode transactions: %w", err)
	}
	return txns, nil
}

func (r *MongoRepository) UpdateTransactionStatus(ctx context.Context, id bson.ObjectID, status model.TransactionStatus) error {
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now().UTC(),
		},
	}
	if status == model.TransactionStatusCompleted {
		now := time.Now().UTC()
		update["$set"].(bson.M)["completed_at"] = now
	}

	result, err := r.transactions.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("update transaction status: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("transaction %s not found", id.Hex())
	}
	return nil
}

// --- Aggregations ---

func (r *MongoRepository) TotalGMV(ctx context.Context, userID uuid.UUID) (*model.Money, error) {
	pipeline := bson.A{
		bson.M{"$match": bson.M{
			"user_id": userID,
			"status":  model.TransactionStatusCompleted,
		}},
		bson.M{"$group": bson.M{
			"_id":      "$amount.currency",
			"total":    bson.M{"$sum": "$amount.amount"},
		}},
	}

	cursor, err := r.transactions.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregate GMV: %w", err)
	}
	defer cursor.Close(ctx)

	var results []struct {
		Currency string  `bson:"_id"`
		Total    float64 `bson:"total"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("decode GMV: %w", err)
	}

	if len(results) == 0 {
		return &model.Money{Amount: 0, Currency: "USD"}, nil
	}

	// Return the first currency's total (simplification for MVP)
	return &model.Money{
		Amount:   results[0].Total,
		Currency: results[0].Currency,
	}, nil
}

func (r *MongoRepository) TotalSavings(ctx context.Context, userID uuid.UUID) (*model.Money, error) {
	pipeline := bson.A{
		bson.M{"$match": bson.M{
			"user_id":                    userID,
			"status":                     model.TransactionStatusCompleted,
			"savings_realized":           bson.M{"$exists": true},
		}},
		bson.M{"$group": bson.M{
			"_id":   "$savings_realized.currency",
			"total": bson.M{"$sum": "$savings_realized.amount"},
		}},
	}

	cursor, err := r.transactions.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregate savings: %w", err)
	}
	defer cursor.Close(ctx)

	var results []struct {
		Currency string  `bson:"_id"`
		Total    float64 `bson:"total"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("decode savings: %w", err)
	}

	if len(results) == 0 {
		return &model.Money{Amount: 0, Currency: "USD"}, nil
	}

	return &model.Money{
		Amount:   results[0].Total,
		Currency: results[0].Currency,
	}, nil
}

// --- GDPR/KVKK ---

func (r *MongoRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	filter := bson.M{"user_id": userID}
	var total int64

	res1, err := r.opportunities.DeleteMany(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("delete opportunities: %w", err)
	}
	total += res1.DeletedCount

	// Delete bids for user's opportunities (via opportunity_id lookup)
	// Note: In production, this would use a join or cascade logic
	res2, err := r.bids.DeleteMany(ctx, filter)
	if err != nil {
		return total, fmt.Errorf("delete bids: %w", err)
	}
	total += res2.DeletedCount

	res3, err := r.transactions.DeleteMany(ctx, filter)
	if err != nil {
		return total, fmt.Errorf("delete transactions: %w", err)
	}
	total += res3.DeletedCount

	return total, nil
}
