package main

import (
	"fmt"
	"net/http"

	"github.com/martini-contrib/render"
)

func getAllChannels(db DB, r render.Render) {
	var channels = &[]Channel{}
	_, err := db.Select(channels, "select * from channel")
	if err != nil {
		r.JSON(http.StatusInternalServerError, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, ocorreu um erro ao selecionar a lista de canais %s.", err)))
		return
	}
	r.JSON(http.StatusOK, channels)
}
