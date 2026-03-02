# fuegoapi-controllers

> REST API controllers scaffolded with Fuego (GET, POST, PUT, PATCH, DELETE)

## app/adapter/in/fuegoapi/get.go

```go
package fuegoapi

import (
	"archetype/app/shared/infrastructure/httpserver"

	"github.com/Ignaciojeria/ioc"
	fuegofw "github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
)

var _ = ioc.Register(NewTemplateGet)

// NewTemplateGet registers a sample GET endpoint.
// It uses *fuego.Server as a dependency, ensuring the server is allocated before this runs.
func NewTemplateGet(s *httpserver.Server) {
	fuegofw.Get(s.Manager, "/hello",
		func(c fuegofw.ContextNoBody) (string, error) {
			return "Hello from Einar IoC Fuego Template!", nil
		},
		option.Summary("newTemplateGet"),
	)
}
```

---

## app/adapter/in/fuegoapi/post.go

```go
package fuegoapi

import (
	"archetype/app/shared/infrastructure/httpserver"

	"github.com/Ignaciojeria/ioc"
	fuegofw "github.com/go-fuego/fuego"
)

var _ = ioc.Register(NewTemplatePost)

type TemplatePostRequest struct {
	Message string `json:"message"`
}

type TemplatePostResponse struct {
	Status string `json:"status"`
}

// NewTemplatePost registers a sample POST endpoint.
func NewTemplatePost(s *httpserver.Server) {
	fuegofw.Post(s.Manager, "/hello",
		func(c fuegofw.ContextWithBody[TemplatePostRequest]) (TemplatePostResponse, error) {
			body, err := c.Body()
			if err != nil {
				return TemplatePostResponse{}, err
			}
			return TemplatePostResponse{
				Status: body.Message + " received",
			}, nil
		})
}
```

---

## app/adapter/in/fuegoapi/put.go

```go
package fuegoapi

import (
	"archetype/app/shared/infrastructure/httpserver"

	"github.com/Ignaciojeria/ioc"
	fuegofw "github.com/go-fuego/fuego"
)

var _ = ioc.Register(NewTemplatePut)

type TemplatePutRequest struct {
	Message string `json:"message"`
}

type TemplatePutResponse struct {
	Status string `json:"status"`
}

// NewTemplatePut registers a sample PUT endpoint.
func NewTemplatePut(s *httpserver.Server) {
	fuegofw.Put(s.Manager, "/hello",
		func(c fuegofw.ContextWithBody[TemplatePutRequest]) (TemplatePutResponse, error) {
			body, err := c.Body()
			if err != nil {
				return TemplatePutResponse{}, err
			}
			return TemplatePutResponse{
				Status: body.Message + " updated",
			}, nil
		})
}
```

---

## app/adapter/in/fuegoapi/patch.go

```go
package fuegoapi

import (
	"archetype/app/shared/infrastructure/httpserver"

	"github.com/Ignaciojeria/ioc"
	fuegofw "github.com/go-fuego/fuego"
)

var _ = ioc.Register(NewTemplatePatch)

type TemplatePatchRequest struct {
	Message string `json:"message"`
}

type TemplatePatchResponse struct {
	Status string `json:"status"`
}

// NewTemplatePatch registers a sample PATCH endpoint.
func NewTemplatePatch(s *httpserver.Server) {
	fuegofw.Patch(s.Manager, "/hello",
		func(c fuegofw.ContextWithBody[TemplatePatchRequest]) (TemplatePatchResponse, error) {
			body, err := c.Body()
			if err != nil {
				return TemplatePatchResponse{}, err
			}
			return TemplatePatchResponse{
				Status: body.Message + " patched",
			}, nil
		})
}
```

---

## app/adapter/in/fuegoapi/delete.go

```go
package fuegoapi

import (
	"archetype/app/shared/infrastructure/httpserver"

	"github.com/Ignaciojeria/ioc"
	fuegofw "github.com/go-fuego/fuego"
)

var _ = ioc.Register(NewTemplateDelete)

// NewTemplateDelete registers a sample DELETE endpoint.
func NewTemplateDelete(s *httpserver.Server) {
	fuegofw.Delete(s.Manager, "/hello/{id}",
		func(c fuegofw.ContextNoBody) (string, error) {
			id := c.PathParam("id")
			return id + " deleted", nil
		})
}
```

---

## Unit tests

When creating a new component, generate tests following this pattern:

### app/adapter/in/fuegoapi/get_test.go

```go
package fuegoapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"archetype/app/shared/configuration"
	"archetype/app/shared/infrastructure/httpserver"
)

func TestNewTemplateGet(t *testing.T) {
	conf := configuration.Conf{
		PORT:         "8083",
		PROJECT_NAME: "test",
		VERSION:      "v1",
	}

	server, err := httpserver.NewServer(conf, nil)
	if err != nil {
		t.Fatalf("unexpected error creating server: %v", err)
	}

	// Register the endpoint using our component logic
	NewTemplateGet(server)

	// Create a simulated test request targeting the endpoint
	req, err := http.NewRequest(http.MethodGet, "/hello", nil)
	if err != nil {
		t.Fatalf("unexpected error building request: %v", err)
	}

	// Record the response directly from the Fuego router's underlying Mux
	recorder := httptest.NewRecorder()
	server.Manager.Mux.ServeHTTP(recorder, req)

	// Assert the status code
	if recorder.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", recorder.Code, http.StatusOK)
	}

	// Fuego string handling directly outputs the string content
	expectedBody := "Hello from Einar IoC Fuego Template!"
	if recorder.Body.String() != expectedBody {
		t.Errorf("handler returned unexpected body: got %s want %s", recorder.Body.String(), expectedBody)
	}
}
```

---

### app/adapter/in/fuegoapi/post_test.go

```go
package fuegoapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"archetype/app/shared/configuration"
	"archetype/app/shared/infrastructure/httpserver"
)

func TestNewTemplatePost(t *testing.T) {
	conf := configuration.Conf{
		PORT:         "8083",
		PROJECT_NAME: "test",
		VERSION:      "v1",
	}

	server, err := httpserver.NewServer(conf, nil)
	if err != nil {
		t.Fatalf("unexpected error creating server: %v", err)
	}

	NewTemplatePost(server)

	reqBody, _ := json.Marshal(TemplatePostRequest{Message: "Einar"})
	req, err := http.NewRequest(http.MethodPost, "/hello", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		t.Fatalf("unexpected error building request: %v", err)
	}

	recorder := httptest.NewRecorder()
	server.Manager.Mux.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", recorder.Code, http.StatusOK)
	}

	expectedBody := `{"status":"Einar received"}` + "\n"
	if recorder.Body.String() != expectedBody {
		t.Errorf("handler returned unexpected body: got %s want %s", recorder.Body.String(), expectedBody)
	}
}

func TestNewTemplatePost_InvalidBody(t *testing.T) {
	conf := configuration.Conf{
		PORT:         "8084",
		PROJECT_NAME: "test-err",
		VERSION:      "v1",
	}

	server, err := httpserver.NewServer(conf, nil)
	if err != nil {
		t.Fatalf("unexpected error creating server: %v", err)
	}

	NewTemplatePost(server)

	// Send invalid JSON body syntax
	req, err := http.NewRequest(http.MethodPost, "/hello", bytes.NewBuffer([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		t.Fatalf("unexpected error building request: %v", err)
	}

	recorder := httptest.NewRecorder()
	server.Manager.Mux.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("handler should return bad request: got %v want %v", recorder.Code, http.StatusBadRequest)
	}
}
```

---

### app/adapter/in/fuegoapi/put_test.go

```go
package fuegoapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"archetype/app/shared/configuration"
	"archetype/app/shared/infrastructure/httpserver"
)

func TestNewTemplatePut(t *testing.T) {
	conf := configuration.Conf{
		PORT:         "8085",
		PROJECT_NAME: "test",
		VERSION:      "v1",
	}

	server, err := httpserver.NewServer(conf, nil)
	if err != nil {
		t.Fatalf("unexpected error creating server: %v", err)
	}

	NewTemplatePut(server)

	reqBody, _ := json.Marshal(TemplatePutRequest{Message: "Einar"})
	req, err := http.NewRequest(http.MethodPut, "/hello", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		t.Fatalf("unexpected error building request: %v", err)
	}

	recorder := httptest.NewRecorder()
	server.Manager.Mux.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", recorder.Code, http.StatusOK)
	}

	expectedBody := `{"status":"Einar updated"}` + "\n"
	if recorder.Body.String() != expectedBody {
		t.Errorf("handler returned unexpected body: got %s want %s", recorder.Body.String(), expectedBody)
	}
}

func TestNewTemplatePut_InvalidBody(t *testing.T) {
	conf := configuration.Conf{
		PORT:         "8086",
		PROJECT_NAME: "test-err",
		VERSION:      "v1",
	}

	server, err := httpserver.NewServer(conf, nil)
	if err != nil {
		t.Fatalf("unexpected error creating server: %v", err)
	}

	NewTemplatePut(server)

	req, err := http.NewRequest(http.MethodPut, "/hello", bytes.NewBuffer([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		t.Fatalf("unexpected error building request: %v", err)
	}

	recorder := httptest.NewRecorder()
	server.Manager.Mux.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("handler should return bad request: got %v want %v", recorder.Code, http.StatusBadRequest)
	}
}
```

---

### app/adapter/in/fuegoapi/patch_test.go

```go
package fuegoapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"archetype/app/shared/configuration"
	"archetype/app/shared/infrastructure/httpserver"
)

func TestNewTemplatePatch(t *testing.T) {
	conf := configuration.Conf{
		PORT:         "8087",
		PROJECT_NAME: "test",
		VERSION:      "v1",
	}

	server, err := httpserver.NewServer(conf, nil)
	if err != nil {
		t.Fatalf("unexpected error creating server: %v", err)
	}

	NewTemplatePatch(server)

	reqBody, _ := json.Marshal(TemplatePatchRequest{Message: "Einar"})
	req, err := http.NewRequest(http.MethodPatch, "/hello", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		t.Fatalf("unexpected error building request: %v", err)
	}

	recorder := httptest.NewRecorder()
	server.Manager.Mux.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", recorder.Code, http.StatusOK)
	}

	expectedBody := `{"status":"Einar patched"}` + "\n"
	if recorder.Body.String() != expectedBody {
		t.Errorf("handler returned unexpected body: got %s want %s", recorder.Body.String(), expectedBody)
	}
}

func TestNewTemplatePatch_InvalidBody(t *testing.T) {
	conf := configuration.Conf{
		PORT:         "8088",
		PROJECT_NAME: "test-err",
		VERSION:      "v1",
	}

	server, err := httpserver.NewServer(conf, nil)
	if err != nil {
		t.Fatalf("unexpected error creating server: %v", err)
	}

	NewTemplatePatch(server)

	req, err := http.NewRequest(http.MethodPatch, "/hello", bytes.NewBuffer([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		t.Fatalf("unexpected error building request: %v", err)
	}

	recorder := httptest.NewRecorder()
	server.Manager.Mux.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("handler should return bad request: got %v want %v", recorder.Code, http.StatusBadRequest)
	}
}
```

---

### app/adapter/in/fuegoapi/delete_test.go

```go
package fuegoapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"archetype/app/shared/configuration"
	"archetype/app/shared/infrastructure/httpserver"
)

func TestNewTemplateDelete(t *testing.T) {
	conf := configuration.Conf{
		PORT:         "8089",
		PROJECT_NAME: "test",
		VERSION:      "v1",
	}

	server, err := httpserver.NewServer(conf, nil)
	if err != nil {
		t.Fatalf("unexpected error creating server: %v", err)
	}

	NewTemplateDelete(server)

	req, err := http.NewRequest(http.MethodDelete, "/hello/123", nil)
	if err != nil {
		t.Fatalf("unexpected error building request: %v", err)
	}

	recorder := httptest.NewRecorder()
	server.Manager.Mux.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", recorder.Code, http.StatusOK)
	}

	expectedBody := "123 deleted"
	if recorder.Body.String() != expectedBody {
		t.Errorf("handler returned unexpected body: got %s want %s", recorder.Body.String(), expectedBody)
	}
}
```
