package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/spf13/viper"
)

func startServer() {
	mux := http.NewServeMux()

	mux.HandleFunc("/v3/update", basicAuth(handleQuery))
	mux.HandleFunc("/nic/update", basicAuth(handleQuery))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", viper.GetInt("Server_Port")),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if viper.Get("Server_Cert") != "" && viper.Get("ServerPrivateKey") != "" {
		log.Printf("Listening (HTTPS) on %v\n", server.Addr)
		log.Fatal(server.ListenAndServeTLS(viper.GetString("Server_Cert"), viper.GetString("Server_PrivateKey")))
	} else {
		log.Printf("Listening (HTTP) on %v\n", server.Addr)
		log.Println("Warning! No certificate and key provided so serving over HTTP! You should never run DDNS_AWS in production this way! Specify certificate and private key (using config or global variables) to enable HTTPS. This will avoid the possibility of someone stealing your username/pass combination.")
		log.Fatal(server.ListenAndServe())
	}
}

func handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		fmt.Fprint(w, "badagent")
		return
	}

	// get required 'name' parameter
	q := r.URL.Query()
	name := q.Get("hostname")
	if name == "" {
		fmt.Fprint(w, "notfqdn")
		return
	}

	// get optional 'ip' parameter or try getting ip from the request
	ip := q.Get("myip")
	if ip == "" {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			fmt.Fprint(w, "dnserr")
		}
		ip = host
	}

	// finally, update record
	status, err := updateRecordIP(&name, &ip)
	if err != nil {
		if status == "NOT_ACTIVE" {
			log.Printf("Error updating record: '%v' with ip:'%v'. Error was: %s: %v ", name, ip, status, err)
			fmt.Fprint(w, "nohost")
		}

		if status == "ROUTE53_UPDATE_FAIL" {
			log.Printf("Error updating record: '%v' with ip:'%v'. Error was: %s: %v ", name, ip, status, err)
			fmt.Fprint(w, "911")
		}

		return
	}

	// return proper response
	switch status {
	case "UPDATE_OK":
		log.Printf("Success: updated '%v' with ip:'%v'\n", name, ip)
		fmt.Fprintf(w, "good %v", ip)
	case "NO_CHANGE":
		log.Printf("No-change: '%v' ip is unchanged: '%v\n'", name, ip)
		fmt.Fprintf(w, "nochg %v", ip)
	default:
		log.Println("Error: 500 Internal Server Error. Unexpected success status.")
		fmt.Fprint(w, "911")
	}
}

func basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if ok {
			userHash := sha256.Sum256([]byte(user))
			passHash := sha256.Sum256([]byte(pass))
			expectedUserHash := sha256.Sum256([]byte(viper.GetString("Server_Username")))
			expectedPassHash := sha256.Sum256([]byte(viper.GetString("Server_Password")))

			userOk := (subtle.ConstantTimeCompare(userHash[:], expectedUserHash[:]) == 1)
			passOk := (subtle.ConstantTimeCompare(passHash[:], expectedPassHash[:]) == 1)

			if userOk && passOk {
				next.ServeHTTP(w, r)
				return
			}
		}
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "badauth", http.StatusUnauthorized)
	})
}
