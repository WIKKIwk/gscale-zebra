package app

import (
	"log"

	"bot/internal/app/commands"
	"bot/internal/config"
	"bot/internal/erp"
	"bot/internal/qtyfile"
	"bot/internal/telegram"
)

type App struct {
	cfg                      config.Config
	tg                       *telegram.Client
	erp                      *erp.Client
	qtyReader                *qtyfile.Client
	log                      *log.Logger
	startInfoMsgByChat       map[int64]int64
	batchPromptMsgByChat     map[int64]int64
	warehousePromptMsgByChat map[int64]int64
	selectionByChat          map[int64]SelectedContext
}

type SelectedContext struct {
	ItemCode  string
	Warehouse string
}

func New(cfg config.Config, logger *log.Logger) *App {
	if logger == nil {
		logger = log.Default()
	}
	return &App{
		cfg:                      cfg,
		tg:                       telegram.New(cfg.TelegramBotToken),
		erp:                      erp.New(cfg.ERPURL, cfg.ERPAPIKey, cfg.ERPAPISecret),
		qtyReader:                qtyfile.New(cfg.ScaleQtyFile),
		log:                      logger,
		startInfoMsgByChat:       make(map[int64]int64),
		batchPromptMsgByChat:     make(map[int64]int64),
		warehousePromptMsgByChat: make(map[int64]int64),
		selectionByChat:          make(map[int64]SelectedContext),
	}
}

func (a *App) deps() commands.Deps {
	return commands.Deps{TG: a.tg, ERP: a.erp}
}
