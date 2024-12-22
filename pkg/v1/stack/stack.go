package stack

import (
	"log"
)

type Stack interface {
	MustInit(p Provider)
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

func (s *stack) MustInit(p Provider) {
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
}

func (a *AbstractProvider) Init() error {
	return nil
}

func (a *AbstractProvider) Close() {
}
