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

	fmt.Printf("Following: %v\n", user.Friendship.Following)
	err = user.Follow()
	e.CheckErr(err)
	fmt.Printf("After func call: Following: %v\n", user.Friendship.Following)

	if !e.UsingSession {
		err = inst.Logout()
		e.CheckErr(err)
	}
}
