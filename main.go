package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	fmt.Println("Avvio del server tramite cmd/server/main.go...")
	cmd := exec.Command("go", "run", "cmd/server/main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Errore durante l'esecuzione del server: %v\n", err)
		os.Exit(1)
	}
}
