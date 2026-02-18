package app

import "strings"

func (a *App) rememberSelection(chatID int64, itemCode, warehouse string) {
	itemCode = strings.TrimSpace(itemCode)
	warehouse = strings.TrimSpace(warehouse)
	if chatID == 0 || itemCode == "" || warehouse == "" {
		return
	}
	a.selectionByChat[chatID] = SelectedContext{ItemCode: itemCode, Warehouse: warehouse}
}

func (a *App) getSelection(chatID int64) (SelectedContext, bool) {
	v, ok := a.selectionByChat[chatID]
	if !ok {
		return SelectedContext{}, false
	}
	if strings.TrimSpace(v.ItemCode) == "" || strings.TrimSpace(v.Warehouse) == "" {
		return SelectedContext{}, false
	}
	return v, true
}

func (a *App) clearSelection(chatID int64) {
	if chatID == 0 {
		return
	}
	delete(a.selectionByChat, chatID)
}
