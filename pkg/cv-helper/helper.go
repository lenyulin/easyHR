package cvhelper

import (
	"context"
	"easyHR/pkg/cv-helper/model"
	"easyHR/pkg/cv-helper/parser"
	"easyHR/pkg/cv-helper/store"
	"easyHR/pkg/logger"
	"time"
)

type CVHelper struct {
	parser *parser.PDFParser
	store  *store.MongoStore
	log    logger.LoggerV1
}

func NewCVHelper(store *store.MongoStore, log logger.LoggerV1) *CVHelper {
	return &CVHelper{
		parser: parser.NewPDFParser(),
		store:  store,
		log:    log,
	}
}

func (h *CVHelper) Run(ch <-chan string, onIdle func()) {
	go func() {
		for {
			select {
			case filePath, ok := <-ch:
				if !ok {
					return
				}
				h.log.Info("CVHelper received file: " + filePath)
				if err := h.processFile(filePath); err != nil {
					h.log.Error("Failed to process file " + filePath + ": " + err.Error())
				} else {
					h.log.Info("Successfully processed file: " + filePath)
				}
			default:
				// Channel is empty
				if onIdle != nil {
					onIdle()
				}
				// Sleep briefly to avoid busy loop if channel remains empty
				time.Sleep(100 * time.Millisecond)

				// Check channel again (blocking wait to resume normal processing)
				select {
				case filePath, ok := <-ch:
					if !ok {
						return
					}
					h.log.Info("CVHelper received file: " + filePath)
					if err := h.processFile(filePath); err != nil {
						h.log.Error("Failed to process file " + filePath + ": " + err.Error())
					} else {
						h.log.Info("Successfully processed file: " + filePath)
					}
				case <-time.After(time.Second):
					// Continue loop to trigger onIdle again if needed
				}
			}
		}
	}()
}

func (h *CVHelper) processFile(filePath string) error {
	// Parse PDF
	content, err := h.parser.Parse(filePath)
	if err != nil {
		return err
	}

	// Create CV model
	cv := &model.CV{
		FilePath:  filePath,
		Content:   content,
		ParsedAt:  time.Now(),
		CreatedAt: time.Now(),
	}

	// Save to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return h.store.Save(ctx, cv)
}
