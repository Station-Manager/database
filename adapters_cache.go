package database

import (
	"github.com/Station-Manager/adapters"
	"github.com/Station-Manager/adapters/converters/common"
	"github.com/Station-Manager/adapters/converters/postgres"
	"github.com/Station-Manager/adapters/converters/sqlite"
)

// initAdapters lazily initializes reusable adapters for model<->type conversions.
// It is safe to call multiple times; initialization runs once.
func (s *Service) initAdapters() {
	s.adaptersOnce.Do(func() {
		// Adapter used when converting from domain types -> DB models (insert/update)
		aToModel := adapters.New()
		// Shared converters
		aToModel.RegisterConverter("Freq", common.TypeToModelFreqConverter)
		aToModel.RegisterConverter("Country", common.TypeToModelStringConverter)
		if s.DatabaseConfig.Driver == SqliteDriver {
			aToModel.RegisterConverter("QsoDate", sqlite.TypeToModelDateConverter)
			aToModel.RegisterConverter("TimeOn", sqlite.TypeToModelTimeConverter)
			aToModel.RegisterConverter("TimeOff", sqlite.TypeToModelTimeConverter)
		} else {
			aToModel.RegisterConverter("QsoDate", postgres.TypeToModelDateConverter)
			aToModel.RegisterConverter("TimeOn", postgres.TypeToModelTimeConverter)
			aToModel.RegisterConverter("TimeOff", postgres.TypeToModelTimeConverter)
		}
		s.adapterToModel = aToModel

		// Adapter used when converting from DB models -> domain types (fetch)
		aFromModel := adapters.New()
		aFromModel.RegisterConverter("Freq", common.ModelToTypeFreqConverter)
		aFromModel.RegisterConverter("Country", common.ModelToTypeStringConverter)
		if s.DatabaseConfig.Driver == SqliteDriver {
			aFromModel.RegisterConverter("QsoDate", sqlite.ModelToTypeDateConverter)
			aFromModel.RegisterConverter("TimeOn", sqlite.ModelToTypeTimeConverter)
			aFromModel.RegisterConverter("TimeOff", sqlite.ModelToTypeTimeConverter)
		} else {
			aFromModel.RegisterConverter("QsoDate", postgres.ModelToTypeDateConverter)
			aFromModel.RegisterConverter("TimeOn", postgres.ModelToTypeTimeConverter)
			aFromModel.RegisterConverter("TimeOff", postgres.ModelToTypeTimeConverter)
		}
		s.adapterFromModel = aFromModel
	})
}
