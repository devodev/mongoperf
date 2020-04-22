package scenario

import (
	"fmt"
	"sync"
	"text/template"
	"time"
)

// Templates .
var (
	ReportTemplate = `
=======================================
  Scenario Report
=======================================

  Generated by mongoperf {{ .Version }}

---------------------------------------
  Config
---------------------------------------
{{ with . -}}
{{ block "config" . }}{{ end }}
{{- end }}
---------------------------------------
  Queries
---------------------------------------
{{ with .Queries -}}
{{ range . -}}
{{ block "query" . }}{{ end }}
{{- end }}
{{- end }}
=======================================
`

	ConfigBlock = `
{{ define "config" }}
    URI:        {{ .URI }}
    Database:   {{ .Database }}
    Collection: {{ .Collection }}
    Parallel:   {{ .Parallel }}
{{ end }}
`

	QueryBlock = `
{{ define "query" }}
  > Name:              {{ .Name }}
    Action:            {{ .Action }}
    QueryCount:        {{ .QueryCount }}
    ChangeCount:       {{ .ChangeCount }}
    DurationTotal:     {{ .DurationTotal }}
    Successful:        {{ if .LastError }}false{{ else }}true{{ end }}
    ErrorCount:        {{ .ErrorCount }}
    LastError:         {{ if .LastError }}{{ .LastError.Error }}{{ else }}nil{{ end }}
{{ end }}
`
)

// Report .
type Report struct {
	Version    string
	URI        string
	Database   string
	Collection string
	Parallel   int
	Queries    map[string]*ReportQuery
}

// ReportQuery .
type ReportQuery struct {
	Name   string
	Action string

	mu            *sync.Mutex
	DurationTotal time.Duration
	QueryCount    int
	ChangeCount   int
	ErrorCount    int
	LastError     error
}

// NewReportQuery .
func NewReportQuery(name, action string) *ReportQuery {
	return &ReportQuery{
		Name:          name,
		Action:        action,
		mu:            &sync.Mutex{},
		DurationTotal: time.Duration(0),
	}
}

// Update .
func (rq *ReportQuery) Update(dur time.Duration, changes int, err error) {
	rq.mu.Lock()
	defer rq.mu.Unlock()
	rq.QueryCount++
	rq.DurationTotal += dur
	rq.ChangeCount += changes
	if err != nil {
		rq.ErrorCount++
		rq.LastError = err
	}
}

// ParseTemplates .
func ParseTemplates(name string, tmpl ...string) (*template.Template, error) {
	if len(tmpl) == 0 {
		return nil, fmt.Errorf("no templates provided")
	}
	var t *template.Template
	for idx, tStr := range tmpl {
		if t == nil {
			t = template.New(fmt.Sprintf("%v-%d", name, idx))
		}
		var err error
		t, err = t.Parse(tStr)
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}
