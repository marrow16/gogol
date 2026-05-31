package logic

import (
	"errors"
	"strconv"
	"strings"
)

type Rule interface {
	NextState(c *Cell) (nextState bool)
	StateChanged(c *Cell) (changed bool)
	Rle() string
	Permutation() int
	Name() string
}

type rule struct {
	name         string
	bornWith     [9]bool
	survivesWith [9]bool
}

func (r rule) NextState(c *Cell) (nextState bool) {
	adjsAlive := c.AdjacentsAlive()
	if c.Alive {
		nextState = r.survivesWith[adjsAlive]
	} else {
		nextState = r.bornWith[adjsAlive]
	}
	return nextState
}

func (r rule) StateChanged(c *Cell) (changed bool) {
	nextState := r.NextState(c)
	return nextState != c.Alive
}

func (r rule) Rle() string {
	var sb strings.Builder
	sb.Grow(18 + 3)
	sb.WriteString("B")
	for i := 0; i < 9; i++ {
		if r.bornWith[i] {
			sb.WriteString(strconv.Itoa(i))
		}
	}
	sb.WriteString("/S")
	for i := 0; i < 9; i++ {
		if r.survivesWith[i] {
			sb.WriteString(strconv.Itoa(i))
		}
	}
	return sb.String()
}

func (r rule) Permutation() int {
	result := 0
	for i := 0; i < 9; i++ {
		if r.bornWith[i] {
			result |= 1 << (i + 9)
		}
		if r.survivesWith[i] {
			result |= 1 << i
		}
	}
	return result
}

func (r rule) Name() string {
	if r.name != "" {
		return r.name
	}
	if rn, ok := rleToName[r.Rle()]; ok {
		return rn
	}
	return "Custom " + r.Rle()
}

func NewRuleFromPermutation(permutation int) (Rule, error) {
	if permutation < 0 || permutation >= 1<<18 {
		return nil, ErrInvalidPermutation
	}
	r := &rule{}
	for i := 0; i < 9; i++ {
		r.bornWith[i] = permutation&(1<<(i+9)) != 0
		r.survivesWith[i] = permutation&(1<<i) != 0
	}
	return r, nil
}

func mustNewRuleRle(name string, rle string) Rule {
	if r, err := NewRuleRle(name, rle); err != nil {
		panic(err)
	} else {
		return r
	}
}

func NewRuleRle(name string, rle string) (Rule, error) {
	result := &rule{
		bornWith:     [9]bool{},
		survivesWith: [9]bool{},
		name:         name,
	}
	parts := strings.Split(rle, "/")
	b := ""
	s := ""
	if len(parts) > 0 {
		if strings.HasPrefix(parts[0], "B") {
			b = parts[0][1:]
		} else if strings.HasPrefix(parts[0], "S") {
			s = parts[0][1:]
		}
	}
	if len(parts) > 1 {
		if strings.HasPrefix(parts[1], "B") {
			b = parts[1][1:]
		} else if strings.HasPrefix(parts[1], "S") {
			s = parts[1][1:]
		}
	}
	for _, ch := range b {
		idx := ch - '0'
		if idx >= 0 && idx <= 8 {
			result.bornWith[idx] = true
		} else {
			return nil, ErrInvalidRule
		}
	}
	for _, ch := range s {
		idx := ch - '0'
		if idx >= 0 && idx <= 8 {
			result.survivesWith[idx] = true
		} else {
			return nil, ErrInvalidRule
		}
	}
	return result, nil
}

var (
	ErrInvalidRule        = errors.New("invalid RLE rule")
	ErrInvalidPermutation = errors.New("invalid permutation")
)

var StandardRule = mustNewRuleRle("Standard", "B3/S23")

var Rules = map[string]Rule{
	"2X2":                     mustNewRuleRle("2X2", "B36/S125"),
	"34 Life":                 mustNewRuleRle("34 Life", "B34/S34"),
	"Amoeba":                  mustNewRuleRle("Amoeba", "B357/S1358"),
	"AntiLife":                mustNewRuleRle("AntiLife", "B0123478/S01234678"),
	"Assimilation":            mustNewRuleRle("Assimilation", "B345/S4567"),
	"Bacteria":                mustNewRuleRle("Bacteria", "B34/S456"),
	"Blinker Life":            mustNewRuleRle("Blinker Life", "B36/S235"),
	"Blinkers":                mustNewRuleRle("Blinkers", "B345/S2"),
	"Bugs":                    mustNewRuleRle("Bugs", "B3567/S15678"),
	"Coagulations":            mustNewRuleRle("Coagulations", "B378/S235678"),
	"Coral":                   mustNewRuleRle("Coral", "B3/S45678"),
	"Corrosion of Conformity": mustNewRuleRle("Corrosion of Conformity", "B3/S124"),
	"Day & Night":             mustNewRuleRle("Day & Night", "B3678/S34678"),
	"Diamoeba":                mustNewRuleRle("Diamoeba", "B35678/S5678"),
	"DotLife":                 mustNewRuleRle("DotLife", "B3/S023"),
	"DryLife":                 mustNewRuleRle("DryLife", "B37/S23"),
	"EightLife":               mustNewRuleRle("EightLife", "B3/S238"),
	"Electrified Maze":        mustNewRuleRle("Electrified Maze", "B45/S12345"),
	"Flock":                   mustNewRuleRle("Flock", "B3/S12"),
	"Fredkin":                 mustNewRuleRle("Fredkin", "B1357/S02468"),
	"Gems Minor":              mustNewRuleRle("Gems Minor", "B34578/S456"),
	"Gems":                    mustNewRuleRle("Gems", "B3457/S4568"),
	"Gnarl":                   mustNewRuleRle("Gnarl", "B1/S1"),
	"H-trees":                 mustNewRuleRle("H-trees", "B1/S012345678"),
	"HiLife":                  mustNewRuleRle("HiLife", "B36/S23"),
	"Holstein":                mustNewRuleRle("Holstein", "B35678/S4678"),
	"HoneyLife":               mustNewRuleRle("HoneyLife", "B38/S238"),
	"Iceballs":                mustNewRuleRle("Iceballs", "B25678/S5678"),
	"InverseLife":             mustNewRuleRle("InverseLife", "B012345678/S34678"),
	"Land Rush":               mustNewRuleRle("Land Rush", "B35/S234578"),
	"Life without death":      mustNewRuleRle("Life without death", "B3/S012345678"),
	"Live Free or Die":        mustNewRuleRle("Live Free or Die", "B2/S0"),
	"Long Life":               mustNewRuleRle("Long Life", "B345/S5"),
	"LowDeath":                mustNewRuleRle("LowDeath", "B368/S238"),
	"LowLife":                 mustNewRuleRle("LowLife", "B3/S13"),
	"Maze with Mice":          mustNewRuleRle("Maze with Mice", "B37/S12345"),
	"Maze":                    mustNewRuleRle("Maze", "B3/S12345"),
	"Mazectric with Mice":     mustNewRuleRle("Mazectric with Mice", "B37/S1234"),
	"Mazectric":               mustNewRuleRle("Mazectric", "B3/S1234"),
	"Move":                    mustNewRuleRle("Move", "B368/S245"),
	"Pedestrian Life":         mustNewRuleRle("Pedestrian Life", "B38/S23"),
	"Plow World":              mustNewRuleRle("Plow World", "B378/S012345678"),
	"Pseudo Life":             mustNewRuleRle("Pseudo Life", "B357/S238"),
	"Replicator":              mustNewRuleRle("Replicator", "B1357/S1357"),
	"Seeds":                   mustNewRuleRle("Seeds", "B2/S"),
	"Serviettes":              mustNewRuleRle("Serviettes", "B234/S"),
	"Slow Blob":               mustNewRuleRle("Slow Blob", "B367/S125678"),
	"SnowLife":                mustNewRuleRle("SnowLife", "B3/S1237"),
	"Stains":                  mustNewRuleRle("Stains", "B3678/S235678"),
	"Standard":                StandardRule,
	"Vote 4/5":                mustNewRuleRle("Vote 4/5", "B4678/S35678"),
	"Vote":                    mustNewRuleRle("Vote", "B5678/S45678"),
	"Walled cities":           mustNewRuleRle("Walled cities", "B45678/S2345"),
	"Water Surface":           mustNewRuleRle("Water Surface", "B34/S23"),
}

var rleToName = map[string]string{
	"B36/S125":           "2X2",
	"B34/S34":            "34 Life",
	"B357/S1358":         "Amoeba",
	"B0123478/S01234678": "AntiLife",
	"B345/S4567":         "Assimilation",
	"B34/S456":           "Bacteria",
	"B36/S235":           "Blinker Life",
	"B345/S2":            "Blinkers",
	"B3567/S15678":       "Bugs",
	"B378/S235678":       "Coagulations",
	"B3/S45678":          "Coral",
	"B3/S124":            "Corrosion of Conformity",
	"B3678/S34678":       "Day & Night",
	"B35678/S5678":       "Diamoeba",
	"B3/S023":            "DotLife",
	"B37/S23":            "DryLife",
	"B3/S238":            "EightLife",
	"B45/S12345":         "Electrified Maze",
	"B3/S12":             "Flock",
	"B1357/S02468":       "Fredkin",
	"B34578/S456":        "Gems Minor",
	"B3457/S4568":        "Gems",
	"B1/S1":              "Gnarl",
	"B1/S012345678":      "H-trees",
	"B36/S23":            "HiLife",
	"B35678/S4678":       "Holstein",
	"B38/S238":           "HoneyLife",
	"B25678/S5678":       "Iceballs",
	"B012345678/S34678":  "InverseLife",
	"B35/S234578":        "Land Rush",
	"B3/S012345678":      "Life without death",
	"B2/S0":              "Live Free or Die",
	"B345/S5":            "Long Life",
	"B368/S238":          "LowDeath",
	"B3/S13":             "LowLife",
	"B37/S12345":         "Maze with Mice",
	"B3/S12345":          "Maze",
	"B37/S1234":          "Mazectric with Mice",
	"B3/S1234":           "Mazectric",
	"B368/S245":          "Move",
	"B38/S23":            "Pedestrian Life",
	"B378/S012345678":    "Plow World",
	"B357/S238":          "Pseudo Life",
	"B1357/S1357":        "Replicator",
	"B2/S":               "Seeds",
	"B234/S":             "Serviettes",
	"B367/S125678":       "Slow Blob",
	"B3/S1237":           "SnowLife",
	"B3678/S235678":      "Stains",
	"B3/S23":             "Standard",
	"B4678/S35678":       "Vote 4/5",
	"B5678/S45678":       "Vote",
	"B45678/S2345":       "Walled cities",
	"B34/S23":            "Water Surface",
}
