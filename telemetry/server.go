package telemetry

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type Server struct {
	Router *mux.Router
	DB     *sql.DB
}

func NewServer(db *sql.DB) *Server {
	server := &Server{
		Router: mux.NewRouter().StrictSlash(true),
		DB:     db,
	}

	server.Router.HandleFunc("/protocol-stats", server.createProtocolStats).Methods("POST")
	server.Router.HandleFunc("/received-messages", server.createReceivedMessages).Methods("POST")
	server.Router.HandleFunc("/waku-messages", server.createWakuMessages).Methods("POST")
	server.Router.HandleFunc("/received-envelope", server.createReceivedEnvelope).Methods("POST")
	server.Router.HandleFunc("/update-envelope", server.updateEnvelope).Methods("POST")
	server.Router.HandleFunc("/health", handleHealthCheck).Methods("GET")

	return server
}

func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

func (s *Server) createReceivedMessages(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var receivedMessages []ReceivedMessage
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&receivedMessages); err != nil {
		log.Println(err)

		err := respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		if err != nil {
			log.Println(err)
		}
		return
	}
	defer r.Body.Close()

	var ids []int
	for _, receivedMessage := range receivedMessages {
		if err := receivedMessage.put(s.DB); err != nil {
			log.Println("could not save message", err, receivedMessage)
			continue
		}
		ids = append(ids, receivedMessage.ID)
	}

	if len(ids) != len(receivedMessages) {
		err := respondWithError(w, http.StatusInternalServerError, "Could not save all record")
		if err != nil {
			log.Println(err)
		}
		return
	}

	err := respondWithJSON(w, http.StatusCreated, receivedMessages)
	if err != nil {
		log.Println(err)
	}

	log.Printf(
		"%s\t%s\t%s",
		r.Method,
		r.RequestURI,
		time.Since(start),
	)
}

func (s *Server) createWakuMessages(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var wakuMessages []WakuMessage
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&wakuMessages); err != nil {
		log.Println(err)

		err := respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		if err != nil {
			log.Println(err)
		}
		return
	}
	defer r.Body.Close()

	var ids []int
	for _, wakuMessage := range wakuMessages {
		if err := wakuMessage.put(s.DB); err != nil {
			log.Println("could not save message", err, wakuMessage)
			continue
		}
		ids = append(ids, wakuMessage.ID)
	}

	if len(ids) != len(wakuMessages) {
		err := respondWithError(w, http.StatusInternalServerError, "Could not save all record")
		if err != nil {
			log.Println(err)
		}
		return
	}

	err := respondWithJSON(w, http.StatusCreated, wakuMessages)
	if err != nil {
		log.Println(err)
	}

	log.Printf(
		"%s\t%s\t%s",
		r.Method,
		r.RequestURI,
		time.Since(start),
	)
}

func (s *Server) createReceivedEnvelope(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var receivedEnvelope ReceivedEnvelope
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&receivedEnvelope); err != nil {
		log.Println(err)

		err := respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		if err != nil {
			log.Println(err)
		}
		return
	}
	defer r.Body.Close()

	err := receivedEnvelope.put(s.DB)
	if err != nil {
		log.Println("could not save envelope", err, receivedEnvelope)
	}

	err = respondWithJSON(w, http.StatusCreated, receivedEnvelope)
	if err != nil {
		log.Println(err)
	}

	log.Printf(
		"%s\t%s\t%s",
		r.Method,
		r.RequestURI,
		time.Since(start),
	)
}

func (s *Server) updateEnvelope(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var receivedEnvelope ReceivedEnvelope
	decoder := json.NewDecoder(r.Body)
	log.Println("Update envelope")
	if err := decoder.Decode(&receivedEnvelope); err != nil {
		log.Println(err)

		err := respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		if err != nil {
			log.Println(err)
		}
		return
	}
	defer r.Body.Close()

	err := receivedEnvelope.updateProcessingError(s.DB)
	if err != nil {
		log.Println("could not update envelope", err, receivedEnvelope)
	}

	err = respondWithJSON(w, http.StatusCreated, receivedEnvelope)
	if err != nil {
		log.Println(err)
	}

	log.Printf(
		"%s\t%s\t%s",
		r.Method,
		r.RequestURI,
		time.Since(start),
	)
}

func (s *Server) createProtocolStats(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var protocolStats ProtocolStats
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&protocolStats); err != nil {
		log.Println(err)

		err := respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		if err != nil {
			log.Println(err)
		}
		return
	}
	defer r.Body.Close()

	peerIDHash := sha256.Sum256([]byte(protocolStats.PeerID))
	protocolStats.PeerID = hex.EncodeToString(peerIDHash[:])

	if err := protocolStats.put(s.DB); err != nil {
		err := respondWithError(w, http.StatusInternalServerError, "Could not save protocol stats")
		if err != nil {
			log.Println(err)
		}
		return
	}

	err := respondWithJSON(w, http.StatusCreated, map[string]string{"error": ""})
	if err != nil {
		log.Println(err)
	}

	log.Printf(
		"%s\t%s\t%s",
		r.Method,
		r.RequestURI,
		time.Since(start),
	)
}

func (s *Server) Start(port int) {
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	log.Printf("Starting server on port %d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), handlers.CORS(originsOk, headersOk, methodsOk)(s.Router)))
}
