package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/XanderD99/disruptor/pkg/db"
)

type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
	config   Config
}

type Config struct {
	Hosts    []string      `env:"HOSTS" default:"localhost:27017"`
	Database string        `env:"DATABASE" default:"disruptor"`
	Auth     AuthConfig    `envPrefix:"AUTH_"`
	Pool     PoolConfig    `envPrefix:"POOL_"`
	Timeout  TimeoutConfig `envPrefix:"TIMEOUT_"`
}

type PoolConfig struct {
	MinSize       uint64        `env:"MIN_SIZE" default:"10"`
	MaxSize       uint64        `env:"MAX_SIZE" default:"100"`
	MaxConnecting uint64        `env:"MAX_CONNECTING" default:"10"`
	MaxIdleTime   time.Duration `env:"MAX_IDLE_TIME" default:"30m"`
}

type TimeoutConfig struct {
	Connect time.Duration `env:"CONNECT" default:"10s"`
	Query   time.Duration `env:"QUERY" default:"30s"`
}

type AuthConfig struct {
	Enabled   bool   `env:"ENABLED" default:"true"`
	Username  string `env:"USERNAME"`
	Password  string `env:"PASSWORD"`
	Mechanism string `env:"MECHANISM" default:"SCRAM-SHA-256"`
	Source    string `env:"SOURCE" default:"admin"`
}

func New(config Config) db.Database {
	return &MongoDB{config: config}
}

func (m *MongoDB) Connect(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, m.config.Timeout.Connect)
	defer cancel()

	opts := options.Client().
		SetHosts(m.config.Hosts).
		SetAppName("disruptor").
		// Connection pool settings
		SetMaxPoolSize(m.config.Pool.MaxSize).             // Max connections in pool
		SetMinPoolSize(m.config.Pool.MinSize).             // Min connections to maintain
		SetMaxConnIdleTime(m.config.Pool.MaxIdleTime).     // Close idle connections
		SetMaxConnecting(m.config.Pool.MaxConnecting).     // Max concurrent connections being established
		SetConnectTimeout(m.config.Timeout.Connect).       // Connection timeout
		SetServerSelectionTimeout(m.config.Timeout.Query). // Server selection timeout
		// Heartbeat and monitoring
		SetHeartbeatInterval(10 * time.Second).
		SetSocketTimeout(60 * time.Second).
		// Compression
		SetCompressors([]string{"snappy", "zlib", "zstd"}).
		// Read preferences for better load distribution
		SetReadPreference(readpref.SecondaryPreferred())

	if m.config.Auth.Enabled {
		opts = opts.SetAuth(options.Credential{
			Username:      m.config.Auth.Username,
			Password:      m.config.Auth.Password,
			AuthMechanism: m.config.Auth.Mechanism,
			AuthSource:    m.config.Auth.Source,
			PasswordSet:   m.config.Auth.Password != "",
		})
	}

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	m.client = client
	m.database = client.Database(m.config.Database)

	return nil
}

func (m *MongoDB) Disconnect() error {
	if m.client == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), m.config.Timeout.Connect)
	defer cancel()
	return m.client.Disconnect(ctx)
}

func (m *MongoDB) Create(ctx context.Context, table string, entity any) error {
	collection := m.database.Collection(table)
	_, err := collection.InsertOne(ctx, entity)
	return err
}

func (m *MongoDB) FindByID(ctx context.Context, table string, id any) (any, error) {
	collection := m.database.Collection(table)

	var result bson.M
	err := collection.FindOne(ctx, bson.M{"id": id}).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (m *MongoDB) FindAll(ctx context.Context, table string, opts ...db.FindOption) ([]any, error) {
	collection := m.database.Collection(table)

	options := &db.FindOptions{}
	for _, opt := range opts {
		opt(options)
	}

	findOpts := m.buildFindOptions(options)

	cursor, err := collection.Find(ctx, options.Filters, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Convert to []any
	items := make([]any, len(results))
	for i, result := range results {
		items[i] = result
	}

	return items, nil
}

func (m *MongoDB) Update(ctx context.Context, table string, entity any) error {
	collection := m.database.Collection(table)

	id, err := db.GetEntityID(entity)
	if err != nil {
		return err
	}

	updateDoc, err := m.toBsonDWithoutID(entity)
	if err != nil {
		return err
	}

	result, err := collection.UpdateOne(
		ctx,
		bson.M{"id": id},
		bson.M{"$set": updateDoc},
	)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("no document found with id = %v", id)
	}

	return nil
}

func (m *MongoDB) Upsert(ctx context.Context, table string, entity any) error {
	collection := m.database.Collection(table)

	id, err := db.GetEntityID(entity)
	if err != nil {
		return err
	}

	updateDoc, err := m.toBsonDWithoutID(entity)
	if err != nil {
		return err
	}

	opts := options.Update().SetUpsert(true)
	_, err = collection.UpdateOne(
		ctx,
		bson.M{"id": id},
		bson.M{"$set": updateDoc},
		opts,
	)

	return err
}

func (m *MongoDB) Delete(ctx context.Context, table string, id any) error {
	collection := m.database.Collection(table)

	// Try string ID first, then numeric
	_, err := collection.DeleteOne(ctx, bson.M{"id": id})

	return err
}

func (m *MongoDB) Count(ctx context.Context, table string, opts ...db.FindOption) (int64, error) {
	collection := m.database.Collection(table)

	options := &db.FindOptions{}
	for _, opt := range opts {
		opt(options)
	}

	return collection.CountDocuments(ctx, options.Filters)
}

// Helper methods
func (m *MongoDB) buildFindOptions(opts *db.FindOptions) *options.FindOptions {
	findOpts := options.Find()

	if opts.Limit > 0 {
		findOpts.SetLimit(int64(opts.Limit))
	}

	if opts.Offset > 0 {
		findOpts.SetSkip(int64(opts.Offset))
	}

	if len(opts.Sort) > 0 {
		sortDoc := bson.D{}
		for field, direction := range opts.Sort {
			dir := 1
			if direction == db.SortDescending {
				dir = -1
			}
			sortDoc = append(sortDoc, bson.E{Key: field, Value: dir})
		}
		findOpts.SetSort(sortDoc)
	}

	if len(opts.Projection) > 0 {
		projection := bson.M{}
		for _, field := range opts.Projection {
			projection[field] = 1
		}
		findOpts.SetProjection(projection)
	}

	return findOpts
}

func (m *MongoDB) toBsonDWithoutID(entity any) (bson.D, error) {
	data, err := bson.Marshal(entity)
	if err != nil {
		return nil, err
	}

	var doc bson.D
	if err := bson.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	// Remove _id if present
	var cleaned bson.D
	for _, elem := range doc {
		if elem.Key != "_id" {
			cleaned = append(cleaned, elem)
		}
	}

	return cleaned, nil
}
