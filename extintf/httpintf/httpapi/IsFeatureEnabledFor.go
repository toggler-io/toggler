package httpapi

import (
	"encoding/json"
	"net/http"
)

type IsFeatureEnabledForReqBody struct {
	Feature string `json:"feature"`
	PilotID string `json:"id"`
}

// swagger:operation GET /rollout/is-feature-enabled.json is-feature-enabled
//
// Returns a single flag state
// ---
// description:
// parameters:
//   - name: feature
//	 in: query
//	 schema:
//	   type: string
//	 example: feature-name
//   - name: id
//	 in: query
//	 schema:
//	   type: string
//	 example: public-uniq-id-of-the-pilot
// responses:
//   '200':
//	 description: Auto generated using Swagger Inspector
//	 content:
//	   application/json:
//		 schema:
//		   type: object
//		   properties:
//			 enrollment:
//			   type: boolean
//		 examples:
//		   '0':
//			 value: |
//			   {"enrollment":false}
//
func (sm *ServeMux) IsFeatureEnabledFor(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var featureName, pilotID string

	q := r.URL.Query()
	featureName = q.Get(`feature`)
	pilotID = q.Get(`id`)

	if pilotID == `` || featureName == `` {
		if handleError(w, parseJSONPayloadForIsFeatureenabled(r, &featureName, &pilotID), http.StatusBadRequest) {
			return
		}
	}

	enrollment, err := sm.UseCases.IsFeatureEnabledFor(featureName, pilotID)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var resp struct {
		Enrollment bool `json:"enrollment"`
	}
	resp.Enrollment = enrollment

	serveJSON(w, &resp)

}

func parseJSONPayloadForIsFeatureenabled(r *http.Request, featureName, pilotID *string) error {
	jsondec := json.NewDecoder(r.Body)
	var payload IsFeatureEnabledForReqBody
	if err := jsondec.Decode(&payload); err != nil {
		return err
	}

	*featureName = payload.Feature
	*pilotID = payload.PilotID
	return nil
}
