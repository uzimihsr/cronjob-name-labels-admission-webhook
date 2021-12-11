package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/uzimihsr/cronjob-name-labels-admission-webhook/webhook"
	v1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

// flags
var (
	certFile string
	keyFile  string
	port     int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cronjob-name-labels-admission-webhook",
	Short: "Kubernetes MutatingAdmissionWebhook to label Jobs owned by CronJob with the value of CronJob name.",
	Long:  `Kubernetes MutatingAdmissionWebhook to label Jobs owned by CronJob with the value of CronJob name.`,
	Args:  cobra.MaximumNArgs(0),
	Run:   main,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	// parse flags
	rootCmd.Flags().StringVar(&certFile, "tls-cert-file", "/tls/tls.crt", "File containing the default x509 Certificate for HTTPS. (CA cert, if any, concatenated after server cert).")
	rootCmd.Flags().StringVar(&keyFile, "tls-private-key-file", "/tls/tls.key", "File containing the default x509 private key matching --tls-cert-file.")
	rootCmd.Flags().IntVar(&port, "port", 443, "port the server listens on")
}

// serve handles the http portion of a request prior to handing to an admit function
func serve(w http.ResponseWriter, r *http.Request, admitFunc func(v1.AdmissionReview) *v1.AdmissionResponse) {

	// load the request body
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

	// parse the request body
	requestObj := &v1.AdmissionReview{}
	err := json.Unmarshal(body, requestObj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// call admitFunc
	var responseObj runtime.Object
	responseAdmissionReview := &v1.AdmissionReview{}
	responseAdmissionReview.SetGroupVersionKind(v1.SchemeGroupVersion.WithKind("AdmissionReview"))
	responseAdmissionReview.Response = admitFunc(*requestObj)
	responseAdmissionReview.Response.UID = requestObj.Request.UID
	responseObj = responseAdmissionReview

	// send response
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
