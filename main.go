package main

import (
	"fmt"
	"time"
)

type Map map[string]interface{}

type Inventory struct {
	LocationId  int       `json:"location_id" validate:"required,gt=0"`
	InventoryId string    `json:"inventory_id"`
	Done        bool      `json:"done"`
	CreatedBy   Person    `json:"created_by" validate:"required,dive"`
	CreatedTime time.Time `json:"created_time" validate:"required"`
	UpdatedBy   Person    `json:"updated_by" validate:"required,dive"`
	UpdatedTime time.Time `json:"updated_time" validate:"required"`
	Buildings   Buildings `json:"buildings" validate:"required,dive"`
	Comment     string    `json:"comment"`
}

func (i *Inventory) TableName() string {
	return "inventories"
}

type Buildings []Building

type SummarizedInventory struct {
	LocationId  int       `json:"location_id" validate:"required,gt=0"`
	InventoryId string    `json:"inventory_id"`
	Done        bool      `json:"done"`
	UpdatedTime time.Time `json:"updated_time" validate:"required"`
}

type Person struct {
	Id    int    `json:"id"`
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type Building struct {
	Name     string    `json:"name" validate:"required"`
	Images   []string  `json:"images"`
	Sections []Section `json:"sections" validate:"unique=Id,dive"`
	Floors   []Floor   `json:"floors" validate:"unique=Id,dive"`
}

type Section struct {
	Id     string   `json:"id" validate:"required"`
	Name   string   `json:"name" validate:"required"`
	Images []string `json:"images"`
}

type Floor struct {
	Id    string `json:"id" validate:"required"`
	Items []Item `json:"items" validate:"required,dive"`
}

type Item struct {
	SectionId string            `json:"section_id" validate:"required"`
	TypeId    string            `json:"type_id" validate:"required,uuid"`
	Quantity  int               `json:"quantity" validate:"required,gt=0"`
	Tags      []string          `json:"tags"`
	Props     map[string]string `json:"props"`
}

func main() {
	var dest []Inventory

	query := SelectQuery{
		From: "inventories",
	}
	err := query.Fill(dest)

	// // err := db.QueryRow(context.Background(), "SELECT buildings FROM inventories WHERE inventory_id = $1", "wallaaaacccd").Scan(&dest)
	// err := SelectRow("inventories", &dest, Map{"inventory_id": "wallaaaacccd"})

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(dest)

	// fmt.Println("Buildings:", dest)

	// dest.InventoryId = "mjau_4"
	// dest.Comment = "Abooo jao"

	// err = UpdateRow("inventories", &dest, Map{"inventory_id": "mjau_4"})

	// // _, err = db.Exec(context.Background(), "INSERT INTO inventories (inventory_id, buildings) VALUES($1, $2)", "mjau_3", dest)

	// if err != nil {
	// 	fmt.Println(err)
	// }

	// doc := Inventory{}
	// err := Load(&doc, Map{"inventory_id": "wallaaaacccd"})
	// fmt.Println(doc.Buildings)

	// row := db.QueryRow("SELECT location_id, inventory_id, buildings FROM eriks_inventory.inventories WHERE inventory_id = 'wallaaaac'")
	// doc := Inventory{}
	// // var val driver.Value

	// err := row.Scan(&doc.LocationId, &doc.InventoryId, &doc.Buildings)

	// if err != nil {
	// 	fmt.Println("Error:", err)
	// }

	// // var t any

	// // err = json.Unmarshal(val.([]byte), &t)

	// // if err != nil {
	// // 	fmt.Println("Error:", err)
	// // }

	// fmt.Println(doc)
	// // fmt.Println(t)
}
