package runs

import (
	"github.com/opst/knitfab-api-types/v2/internal/utils/cmp"
	"github.com/opst/knitfab-api-types/v2/misc/rfctime"
	"github.com/opst/knitfab-api-types/v2/plans"
)

type Summary struct {
	// RunId is the id of the Run.
	RunId string `json:"runId"`

	// Status is the status of the Run.
	//
	// This is one of:
	//
	// - "deactivated": This Run is deactivated. It is not going to be running.
	//
	// - "waiting": This Run is waiting to be running.
	//
	// - "ready": This Run is ready to be running. It is waiting for the worker starts.
	//
	// - "starting": This Run is pulling images, or preparing the environment.
	// The Worker for the Run can be running because of the interval of the periodical health check.
	//
	// - "running": This Run's Worker is running.
	//
	// - "completing": It is observed that the run's worker has stopped successfully.
	//
	// - "aborting": It is observed, or should be done that the run's worker has stopped insuccessfully.
	//
	// - "done": This Run has been finished, successfuly.
	// The Run's output can be used by other Runs.
	//
	// - "failed": This Run has been finished with error.
	//
	// - "invalidated": This run was discarded
	Status string `json:"status"`

	// UpdatedAt is the time of the last update of the Run.
	UpdatedAt rfctime.RFC3339 `json:"updatedAt"`

	// Exit is the exit status of the Run.
	//
	// This is nil if the Run is not finished.
	Exit *Exit `json:"exit,omitempty"`

	// Plan which the Run is created from.
	Plan plans.Summary `json:"plan"`
}

func (s Summary) Equal(o Summary) bool {

	exitEq := (s.Exit == nil && o.Exit == nil) ||
		(s.Exit != nil && o.Exit != nil && s.Exit.Equal(*o.Exit))

	return s.RunId == o.RunId &&
		exitEq &&
		s.Plan.Equal(o.Plan) &&
		s.Status == o.Status &&
		s.UpdatedAt.Equal(o.UpdatedAt)
}

type Exit struct {
	Code    uint8  `json:"code"`
	Message string `json:"message"`
}

func (e Exit) Equal(o Exit) bool {
	return e.Code == o.Code && e.Message == o.Message
}

// Detail is the format for response body from WebAPIs below:
//
// - GET /api/runs/[?...] (as list)
//
// - GET /api/runs/{runId}
//
// - GET /api/runs/{runId}
//
// - PUT /api/runs/{runId}/abort
//
// - PUT /api/runs/{runId}/tearoff
//
// - PUT /api/runs/{runId}/retry
//
// Other Run related WebAPI do not use this for response.
//
// - GET    /api/runs/{runId}/log: text stream (Content-Type: text/plain)
//
// - DELETE /api/runs/{runId}: empty response ("204 No Content" on success)
type Detail struct {
	Summary

	// Inputs are pairs of input mountpoints and inputted Data of the Run.
	Inputs []Assignment `json:"inputs"`

	// Outputs are pairs of output mountpoints and outputted Data of the Run.
	Outputs []Assignment `json:"outputs"`

	// Log is the log point of the Run.
	Log *LogSummary `json:"log"`
}

func (r Detail) Equal(o Detail) bool {

	logEq := (r.Log == nil && o.Log == nil) ||
		(r.Log != nil && o.Log != nil && r.Log.Equal(*o.Log))

	return r.RunId == o.RunId &&
		r.Plan.Equal(o.Plan) &&
		r.Status == o.Status &&
		r.UpdatedAt.Equal(o.UpdatedAt) &&
		cmp.SliceEqualUnordered(r.Inputs, o.Inputs) &&
		cmp.SliceEqualUnordered(r.Outputs, o.Outputs) &&
		logEq
}

type Assignment struct {
	plans.Mountpoint
	KnitId string `json:"knitId"`
}

func (a Assignment) Equal(o Assignment) bool {
	return a.Mountpoint.Equal(o.Mountpoint) && a.KnitId == o.KnitId
}

type LogSummary struct {
	plans.LogPoint
	KnitId string `json:"knitId"`
}

func (l LogSummary) Equal(o LogSummary) bool {
	return l.LogPoint.Equal(o.LogPoint) && l.KnitId == o.KnitId
}
