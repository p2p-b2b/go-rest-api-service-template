package main

import (
	"flag"
	"fmt"

	"github.com/google/uuid"
)

func main() {
	num := flag.Int("n", 10, "number of UUIDs to generate")
	ver := flag.Int("v", 7, "version of UUID to generate. Supported versions: 4, 6, 7")
	flag.Parse()

	// check the flags values
	if *num < 1 || *num > 1000 {
		flag.Usage()
		fmt.Printf("Number of UUIDs to generate must be between 1 and 1000, got: %d\n", *num)
		return
	}

	if *ver != 4 && *ver != 5 && *ver != 6 && *ver != 7 {
		flag.Usage()
		fmt.Printf("Unsupported UUID version: %d. Supported versions are 4, 5, 6, and 7.\n", *ver)
		return
	}

	for i := 0; i < *num; i++ {
		var u uuid.UUID
		var err error

		switch *ver {
		case 7:
			u, err = uuid.NewV7()
			if err != nil {
				panic(err)
			}
		case 6:
			u, err = uuid.NewV6()
			if err != nil {
				panic(err)
			}
		case 4:
			u, err = uuid.NewRandom()
			if err != nil {
				panic(err)
			}
		default:
			fmt.Printf("Unsupported UUID version: %d\n", *ver)
			return
		}

		// Print the generated UUID
		_, err = fmt.Println(u.String())
		if err != nil {
			panic(err)
		}
	}
}
