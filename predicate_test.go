package predicate

import (
	"encoding/json"
	"github.com/okcupid/jsonw"
	"testing"
	)

/*
 These should all compile!

 ["=", "foo", "bar"]
 ["is", "foo"]
 ["not", "foo"]
 ["contains", "foobar", "bar"]

 ["all", ["is", "available"], ["not", "movie"]]
 ["any", ["is", "movie"], ["is", "book"]]

 ["any", ["all", ["is", "movie"], ["is", "available"],
         ["all", ["is", "book"], ["is", "paperback"]]]]

 These should not!
 ["=", "foo"]
 ["any"]
 ["not"]
*/

type Props struct {
	props map[string]interface{}
}

func (p *Props) GetProperty(s string) interface{} {
	v, _ := p.props[s]
	return v
}

func (p *Props) SetProperty(s string, v interface{}) {
	p.props[s] = v
}

func makeWrapper(s string) *jsonw.Wrapper {
	var raw []interface{}
	_ = json.Unmarshal([]byte(s), &raw)
	return jsonw.NewWrapper(raw)	
}

func TestCompileValid(t *testing.T) {
	valid := []string{`["=", "foo", "bar"]`,
		`["is", "foo"]`,
		`["not", "foo"]`,
		`["contains", "foobar", "bar"]`,
		`["all", ["is", "available"], ["not", "movie"]]`,
		`["any", ["is", "movie"], ["is", "book"]]`,
		`["any", ["all", ["is", "movie"], ["is", "available"], ["all", ["is", "book"], ["is", "paperback"]]]]`}

	for _, j := range valid {
		p := makeWrapper(j)
		_, err := CompilePredicate(p)
		if err != nil {
			t.Errorf("%v is a valid rule and should compile!: %v", j, err)
		}
	}


	invalid := []string{`["=", "foo"]`,
		`["any"]`,
		`["not"]`}
	
	for _, j := range invalid {
		_, err := CompilePredicate(makeWrapper(j))
		if err == nil {
			t.Errorf("%v is an invalid rule and should not compile!", j)
		}
	}
}


/*
 Given an interest with the properties:
    
 available => true
 movie => true
 title => "Metropolis"
 year => 1927

 */
func TestPredicate(t *testing.T) {
	pp := &Props{make(map[string]interface{})}

	pp.SetProperty("available", true)
	pp.SetProperty("movie", true)
	pp.SetProperty("title", "Metropolis")
	pp.SetProperty("year", "1927")

	ispred, _ := CompilePredicate(makeWrapper(`["is", "available"]`))
	if !ispred.Test(pp) {
		t.Errorf("Metropolis is available, so predicate failed")
	}

	notpred, _ := CompilePredicate(makeWrapper(`["not", "foo"]`))
	if !notpred.Test(pp) {
		t.Errorf("Metropolis is not foo, but predicate failed")
	}

	eqpred, _ := CompilePredicate(makeWrapper(`["=", "year", "1927"]`))
	if !eqpred.Test(pp) {
		t.Errorf("Metropolis was in 1927, so predicate failed")
	}

	containspred, _ := CompilePredicate(makeWrapper(`["contains", "title", "metro"]`))
	if !containspred.Test(pp) {
		t.Errorf("Metropolis contains 'metro', so predicate failed")
	}

	allpred, _ := CompilePredicate(makeWrapper(`["all", ["is", "available"], ["is", "movie"]]`))
	if !allpred.Test(pp) {
		t.Errorf("Metropolis is available and is movie, but predicate failed")
	}

	anypred, _ := CompilePredicate(makeWrapper(`["any", ["not", "available"], ["is", "movie"]]`))
	if !anypred.Test(pp) {
		t.Errorf("Metropolis was either available, or not a movie, so predicate failed")
	}
}
