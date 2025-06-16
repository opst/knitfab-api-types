package data

import (
	"github.com/opst/knitfab-api-types/v2/internal/utils/cmp"
	"github.com/opst/knitfab-api-types/v2/plans"
	"github.com/opst/knitfab-api-types/v2/runs"
	"github.com/opst/knitfab-api-types/v2/tags"
)

type Summary struct {
	KnitId string     `json:"knitid"`
	Tags   []tags.Tag `json:"tags"`
}

func (s *Summary) Equal(o *Summary) bool {
	return s.KnitId == o.KnitId &&
		cmp.SliceEqualUnordered(s.Tags, o.Tags)
}

// Detail is the format for response body from WebAPIs below:
//
// - GET  /api/data/[?...] (as list)
//
// - POST /api/data/
//
// - PUT  /api/data/{knitId}
//
// Other Data related WebAPI respones do not use this for response.
//
// - GET  /api/data/{knitId} : as binary stream (Content-Type: application/octet-stream)
type Detail struct {
	// KnitId is the id of the Data.
	KnitId string `json:"knitId"`

	// Tags are the tags of the Data.
	Tags []tags.Tag `json:"tags"`

	// Upstream is the upsteram Run and its mountpoint outputs this Data.
	Upstream CreatedFrom `json:"upstream"`

	// Downstreams are the downstream Runs and their mountpoint inputs this Data.
	Downstreams []AssignedTo `json:"downstreams"`

	// Nomination is the nominated Plan and its mountpoint can inputs this Data.
	Nomination []NominatedBy `json:"nomination"`
}

func (d Detail) Equal(o Detail) bool {
	return d.KnitId == o.KnitId &&
		d.Upstream.Equal(o.Upstream) &&
		cmp.SliceEqualUnordered(d.Tags, o.Tags) &&
		cmp.SliceEqualUnordered(d.Downstreams, o.Downstreams) &&
		cmp.SliceEqualUnordered(d.Nomination, o.Nomination)
}

// CreatedFrom represents the source of the data
type CreatedFrom struct {
	// Mountpoint is the mountpoint which created this Data.
	//
	// This and Log are mutually exclusive.
	Mountpoint *plans.Mountpoint `json:"mountpoint,omitempty"`

	// Log is the log point which created this Data.
	//
	// This and Mountpoint are mutually exclusive.
	Log *plans.LogPoint `json:"log,omitempty"`

	// Run is the Run which created this Data.
	Run runs.Summary `json:"run"`
}

func (c CreatedFrom) Equal(o CreatedFrom) bool {
	mountpointEq := (c.Mountpoint == nil && o.Mountpoint == nil) ||
		(c.Mountpoint != nil && o.Mountpoint != nil && c.Mountpoint.Equal(*o.Mountpoint))
	logEq := (c.Log == nil && o.Log == nil) ||
		(c.Log != nil && o.Log != nil && c.Log.Equal(*o.Log))
	return c.Run.Equal(o.Run) && mountpointEq && logEq
}

// assigment representation, looking from data
type AssignedTo struct {
	Mountpoint plans.Mountpoint `json:"mountpoint"`
	Run        runs.Summary     `json:"run"`
}

func (a AssignedTo) Equal(o AssignedTo) bool {
	return a.Run.Equal(o.Run) && a.Mountpoint.Equal(o.Mountpoint)
}

// nomination representation, looking from data
type NominatedBy struct {
	plans.Mountpoint
	Plan plans.Summary `json:"plan"`
}

func (n NominatedBy) Equal(o NominatedBy) bool {
	return n.Plan.Equal(o.Plan) && n.Mountpoint.Equal(o.Mountpoint)
}
