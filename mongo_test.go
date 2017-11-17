package mongo

var collection = "test_collection"

type TestRecord struct {
	Id   string `bson:"_id"`
	Test string `bson:"test"`
}

func (t TestRecord) GetCollection() string {
	return collection
}

// func TestDataStoreCRUD(t *testing.T) {
// 	assert := assert.New(t)

// 	cf := config.Read("../config/testing.json")
// 	as := NewServiceProviderFromConfig(cf)

// 	dataStore := as.Mongo.NewDataStore()
// 	defer dataStore.Close()
// 	assert.NotNil(dataStore, "DataStore should not be nil after service provider")
// 	assert.NotNil(dataStore.Session, "DataStore session should not be nil after service provider")

// 	count, err := dataStore.Count(collection, bson.M{})
// 	assert.NoError(err)
// 	assert.Equal(0, count, "Count empty collection should return 0")

// 	record := TestRecord{
// 		Test: "test-content",
// 	}

// 	err = dataStore.Insert(collection, record)
// 	assert.NoError(err)

// 	count, err = dataStore.Count(collection, bson.M{})
// 	assert.NoError(err)
// 	assert.Equal(1, count, "Count collection should return 1 after insert")

// 	err = dataStore.DropCollection(collection)
// 	assert.NoError(err)
// }
