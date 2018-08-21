package users

import (
	"fmt"
	"net/mail"

	"github.com/nbutton23/zxcvbn-go"
	"golang.org/x/crypto/bcrypt"
)

//bcryptCost is the cost factor used when generating bcrypt password hashes.
//This is a var and not a const so that automated tests can set it to the minimum.
var bcryptCost = 14

//NewUser represents a new user being added to the system
type NewUser struct {
	//Required Fields

	//UserName is the unique screen name for this user
	UserName string `json:"userName"`
	//Password is the user's password
	Password string `json:"password"`

	//Optional Fields

	//PersonalName is the user's personal (first) name
	PersonalName string `json:"personalName"`
	//FamilyName is he user's family (last) name
	FamilyName string `json:"familyName"`
	//Email is the user's email address
	Email string `json:"email"`
	//Mobile is the user's mobile/SMS number
	Mobile string `json:"mobile"`
}

//Validate validates the NewUser
func (nu *NewUser) Validate() error {
	//UserName must be non-zero-length
	if len(nu.UserName) == 0 {
		return fmt.Errorf("userName must be supplied")
	}
	//password must be complex enough
	if len(nu.Password) == 0 {
		return fmt.Errorf("password must be supplied")
	}
	passScore := zxcvbn.PasswordStrength(nu.Password, []string{nu.Email}).Score
	if passScore < 2 {
		return fmt.Errorf("password is not strong enough: score (%d) must be >= 2", passScore)
	}
	//email must be valid if provided
	if len(nu.Email) > 0 {
		if _, err := mail.ParseAddress(nu.Email); err != nil {
			return fmt.Errorf("invalid email address: %v", err)
		}
	}
	return nil
}

//ToUser Validates the NewUser and converts it to a User for storage
func (nu *NewUser) ToUser() (*User, error) {
	if err := nu.Validate(); err != nil {
		return nil, err
	}
	passhash, err := bcrypt.GenerateFromPassword([]byte(nu.Password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("error generating password hash: %v", err)
	}
	return &User{
		UserName:     nu.UserName,
		Email:        nu.Email,
		PasswordHash: passhash,
		PersonalName: nu.PersonalName,
		FamilyName:   nu.FamilyName,
	}, nil
}

//User represents a user account stored in the system
type User struct {
	UserName     string `json:"userName"`
	PasswordHash []byte `json:"-" dynamodbav:"passwordHash"`
	PersonalName string `json:"personalName,omitempty"`
	FamilyName   string `json:"familyName,omitempty"`
	Email        string `json:"-" dynamodbav:"email,omitempty"`
	Mobile       string `json:"-" dynamodbav:"mobile,omitempty"`
}

//Authenticate authenticates the user using the provided password
func (u *User) Authenticate(password []byte) error {
	return bcrypt.CompareHashAndPassword(u.PasswordHash, password)
}

//DummyAuthenticate consumes about the same amount of time
//as user.Authenticate() does, but does nothing. This should
//be used during sign-in when the provided userName is not found,
//so that an attacker can't see a difference in response time
//between an invalid userName and a valid userName with invalid password.
func DummyAuthenticate() {
	bcrypt.GenerateFromPassword([]byte("dummy password"), bcryptCost)
}

//Updates represents updates to a user profile sent by the client.
//Only the fields listed here are updatable.
type Updates struct {
	PersonalName *string `json:"personalName,omitempty"`
	FamilyName   *string `json:"familyName,omitempty"`
	Email        *string `json:"email,omitempty"`
	Mobile       *string `json:"mobile,omitempty"`
}

//Credentials represents a user's sign-in credentials
type Credentials struct {
	UserName string `json:"userName,omitempty"`
	Password string `json:"password,omitempty"`
}
