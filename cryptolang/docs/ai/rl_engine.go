// rl_engine.go – Local RL Driver for Cryptolang
package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"
)

type Config struct {
	Parameters struct {
		Iterations int `json:"iterations"`
	} `json:"parameters"`
}

type Metrics struct {
	CompileTime float64
	CPU         float64
	Mem         float64
	ErrorRate   float64
	Reward      float64
}

func reward(m Metrics) float64 {
	return 0.4*(-m.CompileTime/100.0) + 0.3*(1.0-m.ErrorRate) + 0.3*(-m.CPU)
}

func simulateMetrics(prev Metrics) Metrics {
	noise := func(v float64, pct float64) float64 {
		delta := (rand.Float64()*2 - 1) * pct / 100
		return v * (1 + delta)
	}
	m := Metrics{}
	m.CompileTime = noise(prev.CompileTime, 20)
	m.CPU = noise(prev.CPU, 15)
	m.Mem = noise(prev.Mem, 10)
	m.ErrorRate = noise(prev.ErrorRate, 5)
	m.Reward = reward(m)
	return m
}

func main() {
	rand.Seed(time.Now().UnixNano())

	file, _ := os.ReadFile("rl_config.json")
	var cfg Config
	json.Unmarshal(file, &cfg)

	base := Metrics{CompileTime: 108, CPU: 57.3, Mem: 92, ErrorRate: 0.003}
	base.Reward = reward(base)
	fmt.Printf("Iter 0 → Reward %.3f\n", base.Reward)

	current := base
	for i := 1; i <= cfg.Parameters.Iterations; i++ {
		next := simulateMetrics(current)
		delta := next.Reward - current.Reward
		fmt.Printf("Iter %d → Reward %.3f (Δ%.3f)\n", i, next.Reward, delta)
		current = next
		time.Sleep(150 * time.Millisecond)
	}

	fmt.Println("Simulation complete. Review logs for convergence behavior.")
}
