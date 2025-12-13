package sqlite

import (
	"context"
	"database/sql"
	stderr "errors"
	"fmt"
	"strings"
	"time"

	"github.com/Station-Manager/database/sqlite/adapters"
	"github.com/Station-Manager/database/sqlite/models"
	"github.com/Station-Manager/errors"
	"github.com/Station-Manager/types"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

/**********************************************************************************************************************
 * QSO Methods
 **********************************************************************************************************************/

func (s *Service) InsertQsoWithContext(ctx context.Context, qso types.Qso) (int64, error) {
	const op errors.Op = "sqlite.Service.InsertQsoWithContext"
	if err := checkService(op, s); err != nil {
		return 0, err
	}

	h, err := s.getOpenHandle(op)
	if err != nil {
		return 0, err
	}

	ctx, cancel := s.ensureCtxTimeout(ctx)
	defer cancel()

	model, err := adapters.QsoTypeToModel(qso)
	if err != nil {
		return 0, errors.New(op).Err(err)
	}

	if err = model.Insert(ctx, h, boil.Infer()); err != nil {
		return 0, errors.New(op).Err(err)
	}

	return model.ID, nil
}

func (s *Service) FetchQsoSliceBySessionIDWithContext(ctx context.Context, id int64) ([]types.Qso, error) {
	const op errors.Op = "sqlite.Service.FetchQsoSliceBySessionIDWithContext"
	if err := checkService(op, s); err != nil {
		return nil, err
	}

	if id < 1 {
		return nil, errors.New(op).Msg(errMsgInvalidId)
	}

	h, err := s.getOpenHandle(op)
	if err != nil {
		return nil, err
	}

	ctx, cancel := s.ensureCtxTimeout(ctx)
	defer cancel()

	var mods []qm.QueryMod
	mods = append(mods, models.QsoWhere.SessionID.EQ(id))
	mods = append(mods, qm.OrderBy(models.QsoColumns.CreatedAt+" DESC"))

	slice, err := models.Qsos(mods...).All(ctx, h)
	if err != nil {
		return nil, errors.New(op).Err(err).Msg("Failed to fetch QSOs by session ID.")
	}

	result := make([]types.Qso, 0, len(slice))
	for _, qso := range slice {
		typeQso, er := adapters.QsoModelToType(qso)
		if er != nil {
			s.LoggerService.WarnWith().Int64("qso.id", qso.ID).Err(er).Msg("Failed to adapt QSO for contact history.")
			continue
		}
		result = append(result, typeQso)
	}
	return result, nil
}

func (s *Service) FetchQsoSliceByCallsignWithContext(ctx context.Context, callsign string) ([]types.ContactHistory, error) {
	const op errors.Op = "sqlite.Service.FetchContactHistoryWithContext"
	if err := checkService(op, s); err != nil {
		return nil, err
	}

	callsign = strings.TrimSpace(callsign)
	if callsign == "" {
		return nil, errors.New(op).Msg(errMsgEmptyCallsign)
	}

	h, err := s.getOpenHandle(op)
	if err != nil {
		return nil, err
	}

	ctx, cancel := s.ensureCtxTimeout(ctx)
	defer cancel()

	callsign = fmt.Sprintf("%%%s%%", callsign)

	var mods []qm.QueryMod
	mods = append(mods, models.QsoWhere.Call.LIKE(callsign))
	mods = append(mods, qm.OrderBy(models.QsoColumns.CreatedAt+" DESC"))
	slice, err := models.Qsos(mods...).All(ctx, h)
	if err != nil {
		if stderr.Is(err, sql.ErrNoRows) {
			return nil, errors.ErrNotFound
		}
		return nil, errors.New(op).Err(err).Msg("Failed to fetch contact history.")
	}

	history := make([]types.ContactHistory, 0, len(slice))
	for _, qso := range slice {
		typeQso, er := adapters.QsoModelToType(qso)
		if er != nil {
			s.LoggerService.WarnWith().Int64("qso.id", qso.ID).Err(er).Msg("Failed to adapt QSO for contact history.")
			continue
		}
		item := types.ContactHistory{
			ID:      typeQso.ID,
			Band:    typeQso.Band,
			Freq:    typeQso.Freq,
			Mode:    typeQso.Mode,
			QsoDate: typeQso.QsoDate,
			TimeOn:  typeQso.TimeOn,
			Name:    typeQso.Name,
			Country: typeQso.Country,
			Call:    typeQso.Call,
			RstSent: typeQso.RstSent,
			RstRcvd: typeQso.RstRcvd,
			Notes:   typeQso.Notes,
		}
		history = append(history, item)
	}

	return history, nil
}

func (s *Service) FetchQsoCountByLogbookIdWithContext(ctx context.Context, id int64) (int64, error) {
	const op errors.Op = "sqlite.Service.FetchQsoCountByLogbookIdWithContext"
	if err := checkService(op, s); err != nil {
		return 0, err
	}

	if id < 1 {
		return 0, errors.New(op).Msg(errMsgInvalidId)
	}

	h, err := s.getOpenHandle(op)
	if err != nil {
		return 0, err
	}

	ctx, cancel := s.ensureCtxTimeout(ctx)
	defer cancel()

	count, err := models.Qsos(models.QsoWhere.LogbookID.EQ(id)).Count(ctx, h)
	if err != nil {
		return 0, errors.New(op).Err(err).Msg("Failed to fetch QSO count by logbook ID.")
	}

	return count, nil
}

func (s *Service) UpdateQsoWithContext(ctx context.Context, qso types.Qso) error {
	const op errors.Op = "sqlite.Service.FetchQsoCountByLogbookIdWithContext"
	if err := checkService(op, s); err != nil {
		return err
	}

	if qso.ID < 1 {
		return errors.New(op).Msgf("QSO ID is invalid: %d", qso.ID)
	}

	h, err := s.getOpenHandle(op)
	if err != nil {
		return err
	}

	ctx, cancel := s.ensureCtxTimeout(ctx)
	defer cancel()

	model, err := adapters.QsoTypeToModel(qso)
	if err != nil {
		return errors.New(op).Err(err)
	}

	if _, err = model.Update(ctx, h, boil.Infer()); err != nil {
		return errors.New(op).Err(err)
	}

	return nil
}

/**********************************************************************************************************************
 * ContactedStation Methods
 **********************************************************************************************************************/

func (s *Service) FetchContactedStationByCallsignWithContext(ctx context.Context, callsign string) (types.ContactedStation, error) {
	const op errors.Op = "sqlite.Service.FetchContactedStationByCallsignWithContext"
	if err := checkService(op, s); err != nil {
		return types.ContactedStation{}, err
	}

	callsign = strings.TrimSpace(callsign)
	if callsign == "" {
		return types.ContactedStation{}, errors.New(op).Msg(errMsgEmptyCallsign)
	}

	h, err := s.getOpenHandle(op)
	if err != nil {
		return types.ContactedStation{}, err
	}

	ctx, cancel := s.ensureCtxTimeout(ctx)
	defer cancel()

	model, err := models.ContactedStations(models.ContactedStationWhere.Call.EQ(callsign)).One(ctx, h)
	if err != nil {
		if stderr.Is(err, sql.ErrNoRows) {
			return types.ContactedStation{}, errors.ErrNotFound
		}
		return types.ContactedStation{}, errors.New(op).Err(err)
	}

	contactedStation, err := adapters.ContactedStationModelToType(model)
	if err != nil {
		return types.ContactedStation{}, errors.New(op).Err(err)
	}

	return contactedStation, nil
}

func (s *Service) InsertContactedStationWithContext(ctx context.Context, station types.ContactedStation) (int64, error) {
	const op errors.Op = "sqlite.Service.InsertContactedStationWithContext"
	if err := checkService(op, s); err != nil {
		return 0, err
	}

	h, err := s.getOpenHandle(op)
	if err != nil {
		return 0, err
	}

	ctx, cancel := s.ensureCtxTimeout(ctx)
	defer cancel()

	model, err := adapters.ContactedStationTypeToModel(station)
	if err != nil {
		return 0, errors.New(op).Err(err)
	}
	if err = model.Insert(ctx, h, boil.Infer()); err != nil {
		return 0, errors.New(op).Err(err).Msg("Inserting new contacted station failed.")
	}

	return model.ID, nil
}

func (s *Service) UpdateContactedStationWithContext(ctx context.Context, station types.ContactedStation) error {
	const op errors.Op = "sqlite.Service.UpdateContactedStationWithContext"
	if err := checkService(op, s); err != nil {
		return err
	}

	h, err := s.getOpenHandle(op)
	if err != nil {
		return err
	}

	ctx, cancel := s.ensureCtxTimeout(ctx)
	defer cancel()

	model, err := adapters.ContactedStationTypeToModel(station)
	if err != nil {
		return errors.New(op).Err(err)
	}

	model.ModifiedAt = null.TimeFrom(time.Now())

	_, err = model.Update(ctx, h, boil.Infer())
	if err != nil {
		return errors.New(op).Err(err).Msg("Updating contacted station failed.")
	}

	return nil
}

/**********************************************************************************************************************
 * Country Methods
 **********************************************************************************************************************/

func (s *Service) FetchCountryByCallsignWithContext(ctx context.Context, callsign string) (types.Country, error) {
	const op errors.Op = "sqlite.Service.FetchCountryByCallsignWithContext"
	if err := checkService(op, s); err != nil {
		return types.Country{}, err
	}

	callsign = strings.TrimSpace(callsign)
	if callsign == "" {
		return types.Country{}, errors.New(op).Msg(errMsgEmptyCallsign)
	}

	h, err := s.getOpenHandle(op)
	if err != nil {
		return types.Country{}, err
	}

	ctx, cancel := s.ensureCtxTimeout(ctx)
	defer cancel()

	mods := []qm.QueryMod{
		qm.Where("? LIKE "+models.TableNames.Country+".prefix || '%'", callsign),
		qm.OrderBy("LENGTH(" + models.TableNames.Country + ".prefix) DESC"),
		qm.Limit(1),
	}

	model, err := models.Countries(mods...).One(ctx, h)
	if err != nil {
		if stderr.Is(err, sql.ErrNoRows) {
			return types.Country{}, errors.ErrNotFound
		}
		return types.Country{}, errors.New(op).Err(err)
	}

	country, err := adapters.CountryModelToType(model)
	if err != nil {
		return types.Country{}, errors.New(op).Err(err)
	}

	return country, nil
}

func (s *Service) FetchCountryByNameWithContext(ctx context.Context, name string) (types.Country, error) {
	const op errors.Op = "sqlite.Service.FetchCountryByNameWithContext"
	if err := checkService(op, s); err != nil {
		return types.Country{}, err
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return types.Country{}, errors.New(op).Msg("Country name cannot be empty")
	}

	h, err := s.getOpenHandle(op)
	if err != nil {
		return types.Country{}, err
	}

	ctx, cancel := s.ensureCtxTimeout(ctx)
	defer cancel()

	model, err := models.Countries(models.CountryWhere.Name.EQ(name)).One(ctx, h)
	if err != nil {
		if stderr.Is(err, sql.ErrNoRows) {
			return types.Country{}, errors.ErrNotFound
		}
		return types.Country{}, errors.New(op).Err(err)
	}

	country, err := adapters.CountryModelToType(model)
	if err != nil {
		return types.Country{}, errors.New(op).Err(err)
	}

	return country, nil
}

func (s *Service) InsertCountryWithContext(ctx context.Context, country types.Country) (int64, error) {
	const op errors.Op = "sqlite.Service.InsertCountryWithContext"
	if err := checkService(op, s); err != nil {
		return 0, err
	}

	h, err := s.getOpenHandle(op)
	if err != nil {
		return 0, err
	}

	ctx, cancel := s.ensureCtxTimeout(ctx)
	defer cancel()

	model, err := adapters.CountryTypeToModel(country)
	if err != nil {
		return 0, errors.New(op).Err(err)
	}
	if err = model.Insert(ctx, h, boil.Infer()); err != nil {
		return 0, errors.New(op).Err(err).Msg("Inserting new country failed.")
	}

	return model.ID, nil
}

func (s *Service) UpdateCountryWithContext(ctx context.Context, country types.Country) error {
	const op errors.Op = "sqlite.Service.UpdateCountryWithContext"
	if err := checkService(op, s); err != nil {
		return err
	}

	h, err := s.getOpenHandle(op)
	if err != nil {
		return err
	}

	ctx, cancel := s.ensureCtxTimeout(ctx)
	defer cancel()

	model, err := adapters.CountryTypeToModel(country)
	if err != nil {
		return errors.New(op).Err(err)
	}

	if _, err = model.Update(ctx, h, boil.Infer()); err != nil {
		return errors.New(op).Err(err).Msg("Updating country failed.")
	}

	return nil
}

/**********************************************************************************************************************
 * Logbook Methods
 **********************************************************************************************************************/

func (s *Service) FetchLogbookByIDWithContext(ctx context.Context, id int64) (types.Logbook, error) {
	const op errors.Op = "sqlite.Service.FetchLogbookByIDWithContext"
	if err := checkService(op, s); err != nil {
		return types.Logbook{}, err
	}

	if id < 1 {
		return types.Logbook{}, errors.New(op).Msg(errMsgInvalidId)
	}
	h, err := s.getOpenHandle(op)
	if err != nil {
		return types.Logbook{}, err
	}

	ctx, cancel := s.ensureCtxTimeout(ctx)
	defer cancel()

	model, err := models.FindLogbook(ctx, h, id)
	if err != nil {
		if stderr.Is(err, sql.ErrNoRows) {
			return types.Logbook{}, errors.ErrNotFound
		}
		return types.Logbook{}, errors.New(op).Err(err)
	}

	logbook, err := adapters.LogbookModelToType(model)
	if err != nil {
		return types.Logbook{}, errors.New(op).Err(err)
	}

	return logbook, nil
}

/**********************************************************************************************************************
 * Session Methods
 **********************************************************************************************************************/

func (s *Service) SoftDeleteSessionByIDWithContext(ctx context.Context, id int64) error {
	const op errors.Op = "sqlite.Service.SoftDeleteSessionByIDWithContext"
	if err := checkService(op, s); err != nil {
		return err
	}

	if id < 1 {
		return errors.New(op).Msg(errMsgInvalidId)
	}
	h, err := s.getOpenHandle(op)
	if err != nil {
		return err
	}

	ctx, cancel := s.ensureCtxTimeout(ctx)
	defer cancel()

	model, err := models.FindSession(ctx, h, id)
	if err != nil {
		if stderr.Is(err, sql.ErrNoRows) {
			return errors.ErrNotFound
		}
		return errors.New(op).Err(err)
	}

	if _, err = model.Delete(ctx, h, false); err != nil {
		return errors.New(op).Err(err).Msgf("Failed to soft delete session: %d", id)
	}

	return nil
}

func (s *Service) InsertSessionWithContext(ctx context.Context) (int64, error) {
	const op errors.Op = "sqlite.Service.InsertSessionWithContext"
	if err := checkService(op, s); err != nil {
		return 0, err
	}

	h, err := s.getOpenHandle(op)
	if err != nil {
		return 0, err
	}

	ctx, cancel := s.ensureCtxTimeout(ctx)
	defer cancel()

	session := models.Session{}
	if err = session.Insert(ctx, h, boil.Infer()); err != nil {
		return 0, errors.New(op).Err(err).Msg("Inserting new session failed.")
	}

	return session.ID, nil
}
