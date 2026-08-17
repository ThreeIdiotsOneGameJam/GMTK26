package game

//go:generate stringer -type=ActionType -trimprefix=Action

type ActionType uint8

const (
	ActionUnknown ActionType = iota
	ActionPass
	ActionBuild
	ActionMove
	ActionRecruit
	ActionAttack
)

type BuildActionPayload struct {
	From     Hex          `json:"from"`
	To       Hex          `json:"to"`
	Building BuildingType `json:"building"`
}

type MoveActionPayload struct {
	From Hex `json:"from"`
	To   Hex `json:"to"`
}

type RecruitActionPayload struct {
	From Hex      `json:"from"`
	To   Hex      `json:"to"`
	Unit UnitType `json:"unit"`
}

type AttackActionPayload struct {
	From Hex `json:"from"`
	To   Hex `json:"to"`
}

type ActionResultStatus uint8

const (
	ActionResultUnknown ActionResultStatus = iota
	ActionResultSucceeded
	ActionResultInvalid
	ActionResultInsufficientCoins
	ActionResultContested
	ActionResultBlocked
)

// ActionResult reports the latest effect resolved for one faction. Message is
// intentionally client-ready because it is also useful for server-only
// validation failures that do not map cleanly to a single rule.
type ActionResult struct {
	Round     int32              `json:"round"`
	Type      ActionType         `json:"type"`
	Status    ActionResultStatus `json:"status"`
	Automatic bool               `json:"automatic"`
	From      Hex                `json:"from"`
	To        Hex                `json:"to"`
	Message   string             `json:"message"`
}
