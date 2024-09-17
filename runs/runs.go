package runs

import (
	"github.com/opst/knitfab-api-types/internal/utils/cmp"
	"github.com/opst/knitfab-api-types/misc/rfctime"
	"github.com/opst/knitfab-api-types/plans"
)

type Summary struct {
	RunId     string          `json:"runId"`
	Status    string          `json:"status"`
	UpdatedAt rfctime.RFC3339 `json:"updatedAt"`
	Exit      *Exit           `json:"exit,omitempty"`
	Plan      plans.Summary   `json:"plan"`
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
	Inputs  []Assignment `json:"inputs"`
	Outputs []Assignment `json:"outputs"`
	Log     *LogSummary  `json:"log"`
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
