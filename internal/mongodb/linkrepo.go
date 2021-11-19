package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"urx/internal/service"
)

// Link is service.Link entity for the database.
type Link struct {
	ID    string `mongodb:"_id"`
	URL   string `mongodb:"url"`
	Alias string `mongodb:"alias"`
}

// LinkRepo is the link repository.
type LinkRepo struct {
	links *mongo.Collection
}

// LinkCollectionName is the name of link collection.
const LinkCollectionName = "links"

// NewLinkRepo creates and returns a new LinkRepo instance.
func NewLinkRepo(db *DB) *LinkRepo {
	return &LinkRepo{links: db.Collection(LinkCollectionName)}
}

// Save saves a Link to the database.
func (r *LinkRepo) Save(ctx context.Context, l service.Link) error {
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"_id": l.ID}
	update := bson.M{"$set": l}

	_, err := r.links.UpdateOne(ctx, filter, update, opts)
	return err
}

// FindByURL finds a Link by URL.
func (r *LinkRepo) FindByURL(ctx context.Context, URL string) (l service.Link, err error) {
	err = r.links.FindOne(ctx, bson.M{"url": URL}).Decode(&l)
	if err == mongo.ErrNoDocuments {
		return l, service.ErrLinkNotFound
	}

	return l, err
}

// FindByAlias finds a Link by alias.
func (r *LinkRepo) FindByAlias(ctx context.Context, alias string) (l service.Link, err error) {
	err = r.links.FindOne(ctx, bson.M{"alias": alias}).Decode(&l)
	if err == mongo.ErrNoDocuments {
		return l, service.ErrLinkNotFound
	}

	return l, err
}
