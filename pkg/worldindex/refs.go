package worldindex

import (
	"sort"
	"strings"
)

// RefItem is one inbound or outbound reference.
type RefItem struct {
	Type   Kind   `json:"type"`
	ID     string `json:"id"`
	Name   string `json:"name"`
	Field  string `json:"field"`
	Reason string `json:"reason"`
	Path   string `json:"path"`
}

// RefGroup is the references of one entity type.
type RefGroup struct {
	Type Kind      `json:"type"`
	Refs []RefItem `json:"refs"`
}

// Refs is who points at an entity, and what that entity points at.
type Refs struct {
	Type     Kind       `json:"type"`
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Inbound  []RefGroup `json:"inbound"`
	Outbound []RefGroup `json:"outbound"`
}

// References groups the index edges for one entity.
// The bool is false when the snapshot does not contain it.
func (ix *Index) References(kind Kind, id string) (Refs, bool) {
	id = strings.TrimSpace(id)
	if ix == nil || id == "" || !ix.Has(kind, id) {
		return Refs{}, false
	}
	view := Refs{
		Type:     kind,
		ID:       id,
		Name:     ix.Name(kind, id),
		Inbound:  groupEdges(ix, ix.Inbound(kind, id), true),
		Outbound: groupEdges(ix, ix.Outbound(kind, id), false),
	}
	return view, true
}

func groupEdges(ix *Index, edges []Edge, inbound bool) []RefGroup {
	byType := map[Kind][]RefItem{}
	for _, edge := range edges {
		kind := edge.ToType
		otherID := edge.ToID
		if inbound {
			kind = edge.FromType
			otherID = edge.FromID
		}
		item := RefItem{
			Type:   kind,
			ID:     otherID,
			Name:   ix.Name(kind, otherID),
			Field:  edge.Field,
			Reason: edgeReason(edge),
			Path:   ix.CreatorPath(kind, otherID),
		}
		byType[kind] = append(byType[kind], item)
	}
	kinds := make([]Kind, 0, len(byType))
	for kind := range byType {
		kinds = append(kinds, kind)
	}
	sort.Slice(kinds, func(i, j int) bool { return kinds[i] < kinds[j] })
	groups := make([]RefGroup, 0, len(kinds))
	for _, kind := range kinds {
		refs := byType[kind]
		sort.Slice(refs, func(i, j int) bool {
			if refs[i].ID != refs[j].ID {
				return refs[i].ID < refs[j].ID
			}
			return refs[i].Field < refs[j].Field
		})
		groups = append(groups, RefGroup{Type: kind, Refs: refs})
	}
	return groups
}

func edgeReason(edge Edge) string {
	return string(edge.FromType) + " " + edge.FromID + " " + prettyHow(edge.How)
}

func prettyHow(how string) string {
	for _, prefix := range []string{"instance hidden exit ", "instance exit ", "hidden exit ", "exit "} {
		if name, ok := strings.CutPrefix(how, prefix); ok && name != "" && !strings.Contains(name, " ") {
			return strings.TrimSuffix(prefix, " ") + " '" + name + "'"
		}
	}
	if rest, ok := strings.CutPrefix(how, "quest "); ok {
		return rest
	}
	return how
}
