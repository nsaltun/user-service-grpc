package stack

import (
	"fmt"
	"log"
)

type Stack interface {
	Init(p Provider)
	Close()
}

type stack struct {
	providers []Provider
}

type Provider interface {
	Init() error
	Close()
}

func New() Stack {
	return &stack{}
}

func (s *stack) Init(p Provider) {
	s.providers = append(s.providers, p)
	if err := p.Init(); err != nil {
		log.Panicf("err while init a provider!. err:%v", err.Error())
	}

}
func (s *stack) Close() {
	for _, p := range s.providers {
		p.Close()
	}
}

type AbstractProvider struct {
	Provider
	initCalled  bool
	closeCalled bool
}

func (a *AbstractProvider) Init() error {
	a.initCalled = true
	return nil
}

func (a *AbstractProvider) Close() {
	a.closeCalled = true
	fmt.Println("close for abstract provider")
}
