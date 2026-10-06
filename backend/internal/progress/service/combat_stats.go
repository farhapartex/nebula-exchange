package service

import (
	"encoding/json"
	"errors"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

type attackStats struct {
	Damage     float64 `json:"damage"`
	WindupMS   float64 `json:"windup_ms"`
	ActiveMS   float64 `json:"active_ms"`
	RecoveryMS float64 `json:"recovery_ms"`
}

type combatStats struct {
	MaxHealth float64     `json:"max_health"`
	Punch     attackStats `json:"punch"`
	Kick      attackStats `json:"kick"`
}

func parseCombatStats(document database.JSONDocument) (combatStats, error) {
	var stats combatStats
	if err := json.Unmarshal(document, &stats); err != nil {
		return combatStats{}, err
	}
	if stats.MaxHealth <= 0 || stats.strongestHit() <= 0 {
		return combatStats{}, errors.New("fighter stats need max_health and attacks with damage")
	}
	return stats, nil
}

func (stats combatStats) strongestHit() float64 {
	return max(stats.Punch.Damage, stats.Kick.Damage)
}

func (stats combatStats) highestDamagePerMillisecond() float64 {
	highestRate := 0.0
	for _, attack := range []attackStats{stats.Punch, stats.Kick} {
		cycleMS := attack.WindupMS + attack.ActiveMS + attack.RecoveryMS
		if cycleMS > 0 && attack.Damage/cycleMS > highestRate {
			highestRate = attack.Damage / cycleMS
		}
	}
	return highestRate
}
