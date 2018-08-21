package users

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

//DynamoDBStore is an implementation of the Store interface for AWS DynamoDB
type DynamoDBStore struct {
	client    *dynamodb.DynamoDB
	tableName string
	keyName   string
}

//NewDynamoDBStore constructs a new DynamoDBStore. The table identified by tableName
//should already exist.
func NewDynamoDBStore(client *dynamodb.DynamoDB, tableName string, keyName string) *DynamoDBStore {
	return &DynamoDBStore{
		client:    client,
		tableName: tableName,
		keyName:   keyName,
	}
}

//Get returns the user associated with the provided userName
func (d *DynamoDBStore) Get(userName string) (*User, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(d.tableName),
		Key:       d.getKey(userName),
	}
	result, err := d.client.GetItem(input)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %v", err)
	}
	if result.Item == nil {
		return nil, nil
	}
	user := &User{}
	if err := dynamodbattribute.UnmarshalMap(result.Item, user); err != nil {
		return nil, fmt.Errorf("error decoding user record: %v", err)
	}
	return user, nil
}

//Insert inserts a new user into the store
func (d *DynamoDBStore) Insert(user *User) error {
	vals, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		return fmt.Errorf("error encoding user: %v", err)
	}

	//defensive check: ensure vals contains an entry for the key name
	if _, found := vals[d.keyName]; !found {
		vals[d.keyName] = &dynamodb.AttributeValue{S: aws.String(user.UserName)}
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(d.tableName),
		Item:      vals,
	}
	if _, err := d.client.PutItem(input); err != nil {
		return fmt.Errorf("error inserting user: %v", err)
	}
	return nil
}

//Update updates properties of an existing user
func (d *DynamoDBStore) Update(userName string, updates *Updates) (*User, error) {
	vals, err := dynamodbattribute.MarshalMap(updates)
	if err != nil {
		return nil, fmt.Errorf("error encoding user updates: %v", err)
	}
	if len(vals) == 0 {
		return nil, fmt.Errorf("nothing to update")
	}

	var exprs []string
	exprNames := map[string]*string{}
	exprValues := map[string]*dynamodb.AttributeValue{}
	for k, v := range vals {
		exprs = append(exprs, fmt.Sprintf("%s = %s", "#"+k, ":"+k))
		exprNames["#"+k] = aws.String(k)
		exprValues[":"+k] = v
	}

	input := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(d.tableName),
		Key:                       d.getKey(userName),
		UpdateExpression:          aws.String("SET " + strings.Join(exprs, ", ")),
		ExpressionAttributeNames:  exprNames,
		ExpressionAttributeValues: exprValues,
		ReturnValues:              aws.String("ALL_NEW"),
	}

	result, err := d.client.UpdateItem(input)
	if err != nil {
		return nil, fmt.Errorf("error updating user: %v", err)
	}
	user := &User{}
	if err := dynamodbattribute.UnmarshalMap(result.Attributes, user); err != nil {
		return nil, fmt.Errorf("error decoding user record: %v", err)
	}
	return user, nil
}

//Delete deletes the user
func (d *DynamoDBStore) Delete(userName string) error {
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(d.tableName),
		Key:       d.getKey(userName),
	}
	if _, err := d.client.DeleteItem(input); err != nil {
		return fmt.Errorf("error deleting user: %v", err)
	}
	return nil
}

func (d *DynamoDBStore) getKey(userName string) map[string]*dynamodb.AttributeValue {
	return map[string]*dynamodb.AttributeValue{d.keyName: {S: aws.String(userName)}}
}
