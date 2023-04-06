package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParserGoodInputs(t *testing.T) {
	const (
		YES      = 0
		NO       = 1
		DND_ONLY = 2
	)
	testCases := []struct {
		query       string
		rolls       []int
		expected    int
		success     int
		render      string
		renderBasic string
	}{
		{query: "1", expected: 1},
		{query: "5", expected: 5},
		{query: "5+3", expected: 8},
		{query: "(5+3)", expected: 8},
		{query: "(5-3)", expected: 2},
		{query: "(10-3)/2", expected: 3},
		{query: "(10-3)*2", expected: 14},
		{query: "(10-3)*2", expected: 14},
		{query: "10-3*2", expected: 4},
		{query: "10-(3*2)", expected: 4},
		{query: "d20", rolls: []int{12}, expected: 12},
		{query: "3d20", rolls: []int{12, 10, 3}, expected: 25},
		{query: "d20-18d4k5",
			rolls:    []int{11, 3, 1, 1, 1, 1, 2, 3, 4, 2, 4, 4, 2, 2, 4, 1, 4, 3, 4},
			expected: -9,
			render:   "d20-18d4k5 = **-9**\n- *d20 =* ***11***\n- *18d4k5 (~~3~~ ~~1~~ ~~1~~ ~~1~~ ~~1~~ ~~2~~ ~~3~~ ~~4~~ ~~2~~ 4 4 ~~2~~ ~~2~~ 4 ~~1~~ 4 ~~3~~ 4) =* ***20***",
		},
		{query: "3d20k1", rolls: []int{12, 10, 3}, expected: 12},
		{query: "3d20kh1", rolls: []int{12, 10, 3}, expected: 12},
		{query: "3d20kl1", rolls: []int{12, 10, 3}, expected: 3},
		{query: "3d20d2", rolls: []int{12, 10, 3}, expected: 12},
		{query: "3d20dh2", rolls: []int{12, 10, 3}, expected: 3},
		{query: "3d20dl2", rolls: []int{12, 10, 3}, expected: 12},
		{query: "d20a", rolls: []int{12, 10}, expected: 12, success: DND_ONLY},
		{query: "d20d", rolls: []int{12, 10}, expected: 10, success: DND_ONLY},
		{query: "1d20 for insight",
			rolls:    []int{17},
			expected: 17,
			render:   "1d20 = **17** for insight"},
		{query: "d20+1",
			rolls:    []int{15},
			expected: 16,
			render:   "d20+1 = **16**\n- *d20 =* ***15***"},
		{query: "d20a+3",
			rolls:    []int{16, 5},
			expected: 19,
			success:  DND_ONLY,
			render:   "d20a+3 = **19**\n- *d20a (16 ~~5~~) =* ***16***"},
		{query: "1d12+5",
			rolls:    []int{12},
			expected: 17,
			render:   "1d12+5 = **17**\n- *1d12 =* ***12***"},
		{query: "1d12+5",
			rolls:    []int{1},
			expected: 6,
			render:   "1d12+5 = **6**\n- *1d12 =* ***1***"},
		{query: "2d6+4+10+3d8+1d4+2",
			rolls:    []int{3, 4, 1, 7, 8, 3},
			expected: 42,
			render:   "2d6+4+10+3d8+1d4+2 = **42**\n- *2d6 (3 4) =* ***7***\n- *3d8 (1 7 8) =* ***16***\n- *1d4 =* ***3***"},
		{query: "d%+3+2d%+1d4*5d%k2-d%a",
			success:  DND_ONLY,
			rolls:    []int{56, 40, 30, 2, 21, 38, 16, 55, 3, 21, 31},
			expected: 284,
			render:   "d%+3+2d%+1d4×5d%k2-d%a = **284**\n- *d% =* ***56***\n- *2d% (40 30) =* ***70***\n- *1d4 =* ***2***\n- *5d%k2 (~~21~~ 38 ~~16~~ 55 ~~3~~) =* ***93***\n- *d%a (~~21~~ 31) =* ***31***"},
		{query: "sTaTs",
			success: DND_ONLY,
			rolls:   []int{2, 5, 2, 3, 5, 4, 6, 2, 2, 1, 2, 4, 3, 4, 1, 6, 1, 5, 6, 6, 3, 4, 2, 5},
			render:  "up a new character! Adventure awaits. In the meanwhile, here are your ability scores:\n**17**, **15**, **13**, **12**, **10**, **8**\n- *4d6d1 (~~2~~ 5 2 3) =* ***10***\n- *4d6d1 (5 4 6 ~~2~~) =* ***15***\n- *4d6d1 (2 ~~1~~ 2 4) =* ***8***\n- *4d6d1 (3 4 ~~1~~ 6) =* ***13***\n- *4d6d1 (~~1~~ 5 6 6) =* ***17***\n- *4d6d1 (3 4 ~~2~~ 5) =* ***12***"},
		{query: "death-save",
			success:  DND_ONLY,
			rolls:    []int{1},
			expected: 1,
			render:   "a death saving throw, and suffers **A CRITICAL FAIL!** :coffin:\n- *1d20 =* ***1***"},
		{query: "death save",
			success:  DND_ONLY,
			rolls:    []int{9},
			expected: 9,
			render:   "a death saving throw, and **FAILS** :skull:\n- *1d20 =* ***9***"},
		{query: "deathsave",
			success:  DND_ONLY,
			rolls:    []int{10},
			expected: 10,
			render:   "a death saving throw, and **SUCCEEDS** :thumbsup:\n- *1d20 =* ***10***"},
		{query: "DEATH-save",
			success:  DND_ONLY,
			rolls:    []int{20},
			expected: 20,
			render:   "a death saving throw, and **REGAINS 1 HP!** :star-struck:\n- *1d20 =* ***20***"},
		{query: "2d6+3 piercing, 2d8 radiant, 1d6 fire",
			rolls:  []int{2, 3, 5, 7, 1},
			render: "2d6+3, 2d8, 1d6 = **8** piercing, **12** radiant, **1** fire\n- *2d6 (2 3) =* ***5***\n- *2d8 (5 7) =* ***12***"},
		{query: "d20+8 to hit, 1d8+5+5d6 piercing damage",
			rolls:  []int{14, 4, 3, 2, 5, 1, 6},
			render: "d20+8, 1d8+5+5d6 = **22** to hit, **26** piercing damage\n- *d20 =* ***14***\n- *1d8 =* ***4***\n- *5d6 (3 2 5 1 6) =* ***17***"},
		{query: "(1d12+8 bludgeoning)+(1d8+5d6+1d4 piercing) damage",
			rolls:    []int{3, 5, 3, 5, 3, 6, 2, 4},
			expected: 39,
			render:   "(1d12+8)+(1d8+5d6+1d4) = **39** damage\n- *1d12+8 =* ***11*** *bludgeoning*\n  - *1d12 =* ***3***\n- *1d8+5d6+1d4 =* ***28*** *piercing*\n  - *1d8 =* ***5***\n  - *5d6 (3 5 3 6 2) =* ***19***\n  - *1d4 =* ***4***"},
		{query: "1d20+8*(1d8+5d6+1d4)",
			rolls:    []int{3, 5, 3, 5, 3, 6, 2, 4},
			expected: 227,
			render:   "1d20+8×(1d8+5d6+1d4) = **227**\n- *1d20 =* ***3***\n- *1d8 =* ***5***\n- *5d6 (3 5 3 6 2) =* ***19***\n- *1d4 =* ***4***"},
		{query: "1d20+4 to hit, (1d6+2 slashing)+(2d8 radiant) damage",
			rolls:  []int{16, 3, 6, 2},
			render: "1d20+4, (1d6+2)+(2d8) = **20** to hit, **13** damage\n- *1d20 =* ***16***\n- *1d6+2 =* ***5*** *slashing*\n  - *1d6 =* ***3***\n- *2d8 (6 2) =* ***8*** *radiant*"},
		{query: "hello",
			success: NO},
		{query: "-2",
			success: NO},
		{query: "5+",
			success: NO},
		{query: "/7",
			success: NO},
		{query: "(10-3",
			success: NO},
		{query: "1d20+5",
			rolls:       []int{1},
			expected:    6,
			render:      "1d20+5 = **6** (NAT1! :grimacing:)\n- *1d20 =* ***1***",
			renderBasic: "1d20+5 = **6**\n- *1d20 =* ***1***"},
		{query: "1d20+5",
			rolls:       []int{20},
			expected:    25,
			render:      "1d20+5 = **25** (NAT20! :star-struck:)\n- *1d20 =* ***20***",
			renderBasic: "1d20+5 = **25**\n- *1d20 =* ***20***"},
		{query: "1d20 for insight",
			rolls:       []int{20},
			expected:    20,
			render:      "1d20 = **20** for insight (NAT20! :star-struck:)",
			renderBasic: "1d20 = **20** for insight"},
		{query: "1d20+5 for insight",
			rolls:       []int{20},
			expected:    25,
			render:      "1d20+5 = **25** for insight (NAT20! :star-struck:)\n- *1d20 =* ***20***",
			renderBasic: "1d20+5 = **25** for insight\n- *1d20 =* ***20***"},
		{query: "1d20+4 to hit, (1d6+2 slashing)+(2d8 radiant) damage",
			rolls:       []int{20, 3, 6, 2},
			render:      "1d20+4, (1d6+2)+(2d8) = **24** to hit (NAT20! :star-struck:), **13** damage\n- *1d20 =* ***20***\n- *1d6+2 =* ***5*** *slashing*\n  - *1d6 =* ***3***\n- *2d8 (6 2) =* ***8*** *radiant*",
			renderBasic: "1d20+4, (1d6+2)+(2d8) = **24** to hit, **13** damage\n- *1d20 =* ***20***\n- *1d6+2 =* ***5*** *slashing*\n  - *1d6 =* ***3***\n- *2d8 (6 2) =* ***8*** *radiant*"},
		{query: "1d20+4 to hit, (1d6+2 slashing)+(2d8 radiant) damage, (d20*10-(2d%kl1 discount percentage)*2)/3 feywild encounter",
			rolls:       []int{1, 3, 6, 2, 20, 14, 76},
			render:      "1d20+4, (1d6+2)+(2d8), (d20×10-(2d%kl1)×2)/3 = **5** to hit (NAT1! :grimacing:), **13** damage, **57** feywild encounter\n- *1d20 =* ***1***\n- *1d6+2 =* ***5*** *slashing*\n  - *1d6 =* ***3***\n- *2d8 (6 2) =* ***8*** *radiant*\n- *d20 =* ***20*** (NAT20! :star-struck:)\n- *2d%kl1 (14 ~~76~~) =* ***14*** *discount percentage*",
			renderBasic: "1d20+4, (1d6+2)+(2d8), (d20×10-(2d%kl1)×2)/3 = **5** to hit, **13** damage, **57** feywild encounter\n- *1d20 =* ***1***\n- *1d6+2 =* ***5*** *slashing*\n  - *1d6 =* ***3***\n- *2d8 (6 2) =* ***8*** *radiant*\n- *d20 =* ***20***\n- *2d%kl1 (14 ~~76~~) =* ***14*** *discount percentage*"},
		{query: "3d20",
			rolls:    []int{20, 10, 1},
			expected: 31,
			render:   "3d20 = **31**\n- *3d20 (20 10 1) =* ***31***"},
		{query: "3d20k2",
			rolls:    []int{20, 10, 1},
			expected: 30,
			render:   "3d20k2 = **30**\n- *3d20k2 (20 10 ~~1~~) =* ***30***"},
		{query: "3d20k1",
			rolls:       []int{20, 10, 1},
			expected:    20,
			render:      "3d20k1 = **20** (NAT20! :star-struck:)\n- *3d20k1 (20 ~~10~~ ~~1~~) =* ***20***",
			renderBasic: "3d20k1 = **20**\n- *3d20k1 (20 ~~10~~ ~~1~~) =* ***20***"},
	}
	for _, enableDnd := range []bool{false, true} {
		conf := configuration{EnableDnd5e: enableDnd}
		parse := GetParser(conf)
		for _, testCase := range testCases {
			parsedNode, err := parse(testCase.query)
			message := "Testing case " + testCase.query
			if testCase.success == YES || (testCase.success == DND_ONLY && enableDnd) {
				assert.Nil(t, err, message)
				assert.NotNil(t, parsedNode, message)
				rollerError := ""
				rollerIdx := 0
				roller := func(x int) int {
					ret := 0
					if testCase.rolls == nil {
						rollerError = "Needs mocked rolls"
						return 1001
					}
					rolls := testCase.rolls
					if len(rolls) <= rollerIdx {
						rollerError = "Needs more mocked rolls"
						return 1002
					}
					ret = rolls[rollerIdx]
					rollerIdx++
					if ret < 1 || x < ret {
						rollerError = "Roll out of range"
					}
					return ret
				}
				rolledNode := parsedNode.roll(roller, conf)
				assert.Equal(t, "", rollerError)
				if 0 < rollerIdx && testCase.rolls != nil {
					assert.Equal(t, rollerIdx, len(testCase.rolls))
				}
				assert.Equal(t, testCase.expected, rolledNode.value(), message)
				if testCase.render != "" {
					resultText := rolledNode.renderToplevel()
					if testCase.renderBasic != "" && !enableDnd {
						assert.Equal(t, testCase.renderBasic, resultText, message)
					} else {
						assert.Equal(t, testCase.render, resultText, message)
					}
				}
			} else {
				assert.NotNil(t, err, message)
			}
		}
	}
}
