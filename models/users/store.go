package users

//Store describes what a user store can do
type Store interface {
	Get(userName string) (*User, error)
	Insert(user *User) error
	Update(userName string, updates *Updates) (*User, error)
	Delete(userName string) error
}
