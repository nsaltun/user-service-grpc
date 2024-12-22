package stack

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStacks(t *testing.T) {
	s := New()
	db := newDb()
	second := newSecondProvider()

	//initialize providers
	func() {
		defer s.Close()
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("panic recovered!")
			}
		}()
		s.Init(db)
		s.Init(second)
		assert.True(t, second.IsInitCalled())
	}()

	assert.True(t, db.IsInitCalled())
	assert.True(t, db.IsCloseCalled())
	assert.True(t, second.IsInitCalled())
	assert.True(t, second.IsCloseCalled())

}

type SecondProvider interface {
	Provider
	IsInitCalled() bool  // for testing purpose
	IsCloseCalled() bool // for testing purpose
}

type secondProvider struct {
	AbstractProvider
	initCalled  bool // for testing purpose
	closeCalled bool // for testing purpose
}

func newSecondProvider() SecondProvider {
	return &secondProvider{}
}

func (d *secondProvider) Init() error {
	d.initCalled = true
	fmt.Println("seconddep init")
	return errors.New("err while initing second dep")
}

func (d *secondProvider) IsInitCalled() bool {
	return d.initCalled
}
func (d *secondProvider) IsCloseCalled() bool {
	return d.closeCalled
}

// func (d *secondDep) Close() {
// 	fmt.Println("seconddep closed")
// }

type DBInstance interface {
	Provider
	IsInitCalled() bool  // for testing purpose
	IsCloseCalled() bool // for testing purpose
}
type dbInstance struct {
	initCalled  bool // for testing purpose
	closeCalled bool // for testing purpose
}

func newDb() DBInstance {
	return &dbInstance{}
}

func (d *dbInstance) Init() error {
	d.initCalled = true
	fmt.Println("DBInstance initialized successfully")
	return nil
}

func (d *dbInstance) Close() {
	d.closeCalled = true
	fmt.Println("close for db instance")
}

func (d *dbInstance) IsInitCalled() bool {
	return d.initCalled
}
func (d *dbInstance) IsCloseCalled() bool {
	return d.closeCalled
}
