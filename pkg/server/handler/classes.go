package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/talesmud/talesmud/pkg/classkit"
	"github.com/talesmud/talesmud/pkg/entities/characters"
)

// GetClasses returns the public class catalog. Balance numbers stay off this payload.
func GetClasses(c *gin.Context) {
	base := []classkit.RaceInfo{
		{ID: characters.RaceHuman.ID, Name: characters.RaceHuman.Name, Blurb: characters.RaceHuman.Description},
		{ID: characters.RaceDwarf.ID, Name: characters.RaceDwarf.Name, Blurb: characters.RaceDwarf.Description},
		{ID: characters.RaceElf.ID, Name: characters.RaceElf.Name, Blurb: characters.RaceElf.Description},
		{ID: characters.RaceConstruct.ID, Name: characters.RaceConstruct.Name, Blurb: characters.RaceConstruct.Description},
	}
	c.JSON(http.StatusOK, classkit.ClientView(base))
}
