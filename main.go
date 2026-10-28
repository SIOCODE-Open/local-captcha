package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func main() {
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.LUTC)
	AppConfig = LoadConfig()
	AppContext = NewAppContext()
	port := AppConfig.Port

	log.Printf("Starting server on port %d", port)
	log.Printf("Configured mode: %s (default_score: %.2f)", AppConfig.Mode, AppConfig.DefaultScore)
	log.Printf("Number of configured sites: %d", len(AppConfig.Sites))

	router := mux.NewRouter()

	// Health check
	router.HandleFunc("/healthz", GET_healthz).Methods("GET")

	// Injectable JavaScript SDK
	router.HandleFunc("/recaptcha/api.js", GET_recaptcha_api_js).Methods("GET")
	router.HandleFunc("/captcha.js", GET_recaptcha_api_js).Methods("GET")

	// Interactive challenge UI (iframe)
	router.HandleFunc("/recaptcha/anchor", GET_recaptcha_anchor).Methods("GET")
	router.HandleFunc("/challenge", GET_recaptcha_anchor).Methods("GET")

	// Challenge solve / token issuance
	router.HandleFunc("/challenge/solve", POST_challenge_solve).Methods("POST")
	router.HandleFunc("/recaptcha/reload", POST_challenge_solve).Methods("POST")

	// CAPTCHA verification API (POST and GET supported for compatibility)
	router.HandleFunc("/recaptcha/api/siteverify", HANDLE_siteverify).Methods("POST", "GET")
	router.HandleFunc("/siteverify", HANDLE_siteverify).Methods("POST", "GET")
	router.HandleFunc("/verify", HANDLE_siteverify).Methods("POST", "GET")

	// Testing & inspection API
	router.HandleFunc("/testing/tokens", POST_testing_tokens).Methods("POST")
	router.HandleFunc("/testing/tokens", GET_testing_tokens).Methods("GET")
	router.HandleFunc("/testing/verifications", GET_testing_verifications).Methods("GET")
	router.HandleFunc("/testing/config", GET_testing_config).Methods("GET")
	router.HandleFunc("/testing/config", PUT_testing_config).Methods("PUT", "POST")
	router.HandleFunc("/testing/reset", POST_testing_reset).Methods("POST")

	corsRouter := corsMiddleware(router)
	loggedRouter := accessLogger(corsRouter)

	addr := ":" + strconv.Itoa(port)
	if err := http.ListenAndServe(addr, loggedRouter); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
