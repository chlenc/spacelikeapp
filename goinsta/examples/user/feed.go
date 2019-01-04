// +build ignore

package main

import (
	"fmt"
	"os"

	e "github.com/ahmdrz/goinsta/examples"
)

func main() {
	inst, err := e.InitGoinsta("<target user>")
	e.CheckErr(err)

	user, err := inst.Profiles.ByName(os.Args[0])
	e.CheckErr(err)

	media := user.Feed()

	for media.Next() {
		fmt.Printf("Printing %d items\n", len(media.Items))
		for _, item := range media.Items {
			if len(item.Images.Versions) != 0 {
				fmt.Printf("  %v - %s\n", item.ID, item.Images.Versions[0].URL)
			}
		}
	}
	fmt.Println(media.Error())

	if !e.UsingSession {
		err = inst.Logout()
		e.CheckErr(err)
	}
}
