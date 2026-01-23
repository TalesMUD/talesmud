package handler

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/service"

	graphviz "github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

// GraphNode represents a room node in the graph
type GraphNode struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Area        string  `json:"area"`
	AreaType    string  `json:"areaType"`
	RoomType    string  `json:"roomType"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Z           int32   `json:"z"`
}

// MinimalRoom is a lightweight room structure for the world map
type MinimalRoom struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Area     string        `json:"area"`
	AreaType string        `json:"areaType"`
	Coords   *MinimalCoord `json:"coords,omitempty"`
	Exits    []MinimalExit `json:"exits"`
}

// MinimalCoord represents room coordinates
type MinimalCoord struct {
	X int32 `json:"x"`
	Y int32 `json:"y"`
	Z int32 `json:"z"`
}

// MinimalExit represents an exit for the world map
type MinimalExit struct {
	Name       string `json:"name"`
	Target     string `json:"target"`
	IsCardinal bool   `json:"isCardinal"`
}

// GraphEdge represents a connection between rooms
type GraphEdge struct {
	ID         string `json:"id"`
	Source     string `json:"source"`
	Target     string `json:"target"`
	Label      string `json:"label"`
	ExitType   string `json:"exitType"`
	IsHidden   bool   `json:"isHidden"`
	IsCardinal bool   `json:"isCardinal"`
}

// GraphData represents the complete graph structure
type GraphData struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

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
		allNodes[room.ID], _ = graph.CreateNode(room.Name)
	}

	// second loop to create the edges
	for _, room := range rooms {
		roomNode := allNodes[room.ID]

		for _, exit := range *room.Exits {
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

// isCardinalDirection checks if an exit name is a cardinal direction
func isCardinalDirection(exitName string) bool {
	cardinals := map[string]bool{
		"north": true,
		"south": true,
		"east":  true,
		"west":  true,
	}
	return cardinals[exitName]
}

// calculatePositions assigns X,Y positions to rooms without coordinates
func (handler *WorldRendererHandler) calculatePositions(rooms []*rooms.Room) map[string]*GraphNode {
	const gridScale = 100.0 // pixels per grid unit
	nodes := make(map[string]*GraphNode)
	
	// First pass: create nodes with existing coordinates
	positioned := make(map[string]bool)
	for _, room := range rooms {
		node := &GraphNode{
			ID:          room.ID,
			Name:        room.Name,
			Description: room.Description,
			Area:        room.Area,
			AreaType:    room.AreaType,
			RoomType:    room.RoomType,
			Z:           0,
		}
		
		if room.Coords != nil {
			node.X = float64(room.Coords.X) * gridScale
			node.Y = float64(room.Coords.Y) * gridScale
			node.Z = room.Coords.Z
			positioned[room.ID] = true
		}
		
		nodes[room.ID] = node
	}
	
	// Second pass: auto-position rooms without coordinates based on connections
	// Use BFS to expand from positioned rooms
	changed := true
	maxIterations := 10
	iteration := 0
	
	for changed && iteration < maxIterations {
		changed = false
		iteration++
		
		for _, room := range rooms {
			// Skip if already positioned
			if positioned[room.ID] {
				continue
			}
			
			// Look for positioned neighbors
			if room.Exits != nil {
				for _, exit := range *room.Exits {
					if positioned[exit.Target] && isCardinalDirection(exit.Name) {
						targetNode := nodes[exit.Target]
						currentNode := nodes[room.ID]
						
						// Calculate position based on reverse direction
						switch exit.Name {
						case "north":
							currentNode.X = targetNode.X
							currentNode.Y = targetNode.Y - gridScale
						case "south":
							currentNode.X = targetNode.X
							currentNode.Y = targetNode.Y + gridScale
						case "east":
							currentNode.X = targetNode.X - gridScale
							currentNode.Y = targetNode.Y
						case "west":
							currentNode.X = targetNode.X + gridScale
							currentNode.Y = targetNode.Y
						}
						currentNode.Z = targetNode.Z
						positioned[room.ID] = true
						changed = true
						break
					}
				}
			}
			
			// Also check incoming connections
			if !positioned[room.ID] {
				for _, otherRoom := range rooms {
					if !positioned[otherRoom.ID] || otherRoom.Exits == nil {
						continue
					}
					
					for _, exit := range *otherRoom.Exits {
						if exit.Target == room.ID && isCardinalDirection(exit.Name) {
							sourceNode := nodes[otherRoom.ID]
							currentNode := nodes[room.ID]
							
							// Calculate position based on direction
							switch exit.Name {
							case "north":
								currentNode.X = sourceNode.X
								currentNode.Y = sourceNode.Y + gridScale
							case "south":
								currentNode.X = sourceNode.X
								currentNode.Y = sourceNode.Y - gridScale
							case "east":
								currentNode.X = sourceNode.X + gridScale
								currentNode.Y = sourceNode.Y
							case "west":
								currentNode.X = sourceNode.X - gridScale
								currentNode.Y = sourceNode.Y
							}
							currentNode.Z = sourceNode.Z
							positioned[room.ID] = true
							changed = true
							break
						}
					}
					if positioned[room.ID] {
						break
					}
				}
			}
		}
	}
	
	// Third pass: place remaining unpositioned rooms in a grid
	unpositionedCount := 0
	for _, room := range rooms {
		if !positioned[room.ID] {
			node := nodes[room.ID]
			// Place in a separate area to the side
			node.X = -500 - float64(unpositionedCount%5)*gridScale
			node.Y = float64(unpositionedCount/5) * gridScale
			node.Z = 0
			unpositionedCount++
		}
	}
	
	return nodes
}

// RenderGraphData returns the world map as JSON graph data
func (handler *WorldRendererHandler) RenderGraphData(c *gin.Context) {
	rooms, err := handler.RoomsService.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Calculate positions for all rooms
	nodeMap := handler.calculatePositions(rooms)
	
	// Build nodes array
	nodes := make([]GraphNode, 0, len(nodeMap))
	for _, node := range nodeMap {
		nodes = append(nodes, *node)
	}
	
	// Build edges array
	edges := make([]GraphEdge, 0)
	edgeID := 0
	
	for _, room := range rooms {
		if room.Exits == nil {
			continue
		}
		
		for _, exit := range *room.Exits {
			// Only create edge if target exists
			if _, exists := nodeMap[exit.Target]; exists {
				edge := GraphEdge{
					ID:         room.ID + "-" + exit.Target + "-" + exit.Name,
					Source:     room.ID,
					Target:     exit.Target,
					Label:      exit.Name,
					ExitType:   string(exit.Type),
					IsHidden:   exit.Hidden,
					IsCardinal: isCardinalDirection(exit.Name),
				}
				edges = append(edges, edge)
				edgeID++
			}
		}
	}
	
	graphData := GraphData{
		Nodes: nodes,
		Edges: edges,
	}
	
	c.JSON(http.StatusOK, graphData)
}

// GetMinimalRooms returns a lightweight list of rooms for the world map editor
func (handler *WorldRendererHandler) GetMinimalRooms(c *gin.Context) {
	rooms, err := handler.RoomsService.FindAll()
	if err != nil {
		log.Printf("Error fetching rooms: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Found %d rooms for minimal map", len(rooms))

	minimalRooms := make([]MinimalRoom, 0, len(rooms))

	for _, room := range rooms {
		mr := MinimalRoom{
			ID:       room.ID,
			Name:     room.Name,
			Area:     room.Area,
			AreaType: room.AreaType,
			Exits:    make([]MinimalExit, 0),
		}

		// Add coordinates if present
		if room.Coords != nil {
			mr.Coords = &MinimalCoord{
				X: room.Coords.X,
				Y: room.Coords.Y,
				Z: room.Coords.Z,
			}
		}

		// Add exits
		if room.Exits != nil {
			for _, exit := range *room.Exits {
				mr.Exits = append(mr.Exits, MinimalExit{
					Name:       exit.Name,
					Target:     exit.Target,
					IsCardinal: isCardinalDirection(exit.Name),
				})
			}
		}

		minimalRooms = append(minimalRooms, mr)
	}

	c.JSON(http.StatusOK, gin.H{
		"count": len(minimalRooms),
		"rooms": minimalRooms,
	})
}
