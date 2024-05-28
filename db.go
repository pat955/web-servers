package main

import (
	"fmt"
	"os"
)

func createDB() {
	fmt.Println("creating db")

	f, err := os.Create("./database.json")
	if err != nil {
		panic(err)
	}
	f.Write([]byte(`{"chirps": {}}`))
	f.Close()

}

func deleteDB() {
	fmt.Println("deleting db")
	err := os.Remove("./database.json")
	if err != nil {
		fmt.Println(err)
	}
}
