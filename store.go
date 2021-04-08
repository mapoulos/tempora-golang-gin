package main

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type MeditationRecord struct {
	Pk         string     `dynamodbav:"pk"`
	Sk         string     `dynamodbav:"sk"`
	Meditation Meditation `dynamodbav:"meditation"`
}

type MeditationStore interface {
	SaveMeditation(m Meditation) error
	ListMeditations(userId string) ([]Meditation, error)
	GetMeditation(userId string, id string) (Meditation, error)
	DeleteMeditation(userId string, id string) error
	UpdateMeditation(m Meditation) error
}

type MemoryMeditationStore struct {
	meditationsMap map[string][]Meditation
}

type DynamoMeditationStore struct {
	sess      *session.Session
	svc       *dynamodb.DynamoDB
	tableName string
}

func NewDynamoMeditationStore(tableName string, local bool, createTable bool) DynamoMeditationStore {
	dynamoStore := DynamoMeditationStore{
		tableName: tableName,
	}

	if local {
		dynamoStore.sess = session.Must(session.NewSession(&aws.Config{
			Region:   aws.String("us-east-1"),
			Endpoint: aws.String("http://127.0.0.1:9000"),
			//EndPoint: aws.String("https://dynamodb.us-east-1.amazonaws.com"),
		}))
	} else {
		dynamoStore.sess = session.Must(session.NewSession(&aws.Config{
			Region:   aws.String("us-east-1"),
			Endpoint: aws.String("https://dynamodb.us-east-1.amazonaws.com"),
		}))
	}
	dynamoStore.svc = dynamodb.New(dynamoStore.sess)

	if createTable {
		createTableParams := &dynamodb.CreateTableInput{
			TableName: aws.String(tableName),
			KeySchema: []*dynamodb.KeySchemaElement{
				{AttributeName: aws.String("pk"), KeyType: aws.String("HASH")},
				{AttributeName: aws.String("sk"), KeyType: aws.String("RANGE")},
			},
			AttributeDefinitions: []*dynamodb.AttributeDefinition{
				{AttributeName: aws.String("pk"), AttributeType: aws.String("S")},
				{AttributeName: aws.String("sk"), AttributeType: aws.String("S")},
			},
			BillingMode: aws.String(dynamodb.BillingModePayPerRequest),
		}

		_, err := dynamoStore.svc.CreateTable(createTableParams)

		if err != nil {
			fmt.Println("Error in table creation")
			fmt.Println(err.Error())
		}
	}

	return dynamoStore
}

func mapMeditationToMeditationRecord(m Meditation) MeditationRecord {
	pk := m.UserId
	sk := m.Name + "/" + m.ID

	return MeditationRecord{
		Pk:         pk,
		Sk:         sk,
		Meditation: m,
	}
}

func (store DynamoMeditationStore) SaveMeditation(meditation Meditation) error {
	meditationRecord := mapMeditationToMeditationRecord(meditation)
	meditationAVMap, err := dynamodbattribute.MarshalMap(meditationRecord)
	if err != nil {
		return err
	}

	params := &dynamodb.PutItemInput{
		TableName: aws.String(store.tableName),
		Item:      meditationAVMap,
	}

	_, err = store.svc.PutItem(params)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (store DynamoMeditationStore) ListMeditations(userId string) ([]Meditation, error) {

	params := &dynamodb.QueryInput{
		TableName:              aws.String(store.tableName),
		KeyConditionExpression: aws.String("pk = :userId"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":userId": {
				S: aws.String(userId),
			},
		},
	}

	resp, err := store.svc.Query(params)
	if err != nil {
		fmt.Println(err)
		return []Meditation{}, err
	}

	var meditationRecords []MeditationRecord
	err = dynamodbattribute.UnmarshalListOfMaps(resp.Items, &meditationRecords)

	if err != nil {
		return []Meditation{}, err
	}

	meditations := make([]Meditation, len(meditationRecords))
	for i, m := range meditationRecords {
		meditations[i] = m.Meditation
	}

	return meditations, nil
}

func (store DynamoMeditationStore) GetMeditation(userId string, id string) (Meditation, error) {
	meditations, err := store.ListMeditations(userId)
	if err != nil {
		return Meditation{}, err
	}
	var meditation Meditation
	for _, m := range meditations {
		if m.ID == id {
			meditation = m
		}
	}
	if (Meditation{}) == meditation {
		return Meditation{}, errors.New("No meditation with ID " + id + " was found")
	}
	return meditation, nil
}

func (store DynamoMeditationStore) DeleteMeditation(userId string, id string) error {
	meditation, err := store.GetMeditation(userId, id)
	if err != nil {
		return err
	}
	params := &dynamodb.DeleteItemInput{
		TableName: aws.String(store.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"pk": {
				S: aws.String(userId),
			},
			"sk": {
				S: aws.String(meditation.Name + "/" + meditation.ID),
			},
		},
	}
	resp, err := store.svc.DeleteItem(params)
	fmt.Println(resp)
	if err != nil {
		return err
	}
	return nil
}

func (store DynamoMeditationStore) UpdateMeditation(m Meditation) error {
	oldMeditation, err := store.GetMeditation(m.UserId, m.ID)
	if err != nil {
		return err
	}

	if (Meditation{}) == oldMeditation {
		return errors.New("No meditation with " + m.ID + " found.")
	}

	if m.Name == oldMeditation.Name {
		err = store.SaveMeditation(m)
	} else {
		mAV := mapMeditationToMeditationRecord(m)

		newItem, err := dynamodbattribute.MarshalMap(mAV)

		if err != nil {
			fmt.Println(err.Error())
		}

		params := &dynamodb.TransactWriteItemsInput{
			TransactItems: []*dynamodb.TransactWriteItem{
				{
					Put: &dynamodb.Put{
						TableName: &store.tableName,
						Item:      newItem,
					},
				},
				{
					Delete: &dynamodb.Delete{
						TableName: &store.tableName,
						Key: map[string]*dynamodb.AttributeValue{
							"pk": {
								S: aws.String(oldMeditation.UserId),
							},
							"sk": {
								S: aws.String(oldMeditation.Name + "/" + oldMeditation.ID),
							},
						},
					},
				},
			},
		}
		store.svc.TransactWriteItems(params)

	}

	if err != nil {
		return err
	}

	return nil
}

func NewMemoryMeditationStore() MemoryMeditationStore {
	store := MemoryMeditationStore{
		meditationsMap: map[string][]Meditation{},
	}
	return store
}

func (store MemoryMeditationStore) SaveMeditation(meditation Meditation) error {
	userSlice, ok := store.meditationsMap[meditation.UserId]

	if !ok {
		userSlice = []Meditation{meditation}
	} else {
		userSlice = append(userSlice, meditation)
	}
	store.meditationsMap[meditation.UserId] = userSlice

	return nil
}

func (store MemoryMeditationStore) ListMeditations(userId string) ([]Meditation, error) {
	return store.meditationsMap[userId], nil
}

func (store MemoryMeditationStore) GetMeditation(userId string, id string) (Meditation, error) {
	emptyMeditation := Meditation{
		Name: "",
		URL:  "",
		ID:   "",
	}
	meditations, ok := store.meditationsMap[userId]
	if !ok {
		return emptyMeditation, errors.New("No user with id " + userId + " was found")
	}
	for _, m := range meditations {
		if m.ID == id {
			return m, nil
		}
	}

	return emptyMeditation, errors.New("No meditation with id " + id + " was found")
}

func (store MemoryMeditationStore) DeleteMeditation(userId string, id string) error {
	meditations, ok := store.meditationsMap[userId]
	if !ok {
		return errors.New("No user with id " + userId + " was found")
	}
	idxToDelete := -1
	for i, m := range meditations {
		if m.ID == id {
			idxToDelete = i
		}
	}

	if idxToDelete > -1 {
		finalIdx := len(meditations) - 1
		meditations[idxToDelete], meditations[finalIdx] = meditations[finalIdx], meditations[idxToDelete]
		meditations = meditations[:finalIdx]

		store.meditationsMap[userId] = meditations

		return nil
	}

	return errors.New("No meditation with id " + id + " was found")
}

func (store MemoryMeditationStore) UpdateMeditation(m Meditation) error {
	userId := m.UserId
	meditations, ok := store.meditationsMap[userId]
	if !ok {
		return errors.New("No user with id " + userId + " was found")
	}
	for i, existingMeditation := range meditations {
		if existingMeditation.ID == m.ID {
			meditations[i] = m
			return nil
		}
	}
	return errors.New("No medition with id " + m.ID + " was found")
}
