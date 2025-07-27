package provablyspin

type Player interface {
	GetID() string
	GetName() string
	GetStake() int64
}

type GamePlayer struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Stake int64  `json:"stake"`
}

func (p GamePlayer) GetID() string {
	return p.ID
}

func (p GamePlayer) GetName() string {
	return p.Name
}

func (p GamePlayer) GetStake() int64 {
	return p.Stake
}

type ControlData[T Player] struct {
	Seed    string `json:"seed"`
	Nonce   int    `json:"nonce"`
	Players []T    `json:"players"`
}

type Result[T Player] struct {
	Winner      T                  `json:"winner"`
	SpinFloat   float64            `json:"spin"`
	Nonce       int                `json:"nonce"`
	Ranges      map[string]float64 `json:"ranges"`
	ControlJSON []byte             `json:"control_json"`
}
