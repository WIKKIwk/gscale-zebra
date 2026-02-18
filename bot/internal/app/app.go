package app

import (
	"context"
	"log"
	"sync"

	"bot/internal/app/commands"
	"bot/internal/batchstate"
	"bot/internal/bridgeclient"
	"bot/internal/config"
	"bot/internal/erp"
	"bot/internal/labelprint"
	"bot/internal/telegram"
)

type App struct {
	cfg                      config.Config
	tg                       *telegram.Client
	erp                      *erp.Client
	qtyReader                *bridgeclient.Client
	batchState               *batchstate.Store
	imagePrinter             *labelprint.Service
	log                      *log.Logger
	startInfoMsgByChat       map[int64]int64
	batchPromptMsgByChat     map[int64]int64
	warehousePromptMsgByChat map[int64]int64
	selectionByChat          map[int64]SelectedContext
	itemChoiceByChat         map[int64]itemChoice
	batchChangeMsgByChat     map[int64]int64
	imageAwaitByChat         map[int64]bool

	batchMu     sync.Mutex
	batchNextID int64
	batchByChat map[int64]batchSession
}

type batchSession struct {
	id     int64
	cancel context.CancelFunc
}

type SelectedContext struct {
	ItemCode  string
	ItemName  string
	Warehouse string
}

type itemChoice struct {
	ItemCode string
	ItemName string
}

func New(cfg config.Config, logger *log.Logger) *App {
	if logger == nil {
		logger = log.Default()
	}
	return &App{
		cfg:                      cfg,
		tg:                       telegram.New(cfg.TelegramBotToken),
		erp:                      erp.New(cfg.ERPURL, cfg.ERPAPIKey, cfg.ERPAPISecret),
		qtyReader:                bridgeclient.New(cfg.BridgeStateFile),
		batchState:               batchstate.New(cfg.BridgeStateFile),
		imagePrinter:             labelprint.New(cfg.PrinterDevice, cfg.LabelWidthDots, cfg.LabelHeightDots),
		log:                      logger,
		startInfoMsgByChat:       make(map[int64]int64),
		batchPromptMsgByChat:     make(map[int64]int64),
		warehousePromptMsgByChat: make(map[int64]int64),
		selectionByChat:          make(map[int64]SelectedContext),
		itemChoiceByChat:         make(map[int64]itemChoice),
		batchChangeMsgByChat:     make(map[int64]int64),
		imageAwaitByChat:         make(map[int64]bool),
		batchByChat:              make(map[int64]batchSession),
	}
}

func (a *App) deps() commands.Deps {
	return commands.Deps{TG: a.tg, ERP: a.erp}
}
