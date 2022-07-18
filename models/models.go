package models

type Table struct {
	ID           int `json:"id" db:"id"`
	Pos1         int `json:"pos_1" db:"pos_1"`
	Pos2         int `json:"pos_2" db:"pos_2"`
	Pos3         int `json:"pos_3" db:"pos_3"`
	Pos4         int `json:"pos_4" db:"pos_4"`
	Pos5         int `json:"pos_5" db:"pos_5"`
	Pos6         int `json:"pos_6" db:"pos_6"`
	Pos7         int `json:"pos_7" db:"pos_7"`
	Pos8         int `json:"pos_8" db:"pos_8"`
	Pos9         int `json:"pos_9" db:"pos_9"`
	Pos10        int `json:"pos_10" db:"pos_10"`
	Mafia1       int `json:"mafia_1" db:"mafia_1"`
	Mafia2       int `json:"mafia_2" db:"mafia_2"`
	MafiaBoss    int `json:"mafia_boss" db:"mafia_boss"`
	Sherif       int `json:"sherif" db:"sherif"`
	MafiaStored  bool
	SherifStored bool
	Open         bool
	GameStarted  bool
}
