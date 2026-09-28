package uss_test

import (
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bwawan.com/openuss/sdk/api/scdussv1"
	"github.com/stretchr/testify/require"
)

func TestMakeUSSReport(t *testing.T) {
	recorder := httptest.NewRecorder()

	report := scdussv1.ErrorReport{
		Exchange: scdussv1.ExchangeRecord{
			Url:          "the-url",
			Method:       "the-method",
			RecorderRole: "the-role",
		},
	}

	bytes, _ := json.Marshal(report)

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(bytes)))
	handler := newHandler()
	handler.MakeUSSReport(recorder, request)
	require.Equal(t, http.StatusCreated, recorder.Code)

	var body scdussv1.ErrorReport
	require.NoError(t, json.UnmarshalRead(recorder.Body, &body))
	require.Equal(t, report.Exchange, body.Exchange)
	require.NotZero(t, *body.ReportId)
}
