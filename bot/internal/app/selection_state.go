package app

import "strings"

func (a *App) rememberSelection(chatID int64, itemCode, itemName, warehouse string) {
	itemCode = strings.TrimSpace(itemCode)
	itemName = strings.TrimSpace(itemName)
	warehouse = strings.TrimSpace(warehouse)
	if chatID == 0 || itemCode == "" || warehouse == "" {
		return
	}
	if itemName == "" {
		itemName = itemCode
	}
	a.selectionByChat[chatID] = SelectedContext{ItemCode: itemCode, ItemName: itemName, Warehouse: warehouse}
}

func (a *App) rememberItemChoice(chatID int64, itemCode, itemName string) {
	itemCode = strings.TrimSpace(itemCode)
	itemName = strings.TrimSpace(itemName)
	if chatID == 0 || itemCode == "" {
		return
	}
	if itemName == "" {
		itemName = itemCode
	}
	a.itemChoiceByChat[chatID] = itemChoice{ItemCode: itemCode, ItemName: itemName}
}

func (a *App) itemNameFor(chatID int64, itemCode string) string {
	itemCode = strings.TrimSpace(itemCode)
	if chatID == 0 || itemCode == "" {
		return ""
	}
	v, ok := a.itemChoiceByChat[chatID]
	if !ok {
		return ""
	}
	if strings.TrimSpace(v.ItemCode) != itemCode {
		return ""
	}
	return strings.TrimSpace(v.ItemName)
}

func (a *App) getSelection(chatID int64) (SelectedContext, bool) {
	v, ok := a.selectionByChat[chatID]
	if !ok {
		return SelectedContext{}, false
	}
	if strings.TrimSpace(v.ItemCode) == "" || strings.TrimSpace(v.Warehouse) == "" {
		return SelectedContext{}, false
	}
	if strings.TrimSpace(v.ItemName) == "" {
		v.ItemName = strings.TrimSpace(v.ItemCode)
	}
	return v, true
}

func (a *App) clearSelection(chatID int64) {
	if chatID == 0 {
		return
	}
	delete(a.selectionByChat, chatID)
	delete(a.itemChoiceByChat, chatID)
}
