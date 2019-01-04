package main

import (
	"./goinsta"
	"github.com/gin-gonic/gin"
)

func likeTarget(c *gin.Context) (bool, string) {

	inst := goinsta.New(c.PostForm("login"), c.PostForm("password"))
	if err := inst.Login(); err != nil {
		return true, err.Error()
	}

	inst.Export("~/.goinsta")

	user, err := inst.Profiles.ByName(c.PostForm("target"))

	if err != nil {
		inst.Logout()
		return true, "Аккаунт \"" + c.PostForm("target") + "\" не существует"
	}
	media := user.Feed()
	var counter = 20
	for media.Next() {
		for _, item := range media.Items {
			if counter == 0 {
				break
			}
			item.Like()
			counter--
		}
	}

	inst.Logout()

	return false, ""
}
