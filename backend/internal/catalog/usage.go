package catalog

type ItemUsage struct {
	InputToRecipes  []string `json:"input_to_recipes"`
	InputToUpgrades []string `json:"input_to_upgrades"`
	CraftedBy       []string `json:"crafted_by"`
	UpgradedFrom    []string `json:"upgraded_from"`
	UpgradesInto    []string `json:"upgrades_into"`
	DroppedInZones  []string `json:"dropped_in_zones"`
	SoldAsSKUs      []string `json:"sold_as_skus"`
}

type ItemDetail struct {
	Item
	Usage ItemUsage `json:"usage"`
}

func (snapshot Snapshot) UsageOf(itemID int) ItemUsage {
	usage := ItemUsage{
		InputToRecipes:  []string{},
		InputToUpgrades: []string{},
		CraftedBy:       []string{},
		UpgradedFrom:    []string{},
		UpgradesInto:    []string{},
		DroppedInZones:  []string{},
		SoldAsSKUs:      []string{},
	}
	for _, recipe := range snapshot.Recipes {
		if recipe.OutputItemID == itemID {
			usage.CraftedBy = append(usage.CraftedBy, recipe.ID)
		}
		if containsItem(recipe.Inputs, itemID) {
			usage.InputToRecipes = append(usage.InputToRecipes, recipe.ID)
		}
	}
	for _, upgrade := range snapshot.Upgrades {
		if upgrade.ToItemID == itemID {
			usage.UpgradedFrom = append(usage.UpgradedFrom, upgrade.ID)
		}
		if upgrade.FromItemID == itemID {
			usage.UpgradesInto = append(usage.UpgradesInto, upgrade.ID)
		}
		if containsItem(upgrade.Inputs, itemID) {
			usage.InputToUpgrades = append(usage.InputToUpgrades, upgrade.ID)
		}
	}
	for _, zone := range snapshot.Zones {
		for _, lootEntry := range zone.Loot {
			if lootEntry.ItemID == itemID {
				usage.DroppedInZones = append(usage.DroppedInZones, zone.ID)
				break
			}
		}
	}
	for _, shopItem := range snapshot.ShopItems {
		if containsItem(shopItem.Contents, itemID) {
			usage.SoldAsSKUs = append(usage.SoldAsSKUs, shopItem.SKU)
		}
	}
	return usage
}

func containsItem(quantities []ItemQuantity, itemID int) bool {
	for _, quantity := range quantities {
		if quantity.ItemID == itemID {
			return true
		}
	}
	return false
}
