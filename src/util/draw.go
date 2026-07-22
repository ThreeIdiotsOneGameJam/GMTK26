package util

import (
	"encoding/json"
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func DrawTextSimple(text string, x int32, y int32) {
	rl.DrawText(text, x, y, 10, rl.Black)
}

func DrawJSONSimple(obj any, x int32, y int32) {
	jsonb1, err := json.Marshal(obj)
	if err != nil {
		fmt.Println(err)
		DrawTextSimple("JSON Parse Error", 10, 10)
	}
	DrawTextSimple(string(jsonb1), 10, 10)
}
