package main

import (
	"github.com/initialed85/syncer/internal/args"
	"github.com/initialed85/syncer/internal/utils"
	"github.com/initialed85/syncer/pkg/syncer"
	"log"
)

func main() {
	runArgs := args.ValidateArgs(args.ParseArgs())

	stopFn, err := syncer.Run(
		runArgs.LocalPath,
		runArgs.Rate,
		runArgs.Debounce,
	)

	if err != nil {
		log.Fatal(err)
	}

	defer stopFn()

	utils.WaitForSigInt()
}
