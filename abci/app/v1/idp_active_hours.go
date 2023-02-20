/**
 * Copyright (c) 2018, 2019 National Digital ID COMPANY LIMITED
 *
 * This file is part of NDID software.
 *
 * NDID is the free software: you can redistribute it and/or modify it under
 * the terms of the Affero GNU General Public License as published by the
 * Free Software Foundation, either version 3 of the License, or any later
 * version.
 *
 * NDID is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 * See the Affero GNU General Public License for more details.
 *
 * You should have received a copy of the Affero GNU General Public License
 * along with the NDID source code. If not, see https://www.gnu.org/licenses/agpl.txt.
 *
 * Please contact info@ndid.co.th for any further questions
 *
 */

package app

import (
	"encoding/json"
	"errors"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/tendermint/tendermint/abci/types"
	"google.golang.org/protobuf/proto"

	"github.com/ndidplatform/smart-contract/v8/abci/code"
	"github.com/ndidplatform/smart-contract/v8/abci/utils"
	data "github.com/ndidplatform/smart-contract/v8/protos/data"
)

type ActiveHours struct {
	DayOfWeek     string         `json:"day_of_week"`
	ActivePeriods []ActivePeriod `json:"active_periods"`
	UtcOffset     int            `json:"utc_offset"` // in minutes
}

type ActivePeriod struct {
	Start string `json:"start"` // format "HH:mm:ss"
	End   string `json:"end"`   // format "HH:mm:ss"
}

type SetIdpActiveHoursParam struct {
	ActiveHours []ActiveHours `json:"active_hours"`
}

var validDayOfWeek = map[string]int{
	"sun": 1,
	"mon": 2,
	"tue": 3,
	"wed": 4,
	"thu": 5,
	"fri": 6,
	"sat": 7,
}

var timeFormatRegex = regexp.MustCompile("^(2[0-3]|[01]?[0-9]):([0-5]?[0-9]):([0-5]?[0-9])$")

func convertTimePeriodToSecondsFromMidnight(startTime string, endTime string) (start int, end int, err error) {
	splittedStartTime := strings.Split(startTime, ":")
	splittedEndTime := strings.Split(endTime, ":")

	startTimeHour, err := strconv.Atoi(splittedStartTime[0])
	if err != nil {
		return 0, 0, errors.New("invalid time format (start)")
	}

	startTimeMinute, err := strconv.Atoi(splittedStartTime[1])
	if err != nil {
		return 0, 0, errors.New("invalid time format (start)")
	}

	startTimeSecond, err := strconv.Atoi(splittedStartTime[2])
	if err != nil {
		return 0, 0, errors.New("invalid time format (start)")
	}

	startTimeSecondsFromMidnight := startTimeHour*60*60 + startTimeMinute*60 + startTimeSecond

	endTimeHour, err := strconv.Atoi(splittedEndTime[0])
	if err != nil {
		return 0, 0, errors.New("invalid time format (end)")
	}

	endTimeMinute, err := strconv.Atoi(splittedEndTime[1])
	if err != nil {
		return 0, 0, errors.New("invalid time format (end)")
	}

	endTimeSecond, err := strconv.Atoi(splittedEndTime[2])
	if err != nil {
		return 0, 0, errors.New("invalid time format (end)")
	}

	endTimeSecondsFromMidnight := endTimeHour*60*60 + endTimeMinute*60 + endTimeSecond

	return startTimeSecondsFromMidnight, endTimeSecondsFromMidnight, nil
}

func convertTimePeriodSecondsFromMidnightToString(start int, end int) (startTime string, endTime string) {
	startTimeHour := start / 60 / 60
	startTimeMinute := (start - startTimeHour*60*60) / 60
	startTimeSecond := start - startTimeHour*60*60 - startTimeMinute*60

	startTimeHourStr := strconv.Itoa(startTimeHour)
	if len(startTimeHourStr) == 1 {
		startTimeHourStr = "0" + startTimeHourStr
	}
	startTimeMinuteStr := strconv.Itoa(startTimeMinute)
	if len(startTimeMinuteStr) == 1 {
		startTimeMinuteStr = "0" + startTimeMinuteStr
	}
	startTimeSecondStr := strconv.Itoa(startTimeSecond)
	if len(startTimeSecondStr) == 1 {
		startTimeSecondStr = "0" + startTimeSecondStr
	}

	endTimeHour := end / 60 / 60
	endTimeMinute := (end - endTimeHour*60*60) / 60
	endTimeSecond := end - endTimeHour*60*60 - endTimeMinute*60

	endTimeHourStr := strconv.Itoa(endTimeHour)
	if len(endTimeHourStr) == 1 {
		endTimeHourStr = "0" + endTimeHourStr
	}
	endTimeMinuteStr := strconv.Itoa(endTimeMinute)
	if len(endTimeMinuteStr) == 1 {
		endTimeMinuteStr = "0" + endTimeMinuteStr
	}
	endTimeSecondStr := strconv.Itoa(endTimeSecond)
	if len(endTimeSecondStr) == 1 {
		endTimeSecondStr = "0" + endTimeSecondStr
	}

	startTime = strings.Join(
		[]string{
			startTimeHourStr,
			startTimeMinuteStr,
			startTimeSecondStr,
		},
		":",
	)

	endTime = strings.Join(
		[]string{
			endTimeHourStr,
			endTimeMinuteStr,
			endTimeSecondStr,
		},
		":",
	)

	return startTime, endTime
}

func (app *ABCIApplication) validateSetIdpActiveHours(funcParam SetIdpActiveHoursParam, callerNodeID string, committedState bool) error {
	// TODO: validate is IdP node

	dayOfWeek := make(map[string]bool)
	for _, activeHours := range funcParam.ActiveHours {
		_, valid := validDayOfWeek[activeHours.DayOfWeek]
		if !valid {
			return &ApplicationError{
				Code:    code.UnknownError,
				Message: "invalid day of week",
			}
		}

		_, duplicate := dayOfWeek[activeHours.DayOfWeek]
		if duplicate {
			return &ApplicationError{
				Code:    code.UnknownError,
				Message: "duplicate day of week",
			}
		}

		dayOfWeek[activeHours.DayOfWeek] = true

		if len(activeHours.ActivePeriods) == 0 {
			return &ApplicationError{
				Code:    code.UnknownError,
				Message: "empty active hours",
			}
		}

		periods := make([][2]int, len(activeHours.ActivePeriods))
		for _, activePeriod := range activeHours.ActivePeriods {
			validTimeFormat := timeFormatRegex.MatchString(activePeriod.Start)
			if !validTimeFormat {
				return &ApplicationError{
					Code:    code.UnknownError,
					Message: "invalid time format (start)",
				}
			}

			validTimeFormat = timeFormatRegex.MatchString(activePeriod.End)
			if !validTimeFormat {
				return &ApplicationError{
					Code:    code.UnknownError,
					Message: "invalid time format (end)",
				}
			}

			startTimeSecondsFromMidnight, endTimeSecondsFromMidnight, err :=
				convertTimePeriodToSecondsFromMidnight(activePeriod.Start, activePeriod.End)
			if err != nil {
				return &ApplicationError{
					Code:    code.UnknownError,
					Message: err.Error(),
				}
			}

			// check for time overlap
			for _, period := range periods {
				if startTimeSecondsFromMidnight >= period[0] && startTimeSecondsFromMidnight <= period[1] {
					return &ApplicationError{
						Code:    code.UnknownError,
						Message: "active hours overlap",
					}
				}

				if endTimeSecondsFromMidnight >= period[0] && endTimeSecondsFromMidnight <= period[1] {
					return &ApplicationError{
						Code:    code.UnknownError,
						Message: "active hours overlap",
					}
				}
			}

			periods = append(periods, [2]int{startTimeSecondsFromMidnight, endTimeSecondsFromMidnight})
		}

		// validate UTC offset
		// min: -12 hour
		// max: +14 hour
		if activeHours.UtcOffset < -12*60 || activeHours.UtcOffset > 14*60 {
			return &ApplicationError{
				Code:    code.UnknownError,
				Message: "invalid UTC offset",
			}
		}

	}

	return nil
}

func (app *ABCIApplication) setIdpActiveHoursCheckTx(param []byte, callerNodeID string) types.ResponseCheckTx {
	var funcParam SetIdpActiveHoursParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return ReturnCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateSetIdpActiveHours(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return ReturnCheckTx(appErr.Code, appErr.Message)
		}
		return ReturnCheckTx(code.UnknownError, err.Error())
	}

	return ReturnCheckTx(code.OK, "")
}

func (app *ABCIApplication) setIdpActiveHours(param []byte, callerNodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetIdpActiveHours, Parameter: %s", param)
	var funcParam SetIdpActiveHoursParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateSetIdpActiveHours(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.ReturnDeliverTxLog(appErr.Code, appErr.Message, "")
		}
		return app.ReturnDeliverTxLog(code.UnknownError, err.Error(), "")
	}

	var idpActiveHours data.IdPActiveHours
	idpActiveHours.ActiveHours = make([]*data.ActiveHours, 0)

	inputActiveHours := make([]ActiveHours, len(funcParam.ActiveHours))
	copy(inputActiveHours, funcParam.ActiveHours)

	// sort by day of week starting with "sun"
	sort.Slice(inputActiveHours, func(i, j int) bool {
		return validDayOfWeek[inputActiveHours[i].DayOfWeek] < validDayOfWeek[inputActiveHours[j].DayOfWeek]
	})

	for _, inputActiveHour := range inputActiveHours {
		var activeHourToSet *data.ActiveHours
		activeHourToSet.DayOfWeek = inputActiveHour.DayOfWeek
		activeHourToSet.UtcOffset = int32(inputActiveHour.UtcOffset)
		activeHourToSet.ActivePeriods = make([]*data.ActivePeriod, len(inputActiveHour.ActivePeriods))

		for _, activePeriod := range inputActiveHour.ActivePeriods {
			startTimeSecondsFromMidnight, endTimeSecondsFromMidnight, err :=
				convertTimePeriodToSecondsFromMidnight(activePeriod.Start, activePeriod.End)
			if err != nil {
				return app.ReturnDeliverTxLog(code.UnknownError, err.Error(), "")
			}

			var activePeriodToSet *data.ActivePeriod
			activePeriodToSet.Start = int32(startTimeSecondsFromMidnight)
			activePeriodToSet.End = int32(endTimeSecondsFromMidnight)

			activeHourToSet.ActivePeriods = append(activeHourToSet.ActivePeriods, activePeriodToSet)
		}

		idpActiveHours.ActiveHours = append(idpActiveHours.ActiveHours, activeHourToSet)
	}

	idpActiveHoursKey := idpActiveHoursKeyPrefix + keySeparator + callerNodeID

	newIdpActiveHoursBytes, err := utils.ProtoDeterministicMarshal(&idpActiveHours)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(idpActiveHoursKey), []byte(newIdpActiveHoursBytes))

	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

// TODO: remove/unset

type GetIdpActiveHoursParam struct {
	NodeID string `json:"node_id"`
}

type GetIdpActiveHoursResult struct {
	ActiveHours []ActiveHours `json:"active_hours"`
}

func (app *ABCIApplication) getIdpActiveHours(param []byte) types.ResponseQuery {
	app.logger.Infof("GetIdpActiveHours, Parameter: %s", param)
	var funcParam GetIdpActiveHoursParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	idpActiveHoursKey := idpActiveHoursKeyPrefix + keySeparator + funcParam.NodeID
	idpActiveHoursValue, err := app.state.Get([]byte(idpActiveHoursKey), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if idpActiveHoursValue == nil {
		// FIXME: return empty instead?
		return app.ReturnQuery(nil, "not found", app.state.Height)
	}
	var idpActiveHours data.IdPActiveHours
	err = proto.Unmarshal([]byte(idpActiveHoursValue), &idpActiveHours)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	var retVal GetIdpActiveHoursResult
	retVal.ActiveHours = make([]ActiveHours, 0)

	for _, activeHour := range idpActiveHours.ActiveHours {
		retValActiveHour := ActiveHours{
			DayOfWeek:     activeHour.DayOfWeek,
			ActivePeriods: make([]ActivePeriod, len(activeHour.ActivePeriods)),
			UtcOffset:     int(activeHour.UtcOffset),
		}

		for _, activePeriod := range activeHour.ActivePeriods {
			startTime, endTime :=
				convertTimePeriodSecondsFromMidnightToString(int(activePeriod.Start), int(activePeriod.End))

			retValActivePeriod := ActivePeriod{
				Start: startTime,
				End:   endTime,
			}

			retValActiveHour.ActivePeriods = append(retValActiveHour.ActivePeriods, retValActivePeriod)
		}

		retVal.ActiveHours = append(retVal.ActiveHours, retValActiveHour)
	}

	retValJSON, err := json.Marshal(retVal)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	return app.ReturnQuery(retValJSON, "success", app.state.Height)
}
