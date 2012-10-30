package mandira

import (
	"testing"
)

type M map[string]interface{}

type Test struct {
	template string
	context  interface{}
	expected string
}

func (t *Test) Run(tt *testing.T) {
	output := Render(t.template, t.context)
	if output != t.expected {
		tt.Errorf("%v expected %v, got %v", t.template, t.expected, output)
	}
}

type Data struct {
	A bool
	B string
}

type User struct {
	Name string
	Id   int64
}

type settings struct {
	Allow bool
}

func (u User) Func1() string {
	return u.Name
}

func (u *User) Func2() string {
	return u.Name
}

func (u *User) Func3() (map[string]string, error) {
	return map[string]string{"name": u.Name}, nil
}

func (u *User) Func4() (map[string]string, error) {
	return nil, nil
}

func (u *User) Func5() (*settings, error) {
	return &settings{true}, nil
}

func (u *User) Func6() ([]interface{}, error) {
	var v []interface{}
	v = append(v, &settings{true})
	return v, nil
}

func (u User) Truefunc1() bool {
	return true
}

func (u *User) Truefunc2() bool {
	return true
}

func makeVector(n int) []interface{} {
	var v []interface{}
	for i := 0; i < n; i++ {
		v = append(v, &User{"Mike", 1})
	}
	return v
}

type Category struct {
	Tag         string
	Description string
}

func (c Category) DisplayName() string {
	return c.Tag + " - " + c.Description
}

func TestMustacheEquivalentBasics(t *testing.T) {
	tests := []Test{
		{"Hello, World", nil, "Hello, World"},
		{"Hello, {{name}}", M{"name": "World"}, "Hello, World"},
		{"{{var}}", M{"var": "5 > 2"}, "5 &gt; 2"},
		{"{{{var}}}", M{"var": "5 > 2"}, "5 > 2"},
		{"{{a}}{{b}}{{c}}{{d}}", M{"a": "a", "b": "b", "c": "c", "d": "d"}, "abcd"},
		{"0{{a}}1{{b}}23{{c}}456{{d}}89", M{"a": "a", "b": "b", "c": "c", "d": "d"}, "0a1b23c456d89"},
		{"hello {{! comment }}world", M{}, "hello world"},
		//does not exist
		{`{{dne}}`, M{"name": "world"}, ""},
		{`{{dne}}`, User{"Mike", 1}, ""},
		{`{{dne}}`, &User{"Mike", 1}, ""},
		{`{{#has}}hi{{/has}}`, &User{"Mike", 1}, ""},
		//section tests
		{`{{#A}}{{B}}{{/A}}`, Data{true, "hello"}, "hello"},
		{`{{#A}}{{{B}}}{{/A}}`, Data{true, "5 > 2"}, "5 > 2"},
		{`{{#A}}{{B}}{{/A}}`, Data{true, "5 > 2"}, "5 &gt; 2"},
		{`{{#A}}{{B}}{{/A}}`, Data{false, "hello"}, ""},
		{`{{a}}{{#b}}{{b}}{{/b}}{{c}}`, M{"a": "a", "b": "b", "c": "c"}, "abc"},
		{`{{#A}}{{B}}{{/A}}`, struct{ A []struct{ B string } }{
			[]struct{ B string }{{"a"}, {"b"}, {"c"}}},
			"abc",
		},
		{`{{#A}}{{b}}{{/A}}`, struct{ A []map[string]string }{
			[]map[string]string{{"b": "a"}, {"b": "b"}, {"b": "c"}}}, "abc"},

		{`{{#users}}{{Name}}{{/users}}`, M{"users": []User{{"Mike", 1}}}, "Mike"},

		{`{{#users}}gone{{Name}}{{/users}}`, M{"users": nil}, ""},
		{`{{#users}}gone{{Name}}{{/users}}`, M{"users": (*User)(nil)}, ""},
		{`{{#users}}gone{{Name}}{{/users}}`, M{"users": []User{}}, ""},

		{`{{#users}}{{Name}}{{/users}}`, M{"users": []*User{{"Mike", 1}}}, "Mike"},
		{`{{#users}}{{Name}}{{/users}}`, M{"users": []interface{}{&User{"Mike", 12}}}, "Mike"},
		{`{{#users}}{{Name}}{{/users}}`, M{"users": makeVector(1)}, "Mike"},
		{`{{Name}}`, User{"Mike", 1}, "Mike"},
		{`{{Name}}`, &User{"Mike", 1}, "Mike"},
		{"{{#users}}\n{{Name}}\n{{/users}}", M{"users": makeVector(2)}, "Mike\nMike\n"},
		{"{{#users}}\r\n{{Name}}\r\n{{/users}}", M{"users": makeVector(2)}, "Mike\r\nMike\r\n"},
		//function tests
		{`{{#users}}{{Func1}}{{/users}}`, M{"users": []User{{"Mike", 1}}}, "Mike"},
		{`{{#users}}{{Func1}}{{/users}}`, M{"users": []*User{{"Mike", 1}}}, "Mike"},
		{`{{#users}}{{Func2}}{{/users}}`, M{"users": []*User{{"Mike", 1}}}, "Mike"},

		{`{{#users}}{{#Func3}}{{name}}{{/Func3}}{{/users}}`, M{"users": []*User{{"Mike", 1}}}, "Mike"},
		{`{{#users}}{{#Func4}}{{name}}{{/Func4}}{{/users}}`, M{"users": []*User{{"Mike", 1}}}, ""},
		{`{{#Truefunc1}}abcd{{/Truefunc1}}`, User{"Mike", 1}, "abcd"},
		{`{{#Truefunc1}}abcd{{/Truefunc1}}`, &User{"Mike", 1}, "abcd"},
		{`{{#Truefunc2}}abcd{{/Truefunc2}}`, &User{"Mike", 1}, "abcd"},
		{`{{#Func5}}{{#Allow}}abcd{{/Allow}}{{/Func5}}`, &User{"Mike", 1}, "abcd"},
		{`{{#user}}{{#Func5}}{{#Allow}}abcd{{/Allow}}{{/Func5}}{{/user}}`, M{"user": &User{"Mike", 1}}, "abcd"},
		{`{{#user}}{{#Func6}}{{#Allow}}abcd{{/Allow}}{{/Func6}}{{/user}}`, M{"user": &User{"Mike", 1}}, "abcd"},

		//context chaining
		{`hello {{#section}}{{name}}{{/section}}`, M{"section": map[string]string{"name": "world"}}, "hello world"},
		{`hello {{#section}}{{name}}{{/section}}`, M{"name": "bob", "section": map[string]string{"name": "world"}}, "hello world"},
		{`hello {{#bool}}{{#section}}{{name}}{{/section}}{{/bool}}`, M{"bool": true, "section": map[string]string{"name": "world"}}, "hello world"},
		{`{{#users}}{{canvas}}{{/users}}`, M{"canvas": "hello", "users": []User{{"Mike", 1}}}, "hello"},
		{`{{#categories}}{{DisplayName}}{{/categories}}`, map[string][]*Category{
			"categories": {&Category{"a", "b"}},
		}, "a - b"},
	}
	for _, test := range tests {
		test.Run(t)
	}
}

func TestSample(t *testing.T) {
	tests := []Test{
		{`Hello {{name}}
You have just won ${{value}}!
{{?if in_monaco}}
Well, ${{taxed_value}}, after taxes.
{{/if}}`, M{
			"name":        "Jason",
			"value":       10000,
			"taxed_value": 10000.0,
			"in_monaco":   true,
		},
			`Hello Jason
You have just won $10000!

Well, $10000, after taxes.
`},
		{`Hello {{name}}
You have just won ${{value}}!
{{?if in_monaco}}
Well, ${{taxed_value}}, after taxes.
{{/if}}`, M{
			"name":        "Jason",
			"value":       10000,
			"taxed_value": 10000.0,
			"in_monaco":   false,
		},
			`Hello Jason
You have just won $10000!
`},
	}
	for _, test := range tests {
		test.Run(t)
	}
}

func TestFilters(t *testing.T) {
	// TODO: test date filter, which must be written probably

	names := []string{"john", "bob", "fred"}
	tests := []Test{
		{"{{name}}", M{"name": "Jason"}, "Jason"},
		{"{{name|upper}}", M{"name": "jason"}, "JASON"},
		{"{{name|len}}", M{"name": "jason"}, "5"},
		{"{{name|index(3)}}", M{"name": "jason"}, "o"},
		{"{{name|index(0)}}", M{"name": []string{"john", "bob", "fred"}}, "john"},
		{"{{name|index(0)|upper}}", M{"name": names}, "JOHN"},
		{"{{name|index(1)|title}}", M{"name": names}, "Bob"},
		// index error returns empty string
		{"{{name|index(5)}}", M{"name": names}, ""},
		// index error doesn't blow up later on a filter chain
		{"{{name|index(5)|title}}", M{"name": names}, ""},
		{`{{name|format(">%s<")}}`, M{"name": "jason"}, "&gt;jason&lt;"},
		{`{{{name|format(">%s<")}}}`, M{"name": "jason"}, ">jason<"},
		{`{{names|join(", ")}}`, M{"names": names}, "john, bob, fred"},
		{`{{names|len|divisibleby(2)}}`, M{"names": names}, "false"},
		{`{{names|len|divisibleby(3)}}`, M{"names": names}, "true"},
		{`{{names|join(joiner)}}`, M{"names": names, "joiner": ", "}, "john, bob, fred"},
	}
	for _, test := range tests {
		test.Run(t)
	}
}

func TestIfBlocks(t *testing.T) {
	//names := []string{"john", "bob", "fred"}
	tests := []Test{
		{`{{?if name}}Hello{{/if}}`, M{"name": true}, "Hello"},
		{`{{?if name}}Hello{{/if}}`, M{"name": "hi"}, "Hello"},
		{`{{?if name}}Hello{{/if}}`, M{"name": 1}, "Hello"},
		{`{{?if name}}Hello{{/if}}`, M{"name": []string{"hi"}}, "Hello"},
		{`{{?if name|len > 4}}True{{/if}}`, M{"name": "alex"}, ""},
		{`{{?if name|len > 4}}True{{/if}}`, M{"name": "alexander"}, "True"},
		{`{{?if age|divisibleby(2)}}True{{/if}}`, M{"age": 30}, "True"},
		{`{{?if age|divisibleby(2)}}True{{/if}}`, M{"age": 31}, ""},
		{`{{?if name == "john"}}Yes!{{?else}}No!{{/if}}`, M{"name": "ted"}, "No!"},
		{`{{?if name == "john"}}Yes!{{?else}}No!{{/if}}`, M{"name": "john"}, "Yes!"},
		{`{{?if name == "john" or name == "ted"}}Yes!{{?else}}No!{{/if}}`, M{"name": "john"}, "Yes!"},
		{`{{?if name == "john" or name == "ted"}}Yes!{{?else}}No!{{/if}}`, M{"name": "ted"}, "Yes!"},
		{`{{?if name == "john" or name == "ted"}}Yes!{{?else}}No!{{/if}}`, M{"name": "fred"}, "No!"},
	}

	for _, test := range tests {
		test.Run(t)
	}

}
