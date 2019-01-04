// +build ignore

package main

import (
	"fmt"
	"os"

	e "github.com/ahmdrz/goinsta/examples"
)

func main() {
	inst, err := e.InitGoinsta("<media id>")
	e.CheckErr(err)

	media, err := inst.GetMedia(os.Args[0])
	e.CheckErr(err)

	fmt.Printf("Comments %d:\n", media.Items[0].CommentCount)
	// getting items (images or videos)
	for i := range media.Items {
		// synchonizing...
		media.Items[i].Comments.Sync()

		// Iterating
		for media.Items[i].Comments.Next() {
			for _, c := range media.Items[i].Comments.Items {
				fmt.Printf("   %d - %s\n", c.ID, c.Text)
			}
		}
	}

	if !e.UsingSession {
		err = inst.Logout()
		e.CheckErr(err)
	}
}
