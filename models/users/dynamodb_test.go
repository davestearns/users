package users

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func TestDynamoDBStore(t *testing.T) {
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		t.Fatalf("error creating new AWS session: %v", err)
	}
	client := dynamodb.New(sess)
	store := NewDynamoDBStore(client, "users", "userName")
	if err != nil {
		t.Fatalf("error creating DynamoDBStore: %v", err)
	}

	suffix := time.Now().UnixNano()
	userName := fmt.Sprintf("test-%d", suffix)
	user := &User{
		UserName:     userName,
		PersonalName: "Tester",
		FamilyName:   "Account",
		Email:        "test@test.com",
		Mobile:       "206-555-1212",
	}
	if err := store.Insert(user); err != nil {
		t.Errorf("error inserting new user: %v", err)
	}

	gotUser, err := store.Get(userName)
	if err != nil {
		t.Errorf("error getting previously inserted user %s: %v", userName, err)
	} else {
		if !reflect.DeepEqual(gotUser, user) {
			t.Errorf("fetched user does not match inserted user: expected %+v but got %+v", user, gotUser)
		}
	}

	updates := &Updates{
		FamilyName: aws.String("UPDATED"),
	}
	updatedUser, err := store.Update(userName, updates)
	if err != nil {
		t.Errorf("error updating user %s: %v", userName, err)
	} else {
		if updatedUser.FamilyName != "UPDATED" {
			t.Errorf("returned user did not have updates applied: expected familyName='UPDATED' but got familyName='%s'",
				updatedUser.FamilyName)
		}
	}

	if err := store.Delete(userName); err != nil {
		t.Errorf("error deleting user %s: %v", userName, err)
	}
}
