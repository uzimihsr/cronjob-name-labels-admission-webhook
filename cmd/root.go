/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/uzimihsr/cronjob-labels-admission-webhook/webhook"
	v1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

// var cfgFile string
var (
	certFile string
	keyFile  string
	port     int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cronjob-labels-admission-webhook",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	Args: cobra.MaximumNArgs(0),
	Run:  main,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	// parse flags
	rootCmd.Flags().StringVar(&certFile, "tls-cert-file", "", "File containing the default x509 Certificate for HTTPS. (CA cert, if any, concatenated after server cert).")
	rootCmd.Flags().StringVar(&keyFile, "tls-private-key-file", "", "File containing the default x509 private key matching --tls-cert-file.")
	rootCmd.Flags().IntVar(&port, "port", 443, "port the server listens on")
}

// serve handles the http portion of a request prior to handing to an admit function
func serve(w http.ResponseWriter, r *http.Request, admitFunc func(v1.AdmissionReview) *v1.AdmissionResponse) {

	// load request body
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("contentType=%s, expect application/json", contentType)
		http.Error(w, fmt.Sprintf("contentType=%s, expect application/json", contentType), http.StatusUnsupportedMediaType)
		return
	}

	klog.Info(fmt.Sprintf("handling request: %s", body))

	reqObj := &v1.AdmissionReview{}
	err := json.Unmarshal(body, reqObj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var responseObj runtime.Object
	responseAdmissionReview := &v1.AdmissionReview{}
	responseAdmissionReview.SetGroupVersionKind(v1.SchemeGroupVersion.WithKind("AdmissionReview"))
	responseAdmissionReview.Response = admitFunc(*reqObj)
	responseAdmissionReview.Response.UID = reqObj.Request.UID
	responseObj = responseAdmissionReview

	respBytes, err := json.Marshal(responseObj)
	if err != nil {
		klog.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	klog.Info(fmt.Sprintf("sending response: %s", respBytes))

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(respBytes); err != nil {
		klog.Error(err)
	}
}

func serveLabelJobOwnedByCronJob(w http.ResponseWriter, r *http.Request) {
	serve(w, r, webhook.LabelJobOwnedByCronJob)
}

func main(cmd *cobra.Command, args []string) {
	http.HandleFunc("/label-job-owned-by-cronjob", serveLabelJobOwnedByCronJob)
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}
	err := server.ListenAndServeTLS(certFile, keyFile)
	if err != nil {
		panic(err)
	}
}
