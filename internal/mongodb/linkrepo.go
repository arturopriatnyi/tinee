package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"urx/internal/service"
)

const LinkCollectionName = "links"

// Link is service.Link entity for database.
type Link struct {
	URL string `mongodb:"url"`
	URX string `mongodb:"urx"`
}

// LinkRepo is link repository.
type LinkRepo struct {
	links *mongo.Collection
}

// NewLinkRepo creates and returns a new LinkRepo instance.
func NewLinkRepo(db *DB) *LinkRepo {
	return &LinkRepo{links: db.Collection(LinkCollectionName)}
}

// Save saves Link to database.
func (r *LinkRepo) Save(ctx context.Context, l service.Link) error {
	_, err := r.links.InsertOne(ctx, Link(l))

	return err
}

// FindByURL finds Link by URL.
func (r *LinkRepo) FindByURL(ctx context.Context, URL string) (l service.Link, err error) {
	err = r.links.FindOne(ctx, bson.M{"url": URL}).Decode(&l)
	if err == mongo.ErrNoDocuments {
		return l, service.ErrLinkNotFound
	}

	return l, err
}

// FindByURX finds Link by URX.
func (r *LinkRepo) FindByURX(ctx context.Context, URX string) (l service.Link, err error) {
	err = r.links.FindOne(ctx, bson.M{"urx": URX}).Decode(&l)
	if err == mongo.ErrNoDocuments {
		return l, service.ErrLinkNotFound
	}

	return l, err
}
