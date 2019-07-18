package httpapi

import (
	"encoding/json"
	"net/http"
)

type ClientConfigRequest struct {
	PilotID  string   `json:"id"`
	Features []string `json:"features"`
}

type ClientConfigResponseBody struct {
	States map[string]bool `json:"states"`
}

func (sm *ServeMux) RolloutConfig(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	payloadDecoder := json.NewDecoder(r.Body)

	var requestData ClientConfigRequest

	parseErr := payloadDecoder.Decode(&requestData)

	if parseErr != nil {
		parseErr = nil
		q := r.URL.Query()
		requestData.PilotID = q.Get(`id`)
		requestData.Features = append([]string{}, q[`feature`]...)
		requestData.Features = append(requestData.Features, q[`feature[]`]...)
	}

	states, err := sm.UseCases.GetPilotFlagStates(r.Context(), requestData.PilotID, requestData.Features...)

	if handleError(w, err , http.StatusInternalServerError) {
		return
	}

	serveJSON(w, ClientConfigResponseBody{States: states})
}
