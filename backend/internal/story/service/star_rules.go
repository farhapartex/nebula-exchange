package service

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

type StarRule struct {
	Stars         int `json:"stars"`
	WithinSeconds int `json:"within_seconds"`
}

type StarRules []StarRule

func ParseStarRules(document database.JSONDocument, timeLimitSeconds int) (StarRules, error) {
	var starRules StarRules
	if len(document) == 0 {
		return starRules, nil
	}
	if err := json.Unmarshal(document, &starRules); err != nil {
		return nil, fmt.Errorf("star rules: %w", err)
	}
	for ruleIndex, starRule := range starRules {
		if starRule.Stars < 1 || starRule.Stars > 3 || starRule.WithinSeconds < 1 || starRule.WithinSeconds > timeLimitSeconds {
			return nil, fmt.Errorf("star rule %d must give 1 to 3 stars within 1 to %d seconds", ruleIndex+1, timeLimitSeconds)
		}
		if ruleIndex > 0 {
			previousRule := starRules[ruleIndex-1]
			if starRule.Stars >= previousRule.Stars || starRule.WithinSeconds <= previousRule.WithinSeconds {
				return nil, errors.New("star rules must go from most stars and least time to fewest stars and most time")
			}
		}
	}
	return starRules, nil
}

func (starRules StarRules) StarsForWin(fightSeconds float64) int {
	for _, starRule := range starRules {
		if fightSeconds <= float64(starRule.WithinSeconds) {
			return starRule.Stars
		}
	}
	return 0
}
