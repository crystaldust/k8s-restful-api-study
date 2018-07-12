package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/juzhen/k8s-client-test/model"
	"github.com/juzhen/k8s-client-test/utils"
	"istio.io/api/mixer/v1/config/client"
	"istio.io/istio/mixer/adapter/memquota/config"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
)

func createMemQuota(memquotaName, namespace string, dimensions map[string]string) ([]byte, error) {
	paramsQuota := &config.Params_Quota{
		Name:          fmt.Sprintf("memquota-quota-%s", memquotaName),
		MaxAmount:     1000,
		ValidDuration: time.Second,
		Overrides: []config.Params_Override{
			{
				Dimensions:    dimensions,
				MaxAmount:     1,
				ValidDuration: time.Second * 5,
			},
		},
	}

	restclient, err := utils.CreateRestClient("apis", "config.istio.io", "v1alpha2")
	result := restclient.Get().Resource("memquotas").Namespace(namespace).Name(memquotaName).Do()

	bs, err := result.Raw()
	var resultStatusCode int
	result.StatusCode(&resultStatusCode)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	// If we have the memquota, replace it
	var createReq *rest.Request
	var memquota *model.StructMemQuota

	if resultStatusCode == http.StatusOK {
		e := json.Unmarshal(bs, &memquota)
		if e != nil {
			return nil, e
		}
		memquota.Spec.Quotas = []*config.Params_Quota{paramsQuota}
		createReq = restclient.Put().Resource("memquotas").Namespace(namespace).Name(memquotaName)
	} else {
		memquota = model.MemQuota(memquotaName, namespace, []*config.Params_Quota{paramsQuota})
		createReq = restclient.Post().Resource("memquotas").Namespace(namespace).Name(memquotaName)
	}

	memquotaJsonBytes, _ := json.Marshal(&memquota)
	createReq.Body(memquotaJsonBytes)
	result = createReq.Do()
	return result.Raw()
}

func createRule(name, namespace string, actions []*model.StructAction) ([]byte, error) {

	restclient, err := utils.CreateRestClient("/apis", "config.istio.io", "v1alpha2")
	if err != nil {
		return nil, err
	}

	var result rest.Result
	result = restclient.Get().Resource("rules").Namespace(namespace).Name(name).Do()
	bs, err := result.Raw()
	var resultStatusCode int
	result.StatusCode(&resultStatusCode)

	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	// If we have the quota, replace it
	var createReq *rest.Request
	var rule *model.StructRule

	if resultStatusCode == http.StatusOK {
		e := json.Unmarshal(bs, &rule)
		if e != nil {
			return nil, e
		}
		rule.Spec.Actions = actions
		createReq = restclient.Put().Resource("rules").Namespace(namespace).Name(name)
	} else {
		// Create the quota
		rule = model.Rule(name, namespace, actions)
		createReq = restclient.Post().Resource("rules").Namespace(namespace).Name(name)
	}

	ruleJsonBytes, _ := json.Marshal(&rule)
	createReq.Body(ruleJsonBytes)
	result = createReq.Do()
	return result.Raw()
}

func createQuota(quotaName, namespace string, dimensions map[string]string) ([]byte, error) {
	restclient, err := utils.CreateRestClient("/apis", "config.istio.io", "v1alpha2")
	if err != nil {
		return nil, err
	}

	// req := restclient.Patch(types.JSONPatchType).Name("requestcount")
	// Cannot use 'Get' method to decode the json content, it complaints that 'QuotaList' is not registered
	// This is generated by request.Serilizers.Decoder(which is actually a json serializer: vendor/k8s.io/apimachinery/pkg/runtime/serializer/json/json.go)
	// obj, err := restclient.Get().Resource("quotas").Namespace("lance-test").Do().Get()
	result := restclient.Get().Resource("quotas").Namespace(namespace).Name(quotaName).Do()
	// If we have the quota, replace it

	var createReq *rest.Request
	var resultStatusCode int
	result.StatusCode(&resultStatusCode)

	bs, err := result.Raw()
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	var quota *model.StructQuota

	if resultStatusCode == http.StatusOK {
		e := json.Unmarshal(bs, &quota)
		if e != nil {
			return nil, e
		}
		quota.Spec.Dimensions = dimensions
		createReq = restclient.Put().Resource("quotas").Namespace(namespace).Name(quotaName)
	} else {
		// Create the quota
		quota = model.Quota(quotaName, namespace, dimensions)
		createReq = restclient.Post().Resource("quotas").Namespace(namespace).Name(quotaName)
	}

	quotaJsonBytes, _ := json.Marshal(&quota)
	createReq.Body(quotaJsonBytes)
	result = createReq.Do()
	return result.Raw()
}

func createQuotaSpec(quotaspecName, namespace string, rules []*client.QuotaRule) ([]byte, error) {
	restclient, err := utils.CreateRestClient("/apis", "config.istio.io", "v1alpha2")
	if err != nil {
		return nil, err
	}

	result := restclient.Get().Resource("quotaspecs").Namespace(namespace).Name(quotaspecName).Do()

	var createReq *rest.Request
	var resultStatusCode int
	result.StatusCode(&resultStatusCode)

	bs, err := result.Raw()
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	var quotaspec *model.StructQuotaSpec

	if resultStatusCode == http.StatusOK {
		e := json.Unmarshal(bs, &quotaspec)
		if e != nil {
			return nil, e
		}
		quotaspec.Spec.Rules = rules
		createReq = restclient.Put().Resource("quotaspecs").Namespace(namespace).Name(quotaspecName)
	} else {
		quotaspec = model.QuotaSpec(quotaspecName, namespace, rules)
		createReq = restclient.Post().Resource("quotaspecs").Namespace(namespace).Name(quotaspecName)
	}

	quotaspecJsonBytes, _ := json.Marshal(&quotaspec)
	createReq.Body(quotaspecJsonBytes)

	result = createReq.Do()
	return result.Raw()
}

func createQuotaSpecBinding(quotaspecName, namespace string, binding *client.QuotaSpecBinding) ([]byte, error) {
	restclient, err := utils.CreateRestClient("/apis", "config.istio.io", "v1alpha2")
	if err != nil {
		return nil, err
	}

	result := restclient.Get().Resource("quotaspecbindings").Namespace(namespace).Name(quotaspecName).Do()

	var createReq *rest.Request
	var resultStatusCode int
	result.StatusCode(&resultStatusCode)

	bs, err := result.Raw()
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	var quotaspecBinding *model.StructQuotaSpecBinding

	if resultStatusCode == http.StatusOK {
		e := json.Unmarshal(bs, &quotaspecBinding)
		if e != nil {
			return nil, e
		}
		quotaspecBinding.Spec = binding
		createReq = restclient.Put().Resource("quotaspecbindings").Namespace(namespace).Name(quotaspecName)
	} else {
		quotaspecBinding = model.QuotaSpecBinding(quotaspecName, namespace, binding)
		createReq = restclient.Post().Resource("quotaspecbindings").Namespace(namespace).Name(quotaspecName)
	}

	quotaspecJsonBytes, _ := json.Marshal(&quotaspecBinding)
	createReq.Body(quotaspecJsonBytes)

	result = createReq.Do()
	return result.Raw()
}

func HandleRateLimit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(fmt.Sprintf("Method %s not allowed", r.Method)))
		return
	}

	var chassis model.Chassis

	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to read request body"))
		return
	}
	contentType := r.Header.Get("Content-Type")
	if contentType == "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("application/json is not yet supported"))
		return
		// if e := json.Unmarshal(bs, &chassis); e != nil {
		//     w.WriteHeader(http.StatusInternalServerError)
		//     w.Write([]byte("Failed to parse JSON"))
		//     return
		// }
	} else { // Parse body as YAML
		if e := yaml.Unmarshal(bs, &chassis); e != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to parse YAML"))
			return
		}
	}

	serviceName := r.Header.Get("service_name")
	if serviceName == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("service name is not provided"))
		return
	}

	limit := chassis.Cse.Flowcontrol.Provider.QPS.Limit
	// Assume there is only one key in the map
	var source string
	n := 0
	for k := range limit {
		source = k
		n++
		if n >= 1 {
			break
		}
	}

	namespace := "lance-test"

	quotaName := fmt.Sprintf("requestcount-%s", serviceName)
	dimensions := map[string]string{
		"destination":        "destination.labels[\"app\"] | destination.service | \"unknown\"",
		"destinationVersion": "destination.labels[\"version\"] | \"unknown\"",
		"source":             "source.labels[\"app\"] | source.service | \"unknown\"",
		"sourceVersion":      "source.labels[\"version\"] | \"unknown\"",
	}

	if _, err := createQuota(quotaName, namespace, dimensions); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Create memquota
	memquotaDimensions := map[string]string{
		// "destination":   "ratings",
		// "source":        "reviews",
		//"sourceVersion": "v3",
		"destination": serviceName,
		"source":      source,
		// "sourceVersion": "v3",
	}

	memquotaName := fmt.Sprintf("handler-%s", serviceName)
	if _, err := createMemQuota(memquotaName, namespace, memquotaDimensions); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	actions := []*model.StructAction{
		{
			Handler:   fmt.Sprintf("%s.memquota", memquotaName),
			Instances: []string{"requestcount.quota"},
		},
	}

	ruleName := fmt.Sprintf("quota-%s", serviceName)
	if _, err := createRule(ruleName, namespace, actions); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	quotaspecName := fmt.Sprintf("request-count-%s", serviceName)
	rules := []*client.QuotaRule{
		{
			Quotas: []*client.Quota{
				{
					Quota:  fmt.Sprintf("quota-%s", serviceName),
					Charge: 1,
				},
			},
			// Match:{},
		},
	}

	if _, err := createQuotaSpec(quotaspecName, namespace, rules); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	quotaspecBindingName := fmt.Sprintf("request-count-binding-%s", serviceName)
	binding := &client.QuotaSpecBinding{
		Services: []*client.IstioService{
			{
				Name:      serviceName,
				Namespace: "default",
			},
			{
				Name:      source,
				Namespace: "default",
			},
		},
		QuotaSpecs: []*client.QuotaSpecBinding_QuotaSpecReference{
			{
				Name:      fmt.Sprintf("request-count-%s", serviceName),
				Namespace: namespace,
			},
		},
	}
	if _, err := createQuotaSpecBinding(quotaspecBindingName, namespace, binding); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
}
