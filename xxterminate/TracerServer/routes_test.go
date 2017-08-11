package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"
	"xxterminator-plugin/xxterminate/TracerServer/tracer"
)

/* Used to order request and their corresponding tests. */
type RequestTestPair struct {
	Request *http.Request
	Test    func(*httptest.ResponseRecorder, *testing.T) error
}

/* Testing addTracer with httptest. POST /tracers */
func TestAddTracer(t *testing.T) {
	/* ADDING A TRACER */
	/////////////////////
	var (
		trcrStr = "blahblah"
		URL     = "http://example.com"
		method  = "GET"
		addURL  = "http://127.0.0.1:8081/tracers"
		getURL  = "http://127.0.0.1:8081/tracers/1"
	)
	jsonStr := fmt.Sprintf(`{"TracerString": "%s", "URL": "%s", "Method": "%s"}`,
		trcrStr, URL, method)

	/* Make the POST request. */
	addReq, err := http.NewRequest("POST", addURL, bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		t.Fatalf("tried to build an HTTP request, but got the following error: %+v", err)
	}
	/* ADDING A TRACER */
	/////////////////////

	/* GETING A TRACER */
	/////////////////////
	getReq, err := http.NewRequest("GET", getURL, nil)
	if err != nil {
		t.Fatalf("tried to build an HTTP request but got the following error: %+v", err)
	}
	/* GETTING A TRACER */
	/////////////////////

	/* Create a mapping of the request/test and use the server helper to execute it. */
	tests := make([]RequestTestPair, 2)
	addReqTest := RequestTestPair{addReq, addTest}
	getReqTest := RequestTestPair{getReq, getTest}
	tests[0] = addReqTest
	tests[1] = getReqTest
	serverTestHelper(tests, t)
}

/* Testing deleteTracer. DELETE /tracers/<tracer_id> */
func TestDeleteTracer(t *testing.T) {
	/* ADDING A TRACER */
	/////////////////////
	var (
		trcrStr = "blahblah"
		URL     = "http://example.com"
		method  = "GET"
		delURL  = "http://127.0.0.1:8081/tracers/1"
		addURL  = "http://127.0.0.1:8081/tracers"
		jsonStr = fmt.Sprintf(`{"TracerString": "%s", "URL": "%s", "Method": "%s"}`,
			trcrStr, URL, method)
	)

	t.Logf("sending the following data: %s", jsonStr)
	addReq, err := http.NewRequest("POST", addURL, bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		t.Fatalf("tried to build an HTTP request but got the following error: %+v", err)
	}
	/* ADDING A TRACER */
	/////////////////////

	/* DELETING A TRACER */
	/////////////////////
	delReq, err := http.NewRequest("DELETE", delURL, nil)
	if err != nil {
		t.Fatalf("tried to build an HTTP request but got the following error: %+v", err)
	}

	delTest := func(rr *httptest.ResponseRecorder, t *testing.T) error {
		/* Return variable. */
		var err error

		/* Make sure we are getting the status we are expecting. */
		if status := rr.Code; status != http.StatusAccepted {
			err = fmt.Errorf("DeleteTracer returned the wrong status code. Got %v, but wanted %v", status, http.StatusAccepted)
		} else {
			/* Since we start from a fresh database, this is the expected return from the server. */
			expected := `{"id": "1", "status": "deleted"}`
			if rr.Body.String() != expected {
				err = fmt.Errorf("deleteTracer returned the wrong body in the response. Got %s, but expected %s", rr.Body.String(), expected)
			}
		}

		return err
	}
	/* DELETING A TRACER */
	/////////////////////

	/* GETTING A TRACER */
	/////////////////////
	getReq, err := http.NewRequest("GET", delURL, nil)
	if err != nil {
		t.Fatalf("tried to build an HTTP request but got the following error: %+v", err)
	}

	getNotTest := func(rr *httptest.ResponseRecorder, t *testing.T) error {
		/* Return variable. */
		var err error

		/* Validate we are getting the status we were expecting. */
		if status := rr.Code; status != http.StatusNoContent {
			err = fmt.Errorf("GetTracer returned the wrong status code. Got %v, but wanted %v", status, http.StatusNoContent)
		} else {
			/* Validate the server did not leak any data. */
			got := tracer.Tracer{}
			json.Unmarshal([]byte(rr.Body.String()), &got)

			/* Test to make sure that after we delete the tracer, we can't query for it again. */
			if got.ID != 0 {
				err = fmt.Errorf("getTracer returned the wrong body in the response. Got %+v, but expected %+v", got, tracer.Tracer{})
			}
		}

		return err
	}
	/* GETTING A TRACER */
	/////////////////////

	/* Create a slice of the RequestTestPairs and use the server helper to execute them. */
	tests := make([]RequestTestPair, 4)
	addReqTest := RequestTestPair{addReq, addTest}
	getReqTest := RequestTestPair{getReq, getTest}
	delReqTest := RequestTestPair{delReq, delTest}
	getNotReqTest := RequestTestPair{getReq, getNotTest}
	tests[0] = addReqTest
	tests[1] = getReqTest
	tests[2] = delReqTest
	tests[3] = getNotReqTest
	serverTestHelper(tests, t)
}

/* Testing editTracer. PUT /tracers/<tracer_id>/ */
func TestEditTracer(t *testing.T) {
	/* ADDING A TRACER */
	/////////////////////
	var (
		trcrStr    = "blahblah"
		trcrStrChg = "zahzahzah"
		URL        = "http://example.com"
		urlChg     = "https://example.com"
		method     = "GET"
		methodChg  = "PUT"
		putURL     = "http://127.0.0.1:8081/tracers/1"
		addURL     = "http://127.0.0.1:8081/tracers"
		jsonStr    = fmt.Sprintf(`{"TracerString": "%s", "URL": "%s", "Method": "%s"}`, trcrStr, URL, method)
		putStr     = fmt.Sprintf(`{"TracerString": "%s", "URL": "%s", "Method": "%s"}`, trcrStrChg, urlChg, methodChg)
	)

	t.Logf("sending the following data: %s", jsonStr)
	addReq, err := http.NewRequest("POST", addURL, bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		t.Fatalf("tried to build an HTTP request, but got the following error: %+v", err)
	}
	/* ADDING A TRACER */
	/////////////////////

	/* PUTTING A TRACER */
	/////////////////////
	putReq, err := http.NewRequest("PUT", putURL, bytes.NewBuffer([]byte(putStr)))
	if err != nil {
		t.Fatalf("tried to build an HTTP request, but got the following error: %+v", err)
	}

	putTest := func(rr *httptest.ResponseRecorder, t *testing.T) error {
		/* Return variable. */
		var err error

		/* Make sure the status is what we were expecting. */
		if status := rr.Code; status != http.StatusCreated {
			err = fmt.Errorf("EditTracer returned the wrong status code. Got %v, but wanted %v", status, http.StatusCreated)
		} else {
			/* Validate the server did not leak any data. */
			got := tracer.Tracer{}
			json.Unmarshal([]byte(rr.Body.String()), &got)

			/* Test to make sure the server responds with our updated changes. */
			if got.ID != 1 {
				err = fmt.Errorf("editTracer returned the wrong body ID. Got %+v, but expected %+v", got.ID, 1)
			} else if got.URL.String != urlChg {
				err = fmt.Errorf("editTracer returned the wrong body URL. Got %+v, but expected %+v", got.URL.String, putURL)
			} else if got.Method.String != methodChg {
				err = fmt.Errorf("editTracer returned the wrong body Method. Got %+v, but expected %+v", got.Method.String, methodChg)
			} else if got.TracerString != trcrStrChg {
				err = fmt.Errorf("editTracer returned the wrong body TracerString. Got %+v, but expected %+v", got.TracerString, trcrStrChg)
			}
		}

		return err
	}
	/* PUTTING A TRACER */
	/////////////////////

	/* GETTING A TRACER */
	/////////////////////
	getReq, err := http.NewRequest("GET", putURL, bytes.NewBuffer([]byte(putStr)))
	if err != nil {
		t.Fatalf("tried to build an HTTP request, but got the following error: %+v", err)
	}

	getTest := func(rr *httptest.ResponseRecorder, t *testing.T) error {
		/* Return variable. */
		var err error

		/* Validate we got the status we were expecting */
		if status := rr.Code; status != http.StatusOK {
			err = fmt.Errorf("GetTracer returned the wrong status code. Got %v, but wanted %v", status, http.StatusNoContent)
		} else {
			/* Validate the tracer was the first tracer inserted. */
			got := tracer.Tracer{}
			json.Unmarshal([]byte(rr.Body.String()), &got)

			/* Make sure the retrieved tracer has the updated contents. */
			if got.Method.String != methodChg {
				err = fmt.Errorf("editTracer returned the wrong body Method. Got %+v, but expected %+v", got.Method.String, methodChg)
			} else if got.URL.String != urlChg {
				err = fmt.Errorf("editTracer returned the wrong body URL. Got %+v, but expected %+v", got.URL.String, putURL)
			} else if got.TracerString != trcrStrChg {
				err = fmt.Errorf("editTracer returned the wrong body TracerString. Got %+v, but expected %+v", got.TracerString, trcrStrChg)
			}
		}

		return err
	}
	/* GETTING A TRACER */
	/////////////////////

	tests := make([]RequestTestPair, 3)
	addReqTest := RequestTestPair{addReq, addTest}
	putReqTest := RequestTestPair{putReq, putTest}
	getReqTest := RequestTestPair{getReq, getTest}
	tests[0] = addReqTest
	tests[1] = putReqTest
	tests[2] = getReqTest
	serverTestHelper(tests, t)
}

/* Testing editTracer. PUT /tracers/<tracer_id>/ */
func TestAddEvnt(t *testing.T) {
	/* ADDING A TRACER */
	/////////////////////
	var (
		trcrStr    = "blahblah"
		data       = "dahdata"
		URL        = "http://example.com"
		location   = "dahlocation"
		method     = "GET"
		evntType   = "datevnttype"
		addEvntURL = "http://127.0.0.1:8081/tracers/1"
		addURL     = "http://127.0.0.1:8081/tracers"
		jsonStr    = fmt.Sprintf(`{"TracerString": "%s", "URL": "%s", "Method": "%s"}`,
			trcrStr, URL, method)
		evntStr = fmt.Sprintf(`{"Data": "%s", "Location": "%s", "EventType": "%s"}`,
			data, location, evntType)
	)

	addReq, err := http.NewRequest("POST", addURL, bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		t.Fatalf("tried to build an HTTP request, but got the following error: %+v", err)
	}
	/* ADDING A TRACER */
	/////////////////////

	/* ADDING AN EVENT */
	/////////////////////
	addEvntReq, err := http.NewRequest("POST", addEvntURL, bytes.NewBuffer([]byte(evntStr)))
	if err != nil {
		t.Fatalf("tried to build an HTTP request, but got the following error: %+v", err)
	}

	addEvntTest := func(rr *httptest.ResponseRecorder, t *testing.T) error {
		/* Return variable. */
		var err error

		/* Validate we got the status could that was expected. */
		if status := rr.Code; status != http.StatusOK {
			err = fmt.Errorf("addTracerEvent returned the wrong status code. Got %+v, but expected %+v", status, http.StatusOK)
		} else {
			/* Validate the tracer was the first tracer inserted. */
			got := tracer.TracerEvent{}
			json.Unmarshal([]byte(rr.Body.String()), &got)

			/* Validate the response gave us back the event we added. */
			if got.ID.Int64 != 1 {
				err = fmt.Errorf("addTracerEvent returned the wrong ID. Got %+v, but expected %+v", got.ID, 1)
			} else if got.Data.String != data {
				err = fmt.Errorf("addTracerEvent returned the wrong body data. Got %+v, but expected %+v", got.Data.String, data)
			} else if got.Location.String != location {
				err = fmt.Errorf("addTracerEvent returned the wrong body location. Got %+v, but expected %+v", got.Location.String, location)
			} else if got.EventType.String != evntType {
				err = fmt.Errorf("addTracerEvent returned the wrong body event type. Got %+v, but expected %+v", got.EventType.String, evntType)
			}
		}

		return err
	}
	/* ADDING AN EVENT */
	/////////////////////

	/* GETTING AN EVENT */
	/////////////////////
	getEvntReq, err := http.NewRequest("GET", addEvntURL, nil)
	if err != nil {
		t.Fatalf("tried to build an HTTP request, but got the following error: %+v", err)
	}

	getEvntTest := func(rr *httptest.ResponseRecorder, t *testing.T) error {
		/* Return variable. */
		var err error

		/* Ensure we got the expected status code. */
		if status := rr.Code; status != http.StatusOK {
			err = fmt.Errorf("getTracerEvent returned the wrong status code. Got %+v, but expected %+v", status, http.StatusOK)
		} else {
			/* Validate the tracer was the first tracer inserted. */
			got := tracer.Tracer{}
			json.Unmarshal([]byte(rr.Body.String()), &got)

			/* Make sure we have enough Hits. */
			if len(got.Hits) == 0 {
				err = fmt.Errorf("addTracerEvent didn't have any events to use. Expected 1")
			} else {
				/* Otherwise, grab the event. */
				gotEvnt := got.Hits[0]

				/* Make sure the data we inserted was also the data we received back from the database. */
				if gotEvnt.ID.Int64 != 1 {
					err = fmt.Errorf("addTracerEvent returned the wrong ID. Got %+v, but expected %+v", gotEvnt.ID, 1)
				} else if gotEvnt.Data.String != data {
					err = fmt.Errorf("addTracerEvent returned the wrong body data. Got %+v, but expected %+v", gotEvnt.Data.String, data)
				} else if gotEvnt.Location.String != location {
					err = fmt.Errorf("addTracerEvent returned the wrong body location. Got %+v, but expected %+v", gotEvnt.Location.String, location)
				} else if gotEvnt.EventType.String != evntType {
					err = fmt.Errorf("addTracerEvent returned the wrong body event type. Got %+v, but expected %+v", gotEvnt.EventType.String, evntType)
				}
			}
		}

		return err
	}
	/* GETTING AN EVENT */
	/////////////////////

	tests := make([]RequestTestPair, 3)
	addReqTest := RequestTestPair{addReq, addTest}
	addEvntReqTest := RequestTestPair{addEvntReq, addEvntTest}
	getEvntReqTest := RequestTestPair{getEvntReq, getEvntTest}
	tests[0] = addReqTest
	tests[1] = addEvntReqTest
	tests[2] = getEvntReqTest
}

/* Delete any existing database */
func deleteDatabase(t *testing.T) {
	/* Find the path of this package. */
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Errorf("no caller information, therefore, can't find the database.")
	}
	/* Should be something like $GOPATH/src/xxterminator-plugin/xxtermiate/TracerServer/store/tracer-db.db */
	db := path.Dir(filename) + string(filepath.Separator) + "store" + string(filepath.Separator) + "tracer-db.db"
	/* If the database exists, remove it. It will affect the test. */
	if _, err := os.Stat(db); !os.IsNotExist(err) {
		err := os.Remove(db)
		if err != nil {
			t.Errorf("wasn't able to delete the database at: %s", db)
		}
	}
}

/* A function that takes a slice of RequestTestPairs. Each pair has a request and a
 * test function. Each request is submitted and the corresponding test is run on the
 * response. Tests are run sequence and each test is used to validate the response.
 * This function can be used to chain request/response tests together, for example
 * to test if a particular resource has been deleted or created. An error in the middle
 * of a chain of requests will break since it is likely the following tests will also
 * break. */
func serverTestHelper(tests []RequestTestPair, t *testing.T) {
	/* Delete any existing database entries */
	/* TODO: make a testing database. So that we don't delete a bunch of data when we run tests. */
	deleteDatabase(t)
	/* Open the database because the init method from main.go won't trigger. */
	openDatabase()
	_, handler := configureServer()

	for _, pair := range tests {
		/* For each request/test combo:
		* 1.) send the request
		* 2.) collect the response
		* 3.) run the response on the test method
		* 4.) break on error */
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, pair.Request)
		err := pair.Test(rr, t)
		if err != nil {
			t.Errorf("the following request, %+v, did not pass it's test: %+v", pair.Request, err)
			break
		}
	}
}

/* Commonly used GET request test. */
func getTest(rr *httptest.ResponseRecorder, t *testing.T) error {
	/* Return variable. */
	var err error

	if status := rr.Code; status != http.StatusOK {
		err = fmt.Errorf("GetTracer returned the wrong status code. Got %v, but wanted %v", status, http.StatusNoContent)
	} else {
		/* Validate the tracer was the first tracer inserted. */
		got := tracer.Tracer{}
		json.Unmarshal([]byte(rr.Body.String()), &got)

		/* This test only looks for the tracer just added. The ID should be 1. */
		if got.ID != 1 {
			err = fmt.Errorf("getTracer returned the wrong body in the response. Got ID of %+v, but expected %+v", got.ID, 1)
		}
	}

	/* Return nil to indicate no problems. */
	return err
}

/* Commonly used POST request test. */
func addTest(rr *httptest.ResponseRecorder, t *testing.T) error {
	/* Return variable. */
	var err error

	/* Make sure the status code is 200. */
	if status := rr.Code; status != http.StatusOK {
		err = fmt.Errorf("AddTracer returned the wrong status code: got %v, but wanted %v", status, http.StatusOK)
	} else {
		/* Make sure the body is a valid JSON object. */
		got := tracer.Tracer{}
		json.Unmarshal([]byte(rr.Body.String()), &got)

		/* Sanity checks to make sure the added tracer wasn't empty. */
		if got.ID != 1 {
			err = fmt.Errorf("The inserted tracer has the wrong ID. Expected 1, got: %d", got.ID)
		} else if got.URL.String == "" {
			err = fmt.Errorf("The inserted tracer has the wrong URL. Got nothing, but expected: %s", got.URL.String)
		} else if got.Method.String == "" {
			err = fmt.Errorf("The inserted tracer has the wrong Method. Got: %s", got.Method.String)
		}
	}

	return err
}
