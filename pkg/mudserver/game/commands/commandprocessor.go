package commands

import (
	"log"
	"strings"

	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// CommandProcessor ... global user struct to control logins
type CommandProcessor struct {
	commands map[string]Command
	Help     map[string]string
}

// NewCommandProcessor .. creates a new command processor
func NewCommandProcessor() *CommandProcessor {
	var commandProcessor = &CommandProcessor{
		commands: make(map[string]Command),
		Help:     make(map[string]string),
	}
	// only once?
	commandProcessor.registerCommands()
	return commandProcessor
}

// RegisterCommand ... register
func (commandProcessor *CommandProcessor) RegisterCommand(command Command, desc string, keys ...string) {

	cmds := "["
	for i, key := range keys {
		if i > 0 {
			cmds += ", "
		}
		cmds += key
		commandProcessor.commands[key] = command
	}

	cmds += "]"

	commandProcessor.Help[cmds] = desc
}

// Process ...asd
func (commandProcessor *CommandProcessor) Process(game def.GameCtrl, message *messages.Message) bool {

	parts := strings.Fields(message.Data)

	if len(parts) > 0 {
		if roomActionMatches(game, message) {
			return false
		}

		var key = strings.ToLower(parts[0])
		if val, ok := commandProcessor.commands[key]; ok {

			// support for commands without parameters to enable input like "i did find something" but still support the command "i" for inventory
			if val.Key() != nil && val.Key().Matches(key, message.Data) == false {
				return false
			}

			log.Println("Found command " + key + " executing...")
			return val.Execute(game, message)
		}
	}

	return false

}

func (commandProcessor *CommandProcessor) registerCommands() {

	// Chat commands
	commandProcessor.RegisterCommand(&SayCommand{}, "Speak to the current room: say <message>", "say")
	commandProcessor.RegisterCommand(&TellCommand{}, "Send private message: tell <player> <message>", "tell", "whisper", "pm")
	commandProcessor.RegisterCommand(&PartyCommand{}, "Party commands: party create|invite|accept|decline|leave|list|say or party <message>", "party", "p")
	commandProcessor.RegisterCommand(&ScreamCommand{}, "Scream through the room", "scream")
	commandProcessor.RegisterCommand(&ShrugCommand{}, "Shrug emote", "shrug")
	commandProcessor.RegisterCommand(&SelectCharacterCommand{}, "Select a character, use: sc [charactername]", "sc", "selectcharacter")
	commandProcessor.RegisterCommand(&ListCharactersCommand{}, "List all your characters", "lc", "listcharacters")
	commandProcessor.RegisterCommand(&HelpCommand{processor: commandProcessor}, "Are you really asking?", "h", "help")
	commandProcessor.RegisterCommand(&WhoCommand{}, "List all online players", "who")
	commandProcessor.RegisterCommand(&InventoryCommand{}, "Display your inventory", "inventory", "i")
	commandProcessor.RegisterCommand(&CharacterCommand{}, "Display character stats", "character", "char", "stats")
	commandProcessor.RegisterCommand(&NewCharacterCommand{}, "Create a new character", "newcharacter", "nc")
	commandProcessor.RegisterCommand(&TalkCommand{}, "Talk to an NPC: talk [npc-name]", "talk", "speak")

	// Item commands
	commandProcessor.RegisterCommand(&PickupCommand{}, "Pick up an item: pickup [item]", "pickup", "get", "take")
	commandProcessor.RegisterCommand(&DropCommand{}, "Drop an item: drop [item] [quantity]", "drop")
	commandProcessor.RegisterCommand(&DestroyCommand{}, "Destroy an item: destroy [item]", "destroy", "discard")
	commandProcessor.RegisterCommand(&ExamineCommand{}, "Examine an item: examine [item]", "examine", "inspect")
	commandProcessor.RegisterCommand(&UseCommand{}, "Use a consumable item: use [item]", "use", "eat", "drink", "consume")

	// Trade commands
	commandProcessor.RegisterCommand(&ListCommand{}, "List merchant inventory: list", "list", "shop", "trade")
	commandProcessor.RegisterCommand(&BuyCommand{}, "Buy from merchant: buy [item] [quantity]", "buy")
	commandProcessor.RegisterCommand(&SellCommand{}, "Sell to merchant: sell [item] [quantity]", "sell")
	commandProcessor.RegisterCommand(&ValueCommand{}, "Check item sell price: value [item]", "value", "price")
	commandProcessor.RegisterCommand(&RepairCommand{}, "Repair worn armor at a merchant: repair", "repair")

	// Equipment commands
	commandProcessor.RegisterCommand(&EquipCommand{}, "Equip an item: equip [item]", "equip", "wear")
	commandProcessor.RegisterCommand(&UnequipCommand{}, "Unequip an item: unequip [slot or item]", "unequip", "remove")
	commandProcessor.RegisterCommand(&EquipmentCommand{}, "Show equipped items", "equipment", "eq", "gear")

	// Combat commands
	commandProcessor.RegisterCommand(&AttackCommand{}, "Attack a target: attack [target]", "attack", "a", "hit")
	commandProcessor.RegisterCommand(&DefendCommand{}, "Take defensive stance in combat", "defend", "d", "guard")
	commandProcessor.RegisterCommand(&FleeCommand{}, "Attempt to flee from combat", "flee", "run", "escape")
	commandProcessor.RegisterCommand(&CombatStatusCommand{}, "Show combat status", "status", "cs", "combat")
	commandProcessor.RegisterCommand(&CastCommand{}, "Use a skill in combat: cast <skill> [target]", "cast", "spell")
	commandProcessor.RegisterCommand(&SkillsCommand{}, "Manage skills: skills [equip|unequip] [name]", "skills", "spells", "abilities")
	commandProcessor.RegisterCommand(&SkillShortcutCommand{}, "Quick-cast skill by slot number (in combat): 1, 2, 3, 4", "1", "2", "3", "4")

	// Quest commands
	commandProcessor.RegisterCommand(&QuestLogCommand{}, "Show your quest log", "quests", "ql", "questlog")
	commandProcessor.RegisterCommand(&QuestDetailCommand{}, "Show quest details: quest [name]", "quest")
	commandProcessor.RegisterCommand(&AbandonQuestCommand{}, "Abandon a quest: abandon [name]", "abandon")

	// Respawn commands
	commandProcessor.RegisterCommand(&BindCommand{}, "Bind respawn point: bind", "bind")

	// HP Recovery commands
	commandProcessor.RegisterCommand(&RestCommand{}, "Rest to recover HP faster: rest", "rest")

	// Character progression commands
	commandProcessor.RegisterCommand(&SpendCommand{}, "Spend attribute points: spend <attr> [amount]", "spend")

}
