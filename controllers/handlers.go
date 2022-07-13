// Copyright (c) 2021 Red Hat, Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controllers

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/go-jose/go-jose/v3/json"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	api "github.com/redhat-appstudio/service-provider-integration-operator/api/v1beta1"
	"go.uber.org/zap"
	"go.uber.org/zap/zapio"
)

// OkHandler is a Handler implementation that responds only with http.StatusOK.
// Typically, used for liveness and readiness probes
func OkHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// CallbackSuccessHandler is a Handler implementation that responds with HTML page
// This page is a landing page after successfully completing the OAuth flow.
// Resource file location is prefixed with `../` to be compatible with tests running locally.
func CallbackSuccessHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "../static/callback_success.html")
}

// viewData structure is used to pass parameters during callback_error.html template processing.
type viewData struct {
	Title   string
	Message string
}

// CallbackErrorHandler is a Handler implementation that responds with HTML page
// This page is a landing page after unsuccessfully completing the OAuth flow.
// Resource file location is prefixed with `../` to be compatible with tests running locally.
func CallbackErrorHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	errorMsg := q.Get("error")
	errorDescription := q.Get("error_description")
	data := viewData{
		Title:   errorMsg,
		Message: errorDescription,
	}
	tmpl, _ := template.ParseFiles("../static/callback_error.html")

	err := tmpl.Execute(w, data)
	if err == nil {
		w.WriteHeader(http.StatusOK)
	} else {
		zap.L().Error("failed to process template: %s", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf("Error response returned to OAuth callback: %s. Message: %s ", errorMsg, errorDescription)))
	}

}

// HandleUpload returns Handler implementation that is relied on provided TokenUploader to persist provided credentials
// for some concrete SPIAccessToken.
func HandleUpload(uploader TokenUploader) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, err := WithAuthFromRequestIntoContext(r, r.Context())
		if err != nil {
			LogErrorAndWriteResponse(w, http.StatusUnauthorized, "failed extract authorization information from headers", err)
			return
		}

		vars := mux.Vars(r)
		tokenObjectName := vars["name"]
		tokenObjectNamespace := vars["namespace"]

		if len(tokenObjectName) < 1 || len(tokenObjectNamespace) < 1 {
			LogDebugAndWriteResponse(w, http.StatusInternalServerError, "Incorrect service deployment. Token name and namespace can't be omitted or empty.")
			return
		}

		data := &api.Token{}
		if err := json.NewDecoder(r.Body).Decode(data); err != nil {
			LogErrorAndWriteResponse(w, http.StatusBadRequest, "failed to decode request body as token JSON", err)
			return
		}

		if data.AccessToken == "" {
			LogDebugAndWriteResponse(w, http.StatusBadRequest, "access token can't be omitted or empty")
			return
		}

		if err := uploader.Upload(ctx, tokenObjectName, tokenObjectNamespace, data); err != nil {
			LogErrorAndWriteResponse(w, http.StatusInternalServerError, "failed to upload the token", err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// MiddlewareHandler is a Handler that composed couple of different responsibilities.
// Like:
// - Request logging
// - CORS processing
func MiddlewareHandler(allowedOrigins []string, h http.Handler) http.Handler {
	return handlers.LoggingHandler(&zapio.Writer{Log: zap.L(), Level: zap.InfoLevel},
		handlers.CORS(handlers.AllowedOrigins(allowedOrigins),
			handlers.AllowCredentials(),
			handlers.AllowedHeaders([]string{"Accept", "Accept-Language", "Content-Language", "Origin", "Authorization"}))(h))
}