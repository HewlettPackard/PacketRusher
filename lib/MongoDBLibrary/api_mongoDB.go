package MongoDBLibrary

import (
	"context"
	"encoding/json"
	"my5G-RANTester/lib/MongoDBLibrary/logger"
	"time"

	jsonpatch "github.com/evanphx/json-patch"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client = nil
var dbName string

func SetMongoDB(setdbName string, url string) {

	if Client != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	defer cancel()
	if err != nil {
		//defer cancel()
		logger.MongoDBLog.Panic(err.Error())
	}
	Client = client
	dbName = setdbName
}

func RestfulAPIGetOne(collName string, filter bson.M) map[string]interface{} {

	collection := Client.Database(dbName).Collection(collName)

	var result map[string]interface{}
	collection.FindOne(context.TODO(), filter).Decode(&result)

	return result
}

func RestfulAPIGetMany(collName string, filter bson.M) []map[string]interface{} {
	collection := Client.Database(dbName).Collection(collName)

	var resultArray []map[string]interface{}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	cur, err := collection.Find(ctx, filter)
	defer cancel()
	if err != nil {
		logger.MongoDBLog.Fatal(err)
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var result map[string]interface{}
		err := cur.Decode(&result)
		if err != nil {
			logger.MongoDBLog.Fatal(err)
		}
		resultArray = append(resultArray, result)
	}
	if err := cur.Err(); err != nil {
		logger.MongoDBLog.Fatal(err)
	}

	return resultArray

}

func RestfulAPIPutOne(collName string, filter bson.M, putData map[string]interface{}) bool {
	collection := Client.Database(dbName).Collection(collName)

	var checkItem map[string]interface{}
	collection.FindOne(context.TODO(), filter).Decode(&checkItem)

	if checkItem == nil {
		collection.InsertOne(context.TODO(), putData)
		return false
	} else {
		collection.UpdateOne(context.TODO(), filter, bson.M{"$set": putData})
		return true
	}
}

func RestfulAPIPutOneNotUpdate(collName string, filter bson.M, putData map[string]interface{}) bool {
	collection := Client.Database(dbName).Collection(collName)

	var checkItem map[string]interface{}
	collection.FindOne(context.TODO(), filter).Decode(&checkItem)

	if checkItem == nil {
		collection.InsertOne(context.TODO(), putData)
		return false
	} else {
		// collection.UpdateOne(context.TODO(), filter, bson.M{"$set": putData})
		return true
	}
}

func RestfulAPIPutMany(collName string, filterArray []bson.M, putDataArray []map[string]interface{}) bool {
	collection := Client.Database(dbName).Collection(collName)

	var checkItem map[string]interface{}
	for i, putData := range putDataArray {
		checkItem = nil
		filter := filterArray[i]
		collection.FindOne(context.TODO(), filter).Decode(&checkItem)

		if checkItem == nil {
			collection.InsertOne(context.TODO(), putData)
		} else {
			collection.UpdateOne(context.TODO(), filter, bson.M{"$set": putData})
		}
	}

	if checkItem == nil {
		return false
	} else {
		return true
	}

}

func RestfulAPIDeleteOne(collName string, filter bson.M) {
	collection := Client.Database(dbName).Collection(collName)

	collection.DeleteOne(context.TODO(), filter)
}

func RestfulAPIDeleteMany(collName string, filter bson.M) {
	collection := Client.Database(dbName).Collection(collName)

	collection.DeleteMany(context.TODO(), filter)
}

func RestfulAPIMergePatch(collName string, filter bson.M, patchData map[string]interface{}) bool {
	collection := Client.Database(dbName).Collection(collName)

	var originalData map[string]interface{}
	result := collection.FindOne(context.TODO(), filter)

	if err := result.Decode(&originalData); err != nil { // Data doesn't exist in DB
		return false
	} else {
		delete(originalData, "_id")
		original, _ := json.Marshal(originalData)

		patchDataByte, err := json.Marshal(patchData)
		if err != nil {
			logger.MongoDBLog.Panic(err)
		}

		modifiedAlternative, err := jsonpatch.MergePatch(original, patchDataByte)
		if err != nil {
			logger.MongoDBLog.Panic(err)
		}

		var modifiedData map[string]interface{}

		json.Unmarshal(modifiedAlternative, &modifiedData)
		collection.UpdateOne(context.TODO(), filter, bson.M{"$set": modifiedData})
		return true
	}
}

func RestfulAPIJSONPatch(collName string, filter bson.M, patchJSON []byte) bool {
	collection := Client.Database(dbName).Collection(collName)

	var originalData map[string]interface{}
	result := collection.FindOne(context.TODO(), filter)

	if err := result.Decode(&originalData); err != nil { // Data doesn't exist in DB
		return false
	} else {
		delete(originalData, "_id")
		original, _ := json.Marshal(originalData)

		patch, err := jsonpatch.DecodePatch(patchJSON)
		if err != nil {
			logger.MongoDBLog.Panic(err)
		}

		modified, err := patch.Apply(original)
		if err != nil {
			logger.MongoDBLog.Panic(err)
		}

		var modifiedData map[string]interface{}

		json.Unmarshal(modified, &modifiedData)
		collection.UpdateOne(context.TODO(), filter, bson.M{"$set": modifiedData})
		return true
	}

}

func RestfulAPIJSONPatchExtend(collName string, filter bson.M, patchJSON []byte, dataName string) bool {
	collection := Client.Database(dbName).Collection(collName)

	var originalDataCover map[string]interface{}
	result := collection.FindOne(context.TODO(), filter)

	if err := result.Decode(&originalDataCover); err != nil { // Data does'nt exist in db
		return false
	} else {
		delete(originalDataCover, "_id")
		originalData := originalDataCover[dataName]
		original, _ := json.Marshal(originalData)

		jsonpatch.DecodePatch(patchJSON)
		patch, err := jsonpatch.DecodePatch(patchJSON)
		if err != nil {
			logger.MongoDBLog.Panic(err)
		}

		modified, err := patch.Apply(original)
		if err != nil {
			logger.MongoDBLog.Panic(err)
		}

		var modifiedData map[string]interface{}
		json.Unmarshal(modified, &modifiedData)
		collection.UpdateOne(context.TODO(), filter, bson.M{"$set": bson.M{dataName: modifiedData}})
		return true
	}
}

func RestfulAPIPost(collName string, filter bson.M, postData map[string]interface{}) bool {
	collection := Client.Database(dbName).Collection(collName)

	var checkItem map[string]interface{}
	collection.FindOne(context.TODO(), filter).Decode(&checkItem)

	if checkItem == nil {
		collection.InsertOne(context.TODO(), postData)
		return false
	} else {
		collection.UpdateOne(context.TODO(), filter, bson.M{"$set": postData})
		return true
	}
}
