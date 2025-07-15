package mongo

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"

	"github.com/XanderD99/discord-disruptor/pkg/database"
)

var _ database.Database = (*MongoDB)(nil)

type Config struct {
	Hosts []string `env:"MONGO_HOSTS" default:"localhost:27017"`
	Auth  struct {
		Enabled   bool   `env:"MONGO_AUTH_ENABLED" default:"true"`
		Username  string `env:"MONGO_AUTH_USERNAME"`
		Password  string `env:"MONGO_AUTH_PASSWORD"`
		Mechanism string `env:"MONGO_AUTH_MECHANISM" default:"SCRAM-SHA-256"`
		Source    string `env:"MONGO_AUTH_SOURCE" default:"admin"`
	} `env:"MONGO_AUTH" default:"true"`
	Database string `env:"MONGO_DATABASE" default:"disruptor"`
}

type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database

	config Config
}

func New(config Config) database.Database {
	return &MongoDB{
		config: config,
	}
}

func (m *MongoDB) Close() error {
	if m.client == nil {
		return errors.New("MongoDB client is not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return m.client.Disconnect(ctx)
}

func (m *MongoDB) Open(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	opts := options.Client().SetHosts(m.config.Hosts).SetAppName("disruptor")
	if m.config.Auth.Enabled {
		opts = opts.SetAuth(options.Credential{
			Username:      m.config.Auth.Username,
			Password:      m.config.Auth.Password,
			AuthMechanism: m.config.Auth.Mechanism,
			AuthSource:    m.config.Auth.Source,
			PasswordSet:   true, // Indicate that the password is set.

		})
	}

	if err := opts.Validate(); err != nil {
		return fmt.Errorf("invalid Mongo client options: %w", err)
	}

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return err
	}

	db := client.Database(m.config.Database)
	if err := db.Client().Ping(ctx, nil); err != nil {
		return fmt.Errorf("pinging the DB: %w", err)
	}

	m.client = client
	m.database = db

	return nil
}

func (m *MongoDB) Create(ctx context.Context, entity any) error {
	collection := m.collectionFor(entity)

	_, err := collection.InsertOne(ctx, entity)
	return err
}

func (m *MongoDB) Upsert(ctx context.Context, entity any) error {
	collection := m.collectionFor(entity)
	idValue, err := getEntityID(entity)
	if err != nil {
		return err
	}
	elemType := getElemType(entity)
	typedID, err := getTypedId(idValue, elemType)
	if err != nil {
		return err
	}
	updateDoc, err := toBsonDWithoutID(entity)
	if err != nil {
		return err
	}
	filter := bson.M{"id": typedID}
	update := bson.M{"$set": updateDoc}
	opts := options.Update().SetUpsert(true)
	res, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 && res.UpsertedCount == 0 {
		return fmt.Errorf("no document found with id = %v and no new document created", idValue)
	}
	return nil
}

func (m *MongoDB) Update(ctx context.Context, entity any) error {
	collection := m.collectionFor(entity)

	idValue, err := getEntityID(entity)
	if err != nil {
		return err
	}

	// Exclude `_id` from the update document
	updateDoc, err := toBsonDWithoutID(entity)
	if err != nil {
		return err
	}

	elemType := getElemType(entity)
	typedID, err := getTypedId(idValue, elemType)
	if err != nil {
		return err
	}

	filter := bson.M{"id": typedID}
	update := bson.M{"$set": updateDoc}

	res, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return fmt.Errorf("no document found with id = %v", idValue)
	}
	return nil
}

func (m *MongoDB) Delete(ctx context.Context, id string, entity any) error {
	collection := m.collectionFor(entity)

	// Determine the correct ID type from the entity
	elemType := getElemType(entity)
	typedID, err := getTypedId(id, elemType)
	if err != nil {
		return err
	}
	_, err = collection.DeleteOne(ctx, bson.M{"id": typedID})
	return err
}

func (m *MongoDB) FindAll(ctx context.Context, entity any, opts ...database.FindAllOption) (any, error) {
	entityType := reflect.TypeOf(entity)
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}

	sliceType := reflect.SliceOf(entityType)
	slicePtr := reflect.New(sliceType).Interface()

	collection := m.collectionFor(entity)

	o := database.FindAllOptions{}
	for _, opt := range opts {
		opt(&o)
	}

	findOptions := options.Find()
	if o.Pagination.Limit > 0 {
		findOptions.SetLimit(int64(o.Pagination.Limit))
	}
	if o.Pagination.Offset > 0 {
		findOptions.SetSkip(int64(o.Pagination.Offset))
	}

	if len(o.Sort) > 0 {
		sortDoc := bson.D{}
		for _, s := range o.Sort {
			dir := 1
			if s.Direction == "desc" {
				dir = -1
			}
			sortDoc = append(sortDoc, bson.E{Key: s.Field, Value: dir})
		}
		findOptions.SetSort(sortDoc)
	}

	cursor, err := collection.Find(ctx, o.Filters, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, slicePtr); err != nil {
		return nil, err
	}

	return reflect.ValueOf(slicePtr).Elem().Interface(), nil
}

func (m *MongoDB) FindByID(ctx context.Context, id string, entity any) (any, error) {
	collection := m.collectionFor(entity)

	elemType := getElemType(entity)
	typedID, err := getTypedId(id, elemType)
	if err != nil {
		return nil, err
	}
	result := reflect.New(elemType).Interface()
	err = collection.FindOne(ctx, bson.M{"id": typedID}).Decode(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *MongoDB) collectionFor(entity any) *mongo.Collection {
	t := reflect.TypeOf(entity)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return m.database.Collection(strings.ToLower(t.Name() + "s")) //simple plural name for collections
}

func getEntityID(entity any) (string, error) {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	idField := v.FieldByName("ID")
	if !idField.IsValid() {
		return "", errors.New("ID field not found")
	}

	if idField.Kind() == reflect.String {
		return idField.String(), nil
	}

	if idField.Kind() == reflect.Int || idField.Kind() == reflect.Int64 || idField.Kind() == reflect.Int32 {
		return fmt.Sprintf("%d", idField.Int()), nil
	}

	if idField.Kind() == reflect.Uint || idField.Kind() == reflect.Uint64 || idField.Kind() == reflect.Uint32 {
		return fmt.Sprintf("%d", idField.Uint()), nil
	}

	return "", errors.New("unsupported ID type")
}

func toBsonDWithoutID(entity any) (bson.D, error) {
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

func getElemType(entity any) reflect.Type {
	elemType := reflect.TypeOf(entity)
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}

	return elemType
}

func getTypedId(id string, elemType reflect.Type) (any, error) {
	idField, ok := elemType.FieldByName("ID")
	if !ok {
		return nil, fmt.Errorf("entity does not have an ID field")
	}

	// Convert string ID to correct type
	var typedID any
	switch idField.Type.Kind() {
	case reflect.Int, reflect.Int64:
		if intVal, err := strconv.Atoi(id); err == nil {
			typedID = intVal
		} else {
			return nil, fmt.Errorf("invalid int ID: %v", err)
		}
	case reflect.Uint, reflect.Uint64:
		if uintVal, err := strconv.ParseUint(id, 10, 64); err == nil {
			typedID = uintVal
		} else {
			return nil, fmt.Errorf("invalid uint ID: %v", err)
		}
	case reflect.String:
		typedID = id
	default:
		return nil, fmt.Errorf("unsupported ID type: %s", idField.Type.Kind())
	}

	return typedID, nil
}
