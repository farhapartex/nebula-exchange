package service

import (
	"time"

	"github.com/farhapartex/nebula-exchange/backend/internal/progress/models"
)

const (
	MaximumPauseDuration      = 10 * time.Minute
	reportDelayAllowance      = 2 * time.Minute
	clockSkewAllowance        = 2 * time.Second
	fightTimeAllowanceMS      = 1000
	damageRateAllowanceFactor = 1.25
)

type FightReport struct {
	Outcome     models.FightOutcome
	DurationMS  int
	DamageDealt int
	DamageTaken int
}

type fightReportContext struct {
	TimeLimitMS int
	Player      combatStats
	Enemy       combatStats
	StartedAt   time.Time
	ReportedAt  time.Time
}

func findReportProblem(report FightReport, reportContext fightReportContext) string {
	realElapsed := reportContext.ReportedAt.Sub(reportContext.StartedAt)
	switch {
	case report.DurationMS < 1 || report.DurationMS > reportContext.TimeLimitMS+fightTimeAllowanceMS:
		return "fight time is outside the level time limit"
	case time.Duration(report.DurationMS)*time.Millisecond > realElapsed+clockSkewAllowance:
		return "fight time is longer than the time since the fight started"
	case realElapsed > time.Duration(reportContext.TimeLimitMS)*time.Millisecond+MaximumPauseDuration+reportDelayAllowance:
		return "the fight was paused for longer than allowed"
	case report.DamageDealt < 0 || report.DamageTaken < 0:
		return "damage cannot be negative"
	case float64(report.DamageDealt) > maximumDamage(reportContext.Player, report.DurationMS):
		return "damage dealt is higher than the player can deal in that time"
	case float64(report.DamageTaken) > maximumDamage(reportContext.Enemy, report.DurationMS):
		return "damage taken is higher than the enemy can deal in that time"
	}
	if report.Outcome == models.FightOutcomeWon {
		return findWinProblem(report, reportContext)
	}
	return findLossProblem(report, reportContext)
}

func findWinProblem(report FightReport, reportContext fightReportContext) string {
	switch {
	case float64(report.DamageDealt) < reportContext.Enemy.MaxHealth:
		return "a win needs the enemy knocked out"
	case float64(report.DamageTaken) >= reportContext.Player.MaxHealth:
		return "a knocked out player cannot win"
	case report.DurationMS > reportContext.TimeLimitMS:
		return "a win must happen before time runs out"
	}
	return ""
}

func findLossProblem(report FightReport, reportContext fightReportContext) string {
	isKnockedOut := float64(report.DamageTaken) >= reportContext.Player.MaxHealth
	isTimeUp := report.DurationMS >= reportContext.TimeLimitMS-fightTimeAllowanceMS
	if !isKnockedOut && !isTimeUp {
		return "a loss needs a knockout or the time running out"
	}
	return ""
}

func maximumDamage(attacker combatStats, durationMS int) float64 {
	return float64(durationMS)*attacker.highestDamagePerMillisecond()*damageRateAllowanceFactor + attacker.strongestHit()
}
