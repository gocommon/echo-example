package validator

import (
	"sync"

	"github.com/labstack/echo"

	validator "gopkg.in/go-playground/validator.v9"
)

var _ echo.Validator = &Validator{}

// New New
func New() *Validator {
	v := &Validator{}
	v.pool.New = func() interface{} {
		return validator.New()
	}

	return v
}

// Validator Validator
type Validator struct {
	pool sync.Pool
}

// Validate Validate
func (p *Validator) Validate(i interface{}) error {
	v := p.pool.Get().(*validator.Validate)
	if err := v.Struct(i); err != nil {
		p.pool.Put(v)
		return err
	}

	p.pool.Put(v)
	return nil
}
