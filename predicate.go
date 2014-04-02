package predicate

import (
	"fmt"
	"errors"
	"github.com/okcupid/jsonw"
	"reflect"
	"strings"
	)

type PropertyProvider interface {
	GetProperty(p string) interface{}
	SetProperty(p string, v interface{})
}

type Predicate interface {
	Test(i PropertyProvider) bool
}

type IsPred struct {
	property string
}

type NotPred struct {
	property string
}

type EqPred struct {
	property string
	value interface{}
}

type ContainsPred struct {
	property string
	substr string
}

type AllPred struct {
	predicates []Predicate
}

type AnyPred struct {
	predicates []Predicate
}

func CompilePredicate(json *jsonw.Wrapper) (p Predicate, err error) {
	var operator string
	var l int

	if l, err = json.Len(); (err != nil || l <= 1) {
		return nil, errors.New(fmt.Sprintf("length(%v) of predicate to small", l))
	} else if operator, err = json.AtIndex(0).GetString(); err != nil {
		return nil, errors.New(fmt.Sprintf("Couldn't get predicate operator: %v", err))
	}

  if operator == "any" || operator == "all" {
              preds := make([]Predicate, l - 1)
		for i := 1; i < l; i++ {
			if np, err := CompilePredicate(json.AtIndex(i)); err == nil {
				preds[i-1] = np
			} else {
				err = errors.New(fmt.Sprintf("error compiling '%v' rule: %v", operator, err))
				return nil, err
			}
		}
		if operator == "any" {
			p = &AnyPred{preds}
		} else {
			p = &AllPred{preds}
		}
	} else {
		return compileSimple(l, operator, json)
	}

	return p, err
}

func compileSimple(l int, operator string, json *jsonw.Wrapper) (p Predicate, err error) {
	if (operator == "is" || operator == "not") && l == 2 {
		if property, err := json.AtIndex(1).GetString(); err == nil {
			if operator == "is" {
				return &IsPred{property}, nil
			} else {
				return &NotPred{property}, nil
			}
		} 
		return nil, errors.New(fmt.Sprintf("invalid '%v' rule", operator))
	} else if (operator == "=" && l == 3) {
		if property, err := json.AtIndex(1).GetString(); err == nil {
			if value, err := json.AtIndex(2).GetData(); err == nil {
				return &EqPred{property, value}, nil
			}
		} 
		return nil, errors.New(fmt.Sprintf("invalid '=' rule: %v", err))
	} else if (operator == "contains" && l == 3) {
		if property, err := json.AtIndex(1).GetString(); err == nil {
			if value, err := json.AtIndex(2).GetString(); err == nil {
				return &ContainsPred{property, strings.ToLower(value)}, nil
			}
		} 
		return nil, errors.New(fmt.Sprintf("invalid '=' rule: %v", err))
	} else {
		err = errors.New(fmt.Sprintf("invalid '%v' rule", operator))
	}

	return p, err
}

func (p *IsPred) Test(thing PropertyProvider) bool {
	value := thing.GetProperty(p.property)
	return value == true
}

func (p *IsPred) String() string {
	return fmt.Sprintf("(is '%v'?)", p.property)
}

func (p *NotPred) Test(thing PropertyProvider) bool {
	value := thing.GetProperty(p.property)
	return value == false || value == nil
}

func (p *NotPred) String() string {
	return fmt.Sprintf("(not '%v'?)", p.property)
}

func (p *EqPred) Test(thing PropertyProvider) bool {
	value := thing.GetProperty(p.property)
	return value == p.value
}

func (p *EqPred) String() string {
	return fmt.Sprintf("(= '%v' '%v'?)", p.property, p.value)
}

func (p *ContainsPred) Test(thing PropertyProvider) bool {
	value := thing.GetProperty(p.property)
	propertyValue := reflect.ValueOf(value)
	return strings.Contains(strings.ToLower(propertyValue.String()), p.substr)
}

func (p *ContainsPred) String() string {
	return fmt.Sprintf("(contains? '%v' '%v'?)", p.property, p.substr)
}

func (p *AllPred) Test(thing PropertyProvider) bool {
	for _, p := range p.predicates {
		if !p.Test(thing) {
			return false
		}
	}
	return true
}

func (p *AllPred) String() string {
	return fmt.Sprintf("(all %v?)", p.predicates)
}


func (p *AnyPred) Test(thing PropertyProvider) bool {
	for _, p := range p.predicates {
		if p.Test(thing) {
			return true
		}
	}
	return false
}

func (p *AnyPred) String() string {
	return fmt.Sprintf("(any %v?)", p.predicates)
}
