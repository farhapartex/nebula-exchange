package service

import (
	"testing"
	"time"

	"github.com/farhapartex/nebula-exchange/backend/internal/progress/models"
)

func reportContextAfter(realSeconds float64) fightReportContext {
	startedAt := time.Date(2026, 10, 6, 10, 0, 0, 0, time.UTC)
	return fightReportContext{
		TimeLimitMS: 90_000,
		Player:      combatStats{MaxHealth: 100, Punch: attackStats{Damage: 8, WindupMS: 120, ActiveMS: 80, RecoveryMS: 200}, Kick: attackStats{Damage: 14, WindupMS: 260, ActiveMS: 100, RecoveryMS: 340}},
		Enemy:       combatStats{MaxHealth: 90, Punch: attackStats{Damage: 9, WindupMS: 200, ActiveMS: 80, RecoveryMS: 260}, Kick: attackStats{Damage: 15, WindupMS: 320, ActiveMS: 100, RecoveryMS: 380}},
		StartedAt:   startedAt,
		ReportedAt:  startedAt.Add(time.Duration(realSeconds * float64(time.Second))),
	}
}

func TestHonestReportsAreAccepted(t *testing.T) {
	honestReports := map[string]FightReport{
		"knockout win": {Outcome: models.FightOutcomeWon, DurationMS: 40_000, DamageDealt: 94, DamageTaken: 35},
		"knocked out":  {Outcome: models.FightOutcomeLost, DurationMS: 30_000, DamageDealt: 20, DamageTaken: 100},
		"time ran out": {Outcome: models.FightOutcomeLost, DurationMS: 90_000, DamageDealt: 80, DamageTaken: 40},
	}
	for caseName, report := range honestReports {
		if problem := findReportProblem(report, reportContextAfter(float64(report.DurationMS)/1000+5)); problem != "" {
			t.Fatalf("%s: rejected with %q", caseName, problem)
		}
	}
}

func TestImpossibleReportsAreRejected(t *testing.T) {
	impossibleReports := map[string]struct {
		report      FightReport
		realSeconds float64
	}{
		"win without knocking the enemy out":  {FightReport{Outcome: models.FightOutcomeWon, DurationMS: 40_000, DamageDealt: 50, DamageTaken: 10}, 45},
		"win while knocked out":               {FightReport{Outcome: models.FightOutcomeWon, DurationMS: 40_000, DamageDealt: 95, DamageTaken: 100}, 45},
		"knockout faster than possible":       {FightReport{Outcome: models.FightOutcomeWon, DurationMS: 2_000, DamageDealt: 90, DamageTaken: 0}, 5},
		"fight longer than the time limit":    {FightReport{Outcome: models.FightOutcomeLost, DurationMS: 95_000, DamageDealt: 10, DamageTaken: 40}, 100},
		"fight longer than real time":         {FightReport{Outcome: models.FightOutcomeWon, DurationMS: 60_000, DamageDealt: 92, DamageTaken: 10}, 20},
		"paused past the limit":               {FightReport{Outcome: models.FightOutcomeWon, DurationMS: 60_000, DamageDealt: 92, DamageTaken: 10}, 60 + 30 + 10*60 + 3*60},
		"loss with no knockout and time left": {FightReport{Outcome: models.FightOutcomeLost, DurationMS: 30_000, DamageDealt: 10, DamageTaken: 40}, 35},
		"negative damage":                     {FightReport{Outcome: models.FightOutcomeLost, DurationMS: 90_000, DamageDealt: -5, DamageTaken: 40}, 95},
	}
	for caseName, impossible := range impossibleReports {
		if problem := findReportProblem(impossible.report, reportContextAfter(impossible.realSeconds)); problem == "" {
			t.Fatalf("%s: expected a rejection", caseName)
		}
	}
}
