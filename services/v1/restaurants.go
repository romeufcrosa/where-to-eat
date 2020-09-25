package v1

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/julienschmidt/httprouter"
	domain "github.com/romeufcrosa/where-to-eat/domain/entities"
	"github.com/romeufcrosa/where-to-eat/providers"
)

// FindWhereToEat controller to get a random restaurant
func FindWhereToEat(params Params) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		ctx := req.Context()

		jsonPayload, err := bodyBytes(req)
		if err != nil {
			Error(ctx, w, err)
			return
		}

		searchRequest, err := domain.NewFromJSON(jsonPayload)
		if err != nil {
			Error(ctx, w, err)
			return
		}
		fmt.Printf("Received request: %v\n", searchRequest)

		interactor, err := providers.GetLocator()
		if err != nil {
			Error(ctx, w, err)
			return
		}

		result, err := interactor.FetchRestaurant(ctx, searchRequest)
		if err != nil {
			Error(ctx, w, err)
			return
		}

		Response(ctx, w, result)
	}
}

func bodyBytes(r *http.Request) ([]byte, error) {
	var bodyBytes []byte

	if r.Body == nil {
		// FIXME: Second param should be a error type
		return nil, nil
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	return bodyBytes, nil
}
