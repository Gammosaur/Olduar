package olduar

import (
	"fmt"
	"encoding/json"
	"io/ioutil"
	"math/rand"
)

// Loader for location templates

func LoadLocations() bool {

	files, err := ioutil.ReadDir(MainServerConfig.DirLocations);
	if(err != nil) {
		fmt.Println("Unable to load locations from \""+MainServerConfig.DirLocations+"\"")
		return false
	}

	type LoadingRegion struct {
		Region string 					`json:"region"`
		Description string 				`json:"desc"`
		Locations LocationTemplates 	`json:"locations"`
	}

	//Load locations
	fmt.Println("Loading location files:")
	for _, file := range files {
		data, err := ioutil.ReadFile(MainServerConfig.DirLocations+"/"+file.Name());
		if(err == nil) {
			region := LoadingRegion{}
			err := json.Unmarshal(data,&region)
			if(err == nil && region.Region != "") {
				fmt.Println("\t" + file.Name() + ": (Region: "+region.Region+") " + region.Description);
				_, found := LocationTemplateDirectoryRegions[region.Region];
				if(!found) {
					LocationTemplateDirectoryRegions[region.Region] = make(LocationTemplates,0)
				}
				for _, location := range region.Locations {
					location.Region = region.Region
					//Set action charges to unlimited for 0 value
					for index, action := range location.Actions {
						if(action.Charges == 0) {
							location.Actions[index].Charges = -1 //-1 = unlimited
						}
					}
					//Add location as entry
					if(location.Id != "") {
						LocationTemplateDirectoryEntries[location.Id] = location
					}
					//Add location to region
					LocationTemplateDirectoryRegions[region.Region] = append(LocationTemplateDirectoryRegions[region.Region],location)
				}
			} else {
				fmt.Println("\t" + file.Name() + ": Failed to load")
			}
		} else {
			fmt.Println("\t" + file.Name() + ": Failed to load")
		}
	}

	//Check for amount of regions (must be > 2) & "start" region
	_, found := LocationTemplateDirectoryRegions["start"];
	if(!found) {
		fmt.Println("Error: \"start\" region not found!")
		return false
	}

	fmt.Println()

	return true

}

func CreateLocationFromTemplate(template *LocationTemplate) *Location {
	loc := Location{}

	loc.Name = template.Name
	loc.Description = template.Description
	loc.DescriptionShort = template.DescriptionShort
	loc.Actions = template.Actions
	loc.Exits = template.Exits
	loc.Region = template.Region
	loc.Visited = false

	//Generate items on ground
	loc.Items = make(Inventory,0)
	for _, item := range template.Items {
		if(item.Chance > 0 && item.Chance < rand.Float64()) {
			continue
		}
		if(item.Id != "") {
			finalItem := ItemTemplateDirectory[item.Id].GenerateItem()
			if(finalItem != nil) {
				loc.Items = append(loc.Items,finalItem)
			}
		} else if(item.Group != "") {
			//Give any template from item group
		}
	}

	return &loc
}

func CreateLocationFromRegion(region string) *Location {
	templateBank, found := LocationTemplateDirectoryRegions[region]
	if(!found) {
		return nil
	}
	template := templateBank[rand.Intn(len(templateBank))]
	return CreateLocationFromTemplate(template)
}

func CreateLocationFromEntry(entry string) *Location {
	template, found := LocationTemplateDirectoryEntries[entry]
	if(!found) {
		return nil
	}
	return CreateLocationFromTemplate(template)
}

// Locations types and functions

var LocationTemplateDirectoryEntries = make(map[string]*LocationTemplate)
var LocationTemplateDirectoryRegions = make(map[string]LocationTemplates)

type LocationItemTemplates []*LocationItemTemplate
type LocationItemTemplate struct {
	Id string 		`json:"id"`
	Group string 	`json:"group"`
	Chance float64	`json:"chance"`
}

type LocationTemplates []*LocationTemplate
type LocationTemplate struct {
	Id string 					`json:"id"`
	Name string 				`json:"name"`
	Region string				`json:"region,omitempty"`
	Description string 			`json:"desc"`
	DescriptionShort string 	`json:"desc_short"`
	Actions Actions				`json:"actions,omitempty"`
	Exits LocationExits			`json:"exits,omitempty"`
	Items LocationItemTemplates `json:"items"`
}

