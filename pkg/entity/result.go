/*
	Copyright 2019 whiteblock Inc.
	This file is a part of the genesis.

	Genesis is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	Genesis is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package entity

import (
	"fmt"

	"github.com/imdario/mergo"
)

// ResultType is the type of the result
type ResultType int

// Result is the result of executing the command, contains a type and possibly an error
type Result struct {
	// Error is where the error is stored if this result is not a successful result
	Error error
	// Type is the type of result
	Type ResultType

	// Meta is additional information which can be added for debugging purposes
	Meta map[string]interface{}
}

// IsAllDone checks whether the request is completely finished. If true, then the completion
// protocol should be followed
func (res Result) IsAllDone() bool {
	return res.Error == nil && res.Type == AllDoneType
}

// IsSuccess returns whether or not the result indicates success. Both AllDoneType and
// SuccessType count as being successful
func (res Result) IsSuccess() bool {
	return res.Error == nil
}

// IsFatal returns true if there is an errr and it is marked as a fatal error,
// meaning it should not be reattempted
func (res Result) IsFatal() bool {
	return res.Error != nil && res.Type == FatalType
}

// IsRequeue returns true if this result indicates that the command should be retried at a
// later time
func (res Result) IsRequeue() bool {
	return res.Type == RequeueType || !res.IsSuccess() && !res.IsFatal()
}

// InjectMeta allows for chaining on New...Result for the return statement
func (res Result) InjectMeta(meta map[string]interface{}) Result {
	mergo.Map(&res.Meta, meta)
	return res
}

const (
	//SuccessType is the type of a successful result
	SuccessType ResultType = iota + 1

	// AllDoneType is the type for when all of the commands executed successfully and
	// there are not anymore commands to execute
	AllDoneType

	// TooSoonType is the type of a result from a cmd which tried to execute too soon
	TooSoonType

	// FatalType is the type of a result which indicates a fatal error
	FatalType

	// ErrorType is the generic error type
	ErrorType

	// RequeueType is when there is no error but should requeue something
	RequeueType
)

// NewSuccessResult indicates a successful result
func NewSuccessResult() Result {
	return Result{Type: SuccessType, Error: nil, Meta: map[string]interface{}{}}
}

// NewFatalResult creates a fatal error result. Commands with fatal errors are not retried
func NewFatalResult(err interface{}) Result {
	return Result{Type: FatalType, Error: fmt.Errorf("%v", err), Meta: map[string]interface{}{}}
}

// NewErrorResult creates a result which indicates a non-fatal error. Commands with this result should be requeued.
func NewErrorResult(err interface{}) Result {
	return Result{Type: ErrorType, Error: fmt.Errorf("%v", err), Meta: map[string]interface{}{}}
}

// NewAllDoneResult creates a result for the all done condition
func NewAllDoneResult() Result {
	return Result{Type: AllDoneType, Error: nil, Meta: map[string]interface{}{}}
}

// NewRequeueResult creates a new requeue result non-error
func NewRequeueResult() Result {
	return Result{Type: RequeueType, Error: nil, Meta: map[string]interface{}{}}
}
