package patch

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyIgnoresEmptyPatch(t *testing.T) {

	type Target struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Salary    int    `json:"salary"`
		Extra     string `json:"extra"`
	}
	a := Target{
		FirstName: "Anakin",
		LastName:  "Skywalker",
		Salary:    123,
		Extra:     "unchanged",
	}
	chg, err := Apply(&a, map[string]interface{}{})

	assert.NoError(t, err)
	assert.False(t, chg)
	assert.Equal(t, "Anakin", a.FirstName)
	assert.Equal(t, "Skywalker", a.LastName)
	assert.Equal(t, 123, a.Salary)
	assert.Equal(t, "unchanged", a.Extra)
}

func TestApplyIgnoresUnknownFields(t *testing.T) {
	type Target struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Salary    int    `json:"salary"`
	}

	a := Target{
		FirstName: "Anakin",
		LastName:  "Skywalker",
		Salary:    123,
	}

	data := `{"middle_name":"Sheev", "perk": "pilot"}`
	p := make(map[string]interface{})
	jsonErr := json.Unmarshal([]byte(data), &p)
	assert.NoError(t, jsonErr)

	var chg, err = Apply(&a, p)

	assert.NoError(t, err)
	assert.False(t, chg)
	assert.Equal(t, "Anakin", a.FirstName)   // unchanged
	assert.Equal(t, "Skywalker", a.LastName) // unchanged
	assert.Equal(t, 123, a.Salary)           // unchanged
}

func TestApplyIgnoresUnexportedFields(t *testing.T) {
	type Target struct {
		Exported   string `json:"exported,omitempty"`
		unexported string
	}
	a := Target{
		Exported:   "stormtrooper",
		unexported: "private",
	}

	data := `{"exported":"wookie", "unexported": "leutenant"}`
	p := make(map[string]interface{})
	jsonErr := json.Unmarshal([]byte(data), &p)
	assert.NoError(t, jsonErr)

	chg, err := Apply(&a, p)
	assert.NoError(t, err)
	assert.True(t, chg)
	assert.Equal(t, "wookie", a.Exported)
	assert.Equal(t, "private", a.unexported)
}

func TestApplyZeroValueFields(t *testing.T) {
	type Target struct {
		Name   string `json:"name"`
		Salary int64  `json:"salary"`
		OnDuty bool   `json:"on_duty"`
	}

	var a = Target{
		Name:   "Han Solo",
		Salary: 15,
		OnDuty: true,
	}

	data := `{"name":"Chubacca", "salary": 0, "on_duty": false}`
	p := make(map[string]interface{})
	jsonErr := json.Unmarshal([]byte(data), &p)
	assert.NoError(t, jsonErr)

	chg, err := Apply(&a, p)
	assert.NoError(t, err)
	assert.True(t, chg)
	assert.Equal(t, "Chubacca", a.Name)
	assert.Equal(t, int64(0), a.Salary)
	assert.Equal(t, false, a.OnDuty)
}

func TestApplyDetectsWrongType(t *testing.T) {
	type Target struct {
		Name   string `json:"name"`
		Salary int    `json:"salary"`
	}
	var a = Target{
		Name:   "Anakin Skywalker",
		Salary: 123,
	}

	data := `{"name":"Darth Vader", "salary": "euros"}`
	p := make(map[string]interface{})
	jsonErr := json.Unmarshal([]byte(data), &p)
	assert.NoError(t, jsonErr)

	_, err := Apply(&a, p)
	assert.Error(t, err)
	assert.Equal(t, 123, a.Salary) // unchanged
}

func TestApplySubStructs(t *testing.T) {
	type TargetPerson struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	type Target struct {
		Contact *TargetPerson `json:"contact"`
		Salary  int           `json:"salary"`
	}

	var a = Target{
		Contact: &TargetPerson{
			FirstName: "Anakin",
			LastName:  "Skywalker",
		},
		Salary: 123,
	}

	data := `{"contact": {"first_name":"Darth", "last_name": "Vader"}, "salary": 100500}`
	p := make(map[string]interface{})
	jsonErr := json.Unmarshal([]byte(data), &p)
	assert.NoError(t, jsonErr)

	chg, err := Apply(&a, p)
	assert.NoError(t, err)
	assert.True(t, chg)
	assert.Equal(t, "Darth", a.Contact.FirstName)
	assert.Equal(t, "Vader", a.Contact.LastName)
	assert.Equal(t, 100500, a.Salary)
}

func TestApplyNilSubStructs(t *testing.T) {
	type TargetPerson struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Position  string `json:"position"`
	}
	type Target struct {
		Contact *TargetPerson `json:"contact"`
		Salary  int           `json:"salary"`
	}

	var a = Target{
		Contact: nil,
		Salary:  123,
	}

	data := `{"contact": {"first_name":"Darth", "last_name": "Vader"}}`
	p := make(map[string]interface{})
	jsonErr := json.Unmarshal([]byte(data), &p)
	assert.NoError(t, jsonErr)

	chg, err := Apply(&a, p)
	assert.NoError(t, err)
	assert.True(t, chg)
	assert.Equal(t, a.Contact.FirstName, "Darth")
	assert.Equal(t, a.Contact.LastName, "Vader")
	assert.Equal(t, a.Contact.Position, "")
}

type Position int32

const (
	Position_None    Position = 0
	Position_Padawan Position = 1
	Position_Sith    Position = 2
)

func TestEnums(t *testing.T) {
	type Person struct {
		FirstName string   `json:"first_name"`
		LastName  string   `json:"last_name"`
		Position  Position `json:"position"`
	}

	var a = Person{
		FirstName: "Darth",
		LastName:  "Vader",
		Position:  Position_None,
	}

	data := `{"position": 2}`
	p := make(map[string]interface{})
	jsonErr := json.Unmarshal([]byte(data), &p)
	assert.NoError(t, jsonErr)

	chg, err := Apply(&a, p)
	assert.NoError(t, err)
	assert.True(t, chg)
	assert.Equal(t, "Darth", a.FirstName)
	assert.Equal(t, "Vader", a.LastName)
	assert.Equal(t, Position_Sith, a.Position)
}

func TestSkipConversionErrors(t *testing.T) {
	type Target struct {
		Characters []string `json:"characters"`
	}
	var a = &Target{
		Characters: []string{"Anakin Skywalker"},
	}

	newWrongCharacters := `{"characters":[1,2,3]}`
	p := make(map[string]interface{})
	jsonErr := json.Unmarshal([]byte(newWrongCharacters), &p)
	assert.NoError(t, jsonErr)

	_, err := Apply(&a, p)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "can't convert characters to dst type")
	assert.Equal(t, []string{"Anakin Skywalker"}, a.Characters) // unchanged
}

func TestUpdateArrays(t *testing.T) {
	type Target struct {
		Characters []string `json:"characters"`
	}
	var a = &Target{
		Characters: []string{"Anakin Skywalker"},
	}

	newCharacters := `{"characters":["Darth Vader", "Luke Skywalker"]}`
	p := make(map[string]interface{})
	jsonErr := json.Unmarshal([]byte(newCharacters), &p)
	assert.NoError(t, jsonErr)

	_, err := Apply(&a, p)
	assert.NoError(t, err)
	assert.Equal(t, []string{"Darth Vader", "Luke Skywalker"}, a.Characters)
}

const Red = 0
const Blue = 1
const Green = 2

func TestUpdateIntArrays(t *testing.T) {
	type Target struct {
		LaserColors []int `json:"laser_colors"`
	}
	var a = &Target{
		LaserColors: []int{Green},
	}

	newLaserColors := `{"laser_colors":[0,1]}`
	p := make(map[string]interface{})
	jsonErr := json.Unmarshal([]byte(newLaserColors), &p)
	assert.NoError(t, jsonErr)

	_, err := Apply(&a, p)
	assert.NoError(t, err)
	assert.Equal(t, []int{Red, Blue}, a.LaserColors)
}
