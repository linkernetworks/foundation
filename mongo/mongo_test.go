package mongo

var collection = "test_collection"

type TestRecord struct {
	Id   string `bson:"_id"`
	Test string `bson:"test"`
}

func (t TestRecord) GetCollection() string {
	return collection
}

// func TestContextCRUD(t *testing.T) {
// 	assert := assert.New(t)

// 	cf := config.MustRead("../config/testing.json")
// 	as := NewServiceProviderFromConfig(cf)

// 	context := as.Mongo.NewContext()
// 	defer context.Close()
// 	assert.NotNil(context, "Context should not be nil after service provider")
// 	assert.NotNil(context.Session, "Context session should not be nil after service provider")

// 	count, err := context.Count(collection, bson.M{})
// 	assert.NoError(err)
// 	assert.Equal(0, count, "Count empty collection should return 0")

// 	record := TestRecord{
// 		Test: "test-content",
// 	}

// 	err = context.Insert(collection, record)
// 	assert.NoError(err)

// 	count, err = context.Count(collection, bson.M{})
// 	assert.NoError(err)
// 	assert.Equal(1, count, "Count collection should return 1 after insert")

// 	err = context.DropCollection(collection)
// 	assert.NoError(err)
// }
