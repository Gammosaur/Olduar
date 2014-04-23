package olduar

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"encoding/json"
	"io/ioutil"
	"time"
	"strings"
)

var ActivePlayersCount int = 0
var ActivePlayers map[string]*Player = make(map[string]*Player)
var ActivePlayersByUsername map[string]*Player = make(map[string]*Player)

type Players []*Player

type Player struct {
	//Basic info
	Username string			`json:"username"`
	Password string			`json:"password"`
	Name string 			`json:"name"`

	//Stats
	Health float64 			`json:"health"`
	MaxHealth float64 		`json:"health_max"`
	Inventory Inventory		`json:"inventory"`
	Stats AttributeList		`json:"-"`

	//System properties
	AuthToken string		`json:"-"`
	VotedLocation *Location `json:"-"`
	Room *Room 				`json:"-"`
	LastResponseId int64	`json:"-"`
	LastResponse time.Time	`json:"-"`

	//Equip Slots
	slotLeftHand, slotRightHand, slotHead, slotTorso, slotHands, slotLegs, slotFeet *Item
}

func LoadAllPlayers() {
	files, err := ioutil.ReadDir(MainServerConfig.DirSave+"/players");
	if(err == nil) {
		for _, file := range files {
			player := &Player{}
			player.Load(MainServerConfig.DirSave+"/players/" + file.Name())
		}
		fmt.Println("Loaded",len(ActivePlayers),"players")
	} else {
		fmt.Println("Unable to load players \""+MainServerConfig.DirSave+"/players\"")
	}
}

func (player *Player) Load(filename string) {
	data, err := ioutil.ReadFile(filename);
	if(err == nil) {
		err := json.Unmarshal(data, player)
		if(err == nil) {
			//Items
			for _, item := range player.Inventory {
				item.Load()
			}
			player.Activate()
		}
	}
}

func (player *Player) Save() {
	data, err := json.Marshal(player)
	if(err == nil) {
		err = ioutil.WriteFile(MainServerConfig.DirSave+"/players/"+player.Username+".json", data, 0777)
		if(err != nil) {
			fmt.Println("Failed to save player \""+player.Username+"\":",err)
		} else {
			fmt.Println("Player \""+player.Username+"\" has been saved")
		}
	} else {
		fmt.Println("Failed to serialize player \""+player.Username+"\":",err)
	}
}

func (p *Player) Activate() {
	p.AuthToken = "Basic "+base64.StdEncoding.EncodeToString([]byte(p.Username+":"+p.Password))
	ActivePlayers[p.AuthToken] = p
	ActivePlayersCount = len(ActivePlayers)
	ActivePlayersByUsername[strings.ToLower(p.Username)] = p
}

func (p *Player) Deactivate() {
	if(p.Room != nil) {
		p.Room.Leave(p)
	}
	delete(ActivePlayers,p.AuthToken)
	ActivePlayersCount = len(ActivePlayers)
}

func (p *Player) Ability(target *Npc, skill string) {
	//TODO: Add ability functionality
}

func (p *Player) Attack(target *Npc) {
	//TODO: Add attack functionality
}

func (p *Player) Pickup(entry string) bool {
	item := p.Room.CurrentLocation.Items.Get(entry)
	if(item != nil) {
		var weight float64 = item.Attributes.Weight
		for _, invItem := range p.Inventory {
			weight += invItem.Attributes.Weight
		}
		if(weight <= 1.0) {
			p.Room.CurrentLocation.Items.Remove(item)
			p.Inventory.Add(item)
			p.Room.Tell("You picked up "+item.Attributes.Name+" from ground",p)
			p.Room.TellAllExcept(p.Name+" picked up "+item.Attributes.Name+" from ground",p)
		} else {
			p.Room.Tell("You cannot keep more items in inventory",p)
		}
		return true
	}
	return false
}

func (p *Player) Use(entry string) bool {
	item := p.Inventory.Get(entry)
	if(item != nil) {
		item.Use(p)
		if(item.Attributes.Type == "consumable") {
			p.Inventory.Remove(item)
		}
		return true
	}
	return false
}

func (p *Player) Drop(entry string) bool {
	item := p.Inventory.Get(entry)
	if(item != nil) {
		p.Room.CurrentLocation.Items.Add(item)
		p.Inventory.Remove(item)
		return true
	}
	return false
}

func (p *Player) UnEquip(item *Item) {
	if(item != nil) {
		item.Equipped = false
	}
}

func (p *Player) Equip(entry string) bool {
	item := p.Inventory.Get(entry)
	if(item == nil) {
		return false
	}

	//Equip item and replace slots
	switch item.Attributes.Type {
	case "1hand":
		if(p.slotLeftHand != nil && p.slotLeftHand.Attributes.Type == "2hand") {
			p.UnEquip(p.slotLeftHand)
			p.UnEquip(p.slotRightHand)
			p.slotLeftHand = item
			p.slotRightHand = nil

		} else if(p.slotRightHand == nil) {
			p.slotRightHand = item

		} else {
			p.UnEquip(p.slotLeftHand)
			p.slotLeftHand = item
		}

	case "2hand":
		p.UnEquip(p.slotLeftHand)
		p.UnEquip(p.slotRightHand)
		p.slotLeftHand = item
		p.slotRightHand = item

	case "head":
		p.UnEquip(p.slotHead)
		p.slotHead = item

	case "torso":
		p.UnEquip(p.slotTorso)
		p.slotTorso = item

	case "hands":
		p.UnEquip(p.slotHands)
		p.slotHands = item

	case "legs":
		p.UnEquip(p.slotLegs)
		p.slotLegs = item

	case "feet":
		p.UnEquip(p.slotFeet)
		p.slotFeet = item

	default:
		return false
	}
	item.Equipped = true

	//Count stats
	p.Stats.Reset()
	for _, item := range p.Inventory {
		if(item.Equipped) {
			p.Stats.Append(item.Attributes.Stats)
		}
	}
	return true
}

func (p *Player) Owns(entry string) bool {
	return p.Inventory.Get(entry) != nil
}

func (p *Player) Give(entry string) {
	template, found := ItemTemplateDirectory[entry]
	if(found) {
		p.Inventory.Add(template.GenerateItem())
	}
}

func (p *Player) Heal(value float64) {
	p.Health += value
	if(p.Health > p.MaxHealth) {
		p.Health = p.MaxHealth
	}
}

func (p *Player) Damage(value float64) {
	p.Health -= value
	if(p.Health <= 0) {
		p.Health = 0
	}
}

func PlayerByAuthorization(r *http.Request) (*Player,bool) {
	//Authentication
	authToken, found := r.Header["Authorization"]
	if(!found || len(authToken)!=1) {
		return nil, false
	}

	//Search for player in state
	player, active := ActivePlayers[authToken[0]]
	if(!active) {
		return nil, false
	}

	return player, true
}
