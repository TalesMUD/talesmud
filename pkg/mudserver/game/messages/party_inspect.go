package messages

import (
	"encoding/json"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/portraits"
)

// PartyInspectGear is worn items keyed by slot. Nil marshals as {}.
type PartyInspectGear map[items.ItemSlot]*items.Item

func (g PartyInspectGear) MarshalJSON() ([]byte, error) {
	if g == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(map[items.ItemSlot]*items.Item(g))
}

// PartyInspectMessage is a read-only look at the caller or one of their party members.
type PartyInspectMessage struct {
	MessageResponse
	ID               string           `json:"id"`
	Name             string           `json:"name"`
	Class            string           `json:"class"`
	Level            int32            `json:"level"`
	Portrait         string           `json:"portrait"`
	CurrentHitPoints int32            `json:"currentHitPoints"`
	MaxHitPoints     int32            `json:"maxHitPoints"`
	CurrentMana      int32            `json:"currentMana"`
	MaxMana          int32            `json:"maxMana"`
	AttackPower      int32            `json:"attackPower"`
	Defense          int32            `json:"defense"`
	Equipped         PartyInspectGear `json:"equipped"`
}

var partyInspectSlots = []items.ItemSlot{
	items.ItemSlotHead,
	items.ItemSlotNeck,
	items.ItemSlotChest,
	items.ItemSlotWaist,
	items.ItemSlotHands,
	items.ItemSlotLegs,
	items.ItemSlotBoots,
	items.ItemSlotMainHand,
	items.ItemSlotOffHand,
	items.ItemSlotRing1,
	items.ItemSlotRing2,
}

// NewPartyInspectMessage builds the doll frame. It does not modify ch.
func NewPartyInspectMessage(userID string, ch *characters.Character) *PartyInspectMessage {
	if ch == nil {
		return nil
	}
	id := ""
	if ch.Entity != nil {
		id = ch.ID
	}
	return &PartyInspectMessage{
		MessageResponse: MessageResponse{
			Audience:   MessageAudienceOrigin,
			AudienceID: userID,
			Type:       MessageTypePartyInspect,
		},
		ID:               id,
		Name:             ch.Name,
		Class:            partyInspectClass(ch),
		Level:            ch.Level,
		Portrait:         portraits.ForPlayer(ch),
		CurrentHitPoints: ch.CurrentHitPoints,
		MaxHitPoints:     ch.MaxHitPoints,
		CurrentMana:      ch.CurrentMana,
		MaxMana:          ch.MaxMana,
		AttackPower:      partyInspectAttack(ch),
		Defense:          partyInspectDefense(ch),
		Equipped:         partyInspectGear(ch.EquippedItems),
	}
}

func partyInspectClass(ch *characters.Character) string {
	if name := strings.TrimSpace(ch.Class.Name); name != "" {
		return name
	}
	return ch.Class.ID
}

func partyInspectAttack(ch *characters.Character) int32 {
	power := ch.GetWeaponDamage() + int32(ch.GetPrimaryAttackMod())
	if power < 1 {
		return 1
	}
	return power
}

// partyInspectDefense copies worn items first. GetArmorDefense fills missing durability on the sheet it reads.
func partyInspectDefense(ch *characters.Character) int32 {
	if len(ch.EquippedItems) == 0 {
		return 0
	}
	sheet := *ch
	sheet.EquippedItems = make(map[items.ItemSlot]*items.Item, len(ch.EquippedItems))
	for slot, item := range ch.EquippedItems {
		if item == nil {
			continue
		}
		copied := *item
		sheet.EquippedItems[slot] = &copied
	}
	return sheet.GetArmorDefense()
}

func partyInspectGear(equipped map[items.ItemSlot]*items.Item) PartyInspectGear {
	gear := PartyInspectGear{}
	for _, slot := range partyInspectSlots {
		item := equipped[slot]
		if item == nil {
			continue
		}
		gear[slot] = publicInspectItem(item)
	}
	return gear
}

// publicInspectItem copies a worn item without owner-only data (container contents, lock owner, creator, bind owner).
func publicInspectItem(item *items.Item) *items.Item {
	copied := *item
	copied.Items = nil
	copied.LockedBy = ""
	copied.CreatedBy = ""
	copied.BoundToCharacterID = ""
	return &copied
}
