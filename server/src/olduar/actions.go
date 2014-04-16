package olduar

import (
	"strings"
)

type ActionFunction func (state *GameState, player *Player, config map[string]interface{})

func AppendVariablesToString(str string, player *Player) string {
	str = strings.Replace(str,"%player%",player.Name,-1)
	return str
}

var ActionsDirectory = map[string]ActionFunction {

	"message": func(state *GameState,player *Player,config map[string]interface{}) {
		//Automatically processed - just a placeholder
	},

	"give": func(state *GameState,player *Player,config map[string]interface{}) {
		//Amount of looted items
		amount := 1
		value, found := config["amount"]
		if(found) {
			amount = (int)(value.(float64))
		}

		//Prepare loot table
		table := ItemLootTable{}
		for _, itemConfig := range config["items"].([]interface{}) {
			config := itemConfig.(map[string]interface {})
			item := &ItemLoot{}
			value, found := config["id"]
			if(found) {
				item.Template = ItemTemplateDirectory[value.(string)]

				value, found = config["chance"]
				if(found) {
					item.Chance = value.(float64)
				} else {
					item.Chance = 1.0
				}

				value, found = config["msg_party"]
				if(found) {
					item.MessageParty = value.(string)
				}

				value, found = config["msg_player"]
				if(found) {
					item.MessagePlayer = value.(string)
				}

				table = append(table,item)
			}
		}

		//Get looted items
		items := GetItemsFromLootTable(player,amount,table)
		for _, item := range items {
			player.Inventory = append(player.Inventory,item)
		}
	},

	"effect": func(state *GameState,player *Player,config map[string]interface{}) {
		//Process effect
		fxType, found := config["type"]
		if(found) {
			switch(fxType){
			case "hurt":
				value, found := config["value"]
				if(found) {
					player.Damage((int64)(value.(float64)))
				}
				break
			case "heal":
				value, found := config["value"]
				if(found) {
					player.Heal((int64)(value.(float64)))
				}
				break
			}
		}
	},

}

