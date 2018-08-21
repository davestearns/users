package models

//Validator is implemented by models that can validate themselves
type Validator interface {
	Validate() error
}
