package handler

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"log"
	"net/http"

	"github.com/atla/owndnd/pkg/entities/rooms"
	"github.com/atla/owndnd/pkg/service"
	"github.com/gin-gonic/gin"

	graphviz "github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

//WorldRendererHandler ...
type WorldRendererHandler struct {
	RoomsService service.RoomsService
}

func (handler *WorldRendererHandler) renderGraph(rooms []*rooms.Room) (image.Image, error) {
	g := graphviz.New()
	graph, err := g.Graph()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := graph.Close(); err != nil {
			log.Fatal(err)
		}
		g.Close()
	}()

	allNodes := make(map[string]*cgraph.Node)

	// create all rooms as nodes
	for _, room := range rooms {
		allNodes[room.ID.Hex()], _ = graph.CreateNode(room.Name)
	}

	// second loop to create the edges
	for _, room := range rooms {
		roomNode := allNodes[room.ID.Hex()]

		for _, exit := range room.Exits {
			targetRoomNode := allNodes[exit.Target]

			if roomNode != nil && targetRoomNode != nil {
				e, _ := graph.CreateEdge(exit.Name, roomNode, targetRoomNode)
				e.SetLabel(exit.Name)
			}
		}
	}

	// 1. write encoded PNG data to buffer
	var buf bytes.Buffer
	if err := g.Render(graph, graphviz.PNG, &buf); err != nil {
		log.Fatal(err)
	}

	// 2. get as image.Image instance
	image, err := g.RenderImage(graph)
	if err != nil {
		log.Fatal(err)
	}

	return image, nil
}

//Render renders the
func (handler *WorldRendererHandler) Render(c *gin.Context) {

	// get all rooms
	if rooms, err := handler.RoomsService.FindAll(); err == nil {

		img, _ := handler.renderGraph(rooms)

		//util.HTTPWriteImage(c.Writer, img)

		var b bytes.Buffer
		png.Encode(&b, img)

		imgBase64Str := base64.StdEncoding.EncodeToString(b.Bytes())
		result := "data:image/png;base64," + imgBase64Str

		c.String(http.StatusOK, result)
	}

}

// Embed into an html without PNG file
//img2html := "<html><body><img src=\"data:image/png;base64," + imgBase64Str + "\" /></body></html>"
