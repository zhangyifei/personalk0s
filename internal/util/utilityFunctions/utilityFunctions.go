package utilityFunctions

import (
	"bufio"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

// get user password
func GetUserPassword() (string, error) {

	os.Stderr.WriteString("Password: ")

	if term.IsTerminal(int(os.Stdin.Fd())) {
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return "", err
		}
		password := string(bytePassword)
		os.Stderr.WriteString("\n")
		return strings.TrimSpace(password), nil
	} else {
		reader := bufio.NewReader(os.Stdin)
		password, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(password), nil
	}

}

// get user signum
func GetUserSignum() (string, error) {

	os.Stderr.WriteString("Signum: ")

	reader := bufio.NewReader(os.Stdin)
	username, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(username), nil
}

// This method gets user .crt and .key and returns them as string
func GetCertAndKey(cache_location string) (string, string) {

	_, err := os.Stat(cache_location + "k8s_client.crt")

	// if certification does not already exist, get it from EWS
	if err != nil {

		log.Println("certificate was not found in local cache! Requesting it from EWS...")

		// Get signum and password from the user
		signum, err := GetUserSignum()
		if err != nil {
			log.Fatalln("error occured while prompting for credentials:", err)
		}

		pass, err := GetUserPassword()
		if err != nil {
			log.Fatalln("error occured while prompting for credentials:", err)
		}

		return RequestCertAndKeyFromEWS(cache_location, signum, pass, false)

	} else { // Certificate exists but needs to be checked for expiration

		userCertBytes, err := ioutil.ReadFile(cache_location + "k8s_client.crt")
		if err != nil {
			log.Fatalln("failed to read existing client certificate file")
		}

		userKeyBytes, err := ioutil.ReadFile(cache_location + "k8s_client.key")
		if err != nil {
			log.Fatalln("failed to read existing client key file")
		}

		// Decode the PEM
		block, _ := pem.Decode(userCertBytes)
		if block == nil {
			log.Fatalln("failed to parse existing client certificate")
		}

		// Now parse the certificate
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			log.Fatalln("failed to parse existing client certificate:" + err.Error())
		}

		// if expired, contact EWS
		if (time.Now()).After(cert.NotAfter) {
			log.Println("certificate has expired on:", cert.NotAfter, " Requesting it from EWS...")

			// Get signum and password from the user
			signum, err := GetUserSignum()
			if err != nil {
				log.Fatalln("error occured while prompting for credentials!")
			}

			pass, err := GetUserPassword()
			if err != nil {
				log.Fatalln("error occured while prompting for credentials!")
			}

			return RequestCertAndKeyFromEWS(cache_location, signum, pass, false)

		} else {
			// Certification exists and is valid
			// log.Println("certificate is valid. Expiration date:", cert.NotAfter)

			return string(userCertBytes), string(userKeyBytes)
		}
	}
}

// This method requests user .crt and .key from EWS, saves them in specified location,
// and then returns them as string to later output on stdout for kubectl.
// forceRenew determines whether to force renew the cert and key
func RequestCertAndKeyFromEWS(cache_location string, signum string, pass string, forceRenew bool) (string, string) {

	// Now make a POST request to EWS
	creds := url.Values{
		"userid": {signum},
		"passwd": {pass},
	}

	// TODO: add some timeout logic
	var url string
	if forceRenew {
		url = "https://ews.rnd.gic.ericsson.se/a/?a=ckc&f=yes"
	} else {
		url = "https://ews.rnd.gic.ericsson.se/a/?a=ckc"
	}

	resp, err := http.PostForm(url, creds)
	if err != nil {
		log.Println(err)
	}

	// the response is a map of strings to "ANY" aribtrary type(i.e. empty interface)
	var res map[string]interface{}

	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		log.Println(err)
	}

	// Check if the result is as expected
	if len(res) == 0 {
		log.Fatalln("signum or password incorrect!")
	}

	// Type assert the "status" part of the response(should be map[string]interface{})
	// and then store the returned cert and key of the response
	userCert := res["status"].(map[string]interface{})["clientCertificateData"]
	userKey := res["status"].(map[string]interface{})["clientKeyData"]

	userCertBytes := []byte(fmt.Sprintf("%v", userCert))
	userKeyBytes := []byte(fmt.Sprintf("%v", userKey))

	// Save the received certificate and key
	err = ioutil.WriteFile(cache_location+"k8s_client.crt", userCertBytes, 0600)
	if err != nil {
		log.Fatalln(err)
	}

	err = ioutil.WriteFile(cache_location+"k8s_client.key", userKeyBytes, 0600)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("successfully retrieved and cached user cert and key in:", cache_location)

	return string(userCertBytes), string(userKeyBytes)

}

// This method creates an output in format required by kubectl for authentication
// and returns the output. The required format can be found at:
// https://kubernetes.io/docs/reference/access-authn-authz/authentication/#input-and-output-formats
func CreateOutput(userCert string, userKey string) string {

	// Create a proper response using the user .crt and .key file
	output := make(map[string]interface{})
	output["apiVersion"] = "client.authentication.k8s.io/v1beta1"
	output["kind"] = "ExecCredential"
	output["status"] = map[string]string{
		"clientCertificateData": userCert,
		"clientKeyData":         userKey,
	}

	jsonData, err := json.Marshal(output)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// return the proper output format for kubectl
	return string(jsonData)

}

// checks whether path for caching cert and key (~/.eke/) exists
// or not, if not it creates it and returns the path
func Get_eke_path() string {
	// create the path for eke cache
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	// check the path for EKE_CACHE
	eke_cache := homeDir + "/.eke/"
	if _, err := os.Stat(eke_cache); os.IsNotExist(err) {
		err := os.MkdirAll(eke_cache, 0755)
		if err != nil {
			log.Fatal("error creating path for eke cache:", err)
		}
	}
	return eke_cache
}

// checks whether default path for kubeconfig (~/.kube/) exists
// or not, if not it creates it and returns the path
func Get_kubeconfig_path() string {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	// check the kubeconfig path
	kubeconfig_path := homeDir + "/.kube/"
	if _, err := os.Stat(kubeconfig_path); os.IsNotExist(err) {
		err := os.MkdirAll(kubeconfig_path, 0755)
		if err != nil {
			log.Fatal("error creating path for kubeconfig path:", err)
		}
	}
	return kubeconfig_path
}
