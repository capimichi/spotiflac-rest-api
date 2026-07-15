# Integrazione OpenAPI / Swagger e Refactoring Controller - Piano di Implementazione

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Completare il refactoring architetturale di `spotiflac-rest-api` separando lo strato controller e integrare Swagger UI (OpenAPI) tramite Swaggo per la documentazione interattiva degli endpoint.

**Architecture:** Estrarremo gli endpoint HTTP da `main.go` in controller specializzati (`internal/controllers/`) e un router centralizzato. Aggiungeremo i commenti dichiarativi nei controller e useremo Swaggo per generare e servire la documentazione Swagger UI.

**Tech Stack:** Go 1.26, Gin-Gonic, Swaggo (github.com/swaggo/gin-swagger, github.com/swaggo/files, swag CLI).

## Global Constraints
* Go version: >= 1.26
* Gin-Gonic version: 1.12.0
* Rotta Swagger UI: `/swagger/*any`
* Tutti i test Go devono essere eseguiti con `go test ./...` e devono passare senza errori.

---

### Task 1: Aggiunta Dipendenze e Scaffolding Iniziale

**Files:**
* Modify: `/Users/michele/PycharmProjects/spotiflac-rest-api/go.mod`
* Test: `go.mod` validation and swag CLI verify.

**Interfaces:**
* Consumes: Nessuna
* Produces: Moduli Swaggo importati in `go.mod`.

- [ ] **Step 1: Modificare il file `go.mod` per includere i pacchetti di Swaggo**
  Aggiungere le dipendenze per `gin-swagger` e `files` nel blocco `require`.
  
  ```go
  // Inserire nel blocco require in go.mod:
  github.com/swaggo/files v1.0.1
  github.com/swaggo/gin-swagger v1.6.0
  github.com/swaggo/swag v1.16.4
  ```

- [ ] **Step 2: Eseguire la sincronizzazione dei moduli Go**
  Run: `go mod tidy`
  Expected: Comando eseguito con successo, dipendenze scaricate e aggiornate in `go.sum`.

- [ ] **Step 3: Verificare l'installazione locale di swag CLI**
  Eseguiamo l'installazione globale dello strumento `swag` per assicurarci che sia disponibile per generare i documenti.
  Run: `go install github.com/swaggo/swag/cmd/swag@latest`
  Expected: Esecuzione riuscita senza errori di compilazione.

- [ ] **Step 4: Controllare la versione di swag CLI**
  Run: `swag --version`
  Expected: Mostra la versione correntemente installata (es: `swag version v1.16.x`).

---

### Task 2: Health Controller & Router Setup con Test

**Files:**
* Create: `/Users/michele/PycharmProjects/spotiflac-rest-api/internal/controllers/health_controller.go`
* Create: `/Users/michele/PycharmProjects/spotiflac-rest-api/internal/controllers/router.go`
* Create: `/Users/michele/PycharmProjects/spotiflac-rest-api/internal/controllers/router_test.go`

**Interfaces:**
* Consumes: Swaggo dependencies.
* Produces: `SetupRouter(healthCtrl *HealthController, downloadCtrl *DownloadController) *gin.Engine`

- [ ] **Step 1: Creare l'HealthController con annotazioni Swagger**
  Scrivere il file `/Users/michele/PycharmProjects/spotiflac-rest-api/internal/controllers/health_controller.go`:
  
  ```go
  package controllers

  import (
  	"net/http"
  	"time"

  	"github.com/gin-gonic/gin"
  )

  type HealthController struct{}

  func NewHealthController() *HealthController {
  	return &HealthController{}
  }

  // HealthCheck godoc
  // @Summary      Health Check
  // @Description  Verifica se il server API è attivo e funzionante
  // @Tags         System
  // @Produce      json
  // @Success      200  {object}  map[string]interface{} "status ok"
  // @Router       /api/health [get]
  func (ctrl *HealthController) HealthCheck(c *gin.Context) {
  	c.JSON(http.StatusOK, gin.H{
  		"status":      "ok",
  		"time":        time.Now().Format(time.RFC3339),
  		"environment": gin.Mode(),
  	})
  }
  ```

- [ ] **Step 2: Creare il file Router e configurare Swagger**
  Scrivere il file `/Users/michele/PycharmProjects/spotiflac-rest-api/internal/controllers/router.go` che gestisce il routing e include la Swagger UI. Per ora importiamo fittiziamente la cartella `docs` (la genereremo nei prossimi task).
  
  ```go
  package controllers

  import (
  	"github.com/gin-gonic/gin"
  	swaggerFiles "github.com/swaggo/files"
  	ginSwagger "github.com/swaggo/gin-swagger"
  )

  func SetupRouter(healthCtrl *HealthController) *gin.Engine {
  	r := gin.Default()

  	// CORS Middleware
  	r.Use(func(c *gin.Context) {
  		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
  		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
  		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
  		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

  		if c.Request.Method == "OPTIONS" {
  			c.AbortWithStatus(204)
  			return
  		}
  		c.Next()
  	})

  	r.GET("/api/health", healthCtrl.HealthCheck)
  	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

  	return r
  }
  ```

- [ ] **Step 3: Scrivere il test per verificare l'Health check e la rotta Swagger**
  Scrivere il file `/Users/michele/PycharmProjects/spotiflac-rest-api/internal/controllers/router_test.go`:
  
  ```go
  package controllers

  import (
  	"net/http"
  	"net/http/httptest"
  	"testing"
  )

  func TestRouterEndpoints(t *testing.T) {
  	healthCtrl := NewHealthController()
  	router := SetupRouter(healthCtrl)

  	t.Run("GET /api/health", func(t *testing.T) {
  		w := httptest.NewRecorder()
  		req, _ := http.NewRequest("GET", "/api/health", nil)
  		router.ServeHTTP(w, req)

  		if w.Code != http.StatusOK {
  			t.Errorf("Expected status 200, got %d", w.Code)
  		}
  	})

  	t.Run("GET /swagger/index.html redirects or serves 200", func(t *testing.T) {
  		w := httptest.NewRecorder()
  		req, _ := http.NewRequest("GET", "/swagger/index.html", nil)
  		router.ServeHTTP(w, req)

  		// Should serve status 200 or redirect to index.html (which is served as 200)
  		if w.Code != http.StatusOK && w.Code != http.StatusMovedPermanently {
  			t.Errorf("Expected status 200 or 301, got %d", w.Code)
  		}
  	})
  }
  ```

- [ ] **Step 4: Eseguire il test del router**
  Run: `go test -v ./internal/controllers`
  Expected: PASS per entrambi i test.

---

### Task 3: Download Controller con Test

**Files:**
* Create: `/Users/michele/PycharmProjects/spotiflac-rest-api/internal/controllers/download_controller.go`
* Create: `/Users/michele/PycharmProjects/spotiflac-rest-api/internal/controllers/download_controller_test.go`
* Modify: `/Users/michele/PycharmProjects/spotiflac-rest-api/internal/controllers/router.go`

**Interfaces:**
* Consumes: `internal/services.DownloadService`, `internal/repositories.TaskRepository`
* Produces: `DownloadController` struct e i rispettivi handler integrati in `/Users/michele/PycharmProjects/spotiflac-rest-api/internal/controllers/router.go`.

- [ ] **Step 1: Creare il DownloadController con annotazioni Swagger**
  Scrivere il file `/Users/michele/PycharmProjects/spotiflac-rest-api/internal/controllers/download_controller.go` collegato al `DownloadService` ed alla `TaskRepository`:
  
  ```go
  package controllers

  import (
  	"net/http"

  	"github.com/capimichi/spotiflac-rest-api/internal/models"
  	"github.com/capimichi/spotiflac-rest-api/internal/repositories"
  	"github.com/capimichi/spotiflac-rest-api/internal/services"
  	"github.com/gin-gonic/gin"
  )

  type DownloadController struct {
  	service *services.DownloadService
  	repo    repositories.TaskRepository
  }

  func NewDownloadController(service *services.DownloadService, repo repositories.TaskRepository) *DownloadController {
  	return &DownloadController{
  		service: service,
  		repo:    repo,
  	}
  }

  // DownloadAsync godoc
  // @Summary      Avvia download asincrono
  // @Description  Avvia un task di download in background e restituisce immediatamente l'ID del task
  // @Tags         Download
  // @Accept       json
  // @Produce      json
  // @Param        request  body      models.DownloadRequest  true  "Parametri di download"
  // @Success      202      {object}  map[string]interface{} "Contiene task_id, status e message"
  // @Failure      400      {object}  map[string]string      "Richiesta malformata"
  // @Router       /api/download [post]
  func (ctrl *DownloadController) DownloadAsync(c *gin.Context) {
  	var req models.DownloadRequest
  	if err := c.ShouldBindJSON(&req); err != nil {
  		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
  		return
  	}

  	// Defaults
  	if req.Service == "" {
  		req.Service = "qobuz"
  	}
  	if req.Quality == "" {
  		if req.Service == "qobuz" {
  			req.Quality = "6"
  		} else {
  			req.Quality = "LOSSLESS"
  		}
  	}
  	if req.OutputDir == "" {
  		req.OutputDir = "./downloads"
  	}
  	if req.FilenameFormat == "" {
  		req.FilenameFormat = "{artist} - {title}"
  	}

  	task, err := ctrl.service.DownloadAsync(req)
  	if err != nil {
  		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
  		return
  	}

  	c.JSON(http.StatusAccepted, gin.H{
  		"task_id": task.ID,
  		"status":  task.Status,
  		"message": "Download task started successfully",
  	})
  }

  // DownloadSync godoc
  // @Summary      Avvia download sincrono
  // @Description  Avvia il download e si blocca finché tutti i file non sono stati scaricati e taggati
  // @Tags         Download
  // @Accept       json
  // @Produce      json
  // @Param        request  body      models.DownloadRequest  true  "Parametri di download"
  // @Success      200      {object}  map[string]interface{} "Contiene status e l'elenco dei file"
  // @Failure      400      {object}  map[string]string      "Richiesta malformata"
  // @Failure      500      {object}  map[string]string      "Errore durante il download"
  // @Router       /api/download/sync [post]
  func (ctrl *DownloadController) DownloadSync(c *gin.Context) {
  	var req models.DownloadRequest
  	if err := c.ShouldBindJSON(&req); err != nil {
  		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
  		return
  	}

  	// Defaults
  	if req.Service == "" {
  		req.Service = "qobuz"
  	}
  	if req.Quality == "" {
  		if req.Service == "qobuz" {
  			req.Quality = "6"
  		} else {
  			req.Quality = "LOSSLESS"
  		}
  	}
  	if req.OutputDir == "" {
  		req.OutputDir = "./downloads"
  	}
  	if req.FilenameFormat == "" {
  		req.FilenameFormat = "{artist} - {title}"
  	}

  	files, err := ctrl.service.DownloadSync(req)
  	if err != nil {
  		c.JSON(http.StatusInternalServerError, gin.H{
  			"status": "failed",
  			"error":  err.Error(),
  		})
  		return
  	}

  	c.JSON(http.StatusOK, gin.H{
  		"status": "completed",
  		"files":  files,
  	})
  }

  // GetStatus godoc
  // @Summary      Traccia lo stato del task
  // @Description  Restituisce lo stato attuale di un download asincrono specificando l'ID del task
  // @Tags         Download
  // @Produce      json
  // @Param        id   path      string  true  "ID del task"
  // @Success      200  {object}  models.Task
  // @Failure      404  {object}  map[string]string "Task non trovato"
  // @Router       /api/status/{id} [get]
  func (ctrl *DownloadController) GetStatus(c *gin.Context) {
  	id := c.Param("id")
  	task, exists, err := ctrl.repo.Get(id)
  	if err != nil {
  		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
  		return
  	}
  	if !exists {
  		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
  		return
  	}
  	c.JSON(http.StatusOK, task)
  }
  ```

- [ ] **Step 2: Aggiornare il Router per includere il DownloadController**
  Aggiornare `/Users/michele/PycharmProjects/spotiflac-rest-api/internal/controllers/router.go` in modo che consumi il `DownloadController`.
  
  ```go
  package controllers

  import (
  	"github.com/gin-gonic/gin"
  	swaggerFiles "github.com/swaggo/files"
  	ginSwagger "github.com/swaggo/gin-swagger"
  )

  func SetupRouter(healthCtrl *HealthController, downloadCtrl *DownloadController) *gin.Engine {
  	r := gin.Default()

  	// CORS Middleware
  	r.Use(func(c *gin.Context) {
  		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
  		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
  		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
  		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

  		if c.Request.Method == "OPTIONS" {
  			c.AbortWithStatus(204)
  			return
  		}
  		c.Next()
  	})

  	r.GET("/api/health", healthCtrl.HealthCheck)
  	r.GET("/api/status/:id", downloadCtrl.GetStatus)
  	r.POST("/api/download", downloadCtrl.DownloadAsync)
  	r.POST("/api/download/sync", downloadCtrl.DownloadSync)

  	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

  	return r
  }
  ```

- [ ] **Step 3: Aggiungere i test per il DownloadController**
  Scrivere `/Users/michele/PycharmProjects/spotiflac-rest-api/internal/controllers/download_controller_test.go` utilizzando un mock del repository e servizi o testando i casi limite HTTP.
  
  ```go
  package controllers

  import (
  	"bytes"
  	"encoding/json"
  	"net/http"
  	"net/http/httptest"
  	"testing"

  	"github.com/capimichi/spotiflac-rest-api/internal/models"
  	"github.com/capimichi/spotiflac-rest-api/internal/repositories"
  	"github.com/capimichi/spotiflac-rest-api/internal/services"
  )

  func TestDownloadControllerEndpoints(t *testing.T) {
  	repo := repositories.NewInMemoryTaskRepository()
  	svc := services.NewDownloadService(repo)
  	healthCtrl := NewHealthController()
  	downloadCtrl := NewDownloadController(svc, repo)
  	router := SetupRouter(healthCtrl, downloadCtrl)

  	t.Run("POST /api/download bad request", func(t *testing.T) {
  		w := httptest.NewRecorder()
  		req, _ := http.NewRequest("POST", "/api/download", bytes.NewBufferString("{invalid-json}"))
  		router.ServeHTTP(w, req)

  		if w.Code != http.StatusBadRequest {
  			t.Errorf("Expected status 400, got %d", w.Code)
  		}
  	})

  	t.Run("GET /api/status/:id not found", func(t *testing.T) {
  		w := httptest.NewRecorder()
  		req, _ := http.NewRequest("GET", "/api/status/non-existent-id", nil)
  		router.ServeHTTP(w, req)

  		if w.Code != http.StatusNotFound {
  			t.Errorf("Expected status 404, got %d", w.Code)
  		}
  	})

  	t.Run("POST /api/download valid async request structure", func(t *testing.T) {
  		w := httptest.NewRecorder()
  		reqData := models.DownloadRequest{
  			URL: "https://open.spotify.com/track/4jVnKscdFqT0k2d4Fp4e1F",
  		}
  		jsonVal, _ := json.Marshal(reqData)
  		req, _ := http.NewRequest("POST", "/api/download", bytes.NewBuffer(jsonVal))
  		router.ServeHTTP(w, req)

  		// Svc will attempt actual download or return an error if Spotify connection fails/mocked.
  		// In this test, we expect either a 202 (StatusAccepted) if it initiates download, or 500 if backend download fails.
  		// Either way, HTTP parsing and middleware are verified.
  		if w.Code != http.StatusAccepted && w.Code != http.StatusInternalServerError {
  			t.Errorf("Expected status 202 or 500, got %d", w.Code)
  		}
  	})
  }
  ```

- [ ] **Step 4: Eseguire tutti i test dei controller**
  Run: `go test -v ./internal/controllers`
  Expected: Tutte le prove dei controller passano correttamente.

---

### Task 4: Main Entrypoint & Swagger Code Generation

**Files:**
* Create: `/Users/michele/PycharmProjects/spotiflac-rest-api/cmd/server/main.go`
* Create: `/Users/michele/PycharmProjects/spotiflac-rest-api/internal/controllers/router_swagger.go`
* Modify: `/Users/michele/PycharmProjects/spotiflac-rest-api/main.go` (sarà cancellato o trasformato in redirect/facade)

**Interfaces:**
* Consumes: I controller e servizi scritti precedentemente.
* Produces: Un'applicazione web intera e funzionante compilata con Swagger embedded.

- [ ] **Step 1: Aggiungere i metadati Swagger a livello globale**
  Per evitare conflitti di importazione circolare con il pacchetto `docs`, creiamo un file `router_swagger.go` in `internal/controllers` che importerà `_ "github.com/capimichi/spotiflac-rest-api/docs"`.
  Scrivere il file `/Users/michele/PycharmProjects/spotiflac-rest-api/internal/controllers/router_swagger.go`:
  
  ```go
  package controllers

  // Import anonimo necessario per inizializzare Swagger UI
  import (
  	_ "github.com/capimichi/spotiflac-rest-api/docs"
  )
  ```

- [ ] **Step 2: Scrivere il nuovo entrypoint di produzione**
  Scrivere il file `/Users/michele/PycharmProjects/spotiflac-rest-api/cmd/server/main.go` contenente i metadati globali dell'API ed il setup di persistenza/servizio/routing:
  
  ```go
  package main

  import (
  	"fmt"
  	"os"

  	"github.com/capimichi/spotiflac-rest-api/internal/controllers"
  	"github.com/capimichi/spotiflac-rest-api/internal/repositories"
  	"github.com/capimichi/spotiflac-rest-api/internal/services"
  	"github.com/gin-gonic/gin"
  )

  // @title          SpotiFLAC REST API
  // @version        1.0
  // @description    A lightweight REST API wrapper for the SpotiFLAC backend.
  // @host           localhost:8080
  // @BasePath       /
  func main() {
  	if os.Getenv("GIN_MODE") == "release" {
  		gin.SetMode(gin.ReleaseMode)
  	}

  	repo := repositories.NewInMemoryTaskRepository()
  	svc := services.NewDownloadService(repo)

  	healthCtrl := controllers.NewHealthController()
  	downloadCtrl := controllers.NewDownloadController(svc, repo)

  	r := controllers.SetupRouter(healthCtrl, downloadCtrl)

  	port := os.Getenv("PORT")
  	if port == "" {
  		port = "8080"
  	}

  	fmt.Printf("SpotiFLAC REST API Server listening on port %s...\n", port)
  	if err := r.Run(":" + port); err != nil {
  		fmt.Printf("Failed to run server: %v\n", err)
  	}
  }
  ```

- [ ] **Step 3: Eseguire swag init per autogenerare la cartella `docs/`**
  Run: `swag init -g cmd/server/main.go -o docs/`
  Expected: Creazione con successo dei file `docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml`.

- [ ] **Step 4: Compilare e verificare l'applicazione principale**
  Run: `go build -o spotiflac-server cmd/server/main.go`
  Expected: Compilazione riuscita senza errori.

- [ ] **Step 5: Pulire o sostituire il vecchio file `main.go`**
  Per evitare conflitti con due pacchetti `main` nella radice ed in `cmd/`, rimuoviamo la logica originaria di `/Users/michele/PycharmProjects/spotiflac-rest-api/main.go` e la sostituiamo con un semplice forwarding o lo eliminiamo del tutto se non serve più.
  Modifichiamo `/Users/michele/PycharmProjects/spotiflac-rest-api/main.go` per essere un semplice wrapper del server principale:
  
  ```go
  package main

  import (
  	"fmt"
  	"os"
  	"os/exec"
  )

  func main() {
  	fmt.Println("Avvio del server tramite cmd/server/main.go...")
  	cmd := exec.Command("go", "run", "cmd/server/main.go")
  	cmd.Stdout = os.Stdout
  	cmd.Stderr = os.Stderr
  	if err := cmd.Run(); err != nil {
  		fmt.Printf("Errore durante l'esecuzione del server: %v\n", err)
  		os.Exit(1)
  	}
  }
  ```

- [ ] **Step 6: Eseguire tutti i test e fare la verifica finale**
  Run: `go test -v ./...`
  Expected: PASS per tutti i test dell'intero progetto.
