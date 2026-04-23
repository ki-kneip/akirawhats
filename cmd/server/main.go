package main

import (
	"kikneip.com/akirawhats/src/db"
)

func main() {
	db.Open()
	defer db.Close()

}
