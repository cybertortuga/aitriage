package main

import (
	"fmt"
	"github.com/cybertortuga/aitriage/internal/engine"
)

func main() {
	eng, err := engine.NewEngine(nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Rules length:", len(eng.Rules))
	if len(eng.Rules) > 0 {
		fmt.Println("First rule:", eng.Rules[0].ID)
	}
}
