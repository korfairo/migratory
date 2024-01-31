package main

import (
	"github.com/korfairo/migratory/internal/command"
	_ "github.com/lib/pq"
)

func main() {
	command.Execute()
}
