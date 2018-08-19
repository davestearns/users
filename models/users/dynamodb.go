package users

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

const keyName = "userName"

//DynamoDBStore is an implementation of the Store interface for AWS DynamoDB
type DynamoDBStore struct {
	client    *dynamodb.DynamoDB
	tableName string
}

//NewDynamoDBStore constructs a new DynamoDBStore
func NewDynamoDBStore(client *dynamodb.DynamoDB, tableName string) (*DynamoDBStore, error) {
	if err := ensureTable(client, tableName); err != nil {
		return nil, fmt.Errorf("error ensuring table '%s': %v", tableName, err)
	}
	return &DynamoDBStore{
		client:    client,
		tableName: tableName,
	}, nil
}

func ensureTable(client *dynamodb.DynamoDB, tableName string) error {
	dtInput := &dynamodb.DescribeTableInput{
		TableName: &tableName,
	}
	_, err := client.DescribeTable(dtInput)
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if !ok || awsErr.Code() != dynamodb.ErrCodeResourceNotFoundException {
			return fmt.Errorf("error describing table '%s': %v", tableName, err)
		}
		if awsErr.Code() == dynamodb.ErrCodeResourceNotFoundException {
			ctInput := &dynamodb.CreateTableInput{
				TableName: &tableName,
				AttributeDefinitions: []*dynamodb.AttributeDefinition{
					{
						AttributeName: aws.String(keyName),
						AttributeType: aws.String("S"),
					},
				},
				KeySchema: []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String(keyName),
						KeyType:       aws.String("HASH"),
					},
				},
				ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(5),
					WriteCapacityUnits: aws.Int64(5),
				},
			}
			_, err := client.CreateTable(ctInput)
			if err != nil {
				return fmt.Errorf("error creating table '%s': %v", tableName, err)
			}
			if err := client.WaitUntilTableExists(dtInput); err != nil {
				return fmt.Errorf("error waiting for table '%s' to exist: %v", tableName, err)
			}
		}
	}
	return nil
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
	return map[string]*dynamodb.AttributeValue{keyName: {S: aws.String(userName)}}
}
