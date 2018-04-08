package db

import (
	"SincroNice/types"

	scribble "github.com/nanobox-io/golang-scribble"
)

var dir = "./"

func chk(e error) {
	if e != nil {
		panic(e)
	}
}

// GetFolder :
func GetFolder(id string) (f types.Folder) {
	db, err := scribble.New(dir, nil)
	chk(err)
	err = db.Read("folder", id, &f)
	chk(err)

	return
}

// func main() {

// 	dir := "./"

// 	db, err := scribble.New(dir, nil)
// 	if err != nil {
// 		fmt.Println("Error", err)
// 	}

// 	// Write a fish to the database
// 	for _, name := range []string{"onefish", "twofish", "redfish", "bluefish"} {
// 		db.Write("fish", name, Fish{Name: name})
// 	}

// 	// Read a fish from the database (passing fish by reference)
// 	onefish := Fish{}
// 	if err := db.Read("fish", "onefish", &onefish); err != nil {
// 		fmt.Println("Error", err)
// 	}

// 	// Read all fish from the database, unmarshaling the response.
// 	records, err := db.ReadAll("fish")
// 	if err != nil {
// 		fmt.Println("Error", err)
// 	}

// 	fishies := []Fish{}
// 	for _, f := range records {
// 		fishFound := Fish{}
// 		if err := json.Unmarshal([]byte(f), &fishFound); err != nil {
// 			fmt.Println("Error", err)
// 		}
// 		fishies = append(fishies, fishFound)
// 	}

// // Delete a fish from the database
// if err := db.Delete("fish", "onefish"); err != nil {
// 	fmt.Println("Error", err)
// }

// // Delete all fish from the database
// if err := db.Delete("fish", ""); err != nil {
// 	fmt.Println("Error", err)
// }

// }
